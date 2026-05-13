package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/tomato/backend/einox/model"
	"github.com/tomato/backend/einox/prompt"
	"github.com/tomato/backend/einox/rag"
	"github.com/tomato/backend/einox/tools"
)

type GeneralAgent struct {
	chatModel *model.ChatModel
	rag       *rag.SimpleRAG
	registry  tools.ToolRegistry
}

func NewGeneralAgent(chatModel *model.ChatModel, r *rag.SimpleRAG, registry tools.ToolRegistry) *GeneralAgent {
	return &GeneralAgent{
		chatModel: chatModel,
		rag:       r,
		registry:  registry,
	}
}

func (a *GeneralAgent) Generate(ctx context.Context, userID int64, query string, history []*schema.Message, useKnowledge bool, taskSummary string, userProfile string, episodicContext string) (*schema.Message, error) {
	var relevantTools []tool.BaseTool
	if a.registry != nil {
		// Agentic RAG: 允许 Agent 发现搜索工具
		relevantTools, _ = a.registry.Retrieve(ctx, query, 5)
		toolInfos := make([]*schema.ToolInfo, 0, len(relevantTools))
		for _, t := range relevantTools {
			info, _ := t.Info(ctx)
			if info != nil {
				toolInfos = append(toolInfos, info)
			}
		}
		a.chatModel.BindTools(toolInfos)
	} else {
		a.chatModel.BindTools(nil)
	}

	msgs, err := prompt.Templates.GeneralAssistant.Format(ctx, map[string]any{
		"query":        query,
		"history":      history,
		"user_id":      userID,
		"task_summary": taskSummary,
	})
	if err != nil {
		fmt.Printf("[ERROR] Generate GeneralAssistant format failed: %v\n", err)
		msgs = []*schema.Message{schema.UserMessage(query)}
	}

	systemPrompt := prompt.Config.GeneralAgentSystem
	if systemPrompt == "" {
		systemPrompt = "你是一个智能助手。请作为一个知识渊博的学习助手，以亲切自然的方式回答用户。"
	}
	systemPrompt = strings.ReplaceAll(systemPrompt, "{user_id}", fmt.Sprintf("%d", userID))
	
	if taskSummary != "" {
		systemPrompt = fmt.Sprintf("%s\n\n%s", systemPrompt, prompt.WrapInTag("task_status", taskSummary))
	}
	if userProfile != "" {
		systemPrompt = fmt.Sprintf("%s\n\n%s", systemPrompt, prompt.WrapInTag("user_profile", userProfile))
	}
	if episodicContext != "" {
		systemPrompt = fmt.Sprintf("%s\n\n%s", systemPrompt, prompt.WrapInTag("historical_context", episodicContext))
	}

	finalMsgs := make([]*schema.Message, 0, len(msgs)+1)
	finalMsgs = append(finalMsgs, schema.SystemMessage(systemPrompt))
	finalMsgs = append(finalMsgs, msgs...)

	if a.registry != nil {
		// Zero Trust: 注入 userID
		ctx = context.WithValue(ctx, "user_id", userID)
		return RunToolLoop(ctx, a.chatModel, finalMsgs, relevantTools)
	}
	return a.chatModel.Generate(ctx, finalMsgs)
}

func (a *GeneralAgent) StreamGenerate(ctx context.Context, userID int64, query string, history []*schema.Message, useKnowledge bool, userProfile string, episodicContext string, taskSummary string, chatMode string) (*schema.StreamReader[*schema.Message], error) {
	state := NewAgentState(userID, query, history)
	state.TaskSummary = taskSummary
	state.UserProfile = userProfile
	state.EpisodicContext = episodicContext
	state.UseKnowledge = useKnowledge
	state.ChatMode = chatMode
	return a.StreamGenerateWithState(ctx, state)
}

func (a *GeneralAgent) StreamGenerateWithState(ctx context.Context, state *AgentState) (*schema.StreamReader[*schema.Message], error) {
	var relevantTools []tool.BaseTool
	if a.registry != nil {
		relevantTools, _ = a.registry.Retrieve(ctx, state.Query, 5)
		toolInfos := make([]*schema.ToolInfo, 0, len(relevantTools))
		for _, t := range relevantTools {
			info, _ := t.Info(ctx)
			if info != nil {
				toolInfos = append(toolInfos, info)
			}
		}
		a.chatModel.BindTools(toolInfos)
	} else {
		a.chatModel.BindTools(nil)
	}

	systemPrompt := prompt.Config.GeneralAgentSystem
	if systemPrompt == "" {
		systemPrompt = "你是一个智能助手。请作为一个知识渊博的学习助手，以亲切自然的方式回答用户。"
	}
	systemPrompt = strings.ReplaceAll(systemPrompt, "{user_id}", fmt.Sprintf("%d", state.UserID))

	// 模式决策
	budget := 8000
	items := []prompt.ContextItem{
		{Tag: "user_profile", Content: state.UserProfile, Priority: prompt.P2},
	}

	// 注入知识库上下文 (RAG)
	if state.UseKnowledge && a.rag != nil {
		// 记录当前时间用于 RAG 里的时间感知查询
		timeNow := time.Now().Format("2006-01-02 15:04:05")
		contextStr, docs, _ := a.rag.ProfessionalQuery(ctx, state.Query, timeNow, "通用知识库", state.UserID)
		if len(docs) > 0 {
			items = append(items, prompt.ContextItem{Tag: "knowledge_context", Content: contextStr, Priority: prompt.P2})
		}
	}

	if state.ChatMode == "fast" {
		budget = 1500
		fmt.Printf("[GeneralAgent] 快速模式：精简上下文 (Budget: 1500)\n")
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
	if a.registry != nil {
		return RunGraphReAct(ctx, a.chatModel, state, relevantTools)
	}
	return a.chatModel.Stream(ctx, state.History)
}
