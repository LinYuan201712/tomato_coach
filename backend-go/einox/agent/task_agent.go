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

type TaskAgent struct {
	chatModel *model.ChatModel
	registry  tools.ToolRegistry
}

func NewTaskAgent(chatModel *model.ChatModel, registry tools.ToolRegistry) *TaskAgent {
	return &TaskAgent{
		chatModel: chatModel,
		registry:  registry,
	}
}

func (a *TaskAgent) Generate(ctx context.Context, userID int64, query string, taskSummary string, userProfile string, episodicContext string) (*schema.Message, error) {
	// Tool RAG: 动态检索最相关的工具
	var relevantTools []tool.BaseTool
	if a.registry != nil {
		relevantTools, _ = a.registry.Retrieve(ctx, query, 3)
	}
	toolInfos := make([]*schema.ToolInfo, 0, len(relevantTools))
	for _, t := range relevantTools {
		info, _ := t.Info(ctx)
		if info != nil {
			toolInfos = append(toolInfos, info)
		}
	}
	a.chatModel.BindTools(toolInfos)

	systemPrompt := prompt.Config.TaskAgentSystem
	if systemPrompt == "" {
		systemPrompt = "你是一个任务管理专家。你可以帮助用户创建、查看或完成学习任务。"
	}
	systemPrompt = strings.ReplaceAll(systemPrompt, "{user_id}", fmt.Sprintf("%d", userID))
	systemPrompt = strings.ReplaceAll(systemPrompt, "{task_summary}", taskSummary)
	
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

	// Zero Trust: 确保 userID 注入上下文供工具调用使用
	ctx = context.WithValue(ctx, "user_id", userID)

	return RunToolLoop(ctx, a.chatModel, msgs, relevantTools)
}

func (a *TaskAgent) StreamGenerate(ctx context.Context, userID int64, query string, history []*schema.Message, useKnowledge bool, userProfile string, episodicContext string, taskSummary string, chatMode string) (*schema.StreamReader[*schema.Message], error) {
	state := NewAgentState(userID, query, history)
	state.TaskSummary = taskSummary
	state.UserProfile = userProfile
	state.EpisodicContext = episodicContext
	state.UseKnowledge = useKnowledge
	state.ChatMode = chatMode
	return a.StreamGenerateWithState(ctx, state)
}

func (a *TaskAgent) StreamGenerateWithState(ctx context.Context, state *AgentState) (*schema.StreamReader[*schema.Message], error) {
	// Tool RAG: 动态检索
	var relevantTools []tool.BaseTool
	if a.registry != nil {
		relevantTools, _ = a.registry.Retrieve(ctx, state.Query, 3)
	}
	toolInfos := make([]*schema.ToolInfo, 0, len(relevantTools))
	for _, t := range relevantTools {
		info, _ := t.Info(ctx)
		if info != nil {
			toolInfos = append(toolInfos, info)
		}
	}
	a.chatModel.BindTools(toolInfos)

	systemPrompt := prompt.Config.TaskAgentSystem
	if systemPrompt == "" {
		systemPrompt = "你是一个任务管理专家。你可以帮助用户创建、查看或完成学习任务。"
	}
	systemPrompt = strings.ReplaceAll(systemPrompt, "{user_id}", fmt.Sprintf("%d", state.UserID))
	
	budget := 4000
	items := []prompt.ContextItem{
		{Tag: "user_profile", Content: state.UserProfile, Priority: prompt.P2},
	}

	if state.ChatMode == "fast" {
		budget = 1500
		fmt.Printf("[TaskAgent] 快速模式：精简上下文 (Budget: 1500)\n")
	} else {
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
