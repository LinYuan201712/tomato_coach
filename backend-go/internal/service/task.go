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
	"gorm.io/gorm"
)

// TaskService 任务服务接口
type TaskService interface {
	CreateTask(ctx context.Context, userID int64, req *model.TaskCreateRequest) (*model.TaskResponse, error)
	UpdateTask(ctx context.Context, userID int64, taskID int64, req *model.TaskUpdateRequest) (*model.TaskResponse, error)
	DeleteTask(ctx context.Context, userID int64, taskID int64) error
	GetTaskList(ctx context.Context, userID int64) ([]*model.TaskResponse, error)
	CompleteTask(ctx context.Context, userID int64, taskID int64) error
}

// taskService 任务服务实现
type taskService struct {
	taskRepo     repository.TaskRepository
	userRepo     repository.UserRepository
	currencyRepo repository.UserCurrencyRepository
	idGenerator  *snowflake.Node
	logger       *logger.Logger
}

// NewTaskService 创建新的任务服务
func NewTaskService(
	taskRepo repository.TaskRepository,
	userRepo repository.UserRepository,
	currencyRepo repository.UserCurrencyRepository,
	idGenerator *snowflake.Node,
	logger *logger.Logger,
) TaskService {
	return &taskService{
		taskRepo:     taskRepo,
		userRepo:     userRepo,
		currencyRepo: currencyRepo,
		idGenerator:  idGenerator,
		logger:       logger,
	}
}

// CreateTask 创建任务
func (s *taskService) CreateTask(ctx context.Context, userID int64, req *model.TaskCreateRequest) (*model.TaskResponse, error) {
	// 1. 验证输入
	if req.Duration <= 0 {
		return nil, errors.New(errors.CodeValidationError, "计划时长必须大于0")
	}

	// 2. 生成任务ID
	taskID := s.idGenerator.Generate().Int64()

	// 3. 创建任务
	task := &entity.Task{
		TaskID:         taskID,
		UserID:         userID,
		TaskName:       req.TaskName,
		TaskNote:       req.TaskNote,
		Duration:       req.Duration,
		ActualDuration: 0,
		Status:         constants.TaskStatusUnfinished,
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		s.logger.Error("创建任务失败", zap.Error(err))
		return nil, errors.New(errors.CodeInternalError, "创建任务失败")
	}

	s.logger.Info(fmt.Sprintf("任务创建成功: %d", taskID))

	return s.taskToResponse(task), nil
}

// UpdateTask 更新任务
func (s *taskService) UpdateTask(ctx context.Context, userID int64, taskID int64, req *model.TaskUpdateRequest) (*model.TaskResponse, error) {
	// 1. 查询任务
	task, err := s.taskRepo.FindByTaskID(ctx, taskID)
	if err != nil {
		return nil, errors.New(errors.CodeTaskNotFound, "任务不存在")
	}

	// 2. 检查权限
	if task.UserID != userID {
		return nil, errors.New(errors.CodeForbidden, "无权修改此任务")
	}

	// 3. 更新字段
	if req.TaskName != "" {
		task.TaskName = req.TaskName
	}
	if req.TaskNote != "" {
		task.TaskNote = req.TaskNote
	}
	if req.Duration > 0 {
		task.Duration = req.Duration
	}
	if req.Status != "" {
		task.Status = req.Status
		if req.Status == constants.TaskStatusCompleted {
			now := time.Now()
			task.EndTime = &now
		}
	}
	if req.TaskStatus != "" {
		task.Status = req.TaskStatus
		if req.TaskStatus == constants.TaskStatusCompleted {
			now := time.Now()
			task.EndTime = &now
		}
	}

	// 4. 保存更新
	if err := s.taskRepo.Update(ctx, task); err != nil {
		s.logger.Error("更新任务失败", zap.Error(err))
		return nil, errors.New(errors.CodeInternalError, "更新任务失败")
	}

	s.logger.Info(fmt.Sprintf("任务更新成功: %d", taskID))

	return s.taskToResponse(task), nil
}

// DeleteTask 删除任务
func (s *taskService) DeleteTask(ctx context.Context, userID int64, taskID int64) error {
	// 1. 查询任务
	task, err := s.taskRepo.FindByTaskID(ctx, taskID)
	if err != nil {
		return errors.New(errors.CodeTaskNotFound, "任务不存在")
	}

	// 2. 检查权限
	if task.UserID != userID {
		return errors.New(errors.CodeForbidden, "无权删除此任务")
	}

	// 3. 删除任务
	if err := s.taskRepo.Delete(ctx, task.ID, task); err != nil {
		s.logger.Error("删除任务失败", zap.Error(err))
		return errors.New(errors.CodeInternalError, "删除任务失败")
	}

	s.logger.Info(fmt.Sprintf("任务删除成功: %d", taskID))
	return nil
}

// GetTaskList 获取任务列表
func (s *taskService) GetTaskList(ctx context.Context, userID int64) ([]*model.TaskResponse, error) {
	tasks, err := s.taskRepo.FindByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("查询任务列表失败", zap.Error(err))
		return nil, errors.New(errors.CodeInternalError, "查询任务失败")
	}

	responses := []*model.TaskResponse{}
	for _, task := range tasks {
		responses = append(responses, s.taskToResponse(task))
	}

	return responses, nil
}

// CompleteTask 完成任务（自动+1番茄）
func (s *taskService) CompleteTask(ctx context.Context, userID int64, taskID int64) error {
	// 1. 查询任务
	task, err := s.taskRepo.FindByTaskID(ctx, taskID)
	if err != nil {
		return errors.New(errors.CodeTaskNotFound, "任务不存在")
	}

	// 2. 检查权限
	if task.UserID != userID {
		return errors.New(errors.CodeForbidden, "无权修改此任务")
	}

	// 在事务中删除任务及其关联记录
	err = s.taskRepo.Transaction(ctx, func(tx *gorm.DB) error {
		// 更新状态并记录完成时间
		task.Status = constants.TaskStatusCompleted
		now := time.Now()
		task.EndTime = &now
		if err := s.taskRepo.Update(ctx, task); err != nil {
			s.logger.Error("更新任务状态失败", zap.Error(err))
			return errors.New(errors.CodeInternalError, "完成任务失败")
		}

		// 增加番茄数
		if err := s.userRepo.UpdateTomato(ctx, userID, 1); err != nil {
			s.logger.Error("更新番茄数失败", zap.Error(err))
			return errors.New(errors.CodeInternalError, "完成任务失败")
		}

		return nil
	})

	if err != nil {
		return err
	}

	s.logger.Info(fmt.Sprintf("任务完成，用户%d获得1个番茄", userID))
	return nil
}

// ========== 私有方法 ==========

// taskToResponse 转换任务为响应格式
func (s *taskService) taskToResponse(task *entity.Task) *model.TaskResponse {
	return &model.TaskResponse{
		TaskID:         task.TaskID,
		UserID:         task.UserID,
		TaskName:       task.TaskName,
		TaskNote:       task.TaskNote,
		Duration:       task.Duration,
		ActualDuration: task.ActualDuration,
		Status:         task.Status,
		CreatedAt:      task.CreatedAt,
		StartTime:      task.StartTime,
		EndTime:        task.EndTime,
	}
}
