package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tomato/backend/internal/domain/model"
	"github.com/tomato/backend/internal/pkg/logger"
	"github.com/tomato/backend/internal/service"
)

// TaskHandler 任务处理器
type TaskHandler struct {
	*BaseHandler
	taskService service.TaskService
}

// NewTaskHandler 创建新的任务处理器
func NewTaskHandler(taskService service.TaskService, logger *logger.Logger) *TaskHandler {
	return &TaskHandler{
		BaseHandler: NewBaseHandler(logger),
		taskService: taskService,
	}
}

// CreateTask 创建任务
// @Summary 创建任务
// @Description 创建一个新的学习任务
// @Tags 任务
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body model.TaskCreateRequest true "创建请求"
// @Success 200 {object} Response{data=model.TaskResponse}
// @Failure 400,401,500 {object} Response
// @Router /task [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	var req model.TaskCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "请求参数错误")
		return
	}

	resp, err := h.taskService.CreateTask(c, userID, &req)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, resp)
}

// UpdateTask 更新任务
// @Summary 更新任务
// @Description 更新任务信息
// @Tags 任务
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "任务ID"
// @Param body body model.TaskUpdateRequest true "更新请求"
// @Success 200 {object} Response{data=model.TaskResponse}
// @Failure 400,401,500 {object} Response
// @Router /task/{id} [put]
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	var req model.TaskUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "请求参数错误")
		return
	}

	// 获取任务ID：优先从路径获取，否则从请求体获取
	var taskID int64
	taskIDStr := c.Param("id")
	if taskIDStr != "" {
		taskID, _ = strconv.ParseInt(taskIDStr, 10, 64)
	} else {
		// 兼容前端从 body 传 id
		if req.TaskID > 0 {
			taskID = req.TaskID
		} else if req.TaskId > 0 {
			taskID = req.TaskId
		}
	}

	if taskID <= 0 {
		h.BadRequest(c, "无效的任务ID")
		return
	}

	resp, err := h.taskService.UpdateTask(c, userID, taskID, &req)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, resp)
}

// DeleteTask 删除任务
// @Summary 删除任务
// @Description 删除指定的任务
// @Tags 任务
// @Security Bearer
// @Produce json
// @Param id path int true "任务ID"
// @Success 200 {object} Response
// @Failure 400,401,500 {object} Response
// @Router /task/{id} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	// 获取任务ID：优先从路径获取，否则尝试从请求体获取
	var taskID int64
	taskIDStr := c.Param("id")
	if taskIDStr != "" {
		taskID, _ = strconv.ParseInt(taskIDStr, 10, 64)
	} else {
		// 兼容前端从 body 传 id
		var req model.TaskDeleteRequest
		if err := c.ShouldBindJSON(&req); err == nil {
			if req.TaskID > 0 {
				taskID = req.TaskID
			} else if req.TaskId > 0 {
				taskID = req.TaskId
			}
		}
	}

	if taskID <= 0 {
		h.BadRequest(c, "无效的任务ID")
		return
	}

	if err := h.taskService.DeleteTask(c, userID, taskID); err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, nil)
}

// GetTaskList 获取任务列表
// @Summary 获取任务列表
// @Description 获取当前用户的所有任务列表
// @Tags 任务
// @Security Bearer
// @Produce json
// @Success 200 {object} Response{data=[]model.TaskResponse}
// @Failure 401,500 {object} Response
// @Router /task/list [get]
func (h *TaskHandler) GetTaskList(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	resp, err := h.taskService.GetTaskList(c, userID)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, resp)
}

// CompleteTask 完成任务
// @Summary 完成任务
// @Description 将任务状态修改为已完成并获取奖励
// @Tags 任务
// @Security Bearer
// @Produce json
// @Param id path int true "任务ID"
// @Success 200 {object} Response
// @Failure 400,401,500 {object} Response
// @Router /task/{id}/complete [put]
func (h *TaskHandler) CompleteTask(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		h.BadRequest(c, "无效的任务ID")
		return
	}

	if err := h.taskService.CompleteTask(c, userID, taskID); err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, nil)
}
