package router

import (
	"github.com/gin-gonic/gin"
	"github.com/tomato/backend/internal/handler"
	"github.com/tomato/backend/internal/middleware"
	"github.com/tomato/backend/internal/pkg/auth"
	"github.com/tomato/backend/internal/pkg/logger"
	"github.com/tomato/backend/internal/service"
)

// SetupRoutes 设置所有路由
func SetupRoutes(
	engine *gin.Engine,
	authService service.AuthService,
	userService service.UserService,
	taskService service.TaskService,
	roomService service.RoomService,
	friendService service.FriendService,
	focusService service.FocusService,
	musicService service.BackgroundMusicService,
	configService service.SystemConfigService,
	coachService service.CoachService,
	feishuService service.FeishuService,
	studyReportService service.StudyReportService,
	// 其他服务...
	tokenManager *auth.TokenManager,
	logger *logger.Logger,
) {
	// 创建处理器
	authHandler := handler.NewAuthHandler(authService, logger)
	userHandler := handler.NewUserHandler(userService, logger)
	taskHandler := handler.NewTaskHandler(taskService, logger)
	roomHandler := handler.NewRoomHandler(roomService, logger)
	friendHandler := handler.NewFriendHandler(friendService, logger)
	focusHandler := handler.NewFocusHandler(focusService, logger)
	musicHandler := handler.NewBackgroundMusicHandler(musicService, logger)
	configHandler := handler.NewSystemConfigHandler(configService, logger)
	coachHandler := handler.NewCoachHandler(coachService, logger)
	feishuHandler := handler.NewFeishuHandler(feishuService, logger)
	studyReportHandler := handler.NewStudyReportHandler(studyReportService, logger)

	// 应用全局中间件
	engine.Use(middleware.CORSMiddleware())
	engine.Use(middleware.LoggerMiddleware(logger))
	engine.Use(middleware.ErrorHandlerMiddleware(logger))

	// API路由组
	api := engine.Group("/api")
	{
		// ========== 认证路由（不需要认证）==========
		auth := api.Group("/auth")
		{
			auth.POST("/send-verification-code", authHandler.SendVerificationCode)
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// ========== 需要认证的路由 ==========
		// 应用JWT认证中间件
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(tokenManager, logger))
		{
			// 认证路由
			protected.POST("/auth/logout", authHandler.Logout)

			// 用户路由 (兼容前端 /api/me 和 /api/user/me)
			user := protected.Group("/user")
			{
				user.GET("/me", userHandler.GetUserInfo)
				user.PUT("/me", userHandler.UpdateUserInfo)
				user.GET("/me/privacy", userHandler.GetUserPrivacy)
				user.PUT("/me/privacy", userHandler.UpdateUserPrivacy)
				user.GET("/me/currency", userHandler.GetUserCurrency)
				user.POST("/checkin", userHandler.Checkin)
			}

			me := protected.Group("/me")
			{
				me.GET("", userHandler.GetUserInfo)
				me.PUT("", userHandler.UpdateUserInfo)
				me.GET("/privacy", userHandler.GetUserPrivacy)
				me.PUT("/privacy", userHandler.UpdateUserPrivacy)
				me.GET("/currency", userHandler.GetUserCurrency)
				me.POST("/checkin", userHandler.Checkin)
				me.GET("/checkin/dates", userHandler.GetCheckinDates)
				me.GET("/tasks", taskHandler.GetTaskList) // 前端获取任务列表路径

				// 渠道配置
				feishu := me.Group("/channels/feishu")
				{
					feishu.GET("", feishuHandler.GetConfig)
					feishu.PUT("", feishuHandler.UpdateConfig)
				}
			}

			// 公开用户信息
			users := protected.Group("/users")
			{
				users.GET("/:username", userHandler.GetPublicUserInfo)
				users.POST("/:userId/offline", authHandler.Logout) // 兼容前端离线请求
			}

			// 任务相关路由 (兼容前端 /api/tasks)
			tasks := protected.Group("/tasks")
			{
				tasks.POST("", taskHandler.CreateTask)
				tasks.PUT("/edit", taskHandler.UpdateTask)      // 兼容前端 /api/tasks/edit
				tasks.DELETE("/delete", taskHandler.DeleteTask) // 兼容前端 /api/tasks/delete
				tasks.GET("/list", taskHandler.GetTaskList)
				tasks.PUT("/:id/complete", taskHandler.CompleteTask) // 新增：完成任务接口
			}

			// 房间相关路由 (兼容前端 /api/rooms)
			rooms := protected.Group("/rooms")
			{
				rooms.POST("", roomHandler.CreateRoom)
				rooms.GET("", roomHandler.GetRoomList) // 兼容前端 GET /api/rooms
				rooms.GET("/personal", roomHandler.GetPersonalRoom)
				rooms.GET("/:roomId", roomHandler.GetRoomByID)
				rooms.PUT("/:roomId", roomHandler.UpdateRoom)
				rooms.DELETE("/:roomId", roomHandler.DeleteRoom)
				rooms.POST("/:roomId/join", roomHandler.JoinRoom)
				rooms.POST("/:roomId/leave", roomHandler.LeaveRoom)
				rooms.POST("/:roomId/leave-as-host", roomHandler.LeaveRoom) // 兼容房主退出
				rooms.GET("/:roomId/members", roomHandler.GetRoomMembers)
				rooms.POST("/:roomId/transfer", roomHandler.TransferOwner)
				rooms.POST("/:roomId/kick", roomHandler.KickMember)
				rooms.PUT("/:roomId/status", roomHandler.UpdateMemberStatus) // 状态更新
			}

			// 好友相关路由 (兼容前端 /api/friends)
			friends := protected.Group("/friends")
			{
				friends.GET("", friendHandler.GetFriendList)                 // 兼容前端 GET /api/friends
				friends.POST("/requests", friendHandler.SendFriendRequest)   // 兼容前端 POST /api/friends/requests
				friends.GET("/requests", friendHandler.GetFriendRequests)    // 兼容前端 GET /api/friends/requests
				friends.PUT("/requests", friendHandler.ProcessFriendRequest) // 兼容前端 PUT /api/friends/requests
				friends.POST("/delete", friendHandler.RemoveFriend)          // 兼容前端 POST /api/friends/delete
			}

			// 专注相关路由 (focus)
			focus := protected.Group("/focus")
			{
				focus.POST("", focusHandler.StartFocus) // 兼容前端 POST /api/focus (开始专注)
				focus.POST("/start", focusHandler.StartFocus)
				focus.POST("/stop", focusHandler.StopFocus)
				focus.GET("/records", focusHandler.GetFocusRecords)
				focus.GET("/report/daily", studyReportHandler.GetDailyReport) // 切换到 AI 报告
				focus.GET("/report/weekly", focusHandler.GetWeeklyReport)
				focus.GET("/report/monthly", focusHandler.GetMonthlyReport)
			}

			// 学习报告路由
			reports := protected.Group("/reports")
			{
				reports.GET("/daily", studyReportHandler.GetDailyReport)
				reports.POST("/daily/regenerate", studyReportHandler.RegenerateDailyReport)
			}

			// 背景音乐路由 (background-music)
			music := protected.Group("/background-music")
			{
				music.GET("/list", musicHandler.GetMusicList)
				music.GET("/search", musicHandler.SearchMusic)
				music.GET("/:musicId", musicHandler.GetMusicByID)
			}

			// 系统配置路由 (system)
			system := protected.Group("/system")
			{
				system.GET("/configs", configHandler.GetConfigs)
				system.GET("/config/:key", configHandler.GetConfig)
			}

			// 学习教练路由
			coach := protected.Group("/coach")
			{
				coach.POST("/chat/stream", coachHandler.ChatStream)
				coach.GET("/sessions", coachHandler.GetSessions)
				coach.POST("/sessions", coachHandler.CreateSession)
				coach.PUT("/sessions/:id", coachHandler.UpdateSessionTitle)
				coach.GET("/sessions/:id/history", coachHandler.GetHistory)
				coach.GET("/profile", coachHandler.GetUserProfile)
				coach.PUT("/profile", coachHandler.UpdateUserProfile)
				coach.POST("/knowledge/upload", coachHandler.UploadKnowledge)
				coach.GET("/knowledge/list", coachHandler.ListKnowledge)
				coach.GET("/knowledge/preview", coachHandler.GetKnowledgePreview)
				coach.DELETE("/knowledge/:id", coachHandler.DeleteKnowledge)
				coach.PUT("/knowledge/:id/rename", coachHandler.RenameKnowledge)
				coach.POST("/knowledge/move", coachHandler.MoveKnowledge)
				coach.GET("/folders", coachHandler.ListFolders)
				coach.POST("/folders", coachHandler.CreateFolder)
				coach.DELETE("/folders/:id", coachHandler.DeleteFolder)
			}
		}
	}

	// 健康检查
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, map[string]string{
			"status": "ok",
		})
	})
}
