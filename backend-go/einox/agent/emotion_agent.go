package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/tomato/backend/einox/model"
	"github.com/tomato/backend/einox/prompt"
	"github.com/tomato/backend/einox/tools"
)

type EmotionAgent struct {
	chatModel *model.ChatModel
	registry  tools.ToolRegistry
}

func NewEmotionAgent(chatModel *model.ChatModel, registry tools.ToolRegistry) *EmotionAgent {
	return &EmotionAgent{
		chatModel: chatModel,
		registry:  registry,
	}
}

func (a *EmotionAgent) Generate(ctx context.Context, userID int64, query string, history []*schema.Message, taskSummary string, userProfile string, episodicContext string) (*schema.Message, error) {
	// 如果提问非常简短（纯情绪发泄），直接走温柔陪伴模式
	if len(query) < 60 && !strings.Contains(query, "任务") && !strings.Contains(query, "计划") {
		// 可以在伴侣模式中也注入画像
		companionPrompt := "你是一个极其温柔贴心的学习伴侣桌宠。"
		if userProfile != "" {
			companionPrompt = fmt.Sprintf("%s\n%s", companionPrompt, prompt.WrapInTag("user_profile", userProfile))
		}
		if episodicContext != "" {
			companionPrompt = fmt.Sprintf("%s\n%s", companionPrompt, prompt.WrapInTag("historical_context", episodicContext))
		}
		
		msgs := []*schema.Message{
			schema.SystemMessage(companionPrompt),
			schema.UserMessage(query),
		}
		a.chatModel.BindTools(nil)
		return a.chatModel.Generate(ctx, msgs)
	}

	// Tool RAG
	var relevantTools []tool.BaseTool
	if a.registry != nil {
		relevantTools, _ = a.registry.Retrieve(ctx, query, 5)
	}
	toolInfos := make([]*schema.ToolInfo, 0, len(relevantTools))
	for _, t := range relevantTools {
		info, _ := t.Info(ctx)
		if info != nil {
			toolInfos = append(toolInfos, info)
		}
	}
	a.chatModel.BindTools(toolInfos)

	systemPrompt := prompt.Config.EmotionAgentSystem
	if systemPrompt == "" {
		systemPrompt = "你是一个贴心的学习伴侣。请针对用户的情绪 and 任务压力进行引导。"
	}
	if taskSummary != "" {
		systemPrompt = fmt.Sprintf("%s\n\n%s", systemPrompt, prompt.WrapInTag("task_status", taskSummary))
	}
	if userProfile != "" {
		systemPrompt = fmt.Sprintf("%s\n\n%s", systemPrompt, prompt.WrapInTag("user_profile", userProfile))
	}
	if episodicContext != "" {
		systemPrompt = fmt.Sprintf("%s\n\n%s", systemPrompt, prompt.WrapInTag("historical_context", episodicContext))
	}

	msgs := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(query),
	}

	// Zero Trust: 确保 userID 注入上下文
	ctx = context.WithValue(ctx, "user_id", userID)

	return RunToolLoop(ctx, a.chatModel, msgs, relevantTools)
}

func (a *EmotionAgent) StreamGenerate(ctx context.Context, userID int64, query string, history []*schema.Message, useKnowledge bool, userProfile string, episodicContext string, taskSummary string, chatMode string) (*schema.StreamReader[*schema.Message], error) {
	state := NewAgentState(userID, query, history)
	state.TaskSummary = taskSummary
	state.UserProfile = userProfile
	state.EpisodicContext = episodicContext
	state.UseKnowledge = useKnowledge
	state.ChatMode = chatMode
	return a.StreamGenerateWithState(ctx, state)
}

func (a *EmotionAgent) StreamGenerateWithState(ctx context.Context, state *AgentState) (*schema.StreamReader[*schema.Message], error) {
	// Tool RAG
	var relevantTools []tool.BaseTool
	if a.registry != nil {
		relevantTools, _ = a.registry.Retrieve(ctx, state.Query, 5)
	}
	toolInfos := make([]*schema.ToolInfo, 0, len(relevantTools))
	for _, t := range relevantTools {
		info, _ := t.Info(ctx)
		if info != nil {
			toolInfos = append(toolInfos, info)
		}
	}
	a.chatModel.BindTools(toolInfos)
	systemPrompt := prompt.Config.EmotionAgentSystem
	if systemPrompt == "" {
		systemPrompt = "你是一个极其温柔贴心的学习伴侣。请针对用户的情绪和任务压力进行引导。"
	}
	systemPrompt = strings.ReplaceAll(systemPrompt, "{user_id}", fmt.Sprintf("%d", state.UserID))

	// 模式决策：快速模式使用更紧凑的上下文
	budget := 4000
	items := []prompt.ContextItem{
		{Tag: "user_profile", Content: state.UserProfile, Priority: prompt.P2},
	}

	if state.ChatMode == "fast" {
		budget = 1500 // 快速模式大幅降低预算，强制截断历史
		fmt.Printf("[EmotionAgent] 快速模式：使用精简上下文 (Budget: 1500)\n")
	} else {
		// 思考模式下加入更深入的背景
		items = append(items, 
			prompt.ContextItem{Tag: "task_summary", Content: state.TaskSummary, Priority: prompt.P2},
			prompt.ContextItem{Tag: "episodic_context", Content: state.EpisodicContext, Priority: prompt.P2},
		)
	}

	// 注入执行黑板内容 (Stateful Scratchpad)
	if scratch := state.GetScratchpadString(); scratch != "" {
		items = append(items, prompt.ContextItem{Tag: "scratchpad", Content: scratch, Priority: prompt.P1})
	}

	state.History = prompt.AssembleMessages(systemPrompt, state.Query, state.History, items, budget)
	return RunGraphReAct(ctx, a.chatModel, state, relevantTools)
}
