package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/tomato/backend/internal/pkg/logger"
	"github.com/tomato/backend/internal/service"
)

type CoachHandler struct {
	*BaseHandler
	coachService service.CoachService
}

func NewCoachHandler(coachService service.CoachService, logger *logger.Logger) *CoachHandler {
	return &CoachHandler{
		BaseHandler:  NewBaseHandler(logger),
		coachService: coachService,
	}
}

type ChatRequest struct {
	Message      string `json:"message" binding:"required"`
	SessionID    string `json:"session_id"`
	UseKnowledge bool   `json:"use_knowledge"`
	ChatMode     string `json:"chat_mode"` // fast | thinking
}

func (h *CoachHandler) ChatStream(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = "default"
	}

	// 默认使用 thinking 模式以保持向后兼容
	mode := req.ChatMode
	if mode == "" {
		mode = "thinking"
	}

	stream, err := h.coachService.ChatStream(c, userID, sessionID, req.Message, req.UseKnowledge, mode)
	if err != nil {
		// 记录并返回详细错误
		h.logger.WithError(err).Error("Coach ChatStream failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误", "detail": err.Error()})
		return
	}
	defer stream.Close()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// 设置 SSE 必要的 Header，防止缓冲
	// 显式设置所有 SSE Header，防止任何层级的缓冲
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("X-Accel-Buffering", "no") // 显式告诉 Nginx 等代理不要缓冲
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Flush()

	var fullContent strings.Builder
	var lastUsage interface{}

	// 手动循环读取，确保每一帧都立即发出
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				fmt.Println(">>> CoachHandler: Stream EOF reached")
				// 在流结束时，如果之前抓到了 Usage，再发一次确认帧（可选，但有助于前端更新最终状态）
				if lastUsage != nil {
					finalData := gin.H{"type": "usage", "usage": lastUsage}
					jb, _ := json.Marshal(finalData)
					fmt.Fprintf(c.Writer, "event: message\ndata: %s\n\n", string(jb))
					c.Writer.Flush()
				}
			} else {
				fmt.Printf(">>> CoachHandler: Stream Error: %v\n", err)
			}
			break
		}

		if chunk == nil {
			continue
		}

		// 累加内容用于后续
		if chunk.Content != "" {
			fullContent.WriteString(chunk.Content)
		}

		var reasoning string
		if chunk.ReasoningContent != "" {
			reasoning = chunk.ReasoningContent
		}

		var usage interface{}
		if chunk.ResponseMeta != nil && chunk.ResponseMeta.Usage != nil {
			u := chunk.ResponseMeta.Usage
			usage = gin.H{
				"prompt_tokens":     u.PromptTokens,
				"completion_tokens": u.CompletionTokens,
				"total_tokens":      u.TotalTokens,
			}
			lastUsage = usage // 缓存最新的 Usage
			fmt.Printf("[DEBUG] Captured Token Usage: %+v\n", usage)
		}

		// 构造并发送数据包
		respData := gin.H{
			"type":      "content",
			"content":   chunk.Content,
			"reasoning": reasoning,
			"usage":     usage,
		}
		jsonBytes, _ := json.Marshal(respData)
		
		fmt.Fprintf(c.Writer, "event: message\ndata: %s\n\n", string(jsonBytes))
		c.Writer.Flush() // 每一帧都强制刷新到网络层
	}

	// 最终收尾逻辑
	h.coachService.SaveMessage(c, userID, sessionID, schema.UserMessage(req.Message))
	h.coachService.SaveMessage(c, userID, sessionID, schema.AssistantMessage(fullContent.String(), nil))
	
	// 发送结束标识
	fmt.Fprintf(c.Writer, "event: done\ndata: {\"done\":true}\n\n")
	c.Writer.Flush()
}

func (h *CoachHandler) UploadKnowledge(c *gin.Context) {
	h.logger.Infof("🚀 收到上传请求: Content-Type = %s", c.GetHeader("Content-Type"))

	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.logger.Errorf("❌ 上传失败: 用户未认证")
		h.Unauthorized(c)
		return
	}

	// 尝试手动解析 multipart 确保能捕获具体错误
	file, err := c.FormFile("file")
	if err != nil {
		h.logger.Errorf("❌ 获取上传文件失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取上传文件失败: " + err.Error()})
		return
	}

	h.logger.Infof("📁 正在处理文件: %s (大小: %d)", file.Filename, file.Size)

	f, err := file.Open()
	if err != nil {
		h.logger.Errorf("❌ 打开文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "打开文件失败"})
		return
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		h.logger.Errorf("❌ 读取文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
		return
	}

	folderIDStr := c.DefaultPostForm("folderID", "0")
	var folderID int64
	fmt.Sscanf(folderIDStr, "%d", &folderID)

	err = h.coachService.UploadKnowledge(c, userID, file.Filename, content, folderID)
	if err != nil {
		h.logger.Errorf("❌ UploadKnowledge 业务逻辑失败: %v", err)
		h.Error(c, err)
		return
	}

	h.logger.Infof("✅ 文件 %s 上传并同步 RAG 成功", file.Filename)
	h.Success(c, "上传成功")
}

func (h *CoachHandler) ListKnowledge(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	folderIDStr := c.DefaultQuery("folderID", "0")
	var folderID int64
	fmt.Sscanf(folderIDStr, "%d", &folderID)

	files, err := h.coachService.ListKnowledge(c, userID, folderID)
	if err != nil {
		h.Error(c, err)
		return
	}

	// 转换为前端期望的格式
	type FileResp struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
		Size        int64  `json:"size"`
		CreatedAt   string `json:"createdAt"`
		FolderID    int64  `json:"folderId"`
	}
	res := make([]FileResp, 0)
	for _, f := range files {
		res = append(res, FileResp{
			ID:          f.ID,
			Name:        f.FileName,
			DisplayName: f.DisplayName,
			Size:        f.FileSize,
			CreatedAt:   f.CreatedAt.Format("2006-01-02 15:04"),
			FolderID:    f.FolderID,
		})
	}

	h.Success(c, res)
}

func (h *CoachHandler) DeleteKnowledge(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	docID := c.Param("id")
	err = h.coachService.DeleteKnowledge(c, userID, docID)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, nil)
}

func (h *CoachHandler) RenameKnowledge(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	var req struct {
		NewName string `json:"newName" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var fileID int64
	fmt.Sscanf(c.Param("id"), "%d", &fileID)

	err = h.coachService.RenameKnowledge(c, userID, fileID, req.NewName)
	if err != nil {
		h.Error(c, err)
		return
	}
	h.Success(c, nil)
}

func (h *CoachHandler) MoveKnowledge(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	var req struct {
		FileID         int64 `json:"fileId" binding:"required"`
		TargetFolderID int64 `json:"targetFolderId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.coachService.MoveKnowledge(c, userID, req.FileID, req.TargetFolderID)
	if err != nil {
		h.Error(c, err)
		return
	}
	h.Success(c, nil)
}

func (h *CoachHandler) CreateFolder(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	var req struct {
		Name     string `json:"name" binding:"required"`
		ParentID int64  `json:"parentId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	folder, err := h.coachService.CreateFolder(c, userID, req.Name, req.ParentID)
	if err != nil {
		h.Error(c, err)
		return
	}
	h.Success(c, folder)
}

func (h *CoachHandler) ListFolders(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	var parentID int64
	fmt.Sscanf(c.DefaultQuery("parentId", "0"), "%d", &parentID)

	folders, err := h.coachService.ListFolders(c, userID, parentID)
	if err != nil {
		h.Error(c, err)
		return
	}
	h.Success(c, folders)
}

func (h *CoachHandler) DeleteFolder(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	var folderID int64
	fmt.Sscanf(c.Param("id"), "%d", &folderID)

	err = h.coachService.DeleteFolder(c, userID, folderID)
	if err != nil {
		h.Error(c, err)
		return
	}
	h.Success(c, nil)
}

func (h *CoachHandler) GetKnowledgePreview(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	fileName := strings.TrimSpace(c.Query("fileName"))
	if fileName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少文件名"})
		return
	}

	h.logger.Infof("[Preview] 用户 %d 请求预览文件: [%s]", userID, fileName)

	content, err := h.coachService.GetKnowledgePreview(c, userID, fileName)
	if err != nil {
		h.logger.Errorf("[Preview] 获取预览失败: %v", err)
		h.Error(c, err)
		return
	}

	if content == "" {
		h.logger.Warnf("[Preview] 文件内容为空: [%s]", fileName)
	}

	h.Success(c, gin.H{"content": content})
}

func (h *CoachHandler) GetSessions(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	sessions, err := h.coachService.GetSessions(c, userID)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, sessions)
}

type CreateSessionRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	Title     string `json:"title" binding:"required"`
}

func (h *CoachHandler) CreateSession(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.coachService.CreateSession(c, userID, req.SessionID, req.Title)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, nil)
}

func (h *CoachHandler) GetHistory(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	sessionID := c.Param("id")
	messages, err := h.coachService.GetHistory(c, userID, sessionID)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, messages)
}

type UpdateSessionRequest struct {
	Title string `json:"title" binding:"required"`
}

func (h *CoachHandler) UpdateSessionTitle(c *gin.Context) {
	_, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	sessionID := c.Param("id")
	var req UpdateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.coachService.UpdateSessionTitle(c, sessionID, req.Title)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, nil)
}

func (h *CoachHandler) GetUserProfile(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	user, err := h.coachService.GetUserProfile(c, userID)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, gin.H{
		"goals":            user.Goals,
		"preferred_style":  user.PreferredStyle,
		"profile_lock":     user.ProfileLock,
		"lock_suggestions": user.LockSuggestions,
	})
}

type UpdateProfileRequest struct {
	Goals            string `json:"goals"`
	PreferredStyle   string `json:"preferred_style"`
	ProfileLock      string `json:"profile_lock"`
	ClearSuggestions bool   `json:"clear_suggestions"`
}

func (h *CoachHandler) UpdateUserProfile(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		h.Unauthorized(c)
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.coachService.UpdateUserProfile(c, userID, req.Goals, req.PreferredStyle, req.ProfileLock, req.ClearSuggestions)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, nil)
}
