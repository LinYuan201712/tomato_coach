package agent

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/tomato/backend/einox/model"
	"github.com/tomato/backend/einox/prompt"
)

const MaxToolResponseLen = 2000

const StreamContextKey = "stream_writer"

func TruncateToolResult(res string) string {
	runes := []rune(res)
	if len(runes) <= MaxToolResponseLen {
		return res
	}
	return string(runes[:MaxToolResponseLen]) + "\n\n(内容过长，已自动截断...)"
}

func RunToolLoop(ctx context.Context, chatModel *model.ChatModel, msgs []*schema.Message, allTools []tool.BaseTool) (*schema.Message, error) {
	for i := 0; i < 5; i++ {
		// 防御性截断，确保多轮工具调用不会导致上下文爆炸
		msgs = prompt.TruncateMessages(msgs, 6000)
		
		resp, err := chatModel.Generate(ctx, msgs)
		if err != nil {
			return nil, err
		}
		if len(resp.ToolCalls) == 0 {
			return resp, nil
		}

		msgs = append(msgs, resp)
		for _, tc := range resp.ToolCalls {
			var targetTool tool.BaseTool
			for _, t := range allTools {
				info, err := t.Info(ctx)
				if err == nil && info != nil && info.Name == tc.Function.Name {
					targetTool = t
					break
				}
			}

			if targetTool == nil {
				msgs = append(msgs, &schema.Message{Role: schema.Tool, Content: "工具未找到", ToolCallID: tc.ID})
				continue
			}

			if invokable, ok := targetTool.(tool.InvokableTool); ok {
				args := strings.TrimSpace(tc.Function.Arguments)
				
				// 生命周期治理: 设置 20s 超时
				toolCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
				res, err := invokable.InvokableRun(toolCtx, args)
				cancel()

				if err != nil {
					errMsg := err.Error()
					if err == context.DeadlineExceeded {
						errMsg = "工具执行超时 (20s)，请稍后再试"
					}
					msgs = append(msgs, &schema.Message{Role: schema.Tool, Content: errMsg, ToolCallID: tc.ID})
				} else {
					truncatedRes := TruncateToolResult(res)
					msgs = append(msgs, &schema.Message{Role: schema.Tool, Content: truncatedRes, ToolCallID: tc.ID})
				}
			}
		}
	}
	return nil, fmt.Errorf("tool loop exceeded")
}

func RunStreamToolLoop(ctx context.Context, chatModel *model.ChatModel, msgs []*schema.Message, allTools []tool.BaseTool) (*schema.StreamReader[*schema.Message], error) {
	sr, sw := schema.Pipe[*schema.Message](10)

	go func() {
		defer sw.Close()
		currentMsgs := msgs
		for i := 0; i < 5; i++ {
			// 防御性截断
			currentMsgs = prompt.TruncateMessages(currentMsgs, 6000)

			stream, err := chatModel.Stream(ctx, currentMsgs)
			if err != nil {
				sw.Send(nil, err)
				return
			}

			var fullResp *schema.Message
			for {
				chunk, err := stream.Recv()
				if err != nil {
					if err == io.EOF {
						break
					}
					sw.Send(nil, err)
					return
				}

				sw.Send(chunk, nil)

				if fullResp == nil {
					fullResp = &schema.Message{
						Role:             chunk.Role,
						Content:          chunk.Content,
						ReasoningContent: chunk.ReasoningContent,
						ToolCalls:        chunk.ToolCalls,
					}
				} else {
					fullResp.Content += chunk.Content
					fullResp.ReasoningContent += chunk.ReasoningContent
					if len(chunk.ToolCalls) > 0 {
						// 改进：使用 ID 和 Index 协同定位，处理并行流式 Delta
						for _, tc := range chunk.ToolCalls {
							found := false
							for j := range fullResp.ToolCalls {
								// 匹配逻辑：如果 ID 相同，或者（ID 都为空且 Index 相同）
								idMatch := tc.ID != "" && fullResp.ToolCalls[j].ID == tc.ID
								indexMatch := tc.Index != nil && fullResp.ToolCalls[j].Index != nil && *fullResp.ToolCalls[j].Index == *tc.Index
								
								if idMatch || indexMatch {
									if tc.ID != "" {
										fullResp.ToolCalls[j].ID = tc.ID
									}
									if tc.Function.Name != "" {
										fullResp.ToolCalls[j].Function.Name = tc.Function.Name
									}
									fullResp.ToolCalls[j].Function.Arguments += tc.Function.Arguments
									found = true
									break
								}
							}
							if !found {
								fullResp.ToolCalls = append(fullResp.ToolCalls, tc)
							}
						}
					}
				}
			}
			stream.Close()

			if fullResp == nil || len(fullResp.ToolCalls) == 0 {
				return
			}

			// 调试：打印最终聚合后的工具调用
			currentMsgs = append(currentMsgs, fullResp)
			for _, tc := range fullResp.ToolCalls {
				var targetTool tool.BaseTool
				for _, t := range allTools {
					info, err := t.Info(ctx)
					if err == nil && info != nil && info.Name == tc.Function.Name {
						targetTool = t
						break
					}
				}

				if targetTool == nil {
					currentMsgs = append(currentMsgs, &schema.Message{Role: schema.Tool, Content: "工具未找到", ToolCallID: tc.ID})
					continue
				}

				if invokable, ok := targetTool.(tool.InvokableTool); ok {
					args := strings.TrimSpace(tc.Function.Arguments)
					
					// 生命周期治理: 设置 20s 超时
					toolCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
					res, err := invokable.InvokableRun(toolCtx, args)
					cancel()

					if err != nil {
						errMsg := err.Error()
						if err == context.DeadlineExceeded {
							errMsg = "工具执行超时 (20s)，请稍后再试"
						}
						currentMsgs = append(currentMsgs, &schema.Message{Role: schema.Tool, Content: errMsg, ToolCallID: tc.ID})
					} else {
						truncatedRes := TruncateToolResult(res)
						currentMsgs = append(currentMsgs, &schema.Message{Role: schema.Tool, Content: truncatedRes, ToolCallID: tc.ID})
					}
				}
			}
		}
	}()

	return sr, nil
}
func RunGraphReAct(ctx context.Context, chatModel *model.ChatModel, state *AgentState, allTools []tool.BaseTool) (*schema.StreamReader[*schema.Message], error) {
	sr, sw := schema.Pipe[*schema.Message](10)

	go func() {
		defer sw.Close()

		// 将 StreamWriter 注入 Context，供 Graph 节点使用实现增量更新
		graphCtx := context.WithValue(ctx, StreamContextKey, sw)
		
		builder := NewReActGraphBuilder(chatModel, allTools)
		r, err := builder.Compile(graphCtx)
		if err != nil {
			sw.Send(nil, err)
			return
		}

		// 执行图
		_, err = r.Invoke(graphCtx, state)
		if err != nil {
			sw.Send(nil, err)
			return
		}
	}()

	return sr, nil
}
