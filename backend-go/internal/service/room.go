package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/tomato/backend/internal/domain/constants"
	"github.com/tomato/backend/internal/domain/entity"
	"github.com/tomato/backend/internal/domain/model"
	"github.com/tomato/backend/internal/pkg/errors"
	"github.com/tomato/backend/internal/pkg/logger"
	"github.com/tomato/backend/internal/repository"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RoomService 房间服务接口
type RoomService interface {
	CreateRoom(ctx context.Context, userID int64, req *model.RoomCreateRequest) (*model.RoomResponseDTO, error)
	UpdateRoom(ctx context.Context, userID int64, roomID int64, req *model.RoomUpdateRequest) (*model.RoomResponseDTO, error)
	DeleteRoom(ctx context.Context, userID int64, roomID int64) error
	GetRoomByID(ctx context.Context, roomID int64) (*model.RoomResponseDTO, error)
	GetRoomList(ctx context.Context, page, pageSize int) ([]*model.RoomResponseDTO, int64, error)
	JoinRoom(ctx context.Context, userID int64, roomID int64) error
	LeaveRoom(ctx context.Context, userID int64, roomID int64) error
	GetRoomMembers(ctx context.Context, roomID int64) ([]*model.RoomMemberResponse, error)
	TransferOwner(ctx context.Context, userID int64, roomID int64, newOwnerID int64) error
	KickMember(ctx context.Context, userID int64, roomID int64, targetUserID int64) error
	UpdateMemberStatus(ctx context.Context, userID int64, roomID int64, req *model.RoomMemberStatusUpdateRequest) error
	GetOrCreatePersonalRoom(ctx context.Context, userID int64) (*model.RoomResponseDTO, error)
}

// roomService 房间服务实现
type roomService struct {
	roomRepo    repository.RoomRepository
	userRepo    repository.UserRepository
	idGenerator *snowflake.Node
	logger      *logger.Logger
}

// NewRoomService 创建新的房间服务
func NewRoomService(
	roomRepo repository.RoomRepository,
	userRepo repository.UserRepository,
	idGenerator *snowflake.Node,
	logger *logger.Logger,
) RoomService {
	return &roomService{
		roomRepo:    roomRepo,
		userRepo:    userRepo,
		idGenerator: idGenerator,
		logger:      logger,
	}
}

// CreateRoom 创建房间
func (s *roomService) CreateRoom(ctx context.Context, userID int64, req *model.RoomCreateRequest) (*model.RoomResponseDTO, error) {
	// 1. 兼容性处理
	roomName := req.RoomName
	if roomName == "" {
		roomName = req.RoomName2
	}
	
	maxMembers := req.MaxMembers
	if maxMembers <= 0 {
		maxMembers = req.MaxMembers2
	}
	if maxMembers <= 0 {
		maxMembers = 20 // 默认值
	}

	musicName := req.MusicName
	if musicName == "" {
		musicName = req.MusicName2
	}

	// 2. 验证输入
	if roomName == "" {
		return nil, errors.New(errors.CodeValidationError, "房间名称不能为空")
	}
	if maxMembers < 2 || maxMembers > 100 {
		return nil, errors.New(errors.CodeValidationError, "房间人数必须在2-100之间")
	}

	// 3. 生成房间ID (6位数字，确保兼容前端 regex 和后端 int64)
	// 使用时间戳种子生成随机数
	rand.Seed(time.Now().UnixNano())
	roomID := int64(100000 + rand.Intn(900000))

	// 4. 创建房间
	room := &entity.Room{
		RoomID:       roomID,
		RoomName:     roomName,
		CreatePerson: userID,
		MaxMembers:   maxMembers,
		EndTime:      req.EndTime,
		MusicName:    musicName,
	}

	if err := s.roomRepo.Create(ctx, room); err != nil {
		s.logger.Error("创建房间失败", zap.Error(err))
		return nil, errors.New(errors.CodeInternalError, "创建房间失败")
	}

	// 5. 创建房间成员记录（创建者为房主）
	member := &entity.RoomMember{
		RoomID: roomID,
		UserID: userID,
		Role:   constants.RoomRoleOwner,
		Status: constants.RoomMemberStatusResting,
	}

	if err := s.roomRepo.AddMember(ctx, member); err != nil {
		s.logger.Error("添加房间成员失败", zap.Error(err))
		return nil, errors.New(errors.CodeInternalError, "创建房间失败")
	}

	s.logger.Info(fmt.Sprintf("房间创建成功: %d (创建者: %d)", roomID, userID))

	return s.roomToResponse(room), nil
}

// UpdateRoom 更新房间
func (s *roomService) UpdateRoom(ctx context.Context, userID int64, roomID int64, req *model.RoomUpdateRequest) (*model.RoomResponseDTO, error) {
	// 1. 兼容性处理
	roomName := req.RoomName
	if roomName == "" {
		roomName = req.RoomName2
	}
	
	maxMembers := req.MaxMembers
	if maxMembers <= 0 {
		maxMembers = req.MaxMembers2
	}

	// 2. 查询房间
	room, err := s.roomRepo.FindByRoomID(ctx, roomID)
	if err != nil {
		return nil, errors.New(errors.CodeRoomNotFound, "房间不存在")
	}

	// 3. 检查权限
	if room.CreatePerson != userID {
		return nil, errors.New(errors.CodeNotRoomOwner, "只有房主可以修改房间")
	}

	// 4. 更新字段
	if roomName != "" {
		room.RoomName = roomName
	}
	if maxMembers > 0 {
		room.MaxMembers = maxMembers
	}
	if req.EndTime != nil {
		room.EndTime = req.EndTime
	}
	if req.MusicName != "" {
		room.MusicName = req.MusicName
	} else if req.MusicName2 != "" {
		room.MusicName = req.MusicName2
	}

	// 5. 保存更新
	if err := s.roomRepo.Update(ctx, room); err != nil {
		s.logger.Error("更新房间失败", zap.Error(err))
		return nil, errors.New(errors.CodeInternalError, "更新房间失败")
	}

	s.logger.Info(fmt.Sprintf("房间更新成功: %d", roomID))

	return s.roomToResponse(room), nil
}

// DeleteRoom 删除房间（解散房间）
func (s *roomService) DeleteRoom(ctx context.Context, userID int64, roomID int64) error {
	// 1. 查询房间
	room, err := s.roomRepo.FindByRoomID(ctx, roomID)
	if err != nil {
		return errors.New(errors.CodeRoomNotFound, "房间不存在")
	}

	// 2. 检查权限
	if room.CreatePerson != userID {
		return errors.New(errors.CodeNotRoomOwner, "只有房主可以解散房间")
	}

	// 3. 删除房间所有成员
	if err := s.roomRepo.DeleteMembersByRoomID(ctx, room.RoomID); err != nil {
		s.logger.Error("删除房间成员失败", zap.Error(err))
		return errors.New(errors.CodeInternalError, "解散房间失败：无法移除成员")
	}

	// 4. 删除房间
	if err := s.roomRepo.Delete(ctx, room.ID, room); err != nil {
		s.logger.Error("删除房间记录失败", zap.Error(err))
		return errors.New(errors.CodeInternalError, "解散房间失败：无法删除房间记录")
	}

	s.logger.Info(fmt.Sprintf("房间已解散: %d", roomID))
	return nil
}

// GetRoomByID 获取房间详情
func (s *roomService) GetRoomByID(ctx context.Context, roomID int64) (*model.RoomResponseDTO, error) {
	room, err := s.roomRepo.FindByRoomID(ctx, roomID)
	if err != nil {
		return nil, errors.New(errors.CodeRoomNotFound, "房间不存在")
	}

	return s.roomToResponse(room), nil
}

// GetRoomList 获取房间列表
func (s *roomService) GetRoomList(ctx context.Context, page, pageSize int) ([]*model.RoomResponseDTO, int64, error) {
	rooms, total, err := s.roomRepo.ListRooms(ctx, page, pageSize)
	if err != nil {
		s.logger.Error("查询房间列表失败", zap.Error(err))
		return nil, 0, errors.New(errors.CodeInternalError, "查询房间失败")
	}

	responses := []*model.RoomResponseDTO{}
	for _, room := range rooms {
		responses = append(responses, s.roomToResponse(room))
	}

	return responses, total, nil
}

// JoinRoom 加入房间
func (s *roomService) JoinRoom(ctx context.Context, userID int64, roomID int64) error {
	// 1. 查询房间
	room, err := s.roomRepo.FindByRoomID(ctx, roomID)
	if err != nil {
		return errors.New(errors.CodeRoomNotFound, "房间不存在")
	}

	// 2. 获取当前成员
	members, err := s.roomRepo.GetMembers(ctx, roomID)
	if err != nil {
		return errors.New(errors.CodeInternalError, "获取房间成员失败")
	}

	if len(members) >= room.MaxMembers {
		return errors.New(errors.CodeRoomFull, "房间已满员")
	}

	// 3. 检查是否已在房间中
	inRoom := false
	for _, m := range members {
		if m.UserID == userID {
			inRoom = true
			break
		}
	}
	if inRoom {
		return errors.New(errors.CodeAlreadyInRoom, "已经在房间中")
	}

	// 4. 添加成员
	now := time.Now()
	member := &entity.RoomMember{
		RoomID:   roomID,
		UserID:   userID,
		Role:     constants.RoomRoleMember,
		Status:   constants.RoomMemberStatusResting,
		JoinedAt: &now,
	}

	if err := s.roomRepo.AddMember(ctx, member); err != nil {
		s.logger.Error("加入房间失败", zap.Error(err))
		return errors.New(errors.CodeInternalError, "加入房间失败")
	}

	s.logger.Info(fmt.Sprintf("用户%d加入房间%d", userID, roomID))
	return nil
}

// LeaveRoom 离开房间（处理房主转移逻辑）
func (s *roomService) LeaveRoom(ctx context.Context, userID int64, roomID int64) error {
	// 1. 查询房间
	room, err := s.roomRepo.FindByRoomID(ctx, roomID)
	if err != nil {
		return errors.New(errors.CodeRoomNotFound, "房间不存在")
	}

	// 2. 获取房间成员
	members, err := s.roomRepo.GetMembers(ctx, roomID)
	if err != nil {
		return errors.New(errors.CodeInternalError, "获取房间成员失败")
	}

	// 检查是否在房间中
	inRoom := false
	for _, m := range members {
		if m.UserID == userID {
			inRoom = true
			break
		}
	}
	if !inRoom {
		return errors.New(errors.CodeNotInRoom, "未在房间中")
	}

	isOwner := room.CreatePerson == userID

	// 3. 如果是房主，需要转移房主或解散房间
	if isOwner {
		// 找出最早加入的成员（除了当前用户）
		var newOwner *entity.RoomMember
		for _, m := range members {
			if m.UserID != userID && m.Status == constants.RoomMemberStatusResting || m.Status == constants.RoomMemberStatusFocusing {
				if newOwner == nil || (m.JoinedAt != nil && newOwner.JoinedAt != nil && m.JoinedAt.Before(*newOwner.JoinedAt)) {
					newOwner = m
				}
			}
		}

		// 在事务中进行转移
		err = s.roomRepo.Transaction(ctx, func(tx *gorm.DB) error {
			// 删除当前用户
			if err := s.roomRepo.RemoveMember(ctx, roomID, userID); err != nil {
				return err
			}

			if newOwner != nil {
				// 转移房主
				room.CreatePerson = newOwner.UserID
				newOwner.Role = constants.RoomRoleOwner
				if err := s.roomRepo.Update(ctx, room); err != nil {
					return err
				}
			} else {
				// 没有其他成员，解散房间
				if err := s.roomRepo.Delete(ctx, room.ID, room); err != nil {
					return err
				}
			}

			return nil
		})

		if err != nil {
			s.logger.Error("转移房主失败", zap.Error(err))
			return errors.New(errors.CodeInternalError, "离开房间失败")
		}
	} else {
		// 普通成员，直接删除
		if err := s.roomRepo.RemoveMember(ctx, roomID, userID); err != nil {
			s.logger.Error("移除房间成员失败", zap.Error(err))
			return errors.New(errors.CodeInternalError, "离开房间失败")
		}
	}

	s.logger.Info(fmt.Sprintf("用户%d离开房间%d", userID, roomID))
	return nil
}

// GetRoomMembers 获取房间成员列表
func (s *roomService) GetRoomMembers(ctx context.Context, roomID int64) ([]*model.RoomMemberResponse, error) {
	// 1. 检查房间是否存在
	_, err := s.roomRepo.FindByRoomID(ctx, roomID)
	if err != nil {
		return nil, errors.New(errors.CodeRoomNotFound, "房间不存在")
	}

	// 2. 获取成员列表
	members, err := s.roomRepo.GetMembers(ctx, roomID)
	if err != nil {
		s.logger.Error("获取房间成员失败", zap.Error(err))
		return nil, errors.New(errors.CodeInternalError, "获取成员失败")
	}

	// 3. 转换为响应格式并补充用户名
	responses := []*model.RoomMemberResponse{}
	for _, member := range members {
		username := ""
		user, _ := s.userRepo.FindByUserID(ctx, member.UserID)
		if user != nil {
			username = user.Username
		}

		responses = append(responses, &model.RoomMemberResponse{
			UserID:   member.UserID,
			Username: username,
			Role:     member.Role,
			Status:   member.Status,
			JoinedAt: member.JoinedAt,
		})
	}

	return responses, nil
}

// UpdateMemberStatus 更新成员状态
func (s *roomService) UpdateMemberStatus(ctx context.Context, userID int64, roomID int64, req *model.RoomMemberStatusUpdateRequest) error {
	// 1. 验证用户是否在房间中
	members, err := s.roomRepo.GetMembers(ctx, roomID)
	if err != nil {
		return errors.New(errors.CodeInternalError, "获取房间成员失败")
	}

	var targetMember *entity.RoomMember
	for _, m := range members {
		if m.UserID == userID {
			targetMember = m
			break
		}
	}

	if targetMember == nil {
		return errors.New(errors.CodeNotInRoom, "未在房间中")
	}

	// 2. 更新状态
	targetMember.Status = req.Status
	// 如果前端传递了 focusStartTime，可以保存到成员记录中（如果有这个字段）
	// 目前 RoomMember 实体可能没有这个字段，我们先更新 Status

	if err := s.roomRepo.UpdateMember(ctx, targetMember); err != nil {
		s.logger.Error("更新成员状态失败", zap.Error(err))
		return errors.New(errors.CodeInternalError, "更新状态失败")
	}

	return nil
}

// GetOrCreatePersonalRoom 获取或创建个人自习室
func (s *roomService) GetOrCreatePersonalRoom(ctx context.Context, userID int64) (*model.RoomResponseDTO, error) {
	// 1. 查找用户创建的房间
	rooms, err := s.roomRepo.FindByCreator(ctx, userID)
	if err == nil {
		for _, r := range rooms {
			if r.RoomName == "我的自习室" {
				return s.roomToResponse(r), nil
			}
		}
	}

	// 2. 如果没找到，创建一个默认的
	req := &model.RoomCreateRequest{
		RoomName:   "我的自习室",
		MaxMembers: 10,
	}
	return s.CreateRoom(ctx, userID, req)
}

// TransferOwner 转移房主
func (s *roomService) TransferOwner(ctx context.Context, userID int64, roomID int64, newOwnerID int64) error {
	// 1. 查询房间
	room, err := s.roomRepo.FindByRoomID(ctx, roomID)
	if err != nil {
		return errors.New(errors.CodeRoomNotFound, "房间不存在")
	}

	// 2. 检查权限
	if room.CreatePerson != userID {
		return errors.New(errors.CodeNotRoomOwner, "只有房主可以转移房主")
	}

	// 4. 获取房间成员并检查新房主是否在房间内
	members, err := s.roomRepo.GetMembers(ctx, roomID)
	if err != nil {
		return errors.New(errors.CodeInternalError, "获取房间成员失败")
	}
	inRoom := false
	for _, m := range members {
		if m.UserID == newOwnerID {
			inRoom = true
			break
		}
	}
	if !inRoom {
		return errors.New(errors.CodeNotInRoom, "新房主不在房间中")
	}

	// 4. 转移房主
	room.CreatePerson = newOwnerID
	if err := s.roomRepo.Update(ctx, room); err != nil {
		s.logger.Error("转移房主失败", zap.Error(err))
		return errors.New(errors.CodeInternalError, "转移房主失败")
	}

	s.logger.Info(fmt.Sprintf("房间%d的房主已转移到用户%d", roomID, newOwnerID))
	return nil
}

// KickMember 踢出成员
func (s *roomService) KickMember(ctx context.Context, userID int64, roomID int64, targetUserID int64) error {
	// 1. 查询房间
	room, err := s.roomRepo.FindByRoomID(ctx, roomID)
	if err != nil {
		return errors.New(errors.CodeRoomNotFound, "房间不存在")
	}

	// 2. 检查权限
	if room.CreatePerson != userID {
		return errors.New(errors.CodeNotRoomOwner, "只有房主可以踢出成员")
	}

	// 3. 不能踢出房主自己
	if targetUserID == userID {
		return errors.New(errors.CodeForbidden, "不能踢出自己")
	}

	// 4. 移除成员
	if err := s.roomRepo.RemoveMember(ctx, roomID, targetUserID); err != nil {
		s.logger.Error("踢出成员失败", zap.Error(err))
		return errors.New(errors.CodeInternalError, "踢出成员失败")
	}

	s.logger.Info(fmt.Sprintf("用户%d被踢出房间%d", targetUserID, roomID))
	return nil
}

// ========== 私有方法 ==========

// roomToResponse 转换为响应DTO
func (s *roomService) roomToResponse(room *entity.Room) *model.RoomResponseDTO {
	currentMembers, _ := s.roomRepo.CountMembers(context.Background(), room.RoomID)
	return &model.RoomResponseDTO{
		RoomID:         room.RoomID,
		RoomName:       room.RoomName,
		CreatePerson:   room.CreatePerson,
		MaxMembers:     room.MaxMembers,
		CurrentMembers: int(currentMembers),
		EndTime:        room.EndTime,
	}
}
