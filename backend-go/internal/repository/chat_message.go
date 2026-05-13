package repository

import (
	"context"
	"time"

	"github.com/tomato/backend/internal/domain/entity"
	"gorm.io/gorm"
)

type ChatRepository interface {
	SaveMessage(ctx context.Context, msg *entity.ChatMessage) error
	GetHistory(ctx context.Context, sessionID string, limit int) ([]*entity.ChatMessage, error)
	CreateSession(ctx context.Context, session *entity.ChatSession) error
	GetSessions(ctx context.Context, userID int64) ([]*entity.ChatSession, error)
	UpdateSessionTitle(ctx context.Context, sessionID string, title string) error
	DeleteSession(ctx context.Context, sessionID string) error
	GetSession(ctx context.Context, sessionID string) (*entity.ChatSession, error)
	UpdateSessionSummary(ctx context.Context, sessionID string, summary string, msgCount int) error
	GetMessagesByDateRange(ctx context.Context, userID int64, startDate, endDate time.Time) ([]*entity.ChatMessage, error)
}

type chatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) SaveMessage(ctx context.Context, msg *entity.ChatMessage) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 保存消息
		if err := tx.Create(msg).Error; err != nil {
			return err
		}
		// 2. 更新会话的 updated_at 时间，使其在列表中置顶
		return tx.Model(&entity.ChatSession{}).
			Where("session_id = ?", msg.SessionID).
			Update("updated_at", gorm.Expr("NOW()")).Error
	})
}

func (r *chatRepository) GetHistory(ctx context.Context, sessionID string, limit int) ([]*entity.ChatMessage, error) {
	var msgs []*entity.ChatMessage
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at desc").
		Limit(limit).
		Find(&msgs).Error
	if err != nil {
		return nil, err
	}

	// 逆序排列，使其按时间顺序排列
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}

	return msgs, nil
}

func (r *chatRepository) CreateSession(ctx context.Context, session *entity.ChatSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *chatRepository) GetSessions(ctx context.Context, userID int64) ([]*entity.ChatSession, error) {
	var sessions []*entity.ChatSession
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("updated_at desc").
		Find(&sessions).Error
	return sessions, err
}

func (r *chatRepository) UpdateSessionTitle(ctx context.Context, sessionID string, title string) error {
	return r.db.WithContext(ctx).
		Model(&entity.ChatSession{}).
		Where("session_id = ?", sessionID).
		Update("title", title).Error
}

func (r *chatRepository) DeleteSession(ctx context.Context, sessionID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("session_id = ?", sessionID).Delete(&entity.ChatMessage{}).Error; err != nil {
			return err
		}
		return tx.Where("session_id = ?", sessionID).Delete(&entity.ChatSession{}).Error
	})
}

func (r *chatRepository) GetSession(ctx context.Context, sessionID string) (*entity.ChatSession, error) {
	var session entity.ChatSession
	err := r.db.WithContext(ctx).Where("session_id = ?", sessionID).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *chatRepository) UpdateSessionSummary(ctx context.Context, sessionID string, summary string, msgCount int) error {
	updates := map[string]interface{}{
		"summary":   summary,
		"msg_count": msgCount,
	}
	return r.db.WithContext(ctx).Model(&entity.ChatSession{}).
		Where("session_id = ?", sessionID).
		Updates(updates).Error
}

func (r *chatRepository) GetMessagesByDateRange(ctx context.Context, userID int64, startDate, endDate time.Time) ([]*entity.ChatMessage, error) {
	var msgs []*entity.ChatMessage
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND created_at BETWEEN ?", userID, startDate, endDate).
		Find(&msgs).Error
	return msgs, err
}
