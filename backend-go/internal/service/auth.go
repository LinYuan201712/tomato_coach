package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/snowflake"
	"github.com/tomato/backend/internal/domain/constants"
	"github.com/tomato/backend/internal/domain/entity"
	"github.com/tomato/backend/internal/domain/model"
	"github.com/tomato/backend/internal/pkg/auth"
	"github.com/tomato/backend/internal/pkg/errors"
	"github.com/tomato/backend/internal/pkg/logger"
	"github.com/tomato/backend/internal/repository"
	"gorm.io/gorm"

	"go.uber.org/zap"
)

// AuthService 认证服务接口
type AuthService interface {
	SendVerificationCode(ctx context.Context, req *model.SendVerificationCodeRequest) error
	Register(ctx context.Context, req *model.RegisterRequest) (*model.AuthResponse, error)
	Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error)
	Logout(ctx context.Context, userID int64) error
	ValidateToken(ctx context.Context, token string) (*auth.CustomClaims, error)
}

// authService 认证服务实现
type authService struct {
	userRepo        repository.UserRepository
	currencyRepo    repository.UserCurrencyRepository
	privacyRepo     repository.UserPrivacyRepository
	passwordManager *auth.PasswordManager
	tokenManager    *auth.TokenManager
	idGenerator     *snowflake.Node
	emailSender     EmailSender
	logger          *logger.Logger
}

// NewAuthService 创建新的认证服务
func NewAuthService(
	userRepo repository.UserRepository,
	currencyRepo repository.UserCurrencyRepository,
	privacyRepo repository.UserPrivacyRepository,
	passwordManager *auth.PasswordManager,
	tokenManager *auth.TokenManager,
	idGenerator *snowflake.Node,
	emailSender EmailSender,
	logger *logger.Logger,
) AuthService {
	return &authService{
		userRepo:        userRepo,
		currencyRepo:    currencyRepo,
		privacyRepo:     privacyRepo,
		passwordManager: passwordManager,
		tokenManager:    tokenManager,
		idGenerator:     idGenerator,
		emailSender:     emailSender,
		logger:          logger,
	}
}

// Register 用户注册
func (s *authService) Register(ctx context.Context, req *model.RegisterRequest) (*model.AuthResponse, error) {
	// 1. 验证输入
	if err := s.validateRegisterRequest(req); err != nil {
		return nil, err
	}

	// 2. 检查用户是否已存在
	if err := s.checkUserExists(ctx, req.Username, req.Email, req.Phone); err != nil {
		return nil, err
	}

	if err := s.verifyEmailCode(req.Email, req.VerificationCode); err != nil {
		return nil, err
	}

	// 3. 生成业务ID和密码哈希
	userID := s.idGenerator.Generate().Int64()
	passwordHash, err := s.passwordManager.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("生成密码哈希失败", zap.Error(err))
		return nil, errors.New(errors.CodeInternalError, "注册失败")
	}

	// 4. 在事务中创建用户和相关记录
	var token string
	err = s.userRepo.Transaction(ctx, func(tx *gorm.DB) error {
		// 创建用户
		user := &entity.User{
			UserID:       userID,
			Username:     req.Username,
			Email:        req.Email,
			Phone:        req.Phone,
			PasswordHash: passwordHash,
			Status:       constants.UserStatusOffline,
		}

		if err := tx.Create(user).Error; err != nil {
			// 处理唯一性约束冲突
			if strings.Contains(err.Error(), "Duplicate entry") {
				return errors.New(errors.CodeUsernameExists, "用户信息已存在")
			}
			s.logger.Error("创建用户失败", zap.Error(err))
			return errors.New(errors.CodeInternalError, "注册失败")
		}

		// 创建用户隐私设置
		privacy := &entity.UserPrivacy{
			UserID:             userID,
			ShowBirthday:       constants.PrivacyLevelPublic,
			ShowStudyTime:      constants.PrivacyLevelPublic,
			ShowLocation:       constants.PrivacyLevelPublic,
			AllowFriendRequest: true,
			Searchable:         true,
		}
		if err := tx.Create(privacy).Error; err != nil {
			s.logger.Error("创建隐私设置失败", zap.Error(err))
			return errors.New(errors.CodeInternalError, "注册失败")
		}

		// 创建用户货币记录
		currency := &entity.UserCurrency{
			UserID:   userID,
			Coins:    0,
			CheckDay: 0,
		}
		if err := tx.Create(currency).Error; err != nil {
			s.logger.Error("创建货币记录失败", zap.Error(err))
			return errors.New(errors.CodeInternalError, "注册失败")
		}

		// 生成token
		var tokenErr error
		token, tokenErr = s.tokenManager.GenerateToken(userID, req.Username)
		if tokenErr != nil {
			s.logger.Error("生成token失败", zap.Error(tokenErr))
			return errors.New(errors.CodeInternalError, "注册失败")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	s.logger.Info(fmt.Sprintf("用户注册成功: %s", req.Username))

	return &model.AuthResponse{
		Token:     token,
		TokenType: "Bearer",
		UserID:    userID,
		Username:  req.Username,
	}, nil
}

// Login 用户登录
func (s *authService) Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error) {
	// 1. 查询用户（支持用户名、邮箱或电话）
	user, err := s.findUserByIdentifier(ctx, req.Username)
	if err != nil {
		if errors.IsBusinessError(err) {
			return nil, err
		}
		s.logger.Error("查询用户失败", zap.Error(err))
		return nil, errors.New(errors.CodeUserNotFound, "用户不存在")
	}

	// 2. 验证密码
	if !s.passwordManager.VerifyPassword(user.PasswordHash, req.Password) {
		return nil, errors.New(errors.CodeInvalidPassword, "密码错误")
	}

	// 3. 更新用户状态为在线
	if err := s.userRepo.UpdateStatus(ctx, user.UserID, constants.UserStatusOnline); err != nil {
		s.logger.Error("更新用户状态失败", zap.Error(err))
		return nil, errors.New(errors.CodeInternalError, "登录失败")
	}

	// 4. 生成token
	token, err := s.tokenManager.GenerateToken(user.UserID, user.Username)
	if err != nil {
		s.logger.Error("生成token失败", zap.Error(err))
		return nil, errors.New(errors.CodeInternalError, "登录失败")
	}

	s.logger.Info(fmt.Sprintf("用户登录成功: %s", user.Username))

	return &model.AuthResponse{
		Token:     token,
		TokenType: "Bearer",
		UserID:    user.UserID,
		Username:  user.Username,
	}, nil
}

// Logout 用户登出
func (s *authService) Logout(ctx context.Context, userID int64) error {
	// 更新用户状态为离线
	if err := s.userRepo.UpdateStatus(ctx, userID, constants.UserStatusOffline); err != nil {
		s.logger.Error("更新用户状态失败", zap.Error(err))
		return errors.New(errors.CodeInternalError, "登出失败")
	}

	s.logger.Info(fmt.Sprintf("用户登出成功: userID=%d", userID))
	return nil
}

// ValidateToken 验证token
func (s *authService) ValidateToken(ctx context.Context, token string) (*auth.CustomClaims, error) {
	claims, err := s.tokenManager.ValidateToken(token)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// ========== 私有方法 ==========

// validateRegisterRequest 验证注册请求
func (s *authService) validateRegisterRequest(req *model.RegisterRequest) error {
	if strings.TrimSpace(req.Username) == "" {
		return errors.New(errors.CodeValidationError, "用户名不能为空")
	}

	if strings.TrimSpace(req.Email) == "" {
		return errors.New(errors.CodeValidationError, "邮箱不能为空")
	}

	if err := s.passwordManager.ValidatePassword(req.Password); err != nil {
		return err
	}

	return nil
}

// checkUserExists 检查用户是否已存在
func (s *authService) checkUserExists(ctx context.Context, username, email, phone string) error {
	// 检查用户名
	if user, _ := s.userRepo.FindByUsername(ctx, username); user != nil {
		return errors.New(errors.CodeUsernameExists, constants.ErrMsgUsernameExists)
	}

	// 检查邮箱
	if user, _ := s.userRepo.FindByEmail(ctx, email); user != nil {
		return errors.New(errors.CodeEmailExists, constants.ErrMsgEmailExists)
	}

	// 检查电话
	if strings.TrimSpace(phone) != "" {
		if user, _ := s.userRepo.FindByPhone(ctx, phone); user != nil {
			return errors.New(errors.CodePhoneExists, constants.ErrMsgPhoneExists)
		}
	}

	return nil
}

// findUserByIdentifier 根据用户名/邮箱/电话查找用户
func (s *authService) findUserByIdentifier(ctx context.Context, identifier string) (*entity.User, error) {
	// 先尝试用户名
	user, _ := s.userRepo.FindByUsername(ctx, identifier)
	if user != nil {
		return user, nil
	}

	// 再尝试邮箱
	user, _ = s.userRepo.FindByEmail(ctx, identifier)
	if user != nil {
		return user, nil
	}

	// 最后尝试电话
	user, _ = s.userRepo.FindByPhone(ctx, identifier)
	if user != nil {
		return user, nil
	}

	return nil, errors.New(errors.CodeUserNotFound, "用户不存在")
}
