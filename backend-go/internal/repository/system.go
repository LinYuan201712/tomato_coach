package repository

import (
	"context"

	"github.com/tomato/backend/internal/domain/entity"
	"gorm.io/gorm"
)

// BackgroundMusicRepository 背景音乐Repository接口
type BackgroundMusicRepository interface {
	FindByID(ctx context.Context, musicID int64) (*entity.BackgroundMusic, error)
	List(ctx context.Context, page, pageSize int) ([]*entity.BackgroundMusic, int64, error)
	Search(ctx context.Context, keyword string) ([]*entity.BackgroundMusic, error)
}

// backgroundMusicRepository 背景音乐Repository实现
type backgroundMusicRepository struct {
	*baseRepository
}

// NewBackgroundMusicRepository 创建新的背景音乐Repository
func NewBackgroundMusicRepository(db *gorm.DB) BackgroundMusicRepository {
	return &backgroundMusicRepository{
		baseRepository: &baseRepository{db: db},
	}
}

// FindByID 根据ID查找
func (r *backgroundMusicRepository) FindByID(ctx context.Context, musicID int64) (*entity.BackgroundMusic, error) {
	var music *entity.BackgroundMusic
	err := r.db.WithContext(ctx).First(&music, musicID).Error
	return music, err
}

// List 分页获取列表
func (r *backgroundMusicRepository) List(ctx context.Context, page, pageSize int) ([]*entity.BackgroundMusic, int64, error) {
	var musics []*entity.BackgroundMusic
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.BackgroundMusic{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&musics).Error; err != nil {
		return nil, 0, err
	}

	return musics, total, nil
}

// Search 根据关键字搜索
func (r *backgroundMusicRepository) Search(ctx context.Context, keyword string) ([]*entity.BackgroundMusic, error) {
	var musics []*entity.BackgroundMusic
	err := r.db.WithContext(ctx).Where("music_name LIKE ?", "%"+keyword+"%").Find(&musics).Error
	return musics, err
}

// SystemConfigRepository 系统配置Repository接口
type SystemConfigRepository interface {
	GetByKey(ctx context.Context, key string) (*entity.SystemConfig, error)
	GetAll(ctx context.Context) ([]*entity.SystemConfig, error)
}

// systemConfigRepository 系统配置Repository实现
type systemConfigRepository struct {
	*baseRepository
}

// NewSystemConfigRepository 创建新的系统配置Repository
func NewSystemConfigRepository(db *gorm.DB) SystemConfigRepository {
	return &systemConfigRepository{
		baseRepository: &baseRepository{db: db},
	}
}

// GetByKey 根据Key查找
func (r *systemConfigRepository) GetByKey(ctx context.Context, key string) (*entity.SystemConfig, error) {
	var config *entity.SystemConfig
	err := r.db.WithContext(ctx).Where("config_key = ?", key).First(&config).Error
	return config, err
}

// GetAll 获取所有
func (r *systemConfigRepository) GetAll(ctx context.Context) ([]*entity.SystemConfig, error) {
	var configs []*entity.SystemConfig
	err := r.db.WithContext(ctx).Find(&configs).Error
	return configs, err
}
