package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"
	"github.com/tomato/backend/internal/domain/entity"
	"github.com/tomato/backend/internal/repository"
)

// UserProfile 用户画像
type UserProfile struct {
	LearningGoals  []string `json:"learning_goals"`
	PreferredStyle string   `json:"preferred_style"`
	Strengths      []string `json:"strengths"`
	Weaknesses     []string `json:"weaknesses"`
}

type Summarizer interface {
	Generate(ctx context.Context, msgs []*schema.Message) (*schema.Message, error)
}

// PersistentMemory 持久化记忆管理
type PersistentMemory struct {
	repo       repository.ChatRepository
	userRepo   repository.UserRepository
	summarizer Summarizer
	extractor  *FactExtractor
	episodic   EpisodicMemory
}

func NewPersistentMemory(repo repository.ChatRepository, userRepo repository.UserRepository, summarizer Summarizer, extractor *FactExtractor, episodic EpisodicMemory) *PersistentMemory {
	return &PersistentMemory{
		repo:       repo,
		userRepo:   userRepo,
		summarizer: summarizer,
		extractor:  extractor,
		episodic:   episodic,
	}
}

// GetHistory 获取历史记录（包含分层压缩逻辑）
func (m *PersistentMemory) GetUserRepo() repository.UserRepository {
	return m.userRepo
}

func (m *PersistentMemory) GetHistory(ctx context.Context, userID int64, sessionID string) ([]*schema.Message, error) {
	// 1. 获取会话信息（包含持久化摘要）
	session, err := m.repo.GetSession(ctx, sessionID)
	if err != nil {
		// 如果会话不存在，可能是一个新会话，继续执行
		session = &entity.ChatSession{SessionID: sessionID}
	}

	// 2. 获取最近的历史记录（默认取30条）
	dbMsgs, err := m.repo.GetHistory(ctx, sessionID, 30)
	if err != nil {
		return nil, err
	}

	var history []*schema.Message
	for _, m := range dbMsgs {
		msg := &schema.Message{
			Role:    schema.RoleType(m.Role),
			Content: m.Content,
		}
		if m.Reasoning != "" {
			msg.ReasoningContent = m.Reasoning
		}
		history = append(history, msg)
	}

	// 3. 压缩逻辑：如果最近的历史记录达到25条，则对前20条和旧摘要进行合并压缩
	if len(history) >= 25 {
		newSummary, err := m.compressHistory(ctx, session.Summary, history[:20])
		if err == nil {
			// 将新摘要持久化到数据库
			_ = m.repo.UpdateSessionSummary(ctx, sessionID, newSummary, 0)

			// 返回：新摘要 + 剩余的 5-10 条消息
			res := []*schema.Message{
				schema.SystemMessage(fmt.Sprintf("以下是之前的历史对话摘要：%s", newSummary)),
			}
			res = append(res, history[20:]...)
			return res, nil
		}
	}

	// 4. 如果不需要压缩，返回：持久化摘要（如有）+ 获取到的历史记录
	var res []*schema.Message
	if session.Summary != "" {
		res = append(res, schema.SystemMessage(fmt.Sprintf("以下是之前的历史对话摘要：%s", session.Summary)))
	}

	// 5. 增加情景记忆召回 (Episodic Memory)
	if m.episodic != nil && len(history) > 0 {
		// 使用最后一条用户提问（如果有）进行召回，或者使用当前会话的上下文
		// 这里简单处理：由外部在调用 Agent 时决定是否注入召回内容，
		// 或者在这里尝试为“当前查询”预留位置（但 GetHistory 此时不知道当前 Query）。
		// 
		// 改进：返回历史记录的同时，让外部决定何时调用 Recall。
		// 但为了保持接口兼容，我们可以在 System Message 中提示“系统具备长期记忆”。
	}

	res = append(res, history...)

	return res, nil
}

func (m *PersistentMemory) GetUserProfile(ctx context.Context, userID int64) (string, error) {
	user, err := m.userRepo.FindByUserID(ctx, userID)
	if err != nil || user == nil {
		return "", err
	}
	
	var sb strings.Builder
	sb.WriteString("【用户画像 (长期演变记录)】\n")
	if user.Goals != "" {
		sb.WriteString(fmt.Sprintf("- 核心目标与兴趣：%s\n", user.Goals))
	} else {
		sb.WriteString("- 核心目标与兴趣：尚未明确\n")
	}
	
	if user.PreferredStyle != "" {
		sb.WriteString(fmt.Sprintf("- 学习风格偏好：%s\n", user.PreferredStyle))
	} else {
		sb.WriteString("- 学习风格偏好：通用模式\n")
	}
	
	return sb.String(), nil
}

func (m *PersistentMemory) Recall(ctx context.Context, userID int64, query string) (string, error) {
	if m.episodic == nil {
		return "", nil
	}
	return m.episodic.Recall(ctx, userID, query)
}

// ProcessFacts 异步提取事实并更新画像
func (m *PersistentMemory) ProcessFacts(ctx context.Context, userID int64, query string, reply string) {
	if m.extractor == nil {
		return
	}
	go func() {
		user, err := m.userRepo.FindByUserID(ctx, userID)
		if err != nil || user == nil {
			return
		}

		// 获取当前画像作为提取上下文，防止重复提取
		currentProfile, _ := m.GetUserProfile(ctx, userID)

		facts, err := m.extractor.Extract(ctx, query, reply, currentProfile)
		if err != nil || len(facts) == 0 {
			return
		}

		// 根据锁定状态决定操作
		lockStatus := user.ProfileLock
		if lockStatus == "" {
			lockStatus = "soft" // 默认软锁定
		}

		if lockStatus == "hard" {
			return // 硬锁定，不进行任何更新
		}

		if lockStatus == "soft" {
			// 软锁定：存储建议而不是直接更新
			existingSuggestions := []UserFact{}
			if user.LockSuggestions != "" {
				_ = json.Unmarshal([]byte(user.LockSuggestions), &existingSuggestions)
			}
			
			// 合并新建议 (简单追加)
			existingSuggestions = append(existingSuggestions, facts...)
			
			// 限制建议数量，避免堆积过多
			if len(existingSuggestions) > 10 {
				existingSuggestions = existingSuggestions[len(existingSuggestions)-10:]
			}
			
			suggestionJSON, _ := json.Marshal(existingSuggestions)
			_ = m.userRepo.UpdateLockSuggestions(ctx, userID, string(suggestionJSON))
		} else {
			// Unlocked: 自动更新
			goals := user.Goals
			style := user.PreferredStyle

			for _, f := range facts {
				switch f.Key {
				case "learning_goal":
					switch f.Op {
					case "update":
						goals = f.Value
					case "delete":
						goals = ""
					default: // add
						if !strings.Contains(goals, f.Value) {
							if goals != "" { goals += "；" }
							goals += f.Value
						}
					}
				case "preferred_style":
					switch f.Op {
					case "delete":
						style = ""
					default: // update 或 add 都视为覆盖当前风格
						style = f.Value
					}
				}
			}

			_ = m.userRepo.UpdateProfile(ctx, userID, goals, style)
		}

		// 存储情景记忆
		if m.episodic != nil {
			_ = m.episodic.Store(ctx, userID, "", query, reply)
		}
	}()
}

// SaveMessage 保存单条消息
func (m *PersistentMemory) SaveMessage(ctx context.Context, userID int64, sessionID string, msg *schema.Message) error {
	return m.repo.SaveMessage(ctx, &entity.ChatMessage{
		SessionID: sessionID,
		UserID:    userID,
		Role:      string(msg.Role),
		Content:   msg.Content,
		Reasoning: msg.ReasoningContent,
	})
}

// compressHistory 调用 LLM 生成摘要
func (m *PersistentMemory) compressHistory(ctx context.Context, oldSummary string, history []*schema.Message) (string, error) {
	if m.summarizer == nil {
		return "", fmt.Errorf("summarizer not initialized")
	}

	var prompt string
	if oldSummary != "" {
		prompt = fmt.Sprintf("这是之前的对话摘要：%s\n\n请结合以下新增的对话历史，生成一个新的精简总结（保持在200字以内），保留关键的学习进度、知识点和用户偏好：\n\n", oldSummary)
	} else {
		prompt = "请将以下对话历史精简地总结为一段话，保留关键的学习进度、提到的知识点和用户偏好，不超过200字：\n\n"
	}

	for _, msg := range history {
		prompt += fmt.Sprintf("[%s]: %s\n", msg.Role, msg.Content)
	}

	resp, err := m.summarizer.Generate(ctx, []*schema.Message{
		schema.SystemMessage("你是一个记忆管理助手，负责维护对话的长期记忆。"),
		schema.UserMessage(prompt),
	})
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(resp.Content), nil
}

// GetSessions 获取会话列表
func (m *PersistentMemory) GetSessions(ctx context.Context, userID int64) ([]*entity.ChatSession, error) {
	return m.repo.GetSessions(ctx, userID)
}

// CreateSession 创建新会话
func (m *PersistentMemory) CreateSession(ctx context.Context, userID int64, sessionID string, title string) error {
	return m.repo.CreateSession(ctx, &entity.ChatSession{
		SessionID: sessionID,
		UserID:    userID,
		Title:     title,
	})
}

// UpdateSessionTitle 更新会话标题
func (m *PersistentMemory) UpdateSessionTitle(ctx context.Context, sessionID string, title string) error {
	return m.repo.UpdateSessionTitle(ctx, sessionID, title)
}
