package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"
	"github.com/tomato/backend/einox/model"
)

// UserFact 用户事实信息
type UserFact struct {
	Op    string `json:"op"`    // 操作：add (新增/追加), update (更新/替换), delete (删除)
	Key   string `json:"key"`   // 类别：learning_goal, preferred_style, strength, weakness, other
	Value string `json:"value"` // 具体内容
}

// FactExtractor 事实提取器
type FactExtractor struct {
	chatModel *model.ChatModel
}

func NewFactExtractor(chatModel *model.ChatModel) *FactExtractor {
	return &FactExtractor{
		chatModel: chatModel,
	}
}

// Extract 从最近的对话中提取用户画像相关事实
func (e *FactExtractor) Extract(ctx context.Context, query string, reply string, currentProfile string) ([]UserFact, error) {
	prompt := fmt.Sprintf(`你是一个用户信息提取专家。请从以下对话中提取关于用户的持久事实（学习目标、偏好风格、长处、短处）。

【当前已知的用户画像】:
%s

【新对话内容】:
用户输入：%s
助手回复：%s

请以 JSON 数组格式输出提取到的事实操作指令。
策略指南（非常重要）：
1. 查重比对：如果新对话提到的事实已经存在于上面的【当前已知画像】中，且含义没有发生变化，请直接输出 []，不要重复提取。
2. 模糊合并：如果新对话只是对已有目标的细微补充（例如从“学Go”变成“学Go并发”），请使用 "update" 操作。
3. op 类型说明：
   - "add": 增加完全不同的新信息。
   - "update": 修正或大幅度丰富已有的旧信息。
   - "delete": 用户明确表示放弃某个目标或偏好。

只输出 JSON 数组，如果没有新信息，输出 []。`, currentProfile, query, reply)

	msgs := []*schema.Message{
		schema.SystemMessage("你是一个专业的记忆提取器。"),
		schema.UserMessage(prompt),
	}

	resp, err := e.chatModel.Generate(ctx, msgs)
	if err != nil {
		return nil, err
	}

	content := strings.TrimSpace(resp.Content)
	if !strings.HasPrefix(content, "[") {
		// 尝试修复可能的 Markdown 包裹
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	var facts []UserFact
	err = json.Unmarshal([]byte(content), &facts)
	if err != nil {
		return nil, nil // 解析失败，返回空结果而非错误
	}

	return facts, nil
}
