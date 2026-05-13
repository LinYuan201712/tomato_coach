package model

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/tomato/backend/config"

)

// ChatInner 定义了聊天模型的通用接口
type ChatInner interface {
	Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error)
	Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error)
	BindTools(tools []*schema.ToolInfo) error
}

// NewChatModel 创建一个新的聊天模型 (ARK 默认)
func NewChatModel(ctx context.Context, cfg *config.ARKConfig) (*ChatModel, error) {
	inner, err := NewARKChatModel(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &ChatModel{inner}, nil
}

// NewChatModelFromInner 从已有的内部模型创建 ChatModel
func NewChatModelFromInner(inner ChatInner) *ChatModel {
	return &ChatModel{inner}
}

// ChatModel 封装了聊天模型
type ChatModel struct {
	inner ChatInner
}

func (m *ChatModel) Generate(ctx context.Context, msgs []*schema.Message) (*schema.Message, error) {
	return m.inner.Generate(ctx, msgs)
}

func (m *ChatModel) Stream(ctx context.Context, msgs []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	return m.inner.Stream(ctx, msgs)
}

func (m *ChatModel) BindTools(tools []*schema.ToolInfo) error {
	return m.inner.BindTools(tools)
}

// ARKEmbedder 方舟 (ARK) 文本嵌入模型
type ARKEmbedder struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

func NewEmbedder(ctx context.Context, cfg *config.ARKConfig) (*ARKEmbedder, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("API 密钥为空")
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://ark.cn-beijing.volces.com/api/v3"
	}
	return &ARKEmbedder{
		apiKey:  cfg.APIKey,
		model:   cfg.EmbeddingModel,
		baseURL: baseURL,
		client:  &http.Client{},
	}, nil
}

// EmbedStrings 将一组文本转换为向量
func (e *ARKEmbedder) EmbedStrings(ctx context.Context, texts []string) ([][]float64, error) {
	type embedRequest struct {
		Model string   `json:"model"`
		Input []string `json:"input"`
	}
	type embedResponse struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	reqBody := embedRequest{e.model, texts}
	body, _ := json.Marshal(reqBody)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", e.baseURL+"/embeddings", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := e.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API 错误 [%d]: %s", resp.StatusCode, string(b))
	}
	var res embedResponse
	json.NewDecoder(resp.Body).Decode(&res)
	if len(res.Data) == 0 {
		return nil, errors.New("响应为空")
	}
	embs := make([][]float64, len(res.Data))
	for i, d := range res.Data {
		embs[i] = d.Embedding
	}
	return embs, nil
}

type ARKChatModel struct {
	apiKey    string
	model     string
	baseURL   string
	client    *http.Client
	toolInfos []*schema.ToolInfo
}

func NewARKChatModel(ctx context.Context, cfg *config.ARKConfig) (*ARKChatModel, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("API Key is empty")
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://ark.cn-beijing.volces.com/api/v3"
	}
	return &ARKChatModel{
		apiKey:  cfg.APIKey,
		model:   cfg.ChatModel,
		baseURL: baseURL,
		client:  &http.Client{},
	}, nil
}

type arkMessage struct {
	Role       string        `json:"role"`
	Content    string        `json:"content"`
	ToolCalls  []arkToolCall `json:"tool_calls,omitempty"`
	ToolCallID string        `json:"tool_call_id,omitempty"`
}

type arkToolCall struct {
	Index    int             `json:"index,omitempty"`
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Function arkFunctionCall `json:"function"`
}

type arkFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type arkTool struct {
	Type     string      `json:"type"`
	Function arkFunction `json:"function"`
}

type arkFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

type arkRequest struct {
	Model         string            `json:"model"`
	Messages      []arkMessage      `json:"messages"`
	Tools         []arkTool         `json:"tools,omitempty"`
	Stream        bool              `json:"stream"`
	StreamOptions *arkStreamOptions `json:"stream_options,omitempty"`
}

type arkStreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type arkResponse struct {
	Choices []struct {
		Message struct {
			Role             string        `json:"role"`
			Content          string        `json:"content"`
			ReasoningContent string        `json:"reasoning_content,omitempty"`
			ToolCalls        []arkToolCall `json:"tool_calls,omitempty"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type arkStreamResponse struct {
	Choices []struct {
		Delta struct {
			Role             string        `json:"role"`
			Content          string        `json:"content"`
			ReasoningContent string        `json:"reasoning_content,omitempty"`
			ToolCalls        []arkToolCall `json:"tool_calls,omitempty"`
		} `json:"delta"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
}

func (m *ARKChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
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
			fullMsg.ReasoningContent += msg.ReasoningContent
			if len(msg.ToolCalls) > 0 {
				fullMsg.ToolCalls = msg.ToolCalls
			}
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

func (m *ARKChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	arkMsgs := make([]arkMessage, 0, len(messages))
	for _, msg := range messages {
		amsg := arkMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
		if len(msg.ToolCalls) > 0 {
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
		if msg.Role == schema.Tool {
			amsg.ToolCallID = msg.ToolCallID
		}
		arkMsgs = append(arkMsgs, amsg)
	}
	req := arkRequest{
		Model:    m.model,
		Messages: arkMsgs,
		Stream:   true,
		StreamOptions: &arkStreamOptions{
			IncludeUsage: true,
		},
	}
	if len(m.toolInfos) > 0 {
		req.Tools = m.convertTools()
	}
	sr, sw := schema.Pipe[*schema.Message](10)
	go func() {
		defer sw.Close()
		m.doStreamRequest(ctx, req, sw)
	}()
	return sr, nil
}

func (m *ARKChatModel) BindTools(tools []*schema.ToolInfo) error {
	m.toolInfos = tools
	return nil
}

func (m *ARKChatModel) convertTools() []arkTool {
	ts := make([]arkTool, 0, len(m.toolInfos))
	for _, info := range m.toolInfos {
		var params any
		if info.ParamsOneOf != nil {
			params, _ = info.ParamsOneOf.ToJSONSchema()
		}
		ts = append(ts, arkTool{
			Type: "function",
			Function: arkFunction{
				Name:        info.Name,
				Description: info.Desc,
				Parameters:  params,
			},
		})
	}
	return ts
}

func (m *ARKChatModel) doStreamRequest(ctx context.Context, req arkRequest, sw *schema.StreamWriter[*schema.Message]) error {
	b, _ := json.Marshal(req)
	hreq, _ := http.NewRequestWithContext(ctx, "POST", m.baseURL+"/chat/completions", bytes.NewReader(b))
	hreq.Header.Set("Content-Type", "application/json")
	hreq.Header.Set("Authorization", "Bearer "+m.apiKey)
	hreq.Header.Set("Accept", "text/event-stream")
	hresp, err := m.client.Do(hreq)
	if err != nil {
		return err
	}
	defer hresp.Body.Close()
	reader := hresp.Body
	buf := make([]byte, 4096)
	var leftover string

	// 用于累积流式输出中的工具调用
	accumulatedToolCalls := make(map[int]*schema.ToolCall)

	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		data := leftover + string(buf[:n])
		lines := strings.Split(data, "\n")
		leftover = lines[len(lines)-1]
		lines = lines[:len(lines)-1]
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "data: ") {
				jsonData := strings.TrimPrefix(line, "data: ")
				if jsonData == "[DONE]" {
					return nil
				}
				var sr arkStreamResponse
				if err := json.Unmarshal([]byte(jsonData), &sr); err == nil && len(sr.Choices) > 0 {
					delta := sr.Choices[0].Delta
					msg := &schema.Message{
						Role:             schema.RoleType(delta.Role),
						Content:          delta.Content,
						ReasoningContent: delta.ReasoningContent,
					}
					if len(delta.ToolCalls) > 0 {
						msg.ToolCalls = make([]schema.ToolCall, 0, len(delta.ToolCalls))
						for _, tc := range delta.ToolCalls {
							idx := tc.Index

							atc, ok := accumulatedToolCalls[idx]
							if !ok {
								atc = &schema.ToolCall{
									ID:   tc.ID,
									Type: tc.Type,
									Function: schema.FunctionCall{
										Name:      tc.Function.Name,
										Arguments: tc.Function.Arguments,
									},
								}
								accumulatedToolCalls[idx] = atc
							} else {
								if tc.ID != "" {
									atc.ID = tc.ID
								}
								if tc.Function.Name != "" {
									atc.Function.Name = tc.Function.Name
								}
								atc.Function.Arguments += tc.Function.Arguments
							}
							// 发送当前分片的工具调用信息
							msg.ToolCalls = append(msg.ToolCalls, *atc)
						}
					}
					// 捕获流式输出中的 Usage
					if sr.Usage != nil {
						msg.ResponseMeta = &schema.ResponseMeta{
							Usage: &schema.TokenUsage{
								PromptTokens:     sr.Usage.PromptTokens,
								CompletionTokens: sr.Usage.CompletionTokens,
								TotalTokens:      sr.Usage.TotalTokens,
							},
						}
					}
					sw.Send(msg, nil)
				}
			}
		}
	}
}
