package repository

import (
	"context"
	"time"
	"errors"

	"github.com/tomato/backend/internal/domain/entity"
	"gorm.io/gorm"
)

// BaseRepository 基础Repository接口
type BaseRepository interface {
	Create(ctx context.Context, value interface{}) error
	GetByID(ctx context.Context, id int64, dest interface{}) error
	Update(ctx context.Context, value interface{}) error
	Delete(ctx context.Context, id int64, dest interface{}) error
	List(ctx context.Context, dest interface{}, conds ...interface{}) error
	Count(ctx context.Context, count *int64, conds ...interface{}) error
	Transaction(ctx context.Context, fn func(*gorm.DB) error) error
}

// baseRepository 基础Repository实现
type baseRepository struct {
	db *gorm.DB
}

// NewBaseRepository 创建新的基础Repository
func NewBaseRepository(db *gorm.DB) BaseRepository {
	return &baseRepository{db: db}
}

// Create 创建记录
func (r *baseRepository) Create(ctx context.Context, value interface{}) error {
	return r.db.WithContext(ctx).Create(value).Error
}

// GetByID 根据ID获取记录
func (r *baseRepository) GetByID(ctx context.Context, id int64, dest interface{}) error {
	return r.db.WithContext(ctx).First(dest, id).Error
}

// Update 更新记录
func (r *baseRepository) Update(ctx context.Context, value interface{}) error {
	return r.db.WithContext(ctx).Save(value).Error
}

// Delete 删除记录
func (r *baseRepository) Delete(ctx context.Context, id int64, dest interface{}) error {
	return r.db.WithContext(ctx).Delete(dest, id).Error
}

// List 列出记录
func (r *baseRepository) List(ctx context.Context, dest interface{}, conds ...interface{}) error {
	query := r.db.WithContext(ctx)
	for _, cond := range conds {
		query = query.Where(cond)
	}
	return query.Find(dest).Error
}

// Count 计数
func (r *baseRepository) Count(ctx context.Context, count *int64, conds ...interface{}) error {
	query := r.db.WithContext(ctx)
	for _, cond := range conds {
		query = query.Where(cond)
	}
	return query.Model(&entity.User{}).Count(count).Error
}

// Transaction 事务处理
func (r *baseRepository) Transaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

// UserRepository 用户Repository接口
type UserRepository interface {
	BaseRepository
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByPhone(ctx context.Context, phone string) (*entity.User, error)
	FindByUserID(ctx context.Context, userID int64) (*entity.User, error)
	UpdateStatus(ctx context.Context, userID int64, status string) error
	UpdateTomato(ctx context.Context, userID int64, tomato int) error
	UpdateProfile(ctx context.Context, userID int64, goals string, style string) error
	UpdateProfileLock(ctx context.Context, userID int64, lock string) error
	UpdateLockSuggestions(ctx context.Context, userID int64, suggestions string) error
}

// userRepository 用户Repository实现
type userRepository struct {
	*baseRepository
}

// NewUserRepository 创建新的用户Repository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		baseRepository: &baseRepository{db: db},
	}
}

// FindByUsername 根据用户名查找
func (r *userRepository) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmail 根据邮箱查找
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByPhone 根据电话查找
func (r *userRepository) FindByPhone(ctx context.Context, phone string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByUserID 根据业务ID查找
func (r *userRepository) FindByUserID(ctx context.Context, userID int64) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// UpdateStatus 更新用户状态
func (r *userRepository) UpdateStatus(ctx context.Context, userID int64, status string) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("user_id = ?", userID).Update("status", status).Error
}

// UpdateTomato 更新番茄数
func (r *userRepository) UpdateTomato(ctx context.Context, userID int64, tomato int) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("user_id = ?", userID).Update("tomato", tomato).Error
}

// UpdateProfile 更新用户画像
func (r *userRepository) UpdateProfile(ctx context.Context, userID int64, goals string, style string) error {
	updates := make(map[string]interface{})
	if goals != "" {
		updates["goals"] = goals
	}
	if style != "" {
		updates["preferred_style"] = style
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("user_id = ?", userID).Updates(updates).Error
}

// UpdateProfileLock 更新用户画像锁定状态
func (r *userRepository) UpdateProfileLock(ctx context.Context, userID int64, lock string) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("user_id = ?", userID).Update("profile_lock", lock).Error
}

// UpdateLockSuggestions 更新用户画像建议
func (r *userRepository) UpdateLockSuggestions(ctx context.Context, userID int64, suggestions string) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("user_id = ?", userID).Update("lock_suggestions", suggestions).Error
}

// TaskRepository 任务Repository接口
type TaskRepository interface {
	BaseRepository
	FindByTaskID(ctx context.Context, taskID int64) (*entity.Task, error)
	FindByUserID(ctx context.Context, userID int64) ([]*entity.Task, error)
	FindByStatus(ctx context.Context, userID int64, status string) ([]*entity.Task, error)
	UpdateStatus(ctx context.Context, taskID int64, status string) error
	IncrementActualDuration(ctx context.Context, taskID int64, duration int) error
	ListByDateRange(ctx context.Context, userID int64, startDate, endDate time.Time) ([]*entity.Task, error)
}

// taskRepository 任务Repository实现
type taskRepository struct {
	*baseRepository
}

// NewTaskRepository 创建新的任务Repository
func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{
		baseRepository: &baseRepository{db: db},
	}
}

// FindByTaskID 根据任务ID查找
func (r *taskRepository) FindByTaskID(ctx context.Context, taskID int64) (*entity.Task, error) {
	var task *entity.Task
	err := r.db.WithContext(ctx).Where("task_id = ?", taskID).First(&task).Error
	return task, err
}

// FindByUserID 根据用户ID查找
func (r *taskRepository) FindByUserID(ctx context.Context, userID int64) ([]*entity.Task, error) {
	var tasks []*entity.Task
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&tasks).Error
	return tasks, err
}

// FindByStatus 根据状态查找
func (r *taskRepository) FindByStatus(ctx context.Context, userID int64, status string) ([]*entity.Task, error) {
	var tasks []*entity.Task
	err := r.db.WithContext(ctx).Where("user_id = ? AND status = ?", userID, status).Find(&tasks).Error
	return tasks, err
}

// UpdateStatus 更新任务状态
func (r *taskRepository) UpdateStatus(ctx context.Context, taskID int64, status string) error {
	return r.db.WithContext(ctx).Model(&entity.Task{}).Where("task_id = ?", taskID).Update("status", status).Error
}

// IncrementActualDuration 增加实际时长
func (r *taskRepository) IncrementActualDuration(ctx context.Context, taskID int64, duration int) error {
	return r.db.WithContext(ctx).Model(&entity.Task{}).Where("task_id = ?", taskID).Update("actual_duration", gorm.Expr("actual_duration + ?", duration)).Error
}

// ListByDateRange 获取时间范围内的任务
func (r *taskRepository) ListByDateRange(ctx context.Context, userID int64, startDate, endDate time.Time) ([]*entity.Task, error) {
	var tasks []*entity.Task
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND (created_at BETWEEN ? AND ? OR updated_at BETWEEN ? AND ?)", userID, startDate, endDate, startDate, endDate).
		Find(&tasks).Error
	return tasks, err
}

// RoomRepository 房间Repository接口
type RoomRepository interface {
	BaseRepository
	FindByRoomID(ctx context.Context, roomID int64) (*entity.Room, error)
	FindByCreator(ctx context.Context, creatorID int64) ([]*entity.Room, error)
	GetMembers(ctx context.Context, roomID int64) ([]*entity.RoomMember, error)
	AddMember(ctx context.Context, member *entity.RoomMember) error
	RemoveMember(ctx context.Context, roomID, userID int64) error
	UpdateCreator(ctx context.Context, roomID int64, newCreatorID int64) error
	CountMembers(ctx context.Context, roomID int64) (int64, error)
	ListRooms(ctx context.Context, page, pageSize int) ([]*entity.Room, int64, error)
	UpdateMember(ctx context.Context, member *entity.RoomMember) error
	FindUserRoom(ctx context.Context, userID int64) (*entity.RoomMember, error)
	DeleteMembersByRoomID(ctx context.Context, roomID int64) error
}

// roomRepository 房间Repository实现
type roomRepository struct {
	*baseRepository
}

// NewRoomRepository 创建新的房间Repository
func NewRoomRepository(db *gorm.DB) RoomRepository {
	return &roomRepository{
		baseRepository: &baseRepository{db: db},
	}
}

// FindByRoomID 根据房间ID查找
func (r *roomRepository) FindByRoomID(ctx context.Context, roomID int64) (*entity.Room, error) {
	var room *entity.Room
	err := r.db.WithContext(ctx).Where("room_id = ?", roomID).First(&room).Error
	return room, err
}

// FindByCreator 根据创建者查找
func (r *roomRepository) FindByCreator(ctx context.Context, creatorID int64) ([]*entity.Room, error) {
	var rooms []*entity.Room
	err := r.db.WithContext(ctx).Where("create_person = ?", creatorID).Find(&rooms).Error
	return rooms, err
}

// GetMembers 获取房间成员
func (r *roomRepository) GetMembers(ctx context.Context, roomID int64) ([]*entity.RoomMember, error) {
	var members []*entity.RoomMember
	err := r.db.WithContext(ctx).Where("room_id = ?", roomID).Find(&members).Error
	return members, err
}

// AddMember 添加成员
func (r *roomRepository) AddMember(ctx context.Context, member *entity.RoomMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// RemoveMember 删除成员
func (r *roomRepository) RemoveMember(ctx context.Context, roomID, userID int64) error {
	return r.db.WithContext(ctx).Where("room_id = ? AND user_id = ?", roomID, userID).Delete(&entity.RoomMember{}).Error
}

// UpdateCreator 更新房主
func (r *roomRepository) UpdateCreator(ctx context.Context, roomID int64, newCreatorID int64) error {
	return r.db.WithContext(ctx).Model(&entity.Room{}).Where("room_id = ?", roomID).Update("create_person", newCreatorID).Error
}

// CountMembers 计数成员
func (r *roomRepository) CountMembers(ctx context.Context, roomID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.RoomMember{}).Where("room_id = ?", roomID).Count(&count).Error
	return count, err
}

// ListRooms 分页获取房间列表
func (r *roomRepository) ListRooms(ctx context.Context, page, pageSize int) ([]*entity.Room, int64, error) {
	var rooms []*entity.Room
	var total int64

	offset := (page - 1) * pageSize

	err := r.db.WithContext(ctx).Model(&entity.Room{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&rooms).Error
	return rooms, total, err
}

// UpdateMember 更新成员信息
func (r *roomRepository) UpdateMember(ctx context.Context, member *entity.RoomMember) error {
	return r.db.WithContext(ctx).Save(member).Error
}

// FindUserRoom 查找用户所在的房间
func (r *roomRepository) FindUserRoom(ctx context.Context, userID int64) (*entity.RoomMember, error) {
	var member *entity.RoomMember
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Last(&member).Error
	return member, err
}

// DeleteMembersByRoomID 删除房间的所有成员
func (r *roomRepository) DeleteMembersByRoomID(ctx context.Context, roomID int64) error {
	return r.db.WithContext(ctx).Where("room_id = ?", roomID).Delete(&entity.RoomMember{}).Error
}
