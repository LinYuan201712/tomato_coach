package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/tomato/backend/einox/agent"
	"github.com/tomato/backend/internal/domain/constants"
	"github.com/tomato/backend/internal/domain/entity"
	"github.com/tomato/backend/internal/domain/model"
	"github.com/tomato/backend/internal/pkg/bus"
	"github.com/tomato/backend/internal/pkg/logger"
	"github.com/tomato/backend/internal/repository"
	"go.uber.org/zap"
)

// StudyReportService 学习报告服务接口
type StudyReportService interface {
	GenerateDailyReport(ctx context.Context, userID int64, date time.Time) (*model.StudyReportResponse, error)
	RegenerateDailyReport(ctx context.Context, userID int64, date time.Time) (*model.StudyReportResponse, error)
	GetLatestReport(ctx context.Context, userID int64, reportType string) (*model.StudyReportResponse, error)
	StartCron()
}

type studyReportService struct {
	reportRepo repository.StudyReportRepository
	focusRepo  repository.FocusSessionRepository
	taskRepo   repository.TaskRepository
	chatRepo   repository.ChatRepository
	userRepo   repository.UserRepository
	studyAgent *agent.StudyAgent
	bus        *bus.MessageBus
	logger     *logger.Logger
}

// NewStudyReportService 创建新的学习报告服务
func NewStudyReportService(
	reportRepo repository.StudyReportRepository,
	focusRepo repository.FocusSessionRepository,
	taskRepo repository.TaskRepository,
	chatRepo repository.ChatRepository,
	userRepo repository.UserRepository,
	studyAgent *agent.StudyAgent,
	bus *bus.MessageBus,
	logger *logger.Logger,
) StudyReportService {
	return &studyReportService{
		reportRepo: reportRepo,
		focusRepo:  focusRepo,
		taskRepo:   taskRepo,
		chatRepo:   chatRepo,
		userRepo:   userRepo,
		studyAgent: studyAgent,
		bus:        bus,
		logger:     logger,
	}
}

// GenerateDailyReport 生成每日报告（Step 1: 基础数据聚合）
func (s *studyReportService) GenerateDailyReport(ctx context.Context, userID int64, date time.Time) (*model.StudyReportResponse, error) {
	// 1. 设置时间范围 (00:00:00 - 23:59:59)
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour).Add(-time.Second)

	// 2. 聚合数据
	sessions, err := s.focusRepo.ListByDateRange(ctx, userID, start, end)
	if err != nil {
		s.logger.Error("获取专注记录失败", zap.Error(err))
	}
	tasks, err := s.taskRepo.ListByDateRange(ctx, userID, start, end)
	if err != nil {
		s.logger.Error("获取任务记录失败", zap.Error(err))
	}
	msgs, err := s.chatRepo.GetMessagesByDateRange(ctx, userID, start, end)
	if err != nil {
		s.logger.Error("获取聊天记录失败", zap.Error(err))
	}

	// 3. 计算统计指标
	var totalDuration int64
	var completedTasks int
	sessionBreakdown := make(map[string]int64)
	for _, sess := range sessions {
		if sess.Status == constants.SessionStatusCompleted {
			totalDuration += int64(sess.ActualDuration)
			sessionBreakdown[sess.SessionType]++
		}
	}
	for _, task := range tasks {
		if task.Status == constants.TaskStatusCompleted {
			completedTasks++
		}
	}

	// 4. 构造原始数据摘要 (ChatMessage 内容简单拼接，作为 AI 输入)
	var chatSummaries []string
	for _, m := range msgs {
		if m.Role == "user" {
			chatSummaries = append(chatSummaries, m.Content)
		}
	}

	// 5. 构造元数据快照
	meta := map[string]interface{}{
		"focus_sessions_count": len(sessions),
		"tasks_count":         len(tasks),
		"completed_tasks":     completedTasks,
		"chat_messages":       chatSummaries,
		"total_duration_sec":  totalDuration,
		"breakdown":           sessionBreakdown,
	}
	metaJSON, _ := json.Marshal(meta)

	// 6. 创建报告实体
	report := &entity.StudyReport{
		UserID:         userID,
		ReportType:     constants.ReportTypeDaily,
		ReportDate:     start,
		TotalFocusTime: int(totalDuration / 60), // 转为分钟
		CompletedTasks: completedTasks,
		MetaData:       string(metaJSON),
		Content:        "AI 报告生成中...", // 初始状态
	}

	// 7. 保存到数据库 (如果已存在则更新)
	existing, _ := s.reportRepo.FindByUserIDAndDate(ctx, userID, start, constants.ReportTypeDaily)
	if existing != nil {
		report.ID = existing.ID
		report.CreatedAt = existing.CreatedAt
		_ = s.reportRepo.Update(ctx, report)
	} else {
		_ = s.reportRepo.Create(ctx, report)
	}

	// 8. 生成 AI 报告 (Step 2: Agent 集成)
	if s.studyAgent != nil {
		s.logger.Info("正在通过 AI 生成报告内容...", zap.Int64("user_id", userID))
		aiResp, err := s.studyAgent.GenerateReport(ctx, userID, "daily", string(metaJSON))
		if err == nil && aiResp != nil {
			report.Content = aiResp.Content
			_ = s.reportRepo.UpdateContent(ctx, report.ID, report.Content)
		} else {
			s.logger.Error("AI 生成报告失败", zap.Error(err))
		}
	}

	// 9. 发送飞书推送
	go s.notifyFeishu(userID, report.Content)

	return &model.StudyReportResponse{
		UserID:           userID,
		ReportType:       report.ReportType,
		ReportDate:       report.ReportDate,
		TotalDuration:    totalDuration / 60,
		SessionCount:     int64(len(sessions)),
		CompletedTasks:   int64(completedTasks),
		SessionBreakdown: sessionBreakdown,
		Content:          report.Content,
	}, nil
}

// GetLatestReport 获取最新报告
func (s *studyReportService) GetLatestReport(ctx context.Context, userID int64, reportType string) (*model.StudyReportResponse, error) {
	// 获取最近一天的报告
	now := time.Now()
	var date time.Time
	if reportType == constants.ReportTypeDaily {
		date = now.AddDate(0, 0, -1) // 默认查昨天的
	} else {
		date = now
	}

	report, err := s.reportRepo.FindByUserIDAndDate(ctx, userID, date, reportType)
	if err != nil {
		return nil, err
	}

	// 解析 breakdown (这里需要从 MetaData 解析，或者我们考虑在实体里加字段)
	// 简单处理，暂不解析 breakdown，只返回基础字段
	return &model.StudyReportResponse{
		UserID:         report.UserID,
		ReportType:     report.ReportType,
		ReportDate:     report.ReportDate,
		TotalDuration:  int64(report.TotalFocusTime),
		CompletedTasks: int64(report.CompletedTasks),
		Content:        report.Content,
	}, nil
}

// StartCron 启动定时任务
func (s *studyReportService) StartCron() {
	c := cron.New()
	
	// 每天凌晨 00:05 生成昨日日报
	_, err := c.AddFunc("5 0 * * *", func() {
		s.logger.Info("开始执行每日学习报告生成任务...")
		ctx := context.Background()
		
		// 获取所有活跃用户 (这里简单起见，可以只针对最近 7 天活跃的用户)
		var users []*entity.User
		err := s.userRepo.List(ctx, &users, "deleted = ?", false)
		if err != nil {
			s.logger.Error("获取用户列表失败", zap.Error(err))
			return
		}

		yesterday := time.Now().AddDate(0, 0, -1)
		for _, u := range users {
			_, err := s.GenerateDailyReport(ctx, u.UserID, yesterday)
			if err != nil {
				s.logger.Error("为用户生成日报失败", zap.Int64("user_id", u.UserID), zap.Error(err))
			}
		}
		s.logger.Info("每日学习报告生成任务执行完毕")
	})

	if err != nil {
		s.logger.Error("注册 Cron 任务失败", zap.Error(err))
		return
	}

	c.Start()
	s.logger.Info("学习报告定时任务已启动 (每天 00:05)")
}

// RegenerateDailyReport 重新生成日报
func (s *studyReportService) RegenerateDailyReport(ctx context.Context, userID int64, date time.Time) (*model.StudyReportResponse, error) {
	s.logger.Info("手动重新生成学习日报", zap.Int64("user_id", userID), zap.Any("date", date))
	return s.GenerateDailyReport(ctx, userID, date)
}

// notifyFeishu 发送飞书通知
func (s *studyReportService) notifyFeishu(userID int64, content string) {
	if s.bus == nil || content == "" || content == "AI 报告生成中..." {
		return
	}

	ctx := context.Background()

	// 尝试从用户的历史会话中找一个飞书会话，以获取 ChatID (OpenID)
	// 如果找不到，则无法推送 unsolicited 消息
	sessions, err := s.chatRepo.GetSessions(ctx, userID)
	var targetChatID string
	if err == nil {
		for _, sess := range sessions {
			if strings.HasPrefix(sess.SessionID, "feishu_") {
				targetChatID = strings.TrimPrefix(sess.SessionID, "feishu_")
				break
			}
		}
	}

	if targetChatID == "" {
		s.logger.Warn("找不到用户的飞书 ChatID，跳过推送", zap.Int64("user_id", userID))
		return
	}

	msg := &bus.OutboundMessage{
		ID:        fmt.Sprintf("report-%d-%d", userID, time.Now().Unix()),
		Channel:   "feishu",
		AccountID: fmt.Sprintf("%d", userID),
		ChatID:    targetChatID,
		Content:   fmt.Sprintf("📊 **您的智能学习报告已生成**\n\n%s", content),
	}

	err = s.bus.PublishOutbound(ctx, msg)
	if err != nil {
		s.logger.Error("发布学习报告通知到总线失败", zap.Error(err))
	} else {
		s.logger.Info("学习报告通知已发布到总线", zap.Int64("user_id", userID), zap.String("chat_id", targetChatID))
	}
}
