package model

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/tomato/backend/config"

)

// AliyunChatModel 阿里云百炼 (OpenAI 兼容模式) 聊天模型
type AliyunChatModel struct {
	apiKey    string
	model     string
	baseURL   string
	client    *http.Client
	toolInfos []*schema.ToolInfo
}

func NewAliyunChatModel(ctx context.Context, cfg *config.AliyunConfig) (*AliyunChatModel, error) {
	apiKey := cfg.APIKey
	if apiKey == "" {
		return nil, fmt.Errorf("aliyun api key is required")
	}

	modelName := cfg.ChatModel
	if modelName == "" {
		modelName = "qwen-plus" // 默认使用 qwen-plus
	}
	return &AliyunChatModel{
		apiKey:  apiKey,
		model:   modelName,
		baseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
		client:  &http.Client{},
	}, nil
}

func (m *AliyunChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	// 归一化重构：非流式请求现在在内部调用流式并聚合结果
	sr, err := m.Stream(ctx, messages, opts...)
	if err != nil {
		return nil, err
	}
	defer sr.Close()

	var fullMsg *schema.Message
	for {
		msg, err := sr.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if fullMsg == nil {
			fullMsg = msg
		} else {
			fullMsg.Content += msg.Content
			if msg.ResponseMeta != nil {
				fullMsg.ResponseMeta = msg.ResponseMeta
			}
		}
	}

	if fullMsg == nil {
		return nil, fmt.Errorf("empty response from stream aggregation")
	}

	return fullMsg, nil
}

func (m *AliyunChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	reqBody := m.buildRequest(messages, true) // 强制使用流式请求
	body, _ := json.Marshal(reqBody)

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", m.baseURL+"/chat/completions", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+m.apiKey)

	resp, err := m.client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("Aliyun API Error [%d]: %s", resp.StatusCode, string(b))
	}

	sr, sw := schema.Pipe[*schema.Message](10)
	go func() {
		defer sw.Close()
		defer resp.Body.Close()

		// 增加工具调用状态追踪，解决流式聚合时 ID 丢失导致 JSON 损坏的问题
		type toolCallState struct {
			ID   string
			Name string
		}
		toolStateMap := make(map[int]toolCallState)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" { continue }
			if line == "data: [DONE]" { break }
			if !strings.HasPrefix(line, "data: ") { continue }

			rawJSON := strings.TrimPrefix(line, "data: ")
			var streamRes arkStreamResponse
			if err := json.Unmarshal([]byte(rawJSON), &streamRes); err != nil {
				continue
			}

			var msg *schema.Message
			if len(streamRes.Choices) > 0 {
				delta := streamRes.Choices[0].Delta
				msg = &schema.Message{
					Role:             schema.Assistant,
					Content:          delta.Content,
					ReasoningContent: delta.ReasoningContent,
				}

				if len(delta.ToolCalls) > 0 {
					msg.ToolCalls = make([]schema.ToolCall, 0, len(delta.ToolCalls))
					for _, tc := range delta.ToolCalls {
						idx := tc.Index

						// 获取或更新状态
						state := toolStateMap[idx]
						if tc.ID != "" { state.ID = tc.ID }
						if tc.Function.Name != "" { state.Name = tc.Function.Name }
						toolStateMap[idx] = state

						// 关键修复：如果还没有 ID，绝对不能发送，否则会导致聚合到空的 ID Key 从而损坏 JSON
						if state.ID == "" {
							fmt.Printf("[DEBUG] ToolCall Chunk Ignored (No ID yet for index %d)\n", idx)
							continue
						}

						fmt.Printf("[DEBUG] ToolCall Chunk: Index=%d, ID=%s, Name=%s, ArgsDelta=%s\n", 
							idx, state.ID, state.Name, tc.Function.Arguments)

						msg.ToolCalls = append(msg.ToolCalls, schema.ToolCall{
							Index: &idx,
							ID:    state.ID, 
							Type:  tc.Type,
							Function: schema.FunctionCall{
								Name:      state.Name,
								Arguments: tc.Function.Arguments,
							},
						})
					}
				}
			}

			if streamRes.Usage != nil {
				if msg == nil { msg = &schema.Message{Role: schema.Assistant} }
				msg.ResponseMeta = &schema.ResponseMeta{
					Usage: &schema.TokenUsage{
						PromptTokens:     streamRes.Usage.PromptTokens,
						CompletionTokens: streamRes.Usage.CompletionTokens,
						TotalTokens:      streamRes.Usage.TotalTokens,
					},
				}
			}

			if msg != nil {
				sw.Send(msg, nil)
			}
		}

		if err := scanner.Err(); err != nil {
			sw.Send(nil, err)
		}
	}()

	return sr, nil
}

func (m *AliyunChatModel) BindTools(tools []*schema.ToolInfo) error {
	m.toolInfos = tools
	return nil
}

func (m *AliyunChatModel) buildRequest(messages []*schema.Message, stream bool) arkRequest {
	arkMsgs := make([]arkMessage, 0, len(messages))
	roles := make([]string, 0, len(messages))
	
	for _, msg := range messages {
		role := "user"
		switch msg.Role {
		case schema.System:
			role = "system"
		case schema.Assistant:
			role = "assistant"
		case schema.Tool:
			role = "tool"
		}
		roles = append(roles, role)
		
		amsg := arkMessage{
			Role:    role,
			Content: msg.Content,
		}
		if msg.Role == schema.Tool {
			amsg.ToolCallID = msg.ToolCallID
		}
		if msg.Role == schema.Assistant && len(msg.ToolCalls) > 0 {
			amsg.ToolCalls = make([]arkToolCall, 0, len(msg.ToolCalls))
			for _, tc := range msg.ToolCalls {
				amsg.ToolCalls = append(amsg.ToolCalls, arkToolCall{
					ID:   tc.ID,
					Type: tc.Type,
					Function: arkFunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				})
			}
		}
		arkMsgs = append(arkMsgs, amsg)
	}

	req := arkRequest{
		Model:    m.model,
		Messages: arkMsgs,
		Stream:   stream,
	}

	if len(m.toolInfos) > 0 {
		req.Tools = make([]arkTool, 0, len(m.toolInfos))
		for _, t := range m.toolInfos {
			var params any
			if t.ParamsOneOf != nil {
				params, _ = t.ParamsOneOf.ToJSONSchema()
			}
			req.Tools = append(req.Tools, arkTool{
				Type: "function",
				Function: arkFunction{
					Name:        t.Name,
					Description: t.Desc,
					Parameters:  params,
				},
			})
		}
	}

	if stream {
		req.StreamOptions = &arkStreamOptions{IncludeUsage: true}
	}

	fmt.Printf("[DEBUG] Aliyun Request Roles: %v\n", roles)
	return req
}
