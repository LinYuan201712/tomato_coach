package workflow

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"github.com/tomato/backend/einox/agent"
	"github.com/tomato/backend/einox/callback"
	"github.com/tomato/backend/einox/model"
	"github.com/tomato/backend/einox/prompt"
	"github.com/tomato/backend/einox/rag"
	"github.com/tomato/backend/einox/tools"

)

type CoachChatInput struct {
	UserID       int64
	Query        string
	History      []*schema.Message
	UseKnowledge bool   // 知识库增强开关
	ChatMode     string // fast | thinking
	TaskSummary     string // 任务概况（背景辐射）
	UserProfile     string // 用户画像上下文
	EpisodicContext string // 召回的历史情景记忆
}

func BuildCoachChatGraph(ctx context.Context, smartModel *model.ChatModel, cheapModel *model.ChatModel, r *rag.SimpleRAG, toolRegistry tools.ToolRegistry, cb *callback.LoggingCallback) (compose.Runnable[*CoachChatInput, *schema.Message], error) {
	// 1. 初始化 Smart Agents (用于 Thinking 模式)
	taskAgent := agent.NewTaskAgent(smartModel, toolRegistry)
	studyAgent := agent.NewStudyAgent(smartModel, r, toolRegistry)
	emotionAgent := agent.NewEmotionAgent(smartModel, toolRegistry)
	generalAgent := agent.NewGeneralAgent(smartModel, r, toolRegistry)

	// 2. 初始化 Cheap Agents (用于 Fast 模式，极致节省 Token 和响应时间)
	cheapTaskAgent := agent.NewTaskAgent(cheapModel, toolRegistry)
	cheapStudyAgent := agent.NewStudyAgent(cheapModel, r, toolRegistry)
	cheapEmotionAgent := agent.NewEmotionAgent(cheapModel, toolRegistry)
	cheapGeneralAgent := agent.NewGeneralAgent(cheapModel, r, toolRegistry)

	plannerAgent := agent.NewPlannerAgent(smartModel) 
	orchestrator := NewOrchestrator(taskAgent, studyAgent, emotionAgent, generalAgent, cheapModel, cb)

	g := compose.NewGraph[*CoachChatInput, *schema.Message]()

	finalStream := func(ctx context.Context, input *CoachChatInput, opts ...any) (*schema.StreamReader[*schema.Message], error) {
		// 路由逻辑
		if input.ChatMode == "fast" {
			fmt.Printf("[CoachChat] 走快速路径 (Fast Path)\n")
			intent := detectIntent(ctx, cheapModel, input)
			return dispatchSimple(ctx, intent, input, cheapTaskAgent, cheapEmotionAgent, cheapStudyAgent, cheapGeneralAgent)
		}

		fmt.Printf("[CoachChat] 走深度思考路径 (Thinking Path)\n")
		// 1. 生成执行计划
		plan, err := plannerAgent.Plan(ctx, input.Query, input.History)
		if err != nil {
			fmt.Printf("[Planner] 规划失败，退回简单分发: %v\n", err)
			intent := detectIntent(ctx, cheapModel, input)
			return dispatchSimple(ctx, intent, input, taskAgent, emotionAgent, studyAgent, generalAgent)
		}

		// 2. 调度执行计划
		fmt.Printf("[Planner] 执行计划生成的步骤数: %d\n", len(plan.Steps))
		ctx = context.WithValue(ctx, "user_id", input.UserID)
		return orchestrator.ExecutePlan(ctx, input, plan)
	}

	coachNode, err := compose.AnyLambda[*CoachChatInput, *schema.Message, any](nil, finalStream, nil, nil)
	if err != nil {
		return nil, err
	}

	err = g.AddLambdaNode("coach", coachNode)
	if err != nil {
		return nil, err
	}
	g.AddEdge(compose.START, "coach")
	g.AddEdge("coach", compose.END)

	return g.Compile(ctx)
}

func dispatchSimple(ctx context.Context, intent string, input *CoachChatInput, taskAgent *agent.TaskAgent, emotionAgent *agent.EmotionAgent, studyAgent *agent.StudyAgent, generalAgent *agent.GeneralAgent) (*schema.StreamReader[*schema.Message], error) {
	ctx = context.WithValue(ctx, "user_id", input.UserID)
	state := agent.NewAgentState(input.UserID, input.Query, input.History)
	state.TaskSummary = input.TaskSummary
	state.UserProfile = input.UserProfile
	state.EpisodicContext = input.EpisodicContext
	state.UseKnowledge = input.UseKnowledge
	state.ChatMode = input.ChatMode

	switch intent {
	case "task":
		return taskAgent.StreamGenerateWithState(ctx, state)
	case "emotion":
		return emotionAgent.StreamGenerateWithState(ctx, state)
	case "study":
		return studyAgent.StreamGenerateWithState(ctx, state)
	default:
		return generalAgent.StreamGenerateWithState(ctx, state)
	}
}

func detectIntent(ctx context.Context, cheapModel *model.ChatModel, input *CoachChatInput) string {
	// 简单的关键词先验判断，减少意图识别的模糊性
	q := strings.ToLower(input.Query)
	if strings.Contains(q, "创建") || strings.Contains(q, "增加") || strings.Contains(q, "任务") || strings.Contains(q, "完成") {
		return "task"
	}

	// 走大模型识别
	systemPrompt := prompt.Config.IntentRecognition
	if systemPrompt == "" {
		systemPrompt = "你是一个意图识别助手。请分析用户的输入并将其归类为以下意图之一：emotion, task, study, general。"
	}

	// 注入画像以帮助识别意图（可选）
	if input.UserProfile != "" {
		systemPrompt = fmt.Sprintf("%s\n\n当前用户背景：%s", systemPrompt, input.UserProfile)
	}

	intentMsgs := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(input.Query),
	}
	cheapModel.BindTools(nil)
	intentResp, err := cheapModel.Generate(ctx, intentMsgs)
	if err != nil {
		return "general"
	}
	content := strings.ToLower(intentResp.Content)
	if strings.Contains(content, "task") { return "task" }
	if strings.Contains(content, "emotion") { return "emotion" }
	if strings.Contains(content, "study") { return "study" }
	return "general"
}
