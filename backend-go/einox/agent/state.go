package agent

import (
	"fmt"
	"sync"

	"github.com/cloudwego/eino/schema"
)

// AgentState 定义了 Agent 执行过程中的全局状态
type AgentState struct {
	// 基础上下文
	UserID int64
	Query  string
	// History 存储对话历史，包括工具调用的往返
	History []*schema.Message

	// 增强上下文 (从数据库或 RAG 召回)
	TaskSummary     string
	UserProfile     string
	EpisodicContext string

	// 调度与执行状态
	Plan           *ExecutionPlan
	CurrentStepIdx int
	// Scratchpad 用于在不同节点/步骤之间传递中间结果
	mu         sync.RWMutex
	Scratchpad map[string]any

	// 内部控制
	MaxReActIterations int
	UseKnowledge       bool   // 是否开启知识库增强
	ChatMode           string // fast | thinking

	// Token 统计 (累加)
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// NewAgentState 初始化一个新的状态
func NewAgentState(userID int64, query string, history []*schema.Message) *AgentState {
	return &AgentState{
		UserID:             userID,
		Query:              query,
		History:            history,
		Scratchpad:         make(map[string]any),
		MaxReActIterations: 5,
	}
}

// GetScratchpadString 将黑板内容格式化为 XML 字符串供 Prompt 使用
func (s *AgentState) GetScratchpadString() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.Scratchpad) == 0 {
		return ""
	}
	var res string
	for k, v := range s.Scratchpad {
		res += fmt.Sprintf("<%s>\n%v\n</%s>\n", k, v, k)
	}
	return res
}

// SetScratchpad 安全地写入黑板
func (s *AgentState) SetScratchpad(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Scratchpad[key] = value
}

// Clone 创建一个状态的浅拷贝，用于多步规划中的子步骤执行，避免并发修改冲突
func (s *AgentState) Clone() *AgentState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 浅拷贝主要字段
	ns := &AgentState{
		UserID:             s.UserID,
		Query:              s.Query,
		History:            s.History,
		TaskSummary:        s.TaskSummary,
		UserProfile:        s.UserProfile,
		EpisodicContext:    s.EpisodicContext,
		Plan:               s.Plan,
		CurrentStepIdx:     s.CurrentStepIdx,
		Scratchpad:         make(map[string]any), // Scratchpad 不应共享，以免干扰
		MaxReActIterations: s.MaxReActIterations,
		UseKnowledge:       s.UseKnowledge,
		ChatMode:           s.ChatMode,
		PromptTokens:       s.PromptTokens,
		CompletionTokens:   s.CompletionTokens,
		TotalTokens:        s.TotalTokens,
	}

	// 复制当前的 Scratchpad 内容
	for k, v := range s.Scratchpad {
		ns.Scratchpad[k] = v
	}

	return ns
}
