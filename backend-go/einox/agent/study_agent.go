package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/tomato/backend/einox/model"
	"github.com/tomato/backend/einox/prompt"
	"github.com/tomato/backend/einox/rag"
	"github.com/tomato/backend/einox/tools"
)

type StudyAgent struct {
	chatModel *model.ChatModel
	rag       *rag.SimpleRAG
	registry  tools.ToolRegistry
}

func (a *StudyAgent) GenerateReport(ctx context.Context, userID int64, reportType string, reportData string) (*schema.Message, error) {
	systemPrompt := "你是一个极其专业、细致且充满鼓励的学习教练。你的任务是根据提供的数据，为用户生成一份具有深度反思和未来指导意义的学习报告。"
	if reportType == "daily" {
		systemPrompt += "\n当前正在生成【昨日学习日报】。"
	} else if reportType == "weekly" {
		systemPrompt += "\n当前正在生成【学习周报】。"
	}

	systemPrompt += "\n\n报告应包含以下板块：\n" +
		"1. 专注之星：总结专注时长和任务完成情况，用肯定的语气表扬坚持。\n" +
		"2. 深度思考：根据昨日对话摘要，提炼学习深度和核心知识点，帮用户梳理思路。\n" +
		"3. 教练点评：指出做得好的地方，并犀利地指出可能的瓶颈（如情绪波动、时长不足）。\n" +
		"4. 明日锦囊：提供 1-2 条具体的、科学的学习建议（如使用费曼技巧、交替学习法等）。" +
		"\n\n请使用 Markdown 格式输出，语言要生动、专业，像一个真实的教练在对话。"

	userQuery := fmt.Sprintf("这是我的学习数据和对话记录，请帮我生成报告：\n\n%s", reportData)

	msgs := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userQuery),
	}

	return a.chatModel.Generate(ctx, msgs)
}

func NewStudyAgent(chatModel *model.ChatModel, r *rag.SimpleRAG, registry tools.ToolRegistry) *StudyAgent {
	return &StudyAgent{
		chatModel: chatModel,
		rag:       r,
		registry:  registry,
	}
}

func (a *StudyAgent) Generate(ctx context.Context, userID int64, query string, history []*schema.Message, useKnowledge bool, userProfile string, episodicContext string) (*schema.Message, error) {
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

	var msgs []*schema.Message
	if useKnowledge && a.rag != nil {
		timeNow := time.Now().Format("2006-01-02 15:04:05")
		contextStr, docs, _ := a.rag.ProfessionalQuery(ctx, query, timeNow, "StudyCoach知识库", userID)
		if len(docs) > 0 {
			systemPrompt := prompt.Config.StudyAgentSystem
			if systemPrompt == "" {
				systemPrompt = "你是一个专业学习教练。"
			}
			if userProfile != "" {
				systemPrompt = fmt.Sprintf("%s\n\n【用户背景知识（仅作参考，请优先回答当前问题，不要强行关联）】\n%s", systemPrompt, userProfile)
			}
			if episodicContext != "" {
				systemPrompt = fmt.Sprintf("%s\n\n%s", systemPrompt, prompt.WrapInTag("historical_context", episodicContext))
			}
			msgs, _ = prompt.FormatStudyMessages(ctx, query, contextStr, history, systemPrompt)
		}
	}

	if len(msgs) == 0 {
		systemPrompt := "你是一个专业学习教练。你会直接基于你的内置知识为用户提供学习辅导。"
		if episodicContext != "" {
			systemPrompt = fmt.Sprintf("%s\n\n%s", systemPrompt, prompt.WrapInTag("historical_context", episodicContext))
		}
		msgs = []*schema.Message{
			schema.SystemMessage(systemPrompt),
			schema.UserMessage(query),
		}
	}

	// Zero Trust: 确保 userID 注入上下文
	ctx = context.WithValue(ctx, "user_id", userID)

	return RunToolLoop(ctx, a.chatModel, msgs, relevantTools)
}

func (a *StudyAgent) StreamGenerate(ctx context.Context, userID int64, query string, history []*schema.Message, useKnowledge bool, userProfile string, episodicContext string, taskSummary string, chatMode string) (*schema.StreamReader[*schema.Message], error) {
	state := NewAgentState(userID, query, history)
	state.UserProfile = userProfile
	state.EpisodicContext = episodicContext
	state.TaskSummary = taskSummary
	state.UseKnowledge = useKnowledge
	state.ChatMode = chatMode
	return a.StreamGenerateWithState(ctx, state)
}

func (a *StudyAgent) StreamGenerateWithState(ctx context.Context, state *AgentState) (*schema.StreamReader[*schema.Message], error) {
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

	systemPrompt := prompt.Config.StudyAgentSystem
	if systemPrompt == "" {
		systemPrompt = "你是一个专业学习教练。"
	}

	var knowledge string
	if state.UseKnowledge {
		if a.rag != nil {
			timeNow := time.Now().Format("2006-01-02 15:04:05")
			contextStr, docs, err := a.rag.ProfessionalQuery(ctx, state.Query, timeNow, "StudyCoach知识库", state.UserID)
			if err == nil && len(docs) > 0 {
				knowledge = contextStr
			}
		}
	}

	// 模式决策
	budget := 8000
	items := []prompt.ContextItem{
		{Tag: "user_profile", Content: state.UserProfile, Priority: prompt.P2},
		{Tag: "knowledge_context", Content: knowledge, Priority: prompt.P2},
	}

	if state.ChatMode == "fast" {
		budget = 1500
		fmt.Printf("[StudyAgent] 快速模式：精简上下文 (Budget: 1500)\n")
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
