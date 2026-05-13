package repository

import (
	"context"
	"time"

	"github.com/tomato/backend/internal/domain/entity"
	"gorm.io/gorm"
)

// StudyReportRepository 学习报告Repository接口
type StudyReportRepository interface {
	BaseRepository
	FindByUserIDAndDate(ctx context.Context, userID int64, date time.Time, reportType string) (*entity.StudyReport, error)
	FindByUserIDAndRange(ctx context.Context, userID int64, startDate, endDate time.Time, reportType string) ([]*entity.StudyReport, error)
	UpdateContent(ctx context.Context, reportID int64, content string) error
}

// studyReportRepository 学习报告Repository实现
type studyReportRepository struct {
	*baseRepository
}

// NewStudyReportRepository 创建新的学习报告Repository
func NewStudyReportRepository(db *gorm.DB) StudyReportRepository {
	return &studyReportRepository{
		baseRepository: &baseRepository{db: db},
	}
}

// FindByUserIDAndDate 根据用户ID和日期查找报告
func (r *studyReportRepository) FindByUserIDAndDate(ctx context.Context, userID int64, date time.Time, reportType string) (*entity.StudyReport, error) {
	var report entity.StudyReport
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND DATE(report_date) = DATE(?) AND report_type = ?", userID, date, reportType).
		First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

// FindByUserIDAndRange 根据用户ID和日期范围查找报告
func (r *studyReportRepository) FindByUserIDAndRange(ctx context.Context, userID int64, startDate, endDate time.Time, reportType string) ([]*entity.StudyReport, error) {
	var reports []*entity.StudyReport
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND report_date BETWEEN ? AND ? AND report_type = ?", userID, startDate, endDate, reportType).
		Find(&reports).Error
	return reports, err
}

// UpdateContent 更新报告内容
func (r *studyReportRepository) UpdateContent(ctx context.Context, reportID int64, content string) error {
	return r.db.WithContext(ctx).Model(&entity.StudyReport{}).
		Where("id = ?", reportID).
		Update("content", content).Error
}
