package service

import (
	"context"
	"fmt"

	"github.com/bwmarrin/snowflake"
	"github.com/tomato/backend/internal/domain/constants"
	"github.com/tomato/backend/internal/domain/entity"
	"github.com/tomato/backend/internal/domain/model"
	"github.com/tomato/backend/internal/pkg/errors"
	"github.com/tomato/backend/internal/pkg/logger"
	"github.com/tomato/backend/internal/repository"

	"go.uber.org/zap"
)

// FriendService 好友服务接口
type FriendService interface {
	SendFriendRequest(ctx context.Context, userID int64, req *model.FriendRequestRequest) error
	GetFriendRequests(ctx context.Context, userID int64, status string) ([]*model.FriendRequestResponse, error)
	ProcessFriendRequest(ctx context.Context, userID int64, req *model.ProcessFriendRequestRequest) error
	GetFriendList(ctx context.Context, userID int64) ([]*model.FriendResponse, error)
	RemoveFriend(ctx context.Context, userID int64, req *model.DeleteFriendRequest) error
}

// friendService 好友服务实现
type friendService struct {
	friendRepo        repository.FriendRepository
	friendRequestRepo repository.FriendRequestRepository
	userRepo          repository.UserRepository
	idGenerator       *snowflake.Node
	logger            *logger.Logger
}

// NewFriendService 创建新的好友服务
func NewFriendService(
	friendRepo repository.FriendRepository,
	friendRequestRepo repository.FriendRequestRepository,
	userRepo repository.UserRepository,
	idGenerator *snowflake.Node,
	logger *logger.Logger,
) FriendService {
	return &friendService{
		friendRepo:        friendRepo,
		friendRequestRepo: friendRequestRepo,
		userRepo:          userRepo,
		idGenerator:       idGenerator,
		logger:            logger,
	}
}

// SendFriendRequest 发送好友申请
func (s *friendService) SendFriendRequest(ctx context.Context, userID int64, req *model.FriendRequestRequest) error {
	// 1. 验证目标用户是否存在
	targetUser, err := s.userRepo.FindByUsername(ctx, req.UserName)
	if err != nil {
		return errors.New(errors.CodeUserNotFound, "目标用户不存在")
	}

	targetUserID := targetUser.ID

	// 2. 不能添加自己
	if targetUserID == userID {
		return errors.New(errors.CodeForbidden, "不能添加自己为好友")
	}

	// 3. 检查是否已是好友
	isFriend := s.friendRepo.IsFriend(ctx, userID, targetUserID)
	if isFriend {
		return errors.New(errors.CodeAlreadyFriend, "已经是好友")
	}

	// 4. 检查是否已有待处理的申请
	existingRequest, _ := s.friendRequestRepo.FindByFromAndTo(ctx, userID, targetUserID)
	if existingRequest != nil && existingRequest.Status == string(constants.FriendRequestStatusPending) {
		return errors.New(errors.CodeFriendRequestExists, "已发送过好友申请")
	}

	// 5. 创建好友申请
	friendReq := &entity.FriendRequest{
		FromUserID:   userID,
		FromUserName: "系统", // 从用户信息获取更合适，但此处简化
		ToUserID:     targetUserID,
		ToUserName:   targetUser.Username,
		Status:       string(constants.FriendRequestStatusPending),
		Message:      req.Message,
	}

	if err := s.friendRequestRepo.Create(ctx, friendReq); err != nil {
		s.logger.Error("创建好友申请失败", zap.Error(err))
		return errors.New(errors.CodeInternalError, "发送好友申请失败")
	}

	s.logger.Info(fmt.Sprintf("用户%d向%d发送好友申请", userID, targetUserID))
	return nil
}

// GetFriendRequests 获取好友申请列表
func (s *friendService) GetFriendRequests(ctx context.Context, userID int64, status string) ([]*model.FriendRequestResponse, error) {
	var requests []*entity.FriendRequest
	var err error

	if status == constants.FriendRequestStatusPending {
		requests, err = s.friendRequestRepo.FindPending(ctx, userID)
	} else {
		// 获取所有申请
		requests, err = s.friendRequestRepo.FindByToUserID(ctx, userID)
	}

	if err != nil {
		s.logger.Error("查询好友申请失败", zap.Error(err))
		return nil, errors.New(errors.CodeInternalError, "获取好友申请失败")
	}

	responses := []*model.FriendRequestResponse{}
	for _, req := range requests {
		responses = append(responses, &model.FriendRequestResponse{
			ID:           req.ID,
			FromUserID:   req.FromUserID,
			FromUserName: req.FromUserName,
			ToUserID:     req.ToUserID,
			ToUserName:   req.ToUserName,
			Status:       req.Status,
			Message:      req.Message,
		})
	}

	return responses, nil
}

// ProcessFriendRequest 处理好友申请
func (s *friendService) ProcessFriendRequest(ctx context.Context, userID int64, req *model.ProcessFriendRequestRequest) error {
	// 1. 查询申请
	friendReq, err := s.friendRequestRepo.FindByFromAndTo(ctx, req.FromUserID, userID)
	if err != nil {
		return errors.New(errors.CodeFriendRequestNotFound, "好友申请不存在")
	}

	// 2. 检查状态
	if friendReq.Status != string(constants.FriendRequestStatusPending) {
		return errors.New(errors.CodeForbidden, "申请已处理")
	}

	// 4. 更新状态
	var newStatus string
	if req.Action == "accept" {
		newStatus = string(constants.FriendRequestStatusAccepted)

		// 互相添加好友关系
		friend1 := &entity.Friend{
			UserID:       friendReq.FromUserID,
			FriendID:     friendReq.ToUserID,
			FriendStatus: string(constants.FriendStatusActive),
		}
		friend2 := &entity.Friend{
			UserID:       friendReq.ToUserID,
			FriendID:     friendReq.FromUserID,
			FriendStatus: string(constants.FriendStatusActive),
		}

		if err := s.friendRepo.AddFriend(ctx, friend1); err != nil {
			s.logger.Error("添加好友关系失败", zap.Error(err))
			return errors.New(errors.CodeInternalError, "处理好友申请失败")
		}

		if err := s.friendRepo.AddFriend(ctx, friend2); err != nil {
			s.logger.Error("添加好友关系失败", zap.Error(err))
			return errors.New(errors.CodeInternalError, "处理好友申请失败")
		}

		s.logger.Info(fmt.Sprintf("用户%d接受了用户%d的好友申请", userID, friendReq.FromUserID))
	} else {
		// 拒绝申请
		newStatus = string(constants.FriendRequestStatusRejected)
	}

	friendReq.Status = newStatus
	if err := s.friendRequestRepo.Update(ctx, friendReq); err != nil {
		s.logger.Error("更新好友申请状态失败", zap.Error(err))
		return errors.New(errors.CodeInternalError, "处理好友申请失败")
	}

	return nil
}

// GetFriendList 获取好友列表
func (s *friendService) GetFriendList(ctx context.Context, userID int64) ([]*model.FriendResponse, error) {
	friends, err := s.friendRepo.FindFriends(ctx, userID)
	if err != nil {
		s.logger.Error("查询好友列表失败", zap.Error(err))
		return nil, errors.New(errors.CodeInternalError, "获取好友列表失败")
	}

	responses := []*model.FriendResponse{}
	for _, f := range friends {
		// 获取好友用户信息
		friendUser, err := s.userRepo.FindByUserID(ctx, f.FriendID)
		if err == nil {
			responses = append(responses, &model.FriendResponse{
				FriendID:     friendUser.ID,
				FriendName:   friendUser.Username,
				FriendStatus: friendUser.Status,
			})
		}
	}

	return responses, nil
}

// RemoveFriend 删除好友
func (s *friendService) RemoveFriend(ctx context.Context, userID int64, req *model.DeleteFriendRequest) error {
	// 1. 根据名字查找好友用户ID
	friendUser, err := s.userRepo.FindByUsername(ctx, req.FriendName)
	if err != nil {
		return errors.New(errors.CodeUserNotFound, "好友不存在")
	}
	friendID := friendUser.ID

	// 2. 检查是否是好友
	isFriend := s.friendRepo.IsFriend(ctx, userID, friendID)
	if !isFriend {
		return errors.New(errors.CodeFriendNotFound, "不是好友")
	}

	// 3. 删除双向好友关系
	if err := s.friendRepo.RemoveFriend(ctx, userID, friendID); err != nil {
		s.logger.Error("删除好友失败", zap.Error(err))
		return errors.New(errors.CodeInternalError, "删除好友失败")
	}

	if err := s.friendRepo.RemoveFriend(ctx, friendID, userID); err != nil {
		s.logger.Error("删除好友失败", zap.Error(err))
		return errors.New(errors.CodeInternalError, "删除好友失败")
	}

	s.logger.Info(fmt.Sprintf("用户%d删除了好友%d", userID, friendID))
	return nil
}
