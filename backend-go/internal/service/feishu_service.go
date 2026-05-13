package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/tomato/backend/config"
	"github.com/tomato/backend/internal/channels"
	"github.com/tomato/backend/internal/domain/entity"
	"github.com/tomato/backend/internal/pkg/bus"
	"github.com/tomato/backend/internal/pkg/logger"
	"github.com/tomato/backend/internal/repository"
	"go.uber.org/zap"
)

type FeishuService interface {
	GetConfig(ctx context.Context, userID uint64) (*entity.UserFeishuConfig, error)
	SaveConfig(ctx context.Context, userID uint64, cfg *entity.UserFeishuConfig) error
	StartUserConnection(ctx context.Context, userID uint64) error
	StopUserConnection(ctx context.Context, userID uint64) error
	InitAllConnections(ctx context.Context) error
}

type feishuService struct {
	repo       repository.UserFeishuConfigRepository
	channelMgr *channels.Manager
	messageBus *bus.MessageBus
	logger     *logger.Logger
}

func NewFeishuService(
	repo repository.UserFeishuConfigRepository,
	channelMgr *channels.Manager,
	messageBus *bus.MessageBus,
	logger *logger.Logger,
) FeishuService {
	return &feishuService{
		repo:       repo,
		channelMgr: channelMgr,
		messageBus: messageBus,
		logger:     logger,
	}
}

func (s *feishuService) GetConfig(ctx context.Context, userID uint64) (*entity.UserFeishuConfig, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *feishuService) SaveConfig(ctx context.Context, userID uint64, cfg *entity.UserFeishuConfig) error {
	cfg.UserID = userID
	
	// 先保存到数据库
	err := s.repo.Upsert(ctx, cfg)
	if err != nil {
		return err
	}

	// 根据 Enabled 状态管理连接
	if cfg.Enabled {
		return s.StartUserConnection(ctx, userID)
	} else {
		return s.StopUserConnection(ctx, userID)
	}
}

func (s *feishuService) StartUserConnection(ctx context.Context, userID uint64) error {
	cfg, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if cfg == nil || !cfg.Enabled {
		return nil
	}

	// 转换为 FeishuChannelConfig
	feishuCfg := config.FeishuChannelConfig{
		Enabled:           cfg.Enabled,
		AppID:             cfg.AppID,
		AppSecret:         cfg.AppSecret,
		VerificationToken: cfg.VerificationToken,
		EncryptKey:        cfg.EncryptKey,
		Domain:            "feishu", // 默认
	}

	accountID := strconv.FormatUint(userID, 10)
	
	// 先停止旧的（如果存在）
	s.channelMgr.Unregister("feishu", accountID)

	// 创建新通道
	feishuChan, err := channels.NewFeishuChannel(accountID, feishuCfg, s.messageBus, s.logger)
	if err != nil {
		return fmt.Errorf("failed to create feishu channel: %w", err)
	}
	
	// 注册并启动
	if err := s.channelMgr.Register(feishuChan); err != nil {
		return err
	}

	return feishuChan.Start(context.Background())
}

func (s *feishuService) StopUserConnection(ctx context.Context, userID uint64) error {
	accountID := strconv.FormatUint(userID, 10)
	return s.channelMgr.Unregister("feishu", accountID)
}

func (s *feishuService) InitAllConnections(ctx context.Context) error {
	configs, err := s.repo.GetActiveConfigs(ctx)
	if err != nil {
		return err
	}

	for _, cfg := range configs {
		s.logger.Info("Initializing Feishu connection for user", zap.Uint64("user_id", cfg.UserID))
		if err := s.StartUserConnection(ctx, cfg.UserID); err != nil {
			s.logger.Error("Failed to start Feishu connection", 
				zap.Uint64("user_id", cfg.UserID), 
				zap.Error(err))
		}
	}
	return nil
}
