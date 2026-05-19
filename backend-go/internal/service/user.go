package service

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/tomato/backend/internal/domain/constants"
	"github.com/tomato/backend/internal/domain/entity"
	"github.com/tomato/backend/internal/domain/model"
	"github.com/tomato/backend/internal/pkg/auth"
	"github.com/tomato/backend/internal/pkg/errors"
	"github.com/tomato/backend/internal/pkg/logger"
	"github.com/tomato/backend/internal/repository"

	"go.uber.org/zap"
)

// UserService 用户服务接口
type UserService interface {
	GetUserInfo(ctx context.Context, userID int64) (*model.UserInfoResponse, error)
	GetUserByID(ctx context.Context, userID int64) (*model.UserResponse, error)
	GetPublicUserInfo(ctx context.Context, username string, viewerID int64) (*model.PublicUserResponse, error)
	UpdateUserInfo(ctx context.Context, userID int64, req *model.UpdateUserRequest) (*model.UserInfoResponse, error)
	GetUserPrivacy(ctx context.Context, userID int64) (*model.UserPrivacyResponse, error)
	UpdateUserPrivacy(ctx context.Context, userID int64, req *model.UpdateUserPrivacyRequest) (*model.UserPrivacyResponse, error)
	GetUserCurrency(ctx context.Context, userID int64) (*model.CurrencyResponse, error)
	Checkin(ctx context.Context, userID int64) error
	GetCheckinCount(ctx context.Context, userID int64) (int, error)
	GetCheckinDates(ctx context.Context, userID int64) ([]string, error)
}

// userService 用户服务实现
type userService struct {
	userRepo        repository.UserRepository
	privacyRepo     repository.UserPrivacyRepository
	currencyRepo    repository.UserCurrencyRepository
	friendRepo      repository.FriendRepository
	checkinRepo     repository.CheckinRepository
	roomRepo        repository.RoomRepository
	passwordManager *auth.PasswordManager
	idGenerator     *snowflake.Node
	logger          *logger.Logger
}

// NewUserService 创建新的用户服务
func NewUserService(
	userRepo repository.UserRepository,
	privacyRepo repository.UserPrivacyRepository,
	currencyRepo repository.UserCurrencyRepository,
	friendRepo repository.FriendRepository,
	checkinRepo repository.CheckinRepository,
	roomRepo repository.RoomRepository,
	passwordManager *auth.PasswordManager,
	idGenerator *snowflake.Node,
	logger *logger.Logger,
) UserService {
	return &userService{
		userRepo:        userRepo,
		privacyRepo:     privacyRepo,
		currencyRepo:    currencyRepo,
		friendRepo:      friendRepo,
		checkinRepo:     checkinRepo,
		roomRepo:        roomRepo,
		passwordManager: passwordManager,
		idGenerator:     idGenerator,
		logger:          logger,
	}
}

// GetUserInfo 获取当前用户信息
func (s *userService) GetUserInfo(ctx context.Context, userID int64) (*model.UserInfoResponse, error) {
	user, err := s.userRepo.FindByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("查询用户失败", zap.Error(err))
		return nil, errors.New(errors.CodeUserNotFound, "用户不存在")
	}

	// 查询当前所在房间
	var currentRoomID int64
	member, err := s.roomRepo.FindUserRoom(ctx, userID)
	if err == nil && member != nil {
		currentRoomID = member.RoomID
	}

	return &model.UserInfoResponse{
		UserID:        user.UserID,
		Username:      user.Username,
		Status:        user.Status,
		Email:         user.Email,
		Phone:         user.Phone,
		Sex:           user.Sex,
		Birthday:      user.Birthday,
		Tomato:        user.Tomato,
		Province:      user.Province,
		Avatar:        user.Avatar,
		CurrentRoomID: currentRoomID,
	}, nil
}

// GetUserByID 根据ID获取用户信息
func (s *userService) GetUserByID(ctx context.Context, userID int64) (*model.UserResponse, error) {
	user, err := s.userRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New(errors.CodeUserNotFound, "用户不存在")
	}

	return &model.UserResponse{
		UserID:   user.UserID,
		Username: user.Username,
		Status:   user.Status,
		Sex:      user.Sex,
		Birthday: user.Birthday,
		Tomato:   user.Tomato,
		Province: user.Province,
		Avatar:   user.Avatar,
	}, nil
}

// GetPublicUserInfo 获取公开的用户信息（考虑隐私设置和好友关系）
func (s *userService) GetPublicUserInfo(ctx context.Context, username string, viewerID int64) (*model.PublicUserResponse, error) {
	// 查询用户
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, errors.New(errors.CodeUserNotFound, "用户不存在")
	}

	// 获取用户隐私设置
	privacy, err := s.privacyRepo.FindByUserID(ctx, user.UserID)
	if err != nil {
		privacy = &entity.UserPrivacy{
			ShowBirthday:       constants.PrivacyLevelPublic,
			ShowStudyTime:      constants.PrivacyLevelPublic,
			ShowLocation:       constants.PrivacyLevelPublic,
			AllowFriendRequest: true,
			Searchable:         true,
		}
	}

	// 判断查看者与用户的关系
	isSelf := user.UserID == viewerID
	isFriend := false
	if !isSelf {
		friend, _ := s.friendRepo.FindFriend(ctx, viewerID, user.UserID)
		isFriend = friend != nil
	}

	// 根据隐私设置和关系返回信息
	resp := &model.PublicUserResponse{
		UserID:   user.UserID,
		Username: user.Username,
		Status:   user.Status,
		Sex:      user.Sex,
		Tomato:   user.Tomato,
		Avatar:   user.Avatar,
	}

	// 隐私字段处理
	if isSelf || privacy.ShowBirthday == constants.PrivacyLevelPublic ||
		(isFriend && privacy.ShowBirthday == constants.PrivacyLevelFriends) {
		resp.Birthday = user.Birthday
	}

	if isSelf || privacy.ShowLocation == constants.PrivacyLevelPublic ||
		(isFriend && privacy.ShowLocation == constants.PrivacyLevelFriends) {
		resp.Province = user.Province
	}

	return resp, nil
}

// UpdateUserInfo 更新用户信息
func (s *userService) UpdateUserInfo(ctx context.Context, userID int64, req *model.UpdateUserRequest) (*model.UserInfoResponse, error) {
	user, err := s.userRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New(errors.CodeUserNotFound, "用户不存在")
	}

	// 更新用户名（需要检查唯一性）
	if req.Username != "" && req.Username != user.Username {
		if existUser, _ := s.userRepo.FindByUsername(ctx, req.Username); existUser != nil {
			return nil, errors.New(errors.CodeUsernameExists, "用户名已存在")
		}
		user.Username = req.Username
	}

	// 更新密码
	if req.Password != "" {
		if err := s.passwordManager.ValidatePassword(req.Password); err != nil {
			return nil, err
		}
		hash, err := s.passwordManager.HashPassword(req.Password)
		if err != nil {
			s.logger.Error("生成密码哈希失败", zap.Error(err))
			return nil, errors.New(errors.CodeInternalError, "更新失败")
		}
		user.PasswordHash = hash
	}

	// 更新其他字段
	if req.Sex != "" {
		user.Sex = req.Sex
	}
	if req.Birthday != nil {
		user.Birthday = req.Birthday
	}
	if req.Province != "" {
		user.Province = req.Province
	}

	// 保存更新
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("更新用户信息失败", zap.Error(err))
		return nil, errors.New(errors.CodeInternalError, "更新失败")
	}

	s.logger.Info(fmt.Sprintf("用户信息更新成功: %d", userID))

	return &model.UserInfoResponse{
		UserID:   user.UserID,
		Username: user.Username,
		Status:   user.Status,
		Email:    user.Email,
		Phone:    user.Phone,
		Sex:      user.Sex,
		Birthday: user.Birthday,
		Tomato:   user.Tomato,
		Province: user.Province,
		Avatar:   user.Avatar,
	}, nil
}

// GetUserPrivacy 获取用户隐私设置
func (s *userService) GetUserPrivacy(ctx context.Context, userID int64) (*model.UserPrivacyResponse, error) {
	privacy, err := s.privacyRepo.FindByUserID(ctx, userID)
	if err != nil {
		// 如果不存在，返回默认隐私设置
		return &model.UserPrivacyResponse{
			ShowBirthday:       constants.PrivacyLevelPublic,
			ShowStudyTime:      constants.PrivacyLevelPublic,
			ShowLocation:       constants.PrivacyLevelPublic,
			AllowFriendRequest: true,
			Searchable:         true,
		}, nil
	}

	return &model.UserPrivacyResponse{
		ShowBirthday:       privacy.ShowBirthday,
		ShowStudyTime:      privacy.ShowStudyTime,
		ShowLocation:       privacy.ShowLocation,
		AllowFriendRequest: privacy.AllowFriendRequest,
		Searchable:         privacy.Searchable,
	}, nil
}

// UpdateUserPrivacy 更新用户隐私设置
func (s *userService) UpdateUserPrivacy(ctx context.Context, userID int64, req *model.UpdateUserPrivacyRequest) (*model.UserPrivacyResponse, error) {
	privacy, err := s.privacyRepo.FindByUserID(ctx, userID)
	if err != nil {
		privacy = &entity.UserPrivacy{
			UserID:             userID,
			ShowBirthday:       constants.PrivacyLevelPublic,
			ShowStudyTime:      constants.PrivacyLevelPublic,
			ShowLocation:       constants.PrivacyLevelPublic,
			AllowFriendRequest: true,
			Searchable:         true,
		}
	}

	// 更新字段
	if req.ShowBirthday != "" {
		privacy.ShowBirthday = req.ShowBirthday
	}
	if req.ShowStudyTime != "" {
		privacy.ShowStudyTime = req.ShowStudyTime
	}
	if req.ShowLocation != "" {
		privacy.ShowLocation = req.ShowLocation
	}
	privacy.AllowFriendRequest = req.AllowFriendRequest
	privacy.Searchable = req.Searchable

	// 保存更新
	if err := s.privacyRepo.Update(ctx, privacy); err != nil {
		s.logger.Error("更新隐私设置失败", zap.Error(err))
		return nil, errors.New(errors.CodeInternalError, "更新失败")
	}

	s.logger.Info(fmt.Sprintf("隐私设置更新成功: %d", userID))

	return &model.UserPrivacyResponse{
		ShowBirthday:       privacy.ShowBirthday,
		ShowStudyTime:      privacy.ShowStudyTime,
		ShowLocation:       privacy.ShowLocation,
		AllowFriendRequest: privacy.AllowFriendRequest,
		Searchable:         privacy.Searchable,
	}, nil
}

// GetUserCurrency 获取用户货币信息
func (s *userService) GetUserCurrency(ctx context.Context, userID int64) (*model.CurrencyResponse, error) {
	currency, err := s.currencyRepo.FindByUserID(ctx, userID)
	if err != nil {
		// 如果不存在，创建新的
		currency = &entity.UserCurrency{
			UserID:   userID,
			Coins:    0,
			CheckDay: 0,
		}
		if err := s.currencyRepo.Create(ctx, currency); err != nil {
			s.logger.Error("创建货币记录失败", zap.Error(err))
			return nil, errors.New(errors.CodeInternalError, "查询失败")
		}
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	hasCheckedToday := false
	if todayCount, err := s.checkinRepo.CountByUserAndDate(ctx, userID, today); err == nil && todayCount > 0 {
		hasCheckedToday = true
	}
	monthCheckDays, err := s.GetCheckinCount(ctx, userID)
	if err != nil {
		monthCheckDays = 0
	}

	return &model.CurrencyResponse{
		UserID:            currency.UserID,
		Coins:             currency.Coins,
		CheckDay:          currency.CheckDay,
		MonthCheckDays:    monthCheckDays,
		HasCheckedInToday: hasCheckedToday,
		UpdatedAt:         currency.UpdatedAt.Format("2006-01-02"),
	}, nil
}

// Checkin 用户签到
func (s *userService) Checkin(ctx context.Context, userID int64) error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 检查今天是否已签到
	count, err := s.checkinRepo.CountByUserAndDate(ctx, userID, today)
	if err != nil {
		s.logger.Error("查询签到记录失败", zap.Error(err))
		return errors.New(errors.CodeInternalError, "签到失败")
	}

	if count > 0 {
		return errors.New(errors.CodeValidationError, "今天已签到")
	}

	// 创建签到记录
	record := &entity.CheckinRecord{
		UserID:      userID,
		CheckinDate: today,
	}

	if err := s.checkinRepo.Create(ctx, record); err != nil {
		s.logger.Error("创建签到记录失败", zap.Error(err))
		return errors.New(errors.CodeInternalError, "签到失败")
	}

	// 更新货币记录
	currency, _ := s.currencyRepo.FindByUserID(ctx, userID)
	if currency != nil {
		currency.CheckDay++
		if err := s.currencyRepo.Update(ctx, currency); err != nil {
			s.logger.Error("更新签到天数失败", zap.Error(err))
		}
	}

	s.logger.Info(fmt.Sprintf("用户签到成功: %d", userID))
	return nil
}

// GetCheckinCount 获取本月签到天数
func (s *userService) GetCheckinCount(ctx context.Context, userID int64) (int, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)

	count, err := s.checkinRepo.CountByUserAndDateRange(ctx, userID, startOfMonth, endOfMonth)
	if err != nil {
		s.logger.Error("查询签到天数失败", zap.Error(err))
		return 0, errors.New(errors.CodeInternalError, "查询失败")
	}

	return count, nil
}

// GetCheckinDates 获取本月所有签到日期
func (s *userService) GetCheckinDates(ctx context.Context, userID int64) ([]string, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)

	records, err := s.checkinRepo.ListByUserAndDateRange(ctx, userID, startOfMonth, endOfMonth)
	if err != nil {
		s.logger.Error("查询签到日期失败", zap.Error(err))
		return nil, errors.New(errors.CodeInternalError, "查询失败")
	}

	dates := make([]string, 0)
	for _, record := range records {
		dates = append(dates, record.CheckinDate.Format("2006-01-02"))
	}

	return dates, nil
}
