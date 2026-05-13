package service

import (
	"context"

	"github.com/tomato/backend/internal/domain/entity"
	"github.com/tomato/backend/internal/domain/model"
	"github.com/tomato/backend/internal/pkg/errors"
	"github.com/tomato/backend/internal/pkg/logger"
	"github.com/tomato/backend/internal/repository"

	"go.uber.org/zap"
)

// BackgroundMusicService 背景音乐服务接口
type BackgroundMusicService interface {
	GetMusicList(ctx context.Context, page, pageSize int) ([]*model.BackgroundMusicResponse, int64, error)
	SearchMusic(ctx context.Context, keyword string) ([]*model.BackgroundMusicResponse, error)
	GetMusicByID(ctx context.Context, musicID int64) (*model.BackgroundMusicResponse, error)
}

// backgroundMusicService 背景音乐服务实现
type backgroundMusicService struct {
	musicRepo repository.BackgroundMusicRepository
	logger    *logger.Logger
}

// NewBackgroundMusicService 创建新的背景音乐服务
func NewBackgroundMusicService(
	musicRepo repository.BackgroundMusicRepository,
	logger *logger.Logger,
) BackgroundMusicService {
	return &backgroundMusicService{
		musicRepo: musicRepo,
		logger:    logger,
	}
}

// GetMusicList 获取音乐列表
func (s *backgroundMusicService) GetMusicList(ctx context.Context, page, pageSize int) ([]*model.BackgroundMusicResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	musics, total, err := s.musicRepo.List(ctx, page, pageSize)
	if err != nil {
		s.logger.Error("查询音乐列表失败", zap.Error(err))
		return nil, 0, errors.New(errors.CodeInternalError, "获取音乐列表失败")
	}

	var responses []*model.BackgroundMusicResponse
	for _, music := range musics {
		responses = append(responses, s.musicToResponse(music))
	}

	return responses, total, nil
}

// SearchMusic 搜索音乐
func (s *backgroundMusicService) SearchMusic(ctx context.Context, keyword string) ([]*model.BackgroundMusicResponse, error) {
	if keyword == "" {
		return []*model.BackgroundMusicResponse{}, nil
	}

	musics, err := s.musicRepo.Search(ctx, keyword)
	if err != nil {
		s.logger.Error("搜索音乐失败", zap.Error(err))
		return nil, errors.New(errors.CodeInternalError, "搜索音乐失败")
	}

	var responses []*model.BackgroundMusicResponse
	for _, music := range musics {
		responses = append(responses, s.musicToResponse(music))
	}

	return responses, nil
}

// GetMusicByID 根据ID获取音乐
func (s *backgroundMusicService) GetMusicByID(ctx context.Context, musicID int64) (*model.BackgroundMusicResponse, error) {
	music, err := s.musicRepo.FindByID(ctx, musicID)
	if err != nil {
		return nil, errors.New(errors.CodeInternalError, "音乐不存在")
	}

	return s.musicToResponse(music), nil
}

// ========== 私有方法 ==========

// musicToResponse 转换为响应格式
func (s *backgroundMusicService) musicToResponse(music *entity.BackgroundMusic) *model.BackgroundMusicResponse {
	return &model.BackgroundMusicResponse{
		ID:        music.ID,
		MusicName: music.MusicName,
		AudioURL:  music.AudioURL,
		Price:     music.Price,
		IsFree:    music.IsFree,
		Duration:  music.Duration,
		CreatedAt: music.CreatedAt,
	}
}

// SystemConfigService 系统配置服务接口
type SystemConfigService interface {
	GetConfigs(ctx context.Context) (map[string]interface{}, error)
	GetConfig(ctx context.Context, configKey string) (interface{}, error)
}

// systemConfigService 系统配置服务实现
type systemConfigService struct {
	configRepo repository.SystemConfigRepository
	logger     *logger.Logger
}

// NewSystemConfigService 创建新的系统配置服务
func NewSystemConfigService(
	configRepo repository.SystemConfigRepository,
	logger *logger.Logger,
) SystemConfigService {
	return &systemConfigService{
		configRepo: configRepo,
		logger:     logger,
	}
}

// GetConfigs 获取所有配置
func (s *systemConfigService) GetConfigs(ctx context.Context) (map[string]interface{}, error) {
	configs, err := s.configRepo.GetAll(ctx)
	if err != nil {
		s.logger.Error("查询系统配置失败", zap.Error(err))
		return nil, errors.New(errors.CodeInternalError, "获取配置失败")
	}

	result := make(map[string]interface{})
	for _, cfg := range configs {
		result[cfg.ConfigKey] = cfg.ConfigValue
	}

	return result, nil
}

// GetConfig 获取指定配置
func (s *systemConfigService) GetConfig(ctx context.Context, configKey string) (interface{}, error) {
	config, err := s.configRepo.GetByKey(ctx, configKey)
	if err != nil {
		return nil, errors.New(errors.CodeInternalError, "配置不存在")
	}

	return config.ConfigValue, nil
}
