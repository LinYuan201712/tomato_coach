// Package prompt 提供提示词模板组件
package prompt

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"gopkg.in/yaml.v3"
)

// PromptConfig 提示词配置结构
type PromptConfig struct {
	IntentRecognition    string `yaml:"intent_recognition"`
	TaskAgentSystem      string `yaml:"task_agent_system"`
	EmotionAgentSystem   string `yaml:"emotion_agent_system"`
	StudyAgentSystem     string `yaml:"study_agent_system"`
	GeneralAgentSystem   string `yaml:"general_agent_system"`
	RAGAssistant         string `yaml:"rag_assistant"`
	PlannerAgentSystem   string `yaml:"planner_agent_system"`
}

var Config PromptConfig

// Templates 预定义的提示词模板
var Templates = struct {
	GeneralAssistant     *prompt.DefaultChatTemplate
	EmotionCompanion     *prompt.DefaultChatTemplate
	IntentRecognition    *prompt.DefaultChatTemplate
	RAGAssistant         *prompt.DefaultChatTemplate
	QueryRewriting       *prompt.DefaultChatTemplate
	QueryExtraction      *prompt.DefaultChatTemplate
}{}

func init() {
	// 默认初始化（硬编码兜底）
	initDefaultTemplates()

	// 尝试从 YAML 加载
	err := LoadConfig("einox/config/prompts.yaml")
	if err == nil {
		updateTemplatesFromConfig()
	}
}

func LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, &Config)
}

func initDefaultTemplates() {
	Templates.IntentRecognition = prompt.FromMessages(schema.FString,
		schema.SystemMessage(`你是一个意图识别助手。`),
		schema.UserMessage("{query}"),
	)
	
	Templates.GeneralAssistant = prompt.FromMessages(schema.FString,
		schema.SystemMessage(`你是一个智能助手。`),
		schema.MessagesPlaceholder("history", true),
		schema.UserMessage("{query}"),
	)

	Templates.RAGAssistant = prompt.FromMessages(schema.FString,
		schema.SystemMessage(`你是一个基于知识库的问答助手。参考资料：{context}`),
		schema.MessagesPlaceholder("history", true),
		schema.UserMessage("{query}"),
	)

	Templates.EmotionCompanion = prompt.FromMessages(schema.FString,
		schema.SystemMessage(`你是一个极其温柔贴心的学习伴侣桌宠。`),
		schema.MessagesPlaceholder("history", true),
		schema.UserMessage("{query}"),
	)
}

func updateTemplatesFromConfig() {
	if Config.IntentRecognition != "" {
		Templates.IntentRecognition = prompt.FromMessages(schema.FString,
			schema.SystemMessage(Config.IntentRecognition),
			schema.UserMessage("{query}"),
		)
	}
	if Config.GeneralAgentSystem != "" {
		Templates.GeneralAssistant = prompt.FromMessages(schema.FString,
			schema.SystemMessage(Config.GeneralAgentSystem),
			schema.MessagesPlaceholder("history", true),
			schema.UserMessage("{query}"),
		)
	}
	if Config.EmotionAgentSystem != "" {
		Templates.EmotionCompanion = prompt.FromMessages(schema.FString,
			schema.SystemMessage(Config.EmotionAgentSystem),
			schema.MessagesPlaceholder("history", true),
			schema.UserMessage("{query}"),
		)
	}
	if Config.RAGAssistant != "" {
		Templates.RAGAssistant = prompt.FromMessages(schema.FString,
			schema.SystemMessage(Config.RAGAssistant),
			schema.MessagesPlaceholder("history", true),
			schema.UserMessage("{query}"),
		)
	}
}

// WrapInTag 将内容包装在 XML 标签中，用于结构化 Prompt 工程
func WrapInTag(tag, content string) string {
	if content == "" {
		return ""
	}
	return fmt.Sprintf("<%s>\n%s\n</%s>", tag, content, tag)
}

// AssembleMessages 统一的消息组装入口
func AssembleMessages(systemPrompt string, query string, history []*schema.Message, items []ContextItem, budget int) []*schema.Message {
	if budget <= 0 {
		budget = 6000
	}
	cm := NewContextManager(budget)
	return cm.Assemble(systemPrompt, query, history, items)
}

// CountTokens 计算消息列表的 Token 数 (Legacy)
func CountTokens(msgs []*schema.Message) int {
	cm := NewContextManager(0)
	return cm.CountMessagesTokens(msgs)
}

// TruncateString 截断字符串 (基于字符，Legacy)
func TruncateString(s string, maxChars int) string {
	if len(s) <= maxChars {
		return s
	}
	return s[:maxChars] + "... [已截断]"
}

// TruncateMessages 根据 Token 预算截断历史记录 (Legacy)
func TruncateMessages(msgs []*schema.Message, budget int) []*schema.Message {
	cm := NewContextManager(budget)
	// 这里我们需要一个简单的截断逻辑，不涉及系统提示词组装
	if cm.CountMessagesTokens(msgs) <= budget {
		return msgs
	}

	var systemMsgs []*schema.Message
	var otherMsgs []*schema.Message

	for _, m := range msgs {
		if m.Role == schema.System {
			systemMsgs = append(systemMsgs, m)
		} else {
			otherMsgs = append(otherMsgs, m)
		}
	}

	if len(otherMsgs) == 0 {
		return systemMsgs
	}

	currentTokens := cm.CountMessagesTokens(systemMsgs)
	var keep []*schema.Message
	for i := len(otherMsgs) - 1; i >= 0; i-- {
		m := otherMsgs[i]
		tokens := cm.CountMessagesTokens([]*schema.Message{m})
		if currentTokens+tokens > budget && len(keep) > 0 {
			break
		}
		keep = append([]*schema.Message{m}, keep...)
		currentTokens += tokens
	}

	return append(systemMsgs, keep...)
}

// FormatRAGMessages 格式化 RAG 消息
func FormatRAGMessages(ctx context.Context, query string, context string, history []*schema.Message) ([]*schema.Message, error) {
	// 增强提示词，确保引用格式稳定
	enhancedContext := fmt.Sprintf("%s\n\n注意：请确保在引述资料后标注 [Ref n]，并在回答末尾完整列出来源清单。必须严格按照 [[文件名]] 格式输出引证。", context)
	
	items := []ContextItem{
		{Tag: "context", Content: enhancedContext, Priority: P2},
	}
	
	systemPrompt := "你是一个基于知识库的问答助手。"
	if Config.RAGAssistant != "" {
		systemPrompt = Config.RAGAssistant
	}
	
	return AssembleMessages(systemPrompt, query, history, items, 6000), nil
}

// FormatStudyMessages 专门为学习教练设计的格式化工具
func FormatStudyMessages(ctx context.Context, query string, context string, history []*schema.Message, systemPrompt string) ([]*schema.Message, error) {
	items := []ContextItem{
		{Tag: "knowledge_context", Content: context, Priority: P2},
	}
	
	// 注意：这里的 systemPrompt 已经是传入的完整指令了
	return AssembleMessages(systemPrompt, query, history, items, 6000), nil
}
