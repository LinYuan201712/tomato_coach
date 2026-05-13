package repository

import (
	"context"
	"time"

	"github.com/tomato/backend/internal/domain/entity"
	"gorm.io/gorm"
)

// FriendRepository 好友Repository接口
type FriendRepository interface {
	FindFriend(ctx context.Context, userID, friendID int64) (*entity.Friend, error)
	FindFriends(ctx context.Context, userID int64) ([]*entity.Friend, error)
	AddFriend(ctx context.Context, friend *entity.Friend) error
	RemoveFriend(ctx context.Context, userID, friendID int64) error
	IsFriend(ctx context.Context, userID, friendID int64) bool
}

// friendRepository 好友Repository实现
type friendRepository struct {
	*baseRepository
}

// NewFriendRepository 创建新的好友Repository
func NewFriendRepository(db *gorm.DB) FriendRepository {
	return &friendRepository{
		baseRepository: &baseRepository{db: db},
	}
}

// FindFriend 查找好友关系
func (r *friendRepository) FindFriend(ctx context.Context, userID, friendID int64) (*entity.Friend, error) {
	var friend *entity.Friend
	err := r.db.WithContext(ctx).Where("user_id = ? AND friend_id = ?", userID, friendID).First(&friend).Error
	return friend, err
}

// FindFriends 查找好友列表
func (r *friendRepository) FindFriends(ctx context.Context, userID int64) ([]*entity.Friend, error) {
	var friends []*entity.Friend
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&friends).Error
	return friends, err
}

// AddFriend 添加好友
func (r *friendRepository) AddFriend(ctx context.Context, friend *entity.Friend) error {
	return r.db.WithContext(ctx).Create(friend).Error
}

// RemoveFriend 删除好友
func (r *friendRepository) RemoveFriend(ctx context.Context, userID, friendID int64) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND friend_id = ?", userID, friendID).Delete(&entity.Friend{}).Error
}

// IsFriend 判断是否是好友
func (r *friendRepository) IsFriend(ctx context.Context, userID, friendID int64) bool {
	friend, _ := r.FindFriend(ctx, userID, friendID)
	return friend != nil
}

// FriendRequestRepository 好友请求Repository接口
type FriendRequestRepository interface {
	Create(ctx context.Context, req *entity.FriendRequest) error
	FindByID(ctx context.Context, id int64) (*entity.FriendRequest, error)
	FindPending(ctx context.Context, toUserID int64) ([]*entity.FriendRequest, error)
	FindByToUserID(ctx context.Context, toUserID int64) ([]*entity.FriendRequest, error)
	FindByFromAndTo(ctx context.Context, fromUserID, toUserID int64) (*entity.FriendRequest, error)
	Update(ctx context.Context, req *entity.FriendRequest) error
	Delete(ctx context.Context, id int64) error
}

// friendRequestRepository 好友请求Repository实现
type friendRequestRepository struct {
	*baseRepository
}

// NewFriendRequestRepository 创建新的好友请求Repository
func NewFriendRequestRepository(db *gorm.DB) FriendRequestRepository {
	return &friendRequestRepository{
		baseRepository: &baseRepository{db: db},
	}
}

// Create 创建
func (r *friendRequestRepository) Create(ctx context.Context, req *entity.FriendRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

// FindByID 根据ID查找
func (r *friendRequestRepository) FindByID(ctx context.Context, id int64) (*entity.FriendRequest, error) {
	var req *entity.FriendRequest
	err := r.db.WithContext(ctx).First(&req, id).Error
	return req, err
}

// FindPending 查找待处理的申请
func (r *friendRequestRepository) FindPending(ctx context.Context, toUserID int64) ([]*entity.FriendRequest, error) {
	var reqs []*entity.FriendRequest
	err := r.db.WithContext(ctx).Where("to_user_id = ? AND status = ?", toUserID, "待处理").Find(&reqs).Error
	return reqs, err
}

// FindByToUserID 查找所有接收到的申请
func (r *friendRequestRepository) FindByToUserID(ctx context.Context, toUserID int64) ([]*entity.FriendRequest, error) {
	var reqs []*entity.FriendRequest
	err := r.db.WithContext(ctx).Where("to_user_id = ?", toUserID).Find(&reqs).Error
	return reqs, err
}

// FindByFromAndTo 查找特定的申请
func (r *friendRequestRepository) FindByFromAndTo(ctx context.Context, fromUserID, toUserID int64) (*entity.FriendRequest, error) {
	var req *entity.FriendRequest
	err := r.db.WithContext(ctx).Where("from_user_id = ? AND to_user_id = ?", fromUserID, toUserID).First(&req).Error
	return req, err
}

// Update 更新
func (r *friendRequestRepository) Update(ctx context.Context, req *entity.FriendRequest) error {
	return r.db.WithContext(ctx).Save(req).Error
}

// Delete 删除
func (r *friendRequestRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&entity.FriendRequest{}, id).Error
}

// FocusSessionRepository 专注会话Repository接口
type FocusSessionRepository interface {
	Create(ctx context.Context, session *entity.FocusSession) error
	FindByID(ctx context.Context, sessionID int64) (*entity.FocusSession, error)
	FindByUserID(ctx context.Context, userID int64) ([]*entity.FocusSession, error)
	FindUnfinished(ctx context.Context, userID int64) (*entity.FocusSession, error)
	Update(ctx context.Context, session *entity.FocusSession) error
	GetTotalDuration(ctx context.Context, userID int64) (int, error)
	ListByDateRange(ctx context.Context, userID int64, startDate, endDate time.Time) ([]*entity.FocusSession, error)
}

// focusSessionRepository 专注会话Repository实现
type focusSessionRepository struct {
	*baseRepository
}

// NewFocusSessionRepository 创建新的专注会话Repository
func NewFocusSessionRepository(db *gorm.DB) FocusSessionRepository {
	return &focusSessionRepository{
		baseRepository: &baseRepository{db: db},
	}
}

// Create 创建
func (r *focusSessionRepository) Create(ctx context.Context, session *entity.FocusSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// FindByID 根据ID查找
func (r *focusSessionRepository) FindByID(ctx context.Context, sessionID int64) (*entity.FocusSession, error) {
	var session *entity.FocusSession
	err := r.db.WithContext(ctx).First(&session, sessionID).Error
	return session, err
}

// FindByUserID 查找用户的会话
func (r *focusSessionRepository) FindByUserID(ctx context.Context, userID int64) ([]*entity.FocusSession, error) {
	var sessions []*entity.FocusSession
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&sessions).Error
	return sessions, err
}

// FindUnfinished 查找未完成的会话
func (r *focusSessionRepository) FindUnfinished(ctx context.Context, userID int64) (*entity.FocusSession, error) {
	var session *entity.FocusSession
	err := r.db.WithContext(ctx).Where("user_id = ? AND status = ?", userID, "进行中").First(&session).Error
	return session, err
}

// Update 更新
func (r *focusSessionRepository) Update(ctx context.Context, session *entity.FocusSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

// GetTotalDuration 获取总时长
func (r *focusSessionRepository) GetTotalDuration(ctx context.Context, userID int64) (int, error) {
	var total int
	err := r.db.WithContext(ctx).Model(&entity.FocusSession{}).
		Where("user_id = ? AND status = ?", userID, "已完成").
		Select("COALESCE(SUM(duration), 0)").Scan(&total).Error
	return total, err
}

// ListByDateRange 获取时间范围内的会话
func (r *focusSessionRepository) ListByDateRange(ctx context.Context, userID int64, startDate, endDate time.Time) ([]*entity.FocusSession, error) {
	var sessions []*entity.FocusSession
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND start_time BETWEEN ? AND ?", userID, startDate, endDate).
		Find(&sessions).Error
	return sessions, err
}

// UserPrivacyRepository 用户隐私Repository接口
type UserPrivacyRepository interface {
	FindByUserID(ctx context.Context, userID int64) (*entity.UserPrivacy, error)
	Create(ctx context.Context, privacy *entity.UserPrivacy) error
	Update(ctx context.Context, privacy *entity.UserPrivacy) error
}

// userPrivacyRepository 用户隐私Repository实现
type userPrivacyRepository struct {
	*baseRepository
}

// NewUserPrivacyRepository 创建新的用户隐私Repository
func NewUserPrivacyRepository(db *gorm.DB) UserPrivacyRepository {
	return &userPrivacyRepository{
		baseRepository: &baseRepository{db: db},
	}
}

// FindByUserID 根据用户ID查找
func (r *userPrivacyRepository) FindByUserID(ctx context.Context, userID int64) (*entity.UserPrivacy, error) {
	var privacy *entity.UserPrivacy
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&privacy).Error
	return privacy, err
}

// Create 创建
func (r *userPrivacyRepository) Create(ctx context.Context, privacy *entity.UserPrivacy) error {
	return r.db.WithContext(ctx).Create(privacy).Error
}

// Update 更新
func (r *userPrivacyRepository) Update(ctx context.Context, privacy *entity.UserPrivacy) error {
	return r.db.WithContext(ctx).Save(privacy).Error
}

// UserCurrencyRepository 用户货币Repository接口
type UserCurrencyRepository interface {
	FindByUserID(ctx context.Context, userID int64) (*entity.UserCurrency, error)
	Create(ctx context.Context, currency *entity.UserCurrency) error
	Update(ctx context.Context, currency *entity.UserCurrency) error
	IncrementCoins(ctx context.Context, userID int64, amount int) error
}

// userCurrencyRepository 用户货币Repository实现
type userCurrencyRepository struct {
	*baseRepository
}

// NewUserCurrencyRepository 创建新的用户货币Repository
func NewUserCurrencyRepository(db *gorm.DB) UserCurrencyRepository {
	return &userCurrencyRepository{
		baseRepository: &baseRepository{db: db},
	}
}

// FindByUserID 根据用户ID查找
func (r *userCurrencyRepository) FindByUserID(ctx context.Context, userID int64) (*entity.UserCurrency, error) {
	var currency *entity.UserCurrency
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&currency).Error
	return currency, err
}

// Create 创建
func (r *userCurrencyRepository) Create(ctx context.Context, currency *entity.UserCurrency) error {
	return r.db.WithContext(ctx).Create(currency).Error
}

// Update 更新
func (r *userCurrencyRepository) Update(ctx context.Context, currency *entity.UserCurrency) error {
	return r.db.WithContext(ctx).Save(currency).Error
}

// IncrementCoins 增加硬币
func (r *userCurrencyRepository) IncrementCoins(ctx context.Context, userID int64, amount int) error {
	return r.db.WithContext(ctx).Model(&entity.UserCurrency{}).
		Where("user_id = ?", userID).
		Update("coins", gorm.Expr("coins + ?", amount)).Error
}

// CheckinRepository 签到Repository接口
type CheckinRepository interface {
	Create(ctx context.Context, record *entity.CheckinRecord) error
	CountByUserAndDate(ctx context.Context, userID int64, date time.Time) (int64, error)
	CountByUserAndDateRange(ctx context.Context, userID int64, startDate, endDate time.Time) (int, error)
	ListByUserAndDateRange(ctx context.Context, userID int64, startDate, endDate time.Time) ([]*entity.CheckinRecord, error)
}

// checkinRepository 签到Repository实现
type checkinRepository struct {
	*baseRepository
}

// NewCheckinRepository 创建新的签到Repository
func NewCheckinRepository(db *gorm.DB) CheckinRepository {
	return &checkinRepository{
		baseRepository: &baseRepository{db: db},
	}
}

// Create 创建
func (r *checkinRepository) Create(ctx context.Context, record *entity.CheckinRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// CountByUserAndDate 统计特定日期的签到记录
func (r *checkinRepository) CountByUserAndDate(ctx context.Context, userID int64, date time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.CheckinRecord{}).
		Where("user_id = ? AND DATE(checkin_date) = DATE(?)", userID, date).
		Count(&count).Error
	return count, err
}

// CountByUserAndDateRange 统计日期范围内的签到记录
func (r *checkinRepository) CountByUserAndDateRange(ctx context.Context, userID int64, startDate, endDate time.Time) (int, error) {
	var count64 int64
	err := r.db.WithContext(ctx).Model(&entity.CheckinRecord{}).
		Where("user_id = ? AND checkin_date BETWEEN ? AND ?", userID, startDate, endDate).
		Count(&count64).Error
	return int(count64), err
}

// ListByUserAndDateRange 获取日期范围内的签到记录
func (r *checkinRepository) ListByUserAndDateRange(ctx context.Context, userID int64, startDate, endDate time.Time) ([]*entity.CheckinRecord, error) {
	var records []*entity.CheckinRecord
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND checkin_date BETWEEN ? AND ?", userID, startDate, endDate).
		Find(&records).Error
	return records, err
}
