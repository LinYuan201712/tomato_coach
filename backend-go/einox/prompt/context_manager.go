package prompt

import (
	"fmt"
	"github.com/cloudwego/eino/schema"
	"github.com/pkoukk/tiktoken-go"
)

// Priority 定义上下文优先级
type Priority int

const (
	P0 Priority = iota // 最高优先级：系统指令、当前 Query
	P1                 // 高优先级：短期对话历史
	P2                 // 次要优先级：RAG、背景知识、用户画像等
)

// ContextItem 动态上下文条目
type ContextItem struct {
	Tag      string
	Content  string
	Priority Priority
}

// ContextManager 上下文管理器
type ContextManager struct {
	Budget   int
	encoding string
	tkm      *tiktoken.Tiktoken
}

func NewContextManager(budget int) *ContextManager {
	if budget <= 0 {
		budget = 6000
	}
	
	encoding := "cl100k_base"
	tkm, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		// 降级处理
		return &ContextManager{Budget: budget, encoding: encoding}
	}

	return &ContextManager{
		Budget:   budget,
		encoding: encoding,
		tkm:      tkm,
	}
}

// CountStringTokens 计算字符串的 Token 数
func (m *ContextManager) CountStringTokens(s string) int {
	if m.tkm == nil {
		return len(s) / 4 // 粗略估算
	}
	return len(m.tkm.Encode(s, nil, nil))
}

// CountMessagesTokens 计算消息列表的 Token 数
func (m *ContextManager) CountMessagesTokens(msgs []*schema.Message) int {
	total := 0
	for _, msg := range msgs {
		total += m.CountStringTokens(msg.Content)
		total += m.CountStringTokens(string(msg.Role))
		total += 4 // 每条消息的基础开销
	}
	return total + 3 // 整体结尾开销
}

// Assemble 组装上下文消息列表
func (m *ContextManager) Assemble(systemPrompt string, query string, history []*schema.Message, items []ContextItem) []*schema.Message {
	p0System := schema.SystemMessage(systemPrompt)
	p0Query := schema.UserMessage(query)
	
	currentTokens := m.CountMessagesTokens([]*schema.Message{p0System, p0Query})

	// 1. 处理 P2 (动态上下文/RAG) - 优先分配
	var p2Msgs []*schema.Message
	for _, item := range items {
		if item.Content == "" {
			continue
		}
		taggedContent := WrapInTag(item.Tag, item.Content)
		msg := schema.SystemMessage(taggedContent)
		tokens := m.CountMessagesTokens([]*schema.Message{msg})
		
		// 策略：优先满足知识，预留 500 给回复
		if currentTokens+tokens > m.Budget-500 {
			fmt.Printf("[PromptManager] ⚠️ 预算不足，跳过条目: %s (Tokens: %d)\n", item.Tag, tokens)
			continue 
		}
		
		p2Msgs = append(p2Msgs, msg)
		currentTokens += tokens
	}

	// 2. 处理 P1 (历史记录) - 使用剩余空间
	var p1Keep []*schema.Message
	for i := len(history) - 1; i >= 0; i-- {
		msg := history[i]
		tokens := m.CountMessagesTokens([]*schema.Message{msg})
		
		if currentTokens+tokens > m.Budget {
			fmt.Printf("[PromptManager] ✂️ 历史记录截断\n")
			break
		}
		
		p1Keep = append([]*schema.Message{msg}, p1Keep...)
		currentTokens += tokens
	}

	// 3. 最终组装
	// 策略：将 P2 (知识库) 放在历史记录之后，紧贴用户 Query，以增强模型的注意力。
	res := make([]*schema.Message, 0, 2+len(p2Msgs)+len(p1Keep))
	res = append(res, p0System)
	res = append(res, p1Keep...)
	res = append(res, p2Msgs...)
	res = append(res, p0Query)

	return res
}
