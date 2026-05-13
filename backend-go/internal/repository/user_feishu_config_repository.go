package repository

import (
	"context"

	"github.com/tomato/backend/internal/domain/entity"
	"gorm.io/gorm"
)

type UserFeishuConfigRepository interface {
	GetByUserID(ctx context.Context, userID uint64) (*entity.UserFeishuConfig, error)
	Upsert(ctx context.Context, config *entity.UserFeishuConfig) error
	GetActiveConfigs(ctx context.Context) ([]*entity.UserFeishuConfig, error)
}

type userFeishuConfigRepository struct {
	db *gorm.DB
}

func NewUserFeishuConfigRepository(db *gorm.DB) UserFeishuConfigRepository {
	return &userFeishuConfigRepository{db: db}
}

func (r *userFeishuConfigRepository) GetByUserID(ctx context.Context, userID uint64) (*entity.UserFeishuConfig, error) {
	var config entity.UserFeishuConfig
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

func (r *userFeishuConfigRepository) Upsert(ctx context.Context, config *entity.UserFeishuConfig) error {
	var existing entity.UserFeishuConfig
	err := r.db.WithContext(ctx).Where("user_id = ?", config.UserID).First(&existing).Error
	if err == nil {
		// 记录已存在，执行更新
		config.ID = existing.ID
		return r.db.WithContext(ctx).Save(config).Error
	} else if err == gorm.ErrRecordNotFound {
		// 记录不存在，执行插入
		return r.db.WithContext(ctx).Create(config).Error
	}
	return err
}

func (r *userFeishuConfigRepository) GetActiveConfigs(ctx context.Context) ([]*entity.UserFeishuConfig, error) {
	var configs []*entity.UserFeishuConfig
	err := r.db.WithContext(ctx).Where("enabled = ?", true).Find(&configs).Error
	return configs, err
}
