package service

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/tomato/backend/internal/domain/constants"
	"github.com/tomato/backend/internal/domain/entity"
	"github.com/tomato/backend/internal/domain/model"
	"github.com/tomato/backend/internal/pkg/errors"
	"github.com/tomato/backend/internal/pkg/logger"
	"github.com/tomato/backend/internal/repository"

	"go.uber.org/zap"
)

// FocusService 专注服务接口
type FocusService interface {
	StartFocus(ctx context.Context, userID int64, req *model.StartFocusRequest) (*model.FocusResponse, error)
	StopFocus(ctx context.Context, userID int64, sessionID int64) (*model.StopFocusResponse, error)
	GetFocusRecords(ctx context.Context, userID int64, days int) ([]*model.FocusRecordResponse, error)
	GetDailyReport(ctx context.Context, userID int64) (*model.StudyReportResponse, error)
	GetWeeklyReport(ctx context.Context, userID int64) (*model.StudyReportResponse, error)
	GetMonthlyReport(ctx context.Context, userID int64) (*model.StudyReportResponse, error)
}

// focusService 专注服务实现
type focusService struct {
	focusRepo   repository.FocusSessionRepository
	userRepo    repository.UserRepository
	taskRepo    repository.TaskRepository
	roomRepo    repository.RoomRepository
	idGenerator *snowflake.Node
	logger      *logger.Logger
}

// NewFocusService 创建新的专注服务
func NewFocusService(
	focusRepo repository.FocusSessionRepository,
	userRepo repository.UserRepository,
	taskRepo repository.TaskRepository,
	roomRepo repository.RoomRepository,
	idGenerator *snowflake.Node,
	logger *logger.Logger,
) FocusService {
	return &focusService{
		focusRepo:   focusRepo,
		userRepo:    userRepo,
		taskRepo:    taskRepo,
		roomRepo:    roomRepo,
		idGenerator: idGenerator,
		logger:      logger,
	}
}

// StartFocus 开始专注
func (s *focusService) StartFocus(ctx context.Context, userID int64, req *model.StartFocusRequest) (*model.FocusResponse, error) {
	_, err := s.userRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New(errors.CodeUserNotFound, "用户不存在")
	}

	// 2. 如果指定了房间或任务，验证其有效性
	if req.RoomID > 0 {
		// TODO: 验证房间存在
	}
	if req.TaskID > 0 {
		// TODO: 验证任务存在
	}

	// 3. 检查是否已经有正在进行的专注会话
	existingSession, _ := s.focusRepo.FindUnfinished(ctx, userID)
	if existingSession != nil {
		// 如果会话时间太久（比如超过 12 小时），认为是由于某种原因未正常结束的“僵尸会话”，自动强制结束它
		if time.Since(existingSession.StartTime) > 12*time.Hour {
			s.logger.Info(fmt.Sprintf("检测到僵尸会话 %d (开始于 %v)，强制结束", existingSession.SessionID, existingSession.StartTime))
			existingSession.Status = constants.SessionStatusCompleted
			now := time.Now()
			existingSession.EndTime = &now
			// 给一个保守的时长，或者干脆给 0
			existingSession.ActualDuration = 0 
			_ = s.focusRepo.Update(ctx, existingSession)
			existingSession = nil // 设置为 nil，后续将创建新会话
		}
	}

	if existingSession != nil {
		// 如果已在其他房间专注，提示用户
		if req.RoomID > 0 && existingSession.RoomID != nil && *existingSession.RoomID != req.RoomID {
			return nil, errors.New(errors.CodeBusinessError, fmt.Sprintf("您已在房间 %d 中开始专注，请先结束该专注", *existingSession.RoomID))
		}
		// 如果在同一个房间，或者是同一个任务，可以选择直接返回现有会话或提示
		if req.TaskID > 0 && existingSession.TaskID != nil && *existingSession.TaskID == req.TaskID {
			return &model.FocusResponse{
				SessionID:   existingSession.SessionID,
				UserID:      userID,
				SessionType: existingSession.SessionType,
				Duration:    int64(existingSession.Duration),
				StartTime:   existingSession.StartTime,
				Status:      existingSession.Status,
			}, nil
		}
	}

	// 4. 创建专注会话
	sessionID := s.idGenerator.Generate().Int64()
	now := time.Now()

	var roomID *int64
	if req.RoomID > 0 {
		roomID = &req.RoomID
	}
	var taskID *int64
	if req.TaskID > 0 {
		taskID = &req.TaskID
	}

	session := &entity.FocusSession{
		SessionID:   sessionID,
		UserID:      userID,
		RoomID:      roomID,
		TaskID:      taskID,
		SessionType: req.SessionType,
		Duration:    int(req.Duration),
		StartTime:   now,
		Status:      constants.SessionStatusProcessing,
	}

	if err := s.focusRepo.Create(ctx, session); err != nil {
		s.logger.Error("创建专注会话失败", zap.Error(err))
		return nil, errors.New(errors.CodeInternalError, "开始专注失败")
	}

	// 4. 更新用户状态为"专注中"
	_ = s.userRepo.UpdateStatus(ctx, userID, constants.UserStatusFocusing)

	// 5. 如果指定了房间，更新成员状态
	if roomID != nil {
		members, err := s.roomRepo.GetMembers(ctx, *roomID)
		if err == nil {
			for _, m := range members {
				if m.UserID == userID {
					m.Status = constants.RoomMemberStatusFocusing
					_ = s.roomRepo.UpdateMember(ctx, m)
					break
				}
			}
		}
	}

	s.logger.Info(fmt.Sprintf("用户%d开始专注: %s", userID, req.SessionType))

	return &model.FocusResponse{
		SessionID:   sessionID,
		UserID:      userID,
		SessionType: session.SessionType,
		Duration:    int64(session.Duration),
		StartTime:   session.StartTime,
		Status:      session.Status,
	}, nil
}

// StopFocus 结束专注
func (s *focusService) StopFocus(ctx context.Context, userID int64, sessionID int64) (*model.StopFocusResponse, error) {
	var session *entity.FocusSession
	var err error

	if sessionID == 0 {
		// 如果未提供 sessionID，尝试查找该用户当前正在进行的会话
		session, err = s.focusRepo.FindUnfinished(ctx, userID)
		if err != nil {
			return nil, errors.New(errors.CodeBusinessError, "当前没有正在进行的专注会话")
		}
		sessionID = session.SessionID
	} else {
		// 查询指定专注会话
		session, err = s.focusRepo.FindByID(ctx, sessionID)
		if err != nil {
			return nil, errors.New(errors.CodeInternalError, "会话不存在")
		}
	}

	// 2. 检查权限
	if session.UserID != userID {
		return nil, errors.New(errors.CodeForbidden, "无权操作此会话")
	}

	// 3. 检查会话状态
	if session.Status != constants.SessionStatusProcessing {
		return nil, errors.New(errors.CodeForbidden, "会话已结束")
	}

	// 4. 计算实际专注时长
	now := time.Now()
	actualDuration := int64(now.Sub(session.StartTime).Seconds())
	
	// 增加安全性检查：实际时长不应超过计划时长的 1.5 倍（或者至少给个合理的上限，比如 4 小时）
	// 计划时长是以分钟为单位的
	maxAllowedSeconds := int64(session.Duration) * 60 + 60 // 允许 1 分钟宽限
	if actualDuration > maxAllowedSeconds {
		actualDuration = maxAllowedSeconds
	}
	if actualDuration < 0 {
		actualDuration = 0
	}

	// 5. 更新会话信息
	session.EndTime = &now
	session.ActualDuration = int(actualDuration)
	session.Status = constants.SessionStatusCompleted

	if err := s.focusRepo.Update(ctx, session); err != nil {
		s.logger.Error("更新专注会话失败", zap.Error(err))
		return nil, errors.New(errors.CodeInternalError, "结束专注失败")
	}

	// 6. 如果有关联任务，更新任务的实际时长并标记为完成
	if session.TaskID != nil && *session.TaskID > 0 {
		task, err := s.taskRepo.FindByTaskID(ctx, *session.TaskID)
		if err == nil {
			task.ActualDuration += int(actualDuration)
			// 自动完成任务
			task.Status = constants.TaskStatusCompleted
			task.EndTime = &now
			if err := s.taskRepo.Update(ctx, task); err != nil {
				s.logger.Error("自动完成任务失败", zap.Error(err))
			} else {
				// 自动完成任务奖励番茄
				_ = s.userRepo.UpdateTomato(ctx, userID, 1)
			}
		}
	}

	// 7. 更新用户状态为"在线"
	_ = s.userRepo.UpdateStatus(ctx, userID, constants.UserStatusOnline)

	// 8. 如果会话关联了房间，更新成员状态
	if session.RoomID != nil {
		members, err := s.roomRepo.GetMembers(ctx, *session.RoomID)
		if err == nil {
			for _, m := range members {
				if m.UserID == userID {
					m.Status = constants.RoomMemberStatusResting
					_ = s.roomRepo.UpdateMember(ctx, m)
					break
				}
			}
		}
	}

	s.logger.Info(fmt.Sprintf("用户%d结束专注会话%d，时长%d秒", userID, sessionID, actualDuration))

	return &model.StopFocusResponse{
		SessionID:      sessionID,
		ActualDuration: actualDuration,
		EndTime:        now,
		Status:         session.Status,
	}, nil
}

// GetFocusRecords 获取专注记录
func (s *focusService) GetFocusRecords(ctx context.Context, userID int64, days int) ([]*model.FocusRecordResponse, error) {
	if days < 1 {
		days = 7 // 默认7天
	}

	// 查询指定天数内的专注记录
	startTime := time.Now().AddDate(0, 0, -days)
	sessions, err := s.focusRepo.FindByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("查询专注记录失败", zap.Error(err))
		return nil, errors.New(errors.CodeInternalError, "获取专注记录失败")
	}

	var records []*model.FocusRecordResponse
	totalDuration := int64(0)

	for _, session := range sessions {
		// 过滤时间范围
		if session.StartTime.Before(startTime) {
			continue
		}

		// 转换为分钟，向上取整（只要开始专注了就算1分钟，方便统计显示）
		duration := int64(0)
		if session.ActualDuration > 0 {
			duration = (int64(session.ActualDuration) + 59) / 60
		}
		
		// 如果记录显示已完成但时长为0，可能是极短时间的快速点击，也给1分钟鼓励
		if duration == 0 && session.Status == constants.SessionStatusCompleted {
			duration = 1
		}

		records = append(records, &model.FocusRecordResponse{
			SessionID:   session.SessionID,
			SessionType: session.SessionType,
			Duration:    duration,
			StartTime:   session.StartTime,
			EndTime:     session.EndTime,
			Status:      session.Status,
		})

		totalDuration += duration
	}

	return records, nil
}

// GetDailyReport 获取每日报告
func (s *focusService) GetDailyReport(ctx context.Context, userID int64) (*model.StudyReportResponse, error) {
	return s.generateReport(ctx, userID, constants.ReportTypeDaily)
}

// GetWeeklyReport 获取周报告
func (s *focusService) GetWeeklyReport(ctx context.Context, userID int64) (*model.StudyReportResponse, error) {
	return s.generateReport(ctx, userID, constants.ReportTypeWeekly)
}

// GetMonthlyReport 获取月报告
func (s *focusService) GetMonthlyReport(ctx context.Context, userID int64) (*model.StudyReportResponse, error) {
	return s.generateReport(ctx, userID, constants.ReportTypeMonthly)
}

// ========== 私有方法 ==========

// generateReport 生成学习报告
func (s *focusService) generateReport(ctx context.Context, userID int64, reportType string) (*model.StudyReportResponse, error) {
	var startTime time.Time
	now := time.Now()

	switch reportType {
	case constants.ReportTypeDaily:
		// 今天
		startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case constants.ReportTypeWeekly:
		// 本周
		weekday := now.Weekday()
		startTime = now.AddDate(0, 0, -int(weekday))
		startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, startTime.Location())
	case constants.ReportTypeMonthly:
		// 本月
		startTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	default:
		return nil, errors.New(errors.CodeValidationError, "报告类型无效")
	}

	// 查询时间范围内的会话
	sessions, err := s.focusRepo.FindByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("查询专注会话失败", zap.Error(err))
		return nil, errors.New(errors.CodeInternalError, "生成报告失败")
	}

	// 统计数据
	var totalDuration int64
	var sessionCount int64
	typeCountMap := make(map[string]int64)

	for _, session := range sessions {
		if session.StartTime.Before(startTime) {
			continue
		}

		sessionCount++
		duration := int64(session.ActualDuration)
		if duration == 0 && session.Status == constants.SessionStatusCompleted {
			duration = int64(session.Duration)
		}
		totalDuration += duration

		if session.SessionType != "" {
			typeCountMap[session.SessionType]++
		}
	}

	return &model.StudyReportResponse{
		UserID:        userID,
		ReportType:    reportType,
		ReportDate:    now,
		TotalDuration: totalDuration,
		SessionCount:  sessionCount,
		AverageDuration: func() int64 {
			if sessionCount == 0 {
				return 0
			}
			return totalDuration / sessionCount
		}(),
		SessionBreakdown: typeCountMap,
	}, nil
}
