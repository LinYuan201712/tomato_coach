package entity

import (
	"time"
)

// UserFeishuConfig 用户飞书配置
type UserFeishuConfig struct {
	ID                uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID            uint64    `gorm:"index;unique;not null" json:"user_id"`
	Enabled           bool      `gorm:"default:false" json:"enabled"`
	AppID             string    `gorm:"size:255" json:"app_id"`
	AppSecret         string    `gorm:"size:255" json:"app_secret"`
	VerificationToken string    `gorm:"size:255" json:"verification_token"`
	EncryptKey        string    `gorm:"size:255" json:"encrypt_key"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (UserFeishuConfig) TableName() string {
	return "user_feishu_configs"
}
