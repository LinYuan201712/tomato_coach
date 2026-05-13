package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"github.com/tomato/backend/config"
	"github.com/tomato/backend/internal/channels"
	"github.com/tomato/backend/internal/domain/entity"
	"github.com/tomato/backend/internal/pkg/auth"
	"github.com/tomato/backend/internal/pkg/bus"
	"github.com/tomato/backend/internal/pkg/logger"
	"github.com/tomato/backend/internal/repository"
	"github.com/tomato/backend/internal/router"
	"github.com/tomato/backend/internal/service"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 1. 加载配置
	cfg, err := config.LoadConfig("./config")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. 初始化日志
	appLogger, err := logger.Initialize(&cfg.Log)
	if err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	defer appLogger.Sync()

	appLogger.Info("应用启动中...")

	// 3. 初始化数据库
	db, err := initDatabase(cfg)
	if err != nil {
		appLogger.Fatal("Failed to connect to database", zap.Error(err))
	}

	appLogger.Info("数据库连接成功")

	// 4. 初始化ID生成器
	idGenerator, err := snowflake.NewNode(1) // node id = 1
	if err != nil {
		appLogger.Fatal("初始化ID生成器失败", zap.Error(err))
	}

	// 5. 初始化认证相关
	passwordManager := auth.NewPasswordManager()
	tokenManager := auth.NewTokenManager(cfg.JWT.Secret, cfg.JWT.Expiration)

	// 6. 初始化Repository
	userRepo := repository.NewUserRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	roomRepo := repository.NewRoomRepository(db)
	friendRepo := repository.NewFriendRepository(db)
	friendRequestRepo := repository.NewFriendRequestRepository(db)
	focusSessionRepo := repository.NewFocusSessionRepository(db)
	userPrivacyRepo := repository.NewUserPrivacyRepository(db)
	userCurrencyRepo := repository.NewUserCurrencyRepository(db)
	checkinRepo := repository.NewCheckinRepository(db)
	musicRepo := repository.NewBackgroundMusicRepository(db)
	configRepo := repository.NewSystemConfigRepository(db)
	chatRepo := repository.NewChatRepository(db)
	feishuRepo := repository.NewUserFeishuConfigRepository(db)
	studyReportRepo := repository.NewStudyReportRepository(db)

	// 7. 初始化Service
	authService := service.NewAuthService(
		userRepo,
		userCurrencyRepo,
		userPrivacyRepo,
		passwordManager,
		tokenManager,
		idGenerator,
		appLogger,
	)

	userService := service.NewUserService(
		userRepo,
		userPrivacyRepo,
		userCurrencyRepo,
		friendRepo,
		checkinRepo,
		roomRepo,
		passwordManager,
		idGenerator,
		appLogger,
	)

	taskService := service.NewTaskService(
		taskRepo,
		userRepo,
		userCurrencyRepo,
		idGenerator,
		appLogger,
	)
	roomService := service.NewRoomService(
		roomRepo,
		userRepo,
		idGenerator,
		appLogger,
	)
	friendService := service.NewFriendService(
		friendRepo,
		friendRequestRepo,
		userRepo,
		idGenerator,
		appLogger,
	)
	focusService := service.NewFocusService(
		focusSessionRepo,
		userRepo,
		taskRepo,
		roomRepo,
		idGenerator,
		appLogger,
	)
	musicService := service.NewBackgroundMusicService(musicRepo, appLogger)
	configService := service.NewSystemConfigService(configRepo, appLogger)
	coachService, err := service.NewCoachService(cfg, taskService, chatRepo, userRepo, studyReportRepo, db, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to init CoachService", zap.Error(err))
		// 初始化失败则直接退出，避免将 nil 服务注入路由导致运行时 panic
	}

	// 8. 初始化Gin引擎
	gin.SetMode(gin.DebugMode)
	engine := gin.New()

	// 9. 初始化消息通道与总线
	messageBus := bus.NewMessageBus()
	channelMgr := channels.NewManager(messageBus, appLogger)

	// 10. 初始化飞书业务服务
	feishuService := service.NewFeishuService(
		feishuRepo,
		channelMgr,
		messageBus,
		appLogger,
	)

	studyReportService := service.NewStudyReportService(
		studyReportRepo,
		focusSessionRepo,
		taskRepo,
		chatRepo,
		userRepo,
		coachService.GetStudyAgent(),
		messageBus,
		appLogger,
	)

	// 11. 设置路由
	router.SetupRoutes(
		engine,
		authService,
		userService,
		taskService,
		roomService,
		friendService,
		focusService,
		musicService,
		configService,
		coachService,
		feishuService,
		studyReportService,
		tokenManager,
		appLogger,
	)

	// 12. 启动异步后台任务与连接
	channelBridge := service.NewChannelBridge(coachService, messageBus, appLogger)
	ctx := context.Background()

	// 启动出站消息分发
	go channelMgr.DispatchOutbound(ctx)

	// 启动桥接器监听
	go func() {
		if err := channelBridge.Start(ctx); err != nil {
			appLogger.Error("ChannelBridge 异常退出", zap.Error(err))
		}
	}()

	// 13. 初始化并启动所有用户自定义的飞书机器人连接
	if err := feishuService.InitAllConnections(ctx); err != nil {
		appLogger.Error("初始化用户飞书连接失败", zap.Error(err))
	}

	// 启动学习报告定时任务
	studyReportService.StartCron()

	// 14. 启动服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: engine,
	}

	appLogger.Info("Starting server on " + srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		appLogger.Fatal("Failed to start server", zap.Error(err))
	}
}

// initDatabase 初始化数据库连接
func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	// 构建DSN
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Database,
	)

	// 打开数据库连接
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, err
	}

	// 获取底层SQL数据库
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(0) // 禁用过期检查

	// 自动迁移Entity
	if err := autoMigrate(db); err != nil {
		return nil, err
	}

	return db, nil
}

// autoMigrate 自动迁移数据库表
func autoMigrate(db *gorm.DB) error {
	// 暂时禁用外键约束进行迁移，避免循环依赖导致创建失败
	db.Exec("SET FOREIGN_KEY_CHECKS = 0;")
	defer db.Exec("SET FOREIGN_KEY_CHECKS = 1;")

	return db.AutoMigrate(
		&entity.User{},
		&entity.Task{},
		&entity.Room{},
		&entity.RoomMember{},
		&entity.Friend{},
		&entity.FriendRequest{},
		&entity.FocusSession{},
		&entity.BackgroundMusic{},
		&entity.SystemConfig{},
		&entity.StudyReport{},
		&entity.UserPrivacy{},
		&entity.UserCurrency{},
		&entity.CheckinRecord{},
		&entity.ChatSession{},
		&entity.ChatMessage{},
		&entity.UserFeishuConfig{},
		&entity.KnowledgeFolder{},
		&entity.KnowledgeFile{},
	)
}
