package workflow

import (
	"context"
	"fmt"
	"io"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/tomato/backend/einox/agent"
	"github.com/tomato/backend/einox/callback"
	"github.com/tomato/backend/einox/model"
)

// Orchestrator 负责根据执行计划调度 Worker Agents
type Orchestrator struct {
	TaskAgent    *agent.TaskAgent
	StudyAgent   *agent.StudyAgent
	EmotionAgent *agent.EmotionAgent
	GeneralAgent *agent.GeneralAgent
	CheapModel   *model.ChatModel
	Callback     *callback.LoggingCallback
}

func NewOrchestrator(task *agent.TaskAgent, study *agent.StudyAgent, emotion *agent.EmotionAgent, general *agent.GeneralAgent, cheap *model.ChatModel, cb *callback.LoggingCallback) *Orchestrator {
	return &Orchestrator{
		TaskAgent:    task,
		StudyAgent:   study,
		EmotionAgent: emotion,
		GeneralAgent: general,
		CheapModel:   cheap,
		Callback:     cb,
	}
}

// ExecutePlan 执行计划并流式返回结果
func (o *Orchestrator) ExecutePlan(ctx context.Context, input *CoachChatInput, plan *agent.ExecutionPlan) (*schema.StreamReader[*schema.Message], error) {
	state := agent.NewAgentState(input.UserID, input.Query, input.History)
	state.TaskSummary = input.TaskSummary
	state.UserProfile = input.UserProfile
	state.EpisodicContext = input.EpisodicContext
	state.UseKnowledge = input.UseKnowledge
	state.ChatMode = input.ChatMode
	state.Plan = plan

	return o.ExecutePlanWithState(ctx, state)
}

// ExecutePlanWithState 使用 AgentState 执行计划
func (o *Orchestrator) ExecutePlanWithState(ctx context.Context, state *agent.AgentState) (*schema.StreamReader[*schema.Message], error) {
	sr, sw := schema.Pipe[*schema.Message](10)

	go func() {
		defer sw.Close()

		if state.Plan == nil || len(state.Plan.Steps) == 0 {
			sw.Send(nil, fmt.Errorf("no execution plan provided"))
			return
		}

		// 构建动态 DAG 图
		g := compose.NewGraph[*agent.AgentState, *agent.AgentState]()
		
		// 1. 添加所有步骤作为节点
		for _, step := range state.Plan.Steps {
			stepCopy := step // 闭包捕获
			node, err := compose.AnyLambda(func(ctx context.Context, s *agent.AgentState, opts ...any) (*agent.AgentState, error) {
				// 为每个步骤开启一个 Span 追踪
				stepCtx := ctx
				var stepInfo *callback.RunInfo
				if o.Callback != nil {
					stepInfo = &callback.RunInfo{Name: fmt.Sprintf("Step: %s", stepCopy.ID)}
					stepCtx = o.Callback.OnStart(ctx, stepInfo, nil)
					defer o.Callback.OnEnd(stepCtx, stepInfo, nil)
				}

				// 发送步骤开始标记到流
				if swVal := ctx.Value(agent.StreamContextKey); swVal != nil {
					if sw, ok := swVal.(*schema.StreamWriter[*schema.Message]); ok {
						sw.Send(&schema.Message{
							Role:    schema.Assistant,
							Content: fmt.Sprintf("\n> [正在执行: %s]\n", stepCopy.ID),
						}, nil)
					}
				}

				var stepSr *schema.StreamReader[*schema.Message]
				var err error
				
				// 准备步骤特有的 State，使用 Clone 避免 RWMutex 拷贝警告且保证并发安全
				stepState := s.Clone()
				stepState.Query = stepCopy.Query

				// 优化：多步规划中的子步骤不需要沉重的历史负担
				// 核心优化：仅保留最近一轮历史以维持指代消解能力，其余截断
				if len(s.History) > 2 {
					stepState.History = s.History[len(s.History)-2:]
				} else {
					stepState.History = s.History
				}

				// 如果是任务 Agent，摘要背景是多余的，因为它会实时查库
				if stepCopy.Agent == "task" {
					stepState.TaskSummary = "" 
				}
				// 强制降低子步骤的记忆压力
				if len(stepState.EpisodicContext) > 500 {
					stepState.EpisodicContext = stepState.EpisodicContext[:500] + "...(truncated)"
				}

				switch stepCopy.Agent {
				case "task":
					stepSr, err = o.TaskAgent.StreamGenerateWithState(stepCtx, stepState)
				case "study":
					stepSr, err = o.StudyAgent.StreamGenerateWithState(stepCtx, stepState)
				case "emotion":
					stepSr, err = o.EmotionAgent.StreamGenerateWithState(stepCtx, stepState)
				default:
					stepSr, err = o.GeneralAgent.StreamGenerateWithState(stepCtx, stepState)
				}

				if err != nil {
					return nil, fmt.Errorf("step %s failed: %w", stepCopy.ID, err)
				}

				// 合并子步骤流到主流
				var fullContent string
				if stepSr != nil {
					for {
						msg, err := stepSr.Recv()
						if err != nil {
							if err == io.EOF {
								break
							}
							return nil, err
						}
						swVal := ctx.Value(agent.StreamContextKey)
						if sw, ok := swVal.(*schema.StreamWriter[*schema.Message]); ok {
							if msg.Content != "" || msg.ResponseMeta != nil {
								sw.Send(msg, nil)
							}
						}
						if msg.ResponseMeta != nil && msg.ResponseMeta.Usage != nil {
							state.PromptTokens += msg.ResponseMeta.Usage.PromptTokens
							state.CompletionTokens += msg.ResponseMeta.Usage.CompletionTokens
							state.TotalTokens += msg.ResponseMeta.Usage.TotalTokens
						}
						fullContent += msg.Content
					}
					stepSr.Close()
				}

				// 安全地将结果存入 Scratchpad (带脱水蒸馏逻辑)
				contentToStore := fullContent
				if len([]rune(fullContent)) > 300 && o.CheapModel != nil {
					fmt.Printf("[Orchestrator] 步骤 %s 结果过长 (%d 字)，启动级联脱水...\n", stepCopy.ID, len([]rune(fullContent)))
					summaryResp, err := o.CheapModel.Generate(ctx, []*schema.Message{
						schema.SystemMessage("你是一个任务上下文压缩专家。请为 Agent 的执行结果生成简短摘要（100字以内），作为后续步骤的参考。要求：保留所有核心结论、数字和关键实体。"),
						schema.UserMessage(fullContent),
					})
					if err == nil && summaryResp.Content != "" {
						contentToStore = "(摘要版) " + summaryResp.Content
						fmt.Printf("[Orchestrator] 脱水完成，压缩至 %d 字\n", len([]rune(contentToStore)))
					}
				}

				s.SetScratchpad(stepCopy.ID, contentToStore)
				return s, nil
			}, nil, nil, nil)
			if err != nil {
				sw.Send(nil, err)
				return
			}
			err = g.AddLambdaNode(stepCopy.ID, node)
			if err != nil {
				sw.Send(nil, err)
				return
			}
		}

		// 2. 根据依赖关系添加边
		for _, step := range state.Plan.Steps {
			if len(step.Dependencies) == 0 {
				g.AddEdge(compose.START, step.ID)
			} else {
				for _, dep := range step.Dependencies {
					g.AddEdge(dep, step.ID)
				}
			}
			// 标记有出边的节点，如果没有出边，则连向 END
			// 但 Eino Graph 如果没有连向 END 的边，Compile 可能报错或无法结束
		}
		
		// 找到所有没有后继节点的节点，连向 END
		for _, step := range state.Plan.Steps {
			isPredecessor := false
			for _, other := range state.Plan.Steps {
				for _, dep := range other.Dependencies {
					if dep == step.ID {
						isPredecessor = true
						break
					}
				}
				if isPredecessor { break }
			}
			if !isPredecessor {
				g.AddEdge(step.ID, compose.END)
			}
		}

		r, err := g.Compile(ctx)
		if err != nil {
			sw.Send(nil, err)
			return
		}

		// 执行
		graphCtx := context.WithValue(ctx, agent.StreamContextKey, sw)
		_, err = r.Invoke(graphCtx, state)
		if err != nil {
			sw.Send(nil, err)
			return
		}
		// 执行完成后，将最终聚合的 Token 使用量发送到流（作为元数据）
		sw.Send(&schema.Message{
			Role: schema.Assistant,
			ResponseMeta: &schema.ResponseMeta{
				Usage: &schema.TokenUsage{
					PromptTokens:     state.PromptTokens,
					CompletionTokens: state.CompletionTokens,
					TotalTokens:      state.TotalTokens,
				},
			},
		}, nil)
	}()

	return sr, nil
}
