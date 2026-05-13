package service

import (
	"context"
	"testing"

	"github.com/tomato/backend/internal/domain/constants"
	"github.com/tomato/backend/internal/domain/model"
)

// MockTaskSvcForSummary 模拟任务服务
type MockTaskSvcForSummary struct {
	TaskService
	tasks []*model.TaskResponse
}

func (m *MockTaskSvcForSummary) GetTaskList(ctx context.Context, userID int64) ([]*model.TaskResponse, error) {
	return m.tasks, nil
}

func TestGetTaskSummary(t *testing.T) {
	svc := &coachService{
		taskService: &MockTaskSvcForSummary{
			tasks: []*model.TaskResponse{
				{Status: constants.TaskStatusUnfinished},
				{Status: constants.TaskStatusProcessing},
				{Status: constants.TaskStatusCompleted},
			},
		},
	}

	ctx := context.Background()
	summary := svc.getTaskSummary(ctx, 1)
	expected := "当前有 2 个未完成的任务。"

	if summary != expected {
		t.Errorf("Expected summary '%s', got '%s'", expected, summary)
	}

	// 测试全完成情况
	svc.taskService = &MockTaskSvcForSummary{
		tasks: []*model.TaskResponse{
			{Status: constants.TaskStatusCompleted},
		},
	}
	summary = svc.getTaskSummary(ctx, 1)
	if summary != "当前没有未完成的任务。" {
		t.Errorf("Expected empty summary message, got '%s'", summary)
	}
}
