package service

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/tomato/backend/config"
	"github.com/tomato/backend/einox/agent"
	"github.com/tomato/backend/einox/callback"
	"github.com/tomato/backend/einox/memory"
	"github.com/tomato/backend/einox/model"
	"github.com/tomato/backend/einox/rag"
	"github.com/tomato/backend/einox/tools"
	"github.com/tomato/backend/einox/workflow"
	"github.com/tomato/backend/internal/domain/constants"
	"github.com/tomato/backend/internal/domain/entity"
	"github.com/tomato/backend/internal/pkg/errors"
	"github.com/tomato/backend/internal/pkg/logger"
	"github.com/tomato/backend/internal/repository"
	"gorm.io/gorm"
	"os"
	"path/filepath"
)

// CoachService 学习教练服务
type CoachService interface {
	ChatStream(ctx context.Context, userID int64, sessionID string, message string, useKnowledge bool, chatMode string) (*schema.StreamReader[*schema.Message], error)
	SaveMessage(ctx context.Context, userID int64, sessionID string, msg *schema.Message) error

	// 会话管理
	GetSessions(ctx context.Context, userID int64) ([]*entity.ChatSession, error)
	CreateSession(ctx context.Context, userID int64, sessionID string, title string) error
	UpdateSessionTitle(ctx context.Context, sessionID string, title string) error

	UploadKnowledge(ctx context.Context, userID int64, fileName string, content []byte, folderID int64) error
	ListKnowledge(ctx context.Context, userID int64, folderID int64) ([]*entity.KnowledgeFile, error)
	DeleteKnowledge(ctx context.Context, userID int64, docID string) error
	RenameKnowledge(ctx context.Context, userID int64, fileID int64, newName string) error
	MoveKnowledge(ctx context.Context, userID int64, fileID int64, targetFolderID int64) error
	GetHistory(ctx context.Context, userID int64, sessionID string) ([]*schema.Message, error)
	GetKnowledgePreview(ctx context.Context, userID int64, fileName string) (string, error)

	// 文件夹管理
	CreateFolder(ctx context.Context, userID int64, name string, parentID int64) (*entity.KnowledgeFolder, error)
	ListFolders(ctx context.Context, userID int64, parentID int64) ([]*entity.KnowledgeFolder, error)
	DeleteFolder(ctx context.Context, userID int64, folderID int64) error

	// 画像管理
	GetUserProfile(ctx context.Context, userID int64) (*entity.User, error)
	UpdateUserProfile(ctx context.Context, userID int64, goals, style, lock string, clearSuggestions bool) error

	GetStudyAgent() *agent.StudyAgent
}

type coachService struct {
	chatModel   *model.ChatModel
	rag         *rag.SimpleRAG
	memory      *memory.PersistentMemory
	db          *gorm.DB
	taskService TaskService
	graph       compose.Runnable[*workflow.CoachChatInput, *schema.Message]
	studyAgent  *agent.StudyAgent
	mineru      *rag.MinerUWorker
	logger      *logger.Logger
}

func NewCoachService(cfg *config.Config, taskSvc TaskService, chatRepo repository.ChatRepository, userRepo repository.UserRepository, reportRepo repository.StudyReportRepository, db *gorm.DB, logger *logger.Logger) (CoachService, error) {
	ctx := context.Background()

	var smartModel, cheapModel *model.ChatModel
	var err error

	// 0. 初始化全局回调处理器 (Langfuse + Logging)
	var lfCallback *callback.LangfuseCallback
	isValidKey := func(k string) bool {
		return k != "" && !strings.HasPrefix(k, "$") && !strings.HasPrefix(k, "<")
	}
	if cfg.Langfuse.Enabled && isValidKey(cfg.Langfuse.PublicKey) && isValidKey(cfg.Langfuse.SecretKey) {
		lfCallback = callback.NewLangfuseCallback(cfg.Langfuse.PublicKey, cfg.Langfuse.SecretKey, cfg.Langfuse.BaseURL)
		logger.Info("Langfuse observability enabled")
	} else {
		logger.Info("Langfuse observability disabled")
	}
	cbHandler := callback.NewModelCallbackHandler(cfg.Server.Environment == "dev", lfCallback)

	// 初始化聊天模型
	if cfg.Aliyun.APIKey != "" {
		smartCfg := &config.AliyunConfig{APIKey: cfg.Aliyun.APIKey, ChatModel: cfg.Aliyun.ChatModel}
		cheapCfg := &config.AliyunConfig{APIKey: cfg.Aliyun.APIKey, ChatModel: cfg.Aliyun.CheapModel}

		smartInner, err := model.NewAliyunChatModel(ctx, smartCfg)
		if err != nil {
			return nil, fmt.Errorf("创建阿里云 Smart 模型失败: %w", err)
		}
		cheapInner, err := model.NewAliyunChatModel(ctx, cheapCfg)
		if err != nil {
			return nil, fmt.Errorf("创建阿里云 Cheap 模型失败: %w", err)
		}

		smartModel = model.NewChatModelFromInner(
			model.NewCallbackDecorator(
				model.NewRetryAndCircuitBreakerDecorator(smartInner, "Aliyun-Smart"),
				"Aliyun-Smart",
				cbHandler,
			),
		)
		cheapModel = model.NewChatModelFromInner(
			model.NewCallbackDecorator(
				model.NewRetryAndCircuitBreakerDecorator(cheapInner, "Aliyun-Cheap"),
				"Aliyun-Cheap",
				cbHandler,
			),
		)

		logger.Infof("使用阿里云聊天模型 (Smart: %s, Cheap: %s)", cfg.Aliyun.ChatModel, cfg.Aliyun.CheapModel)
	} else {
		smartCfg := &config.ARKConfig{
			APIKey:         cfg.ARK.APIKey,
			ChatModel:      cfg.ARK.ChatModel,
			EmbeddingModel: cfg.ARK.EmbeddingModel,
			BaseURL:        cfg.ARK.BaseURL,
		}
		cheapCfg := &config.ARKConfig{
			APIKey:         cfg.ARK.APIKey,
			ChatModel:      cfg.ARK.CheapModel,
			EmbeddingModel: cfg.ARK.EmbeddingModel,
			BaseURL:        cfg.ARK.BaseURL,
		}

		smartInner, err := model.NewARKChatModel(ctx, smartCfg)
		if err != nil {
			return nil, fmt.Errorf("创建方舟 Smart 模型失败: %w", err)
		}
		cheapInner, err := model.NewARKChatModel(ctx, cheapCfg)
		if err != nil {
			return nil, fmt.Errorf("创建方舟 Cheap 模型失败: %w", err)
		}

		smartModel = model.NewChatModelFromInner(
			model.NewCallbackDecorator(
				model.NewRetryAndCircuitBreakerDecorator(smartInner, "ARK-Smart"),
				"ARK-Smart",
				cbHandler,
			),
		)
		cheapModel = model.NewChatModelFromInner(
			model.NewCallbackDecorator(
				model.NewRetryAndCircuitBreakerDecorator(cheapInner, "ARK-Cheap"),
				"ARK-Cheap",
				cbHandler,
			),
		)

		logger.Infof("使用方舟聊天模型 (Smart: %s, Cheap: %s)", cfg.ARK.ChatModel, cfg.ARK.CheapModel)
	}

	// 初始化向量化模型
	var embedder rag.Embedder
	if cfg.Aliyun.APIKey != "" {
		aliyunCfg := &config.AliyunConfig{
			APIKey:         cfg.Aliyun.APIKey,
			EmbeddingModel: cfg.Aliyun.EmbeddingModel,
		}
		embedder, err = model.NewAliyunEmbedder(ctx, aliyunCfg)
		if err != nil {
			return nil, fmt.Errorf("创建阿里云向量化模型失败: %w", err)
		}
	} else {
		arkCfg := &config.ARKConfig{
			APIKey:         cfg.ARK.APIKey,
			EmbeddingModel: cfg.ARK.EmbeddingModel,
			BaseURL:        cfg.ARK.BaseURL,
		}
		embedder, err = model.NewEmbedder(ctx, arkCfg)
		if err != nil {
			return nil, fmt.Errorf("创建方舟向量化模型失败: %w", err)
		}
	}

	// 1. 存储层初始化
	var store rag.VectorStore
	esStore, err := rag.NewElasticsearchStore(&cfg.Elastic)
	if err != nil {
		logger.WithError(err).Error("初始化 Elasticsearch 失败，降级使用 MemoryStore")
		store = rag.NewMemoryVectorStore()
	} else {
		store = esStore
	}

	// 2. 创建 RAG (核心检索依然使用 Smart Model 关联，但目前 RAG 接口只需 embedder 和 store)
	simpleRAG := rag.NewSimpleRAG(store, embedder, smartModel, cfg.RAG.TopK, cfg.RAG.HybridWeight)

	if cfg.Aliyun.APIKey != "" {
		reranker := rag.NewAliyunReranker(cfg.Aliyun.APIKey)
		simpleRAG.SetReranker(reranker)
		logger.Info("已启用阿里云 Rerank (gte-rerank-v2)")
	}

	// 3. 初始化持久化记忆 (摘要生成和事实提取使用 Cheap Model 以降低成本)
	episodicMemory := memory.NewVectorEpisodicMemory(store, embedder)
	extractor := memory.NewFactExtractor(cheapModel)
	persistentMemory := memory.NewPersistentMemory(chatRepo, userRepo, cheapModel, extractor, episodicMemory)

	// 4. 初始化工具注册表 (Tool RAG)
	toolRegistry := tools.NewSemanticToolRegistry(embedder)
	allTools := tools.GetAllTools(taskSvc, userRepo, simpleRAG, reportRepo)
	for _, t := range allTools {
		_ = toolRegistry.Register(ctx, t)
	}

	// 5. 初始化工作流 (注入 Smart 和 Cheap 模型，以及工具注册表)
	graph, err := workflow.BuildCoachChatGraph(ctx, smartModel, cheapModel, simpleRAG, toolRegistry, cbHandler)
	if err != nil {
		return nil, fmt.Errorf("构建 CoachChat 图失败: %w", err)
	}

	studyAgent := agent.NewStudyAgent(smartModel, simpleRAG, toolRegistry)
	mineruWorker := rag.NewMinerUWorker(&cfg.MinerU)

	return &coachService{
		chatModel:   smartModel,
		rag:         simpleRAG,
		memory:      persistentMemory,
		db:          db,
		taskService: taskSvc,
		graph:       graph,
		studyAgent:  studyAgent,
		mineru:      mineruWorker,
		logger:      logger,
	}, nil
}

func (s *coachService) getTaskSummary(ctx context.Context, userID int64) string {
	tasks, err := s.taskService.GetTaskList(ctx, userID)
	if err != nil {
		return "无法获取任务状态"
	}
	unfinishedCount := 0
	for _, t := range tasks {
		if t.Status != constants.TaskStatusCompleted {
			unfinishedCount++
		}
	}
	if unfinishedCount == 0 {
		return "当前没有未完成的任务。"
	}
	return fmt.Sprintf("当前有 %d 个未完成的任务。", unfinishedCount)
}

func (s *coachService) ChatStream(ctx context.Context, userID int64, sessionID string, message string, useKnowledge bool, chatMode string) (*schema.StreamReader[*schema.Message], error) {
	// 1. 生成全局 Trace ID 并注入用户 ID 到上下文
	traceID := uuid.New().String()
	ctx = callback.SetTraceInfo(ctx, traceID, fmt.Sprintf("%d", userID))
	ctx = context.WithValue(ctx, "user_id", userID) // 关键修复：显式注入 userID 供工具调用
	if useKnowledge {
		allowedFiles, err := s.getActiveKnowledgeFileSet(ctx, userID)
		if err != nil {
			s.logger.WithError(err).Warn("获取知识库文件清单失败")
		} else {
			ctx = context.WithValue(ctx, rag.AllowedKnowledgeFilesContextKey, allowedFiles)
		}
		if err := s.cleanupStaleKnowledgeIndex(ctx, userID); err != nil {
			s.logger.WithError(err).Warn("清理过期知识库索引失败")
		}
	}

	history, err := s.memory.GetHistory(ctx, userID, sessionID)
	if err != nil {
		s.logger.WithError(err).Error("Get history failed")
	}

	profile, _ := s.memory.GetUserProfile(ctx, userID)
	episodic, _ := s.memory.Recall(ctx, userID, message)

	input := &workflow.CoachChatInput{
		UserID:          userID,
		Query:           message,
		History:         history,
		UseKnowledge:    useKnowledge,
		ChatMode:        chatMode,
		TaskSummary:     s.getTaskSummary(ctx, userID),
		UserProfile:     profile,
		EpisodicContext: episodic,
	}

	if len(history) == 0 && message != "" {
		go s.autoGenerateTitle(context.Background(), message, sessionID)
	}

	sr, err := s.graph.Stream(ctx, input)
	if err != nil {
		return nil, err
	}

	resSr, sw := schema.Pipe[*schema.Message](10)
	go func() {
		defer sw.Close()
		defer sr.Close()
		var fullMsg *schema.Message
		for {
			msg, err := sr.Recv()
			if err != nil {
				if err == io.EOF {
					fmt.Println(">>> CoachService: Graph stream EOF reached")
					if fullMsg != nil {
						// 流结束时异步提取事实并保存
						s.memory.ProcessFacts(context.Background(), userID, message, fullMsg.Content)
						s.SaveMessage(context.Background(), userID, sessionID, schema.UserMessage(message))
						s.SaveMessage(context.Background(), userID, sessionID, fullMsg)
					}
				} else {
					fmt.Printf(">>> CoachService: Graph stream error: %v\n", err)
					sw.Send(nil, err)
				}
				break
			}

			if msg != nil && msg.Content != "" {
				fmt.Printf(">>> CoachService: Forwarding to pipe: [%s]\n", msg.Content)
			}
			sw.Send(msg, nil)
			if fullMsg == nil {
				fullMsg = &schema.Message{
					Role:             msg.Role,
					Content:          msg.Content,
					ReasoningContent: msg.ReasoningContent,
					ResponseMeta:     msg.ResponseMeta,
				}
			} else {
				fullMsg.Content += msg.Content
				fullMsg.ReasoningContent += msg.ReasoningContent
				if msg.ResponseMeta != nil {
					fullMsg.ResponseMeta = msg.ResponseMeta
				}
			}
		}
	}()
	return resSr, nil
}

func (s *coachService) autoGenerateTitle(ctx context.Context, query string, sessionID string) {
	titlePrompt := fmt.Sprintf("你是一个对话标题提取助手。请根据用户的第一句话，总结出一个简短的对话标题（不超过10个字），直接输出标题内容，不要带引号或任何解释性文字。用户的话是: \"%s\"", query)
	resp, err := s.chatModel.Generate(ctx, []*schema.Message{
		schema.UserMessage(titlePrompt),
	})
	if err != nil {
		s.logger.WithError(err).Error("自动生成标题失败")
		return
	}
	title := strings.TrimSpace(resp.Content)
	if title != "" {
		s.logger.Printf("生成标题成功: %s -> %s", sessionID, title)
		_ = s.memory.UpdateSessionTitle(ctx, sessionID, title)
	}
}

func (s *coachService) SaveMessage(ctx context.Context, userID int64, sessionID string, msg *schema.Message) error {
	return s.memory.SaveMessage(ctx, userID, sessionID, msg)
}

func (s *coachService) GetSessions(ctx context.Context, userID int64) ([]*entity.ChatSession, error) {
	return s.memory.GetSessions(ctx, userID)
}

func (s *coachService) CreateSession(ctx context.Context, userID int64, sessionID string, title string) error {
	return s.memory.CreateSession(ctx, userID, sessionID, title)
}

func (s *coachService) UpdateSessionTitle(ctx context.Context, sessionID string, title string) error {
	return s.memory.UpdateSessionTitle(ctx, sessionID, title)
}

func (s *coachService) UploadKnowledge(ctx context.Context, userID int64, fileName string, content []byte, folderID int64) error {
	// 1. 保存到数据库获取元数据
	status := constants.KnowledgeStatusActive
	if rag.IsPDF(fileName) {
		status = constants.KnowledgeStatusParsing
	}

	file := &entity.KnowledgeFile{
		UserID:      userID,
		FolderID:    folderID,
		FileName:    fileName,
		DisplayName: fileName,
		FileSize:    int64(len(content)),
		Status:      status,
	}
	// 同名文件重新上传：清理旧数据库记录与 RAG 索引，避免预览读到旧的二进制分片
	_ = s.rag.DeleteFile(ctx, fileName, userID)
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND file_name = ?", userID, fileName).
		Delete(&entity.KnowledgeFile{}).Error; err != nil {
		return fmt.Errorf("清理旧文件记录失败: %w", err)
	}

	if err := s.db.WithContext(ctx).Create(file).Error; err != nil {
		return fmt.Errorf("保存文件元数据失败: %w", err)
	}

	// 2. 如果是 PDF，启动异步解析
	if rag.IsPDF(fileName) {
		go func() {
			bgCtx := context.Background()
			// 保存临时文件供 MinerU SDK 读取
			tmpDir := filepath.Join("tmp", "mineru")
			_ = os.MkdirAll(tmpDir, 0755)
			tmpPath := filepath.Join(tmpDir, fmt.Sprintf("%d_%s", file.ID, fileName))
			if err := os.WriteFile(tmpPath, content, 0644); err != nil {
				s.logger.Errorf("写入临时文件失败: %v", err)
				s.db.Model(file).Update("status", constants.KnowledgeStatusFailed)
				return
			}
			defer os.Remove(tmpPath)

			// 解析
			s.logger.Infof("[MinerU] 开始解析 PDF: %s (ID: %d)", fileName, file.ID)
			markdown, err := s.mineru.Extract(bgCtx, tmpPath)
			if err != nil {
				s.logger.Errorf("[MinerU] 解析失败: %v", err)
				s.db.Model(file).Update("status", constants.KnowledgeStatusFailed)
				return
			}

			// 索引
			metadata := map[string]any{
				"file_name":   fileName,
				"user_id":     userID,
				"file_id":     file.ID,
				"upload_time": file.CreatedAt.Unix(),
				"_extension":  ".md",
			}
			if err := s.rag.AddText(bgCtx, markdown, metadata); err != nil {
				s.logger.Errorf("[RAG] 索引失败: %v", err)
				s.db.Model(file).Update("status", constants.KnowledgeStatusFailed)
				return
			}

			// 更新状态为活跃
			s.db.Model(file).Update("status", constants.KnowledgeStatusActive)
			s.logger.Infof("[MinerU] 解析并索引成功: %s", fileName)
		}()
		return nil
	}

	// 3. 非 PDF：解析 Word/Markdown/文本 后再索引
	text, err := rag.ExtractText(fileName, content)
	if err != nil {
		_ = s.db.WithContext(ctx).Delete(file).Error
		return fmt.Errorf("解析文件内容失败: %w", err)
	}

	metadata := map[string]any{
		rag.MetaFileName:   fileName,
		rag.MetaUserID:     userID,
		"file_id":          file.ID,
		"upload_time":      file.CreatedAt.Unix(),
		rag.MetaFullText:   text,
	}
	if rag.IsDocx(fileName) {
		metadata["_extension"] = ".md"
	}
	s.logger.Infof("[Knowledge] 已解析 %s，文本长度 %d 字符", fileName, len([]rune(text)))

	return s.rag.AddText(ctx, text, metadata)
}

func (s *coachService) ListKnowledge(ctx context.Context, userID int64, folderID int64) ([]*entity.KnowledgeFile, error) {
	var files []*entity.KnowledgeFile
	query := s.db.WithContext(ctx).Where("user_id = ?", userID)
	if folderID > 0 {
		query = query.Where("folder_id = ?", folderID)
	}
	if err := query.Find(&files).Error; err != nil {
		return nil, err
	}

	// 自动同步逻辑：如果数据库中没有记录，但 RAG 中有，则进行同步
	if len(files) == 0 && folderID == 0 {
		docs, err := s.rag.ListDocuments(ctx, userID)
		if err == nil && len(docs) > 0 {
			s.logger.Infof("检测到 RAG 中有 %d 个旧文件，正在同步到数据库...", len(docs))
			for _, d := range docs {
				fileName, _ := d.MetaData["file_name"].(string)
				uploadTime, _ := d.MetaData["upload_time"].(int64)
				if uploadTime == 0 {
					if ut, ok := d.MetaData["upload_time"].(float64); ok {
						uploadTime = int64(ut)
					}
				}

				newFile := &entity.KnowledgeFile{
					UserID:      userID,
					FileName:    fileName,
					DisplayName: fileName,
					FileSize:    int64(len(d.Content)),
					DocID:       d.ID,
				}
				s.db.WithContext(ctx).Create(newFile)
				files = append(files, newFile)
			}
		}
	}

	return files, nil
}

func (s *coachService) DeleteKnowledge(ctx context.Context, userID int64, docID string) error {
	// 1. 删除数据库记录 (如果是数字ID则转换)
	// 这里暂时用文件名作为 docID 兼容之前的逻辑，或者查找记录
	var file entity.KnowledgeFile
	if err := s.db.WithContext(ctx).Where("user_id = ? AND (id = ? OR file_name = ?)", userID, docID, docID).First(&file).Error; err == nil {
		if err := s.db.WithContext(ctx).Delete(&file).Error; err != nil {
			return err
		}
		if err := s.rag.DeleteFile(ctx, file.FileName, userID); err != nil {
			return err
		}
		return nil
	}

	return s.rag.DeleteDocument(ctx, docID)
}

func (s *coachService) cleanupStaleKnowledgeIndex(ctx context.Context, userID int64) error {
	activeFiles, err := s.getActiveKnowledgeFileSet(ctx, userID)
	if err != nil {
		return err
	}

	docs, err := s.rag.ListDocuments(ctx, userID)
	if err != nil {
		return err
	}

	for _, doc := range docs {
		fileName, _ := doc.MetaData[rag.MetaFileName].(string)
		fileName = strings.TrimSpace(fileName)
		if fileName == "" || activeFiles[fileName] {
			continue
		}
		s.logger.Infof("[RAG] 清理数据库中不存在的知识库索引: %s", fileName)
		if err := s.rag.DeleteFile(ctx, fileName, userID); err != nil {
			return err
		}
	}
	return nil
}

func (s *coachService) getActiveKnowledgeFileSet(ctx context.Context, userID int64) (map[string]bool, error) {
	var files []entity.KnowledgeFile
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&files).Error; err != nil {
		return nil, err
	}

	activeFiles := make(map[string]bool, len(files))
	for _, file := range files {
		fileName := strings.TrimSpace(file.FileName)
		if fileName != "" {
			activeFiles[fileName] = true
		}
	}
	return activeFiles, nil
}

func (s *coachService) RenameKnowledge(ctx context.Context, userID int64, fileID int64, newName string) error {
	return s.db.WithContext(ctx).Model(&entity.KnowledgeFile{}).
		Where("id = ? AND user_id = ?", fileID, userID).
		Update("display_name", newName).Error
}

func (s *coachService) MoveKnowledge(ctx context.Context, userID int64, fileID int64, targetFolderID int64) error {
	return s.db.WithContext(ctx).Model(&entity.KnowledgeFile{}).
		Where("id = ? AND user_id = ?", fileID, userID).
		Update("folder_id", targetFolderID).Error
}

func (s *coachService) CreateFolder(ctx context.Context, userID int64, name string, parentID int64) (*entity.KnowledgeFolder, error) {
	folder := &entity.KnowledgeFolder{
		UserID:   userID,
		Name:     name,
		ParentID: parentID,
	}
	if err := s.db.WithContext(ctx).Create(folder).Error; err != nil {
		return nil, err
	}
	return folder, nil
}

func (s *coachService) ListFolders(ctx context.Context, userID int64, parentID int64) ([]*entity.KnowledgeFolder, error) {
	var folders []*entity.KnowledgeFolder
	if err := s.db.WithContext(ctx).Where("user_id = ? AND parent_id = ?", userID, parentID).Find(&folders).Error; err != nil {
		return nil, err
	}
	return folders, nil
}

func (s *coachService) DeleteFolder(ctx context.Context, userID int64, folderID int64) error {
	// 事务处理：删除文件夹及其中的文件引用（或者把文件移到根目录）
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 检查是否存在
		var folder entity.KnowledgeFolder
		if err := tx.Where("id = ? AND user_id = ?", folderID, userID).First(&folder).Error; err != nil {
			return err
		}

		// 2. 将该文件夹下的文件移动到根目录 (folder_id = 0)
		tx.Model(&entity.KnowledgeFile{}).Where("folder_id = ?", folderID).Update("folder_id", 0)

		// 3. 删除文件夹
		return tx.Delete(&folder).Error
	})
}

func (s *coachService) GetHistory(ctx context.Context, userID int64, sessionID string) ([]*schema.Message, error) {
	return s.memory.GetHistory(ctx, userID, sessionID)
}

func (s *coachService) GetKnowledgePreview(ctx context.Context, userID int64, fileName string) (string, error) {
	fileName = strings.TrimSpace(fileName)
	content, err := s.rag.GetFullDocument(ctx, fileName, userID)
	if err != nil {
		return content, err
	}
	content = strings.TrimSpace(content)
	if content != "" && rag.IsBinaryGarbledText(content) {
		s.logger.Warnf("[Preview] 文件 %s 索引内容为旧版二进制乱码，请删除后重新上传", fileName)
		return "", errors.New(errors.CodeValidationError, "该文件为旧版乱码索引，请删除后重新上传 docx")
	}
	if content != "" {
		return content, nil
	}

	var file entity.KnowledgeFile
	dbErr := s.db.WithContext(ctx).
		Where("user_id = ? AND (file_name = ? OR display_name = ?)", userID, fileName, fileName).
		First(&file).Error
	if dbErr != nil || strings.TrimSpace(file.FileName) == "" || file.FileName == fileName {
		return content, err
	}

	return s.rag.GetFullDocument(ctx, file.FileName, userID)
}

func (s *coachService) GetUserProfile(ctx context.Context, userID int64) (*entity.User, error) {
	return s.memory.GetUserRepo().FindByUserID(ctx, userID)
}

func (s *coachService) UpdateUserProfile(ctx context.Context, userID int64, goals, style, lock string, clearSuggestions bool) error {
	if lock != "" {
		_ = s.memory.GetUserRepo().UpdateProfileLock(ctx, userID, lock)
	}
	if goals != "" || style != "" {
		_ = s.memory.GetUserRepo().UpdateProfile(ctx, userID, goals, style)
	}
	if clearSuggestions {
		_ = s.memory.GetUserRepo().UpdateLockSuggestions(ctx, userID, "")
	}
	return nil
}

func (s *coachService) GetStudyAgent() *agent.StudyAgent {
	return s.studyAgent
}
