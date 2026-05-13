package agent

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/tomato/backend/einox/model"
)

// ReActGraphBuilder 封装了构建 ReAct 图的逻辑
type ReActGraphBuilder struct {
	chatModel *model.ChatModel
	tools     []tool.BaseTool
}

func NewReActGraphBuilder(m *model.ChatModel, tools []tool.BaseTool) *ReActGraphBuilder {
	return &ReActGraphBuilder{
		chatModel: m,
		tools:     tools,
	}
}

// Compile 构建并编译 ReAct 状态图
func (b *ReActGraphBuilder) Compile(ctx context.Context) (compose.Runnable[*AgentState, *AgentState], error) {
	g := compose.NewGraph[*AgentState, *AgentState]()

	// 1. Model Node: 负责生成回答或工具调用
	modelNode, err := compose.AnyLambda(func(ctx context.Context, state *AgentState, opts ...any) (*AgentState, error) {
		// 这里执行 Model 调用
		// 如果是流式，我们需要将流合并并发送给输出流（通过 Context 中的 StreamWriter）
		swVal := ctx.Value(StreamContextKey)
		sw, ok := swVal.(*schema.StreamWriter[*schema.Message])

		// 注意：这里为了简化逻辑，我们先处理非流式的增量更新逻辑
		// 真正的流式增量更新需要 Model.Stream
		stream, err := b.chatModel.Stream(ctx, state.History)
		if err != nil {
			return nil, err
		}
		defer stream.Close()

		var fullResp *schema.Message
		for {
			chunk, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				return nil, err
			}

			// 如果有内容或元数据，则通过 StreamWriter 实时发送增量
			if ok && (chunk.Content != "" || chunk.ResponseMeta != nil) {
				sw.Send(chunk, nil)
			}

			if chunk.ResponseMeta != nil && chunk.ResponseMeta.Usage != nil {
				state.PromptTokens += chunk.ResponseMeta.Usage.PromptTokens
				state.CompletionTokens += chunk.ResponseMeta.Usage.CompletionTokens
				state.TotalTokens += chunk.ResponseMeta.Usage.TotalTokens
			}

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
					mergeToolCalls(fullResp, chunk.ToolCalls)
				}
			}
		}

		state.History = append(state.History, fullResp)
		return state, nil
	}, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	err = g.AddLambdaNode("model", modelNode)
	if err != nil {
		return nil, err
	}

	// 2. Tool Node: 负责执行工具
	toolNode, err := compose.AnyLambda(func(ctx context.Context, state *AgentState, opts ...any) (*AgentState, error) {
		lastMsg := state.History[len(state.History)-1]
		if len(lastMsg.ToolCalls) == 0 {
			return state, nil
		}

		for _, tc := range lastMsg.ToolCalls {
			var targetTool tool.BaseTool
			for _, t := range b.tools {
				info, err := t.Info(ctx)
				if err == nil && info != nil && info.Name == tc.Function.Name {
					targetTool = t
					break
				}
			}

			if targetTool == nil {
				state.History = append(state.History, &schema.Message{Role: schema.Tool, Content: "工具未找到", ToolCallID: tc.ID})
				continue
			}

			if invokable, ok := targetTool.(tool.InvokableTool); ok {
				args := strings.TrimSpace(tc.Function.Arguments)
				
				// Zero Trust: 注入 UserID 到上下文
				toolCtx := context.WithValue(ctx, "user_id", state.UserID)
				// 生命周期治理: 设置 20s 超时
				toolCtx, cancel := context.WithTimeout(toolCtx, 20*time.Second)
				
				res, err := invokable.InvokableRun(toolCtx, args)
				cancel() // 及时释放资源

				if err != nil {
					errMsg := err.Error()
					if err == context.DeadlineExceeded {
						errMsg = "工具执行超时 (20s)，请优化查询或稍后再试"
					}
					state.History = append(state.History, &schema.Message{Role: schema.Tool, Content: errMsg, ToolCallID: tc.ID})
				} else {
					truncatedRes := TruncateToolResult(res)
					state.History = append(state.History, &schema.Message{Role: schema.Tool, Content: truncatedRes, ToolCallID: tc.ID})
				}
			}
		}
		return state, nil
	}, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	err = g.AddLambdaNode("tools", toolNode)
	if err != nil {
		return nil, err
	}

	// 3. 边与路由
	g.AddEdge(compose.START, "model")
	
	// 路由：根据是否有工具调用决定是去 tool node 还是结束
	err = g.AddBranch("model", compose.NewGraphBranch[*AgentState](func(ctx context.Context, state *AgentState) (string, error) {
		lastMsg := state.History[len(state.History)-1]
		if len(lastMsg.ToolCalls) > 0 {
			return "tools", nil
		}
		return compose.END, nil
	}, map[string]bool{"tools": true, compose.END: true}))
	if err != nil {
		return nil, err
	}

	g.AddEdge("tools", "model")

	return g.Compile(ctx)
}

func mergeToolCalls(fullResp *schema.Message, newCalls []schema.ToolCall) {
	for _, tc := range newCalls {
		found := false
		for j := range fullResp.ToolCalls {
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
