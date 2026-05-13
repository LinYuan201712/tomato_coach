package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"
	"github.com/tomato/backend/einox/model"
	"github.com/tomato/backend/einox/prompt"
)

// PlanStep 代表执行计划中的一个步骤
type PlanStep struct {
	ID           string   `json:"id"`
	Agent        string   `json:"agent"` // task, study, emotion, general
	Query        string   `json:"query"`
	Dependencies []string `json:"dependencies"`
}

// ExecutionPlan 代表完整的执行计划
type ExecutionPlan struct {
	Steps []PlanStep `json:"steps"`
}

// PlannerAgent 负责将复杂请求拆解为执行计划
type PlannerAgent struct {
	model *model.ChatModel
}

func NewPlannerAgent(m *model.ChatModel) *PlannerAgent {
	return &PlannerAgent{model: m}
}

// Plan 生成执行计划
func (a *PlannerAgent) Plan(ctx context.Context, query string, history []*schema.Message) (*ExecutionPlan, error) {
	systemPrompt := prompt.Config.PlannerAgentSystem
	if systemPrompt == "" {
		return nil, fmt.Errorf("planner agent system prompt not configured")
	}

	msgs := prompt.AssembleMessages(systemPrompt, query, history, nil, 4000)

	resp, err := a.model.Generate(ctx, msgs)
	if err != nil {
		return nil, fmt.Errorf("planner model generation failed: %w", err)
	}

	content := strings.TrimSpace(resp.Content)
	// 尝试清洗 Markdown 代码块标记 (如 ```json ... ```)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var plan ExecutionPlan
	if err := json.Unmarshal([]byte(content), &plan); err != nil {
		return nil, fmt.Errorf("failed to parse execution plan JSON: %w\nContent: %s", err, content)
	}

	if len(plan.Steps) == 0 {
		return nil, fmt.Errorf("generated execution plan has no steps")
	}

	return &plan, nil
}
