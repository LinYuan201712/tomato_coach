package model

import "time"

// ========== 通用模型 ==========

// PageRequest 分页请求
type PageRequest struct {
	PageNum  int `json:"page_num" form:"page_num" binding:"omitempty,min=1"`
	PageSize int `json:"page_size" form:"page_size" binding:"omitempty,min=1,max=100"`
}

// PageResponse 分页响应
type PageResponse struct {
	Total    int64       `json:"total"`
	PageNum  int         `json:"page_num"`
	PageSize int         `json:"page_size"`
	Data     interface{} `json:"data"`
}

// ========== 认证相关 ==========

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username         string `json:"username" binding:"required,min=3,max=50"`
	Email            string `json:"email" binding:"required,email"`
	Phone            string `json:"phone" binding:"omitempty"`
	Password         string `json:"password" binding:"required,min=6,max=128"`
	VerificationCode string `json:"verificationCode" binding:"required,len=6,numeric"`
}

// SendVerificationCodeRequest 鍙戦€侀獙璇佺爜璇锋眰
type SendVerificationCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	Token     string `json:"token"`
	TokenType string `json:"token_type"`
	UserID    int64  `json:"user_id,string"`
	Username  string `json:"username"`
}

// ========== 用户相关 ==========

// UserResponse 用户信息响应
type UserResponse struct {
	UserID   int64      `json:"user_id,string"`
	Username string     `json:"username"`
	Status   string     `json:"status"`
	Sex      string     `json:"sex"`
	Birthday *time.Time `json:"birthday"`
	Tomato   int        `json:"tomato"`
	Province string     `json:"province"`
	Avatar   string     `json:"avatar"`
}

// UserInfoResponse 用户详细信息响应（当前用户）
type UserInfoResponse struct {
	UserID        int64      `json:"user_id,string"`
	Username      string     `json:"username"`
	Status        string     `json:"status"`
	Email         string     `json:"email"`
	Phone         string     `json:"phone"`
	Sex           string     `json:"sex"`
	Birthday      *time.Time `json:"birthday"`
	Tomato        int        `json:"tomato"`
	Province      string     `json:"province"`
	Avatar        string     `json:"avatar"`
	CurrentRoomID int64      `json:"current_room_id,string"`
}

// PublicUserResponse 公开用户信息响应
type PublicUserResponse struct {
	UserID   int64      `json:"user_id,string"`
	Username string     `json:"username"`
	Status   string     `json:"status"`
	Avatar   string     `json:"avatar"`
	Tomato   int        `json:"tomato"`
	Sex      string     `json:"sex"`
	Birthday *time.Time `json:"birthday"`
	Province string     `json:"province"`
}

// UpdateUserRequest 更新用户信息请求
type UpdateUserRequest struct {
	Username string     `json:"username" binding:"omitempty,min=3,max=50"`
	Password string     `json:"password" binding:"omitempty,min=6,max=128"`
	Sex      string     `json:"sex" binding:"omitempty"`
	Birthday *time.Time `json:"birthday" binding:"omitempty"`
	Province string     `json:"province" binding:"omitempty"`
}

// UserPrivacyResponse 用户隐私设置响应
type UserPrivacyResponse struct {
	ShowBirthday       string `json:"show_birthday"`
	ShowStudyTime      string `json:"show_study_time"`
	ShowLocation       string `json:"show_location"`
	AllowFriendRequest bool   `json:"allow_friend_request"`
	Searchable         bool   `json:"searchable"`
}

// UpdateUserPrivacyRequest 更新用户隐私设置请求
type UpdateUserPrivacyRequest struct {
	ShowBirthday       string `json:"show_birthday" binding:"omitempty,oneof=public friends private"`
	ShowStudyTime      string `json:"show_study_time" binding:"omitempty,oneof=public friends private"`
	ShowLocation       string `json:"show_location" binding:"omitempty,oneof=public friends private"`
	AllowFriendRequest bool   `json:"allow_friend_request" binding:"omitempty"`
	Searchable         bool   `json:"searchable" binding:"omitempty"`
}

// CurrencyResponse 用户货币信息响应
type CurrencyResponse struct {
	UserID            int64  `json:"user_id,string"`
	Coins             int    `json:"coins"`
	Tomato            int    `json:"tomato"`
	CheckDay          int    `json:"check_day"`           // 累计签到天数
	MonthCheckDays    int    `json:"month_check_days"`    // 本月签到天数
	HasCheckedInToday bool   `json:"has_checked_in_today"`
	UpdatedAt         string `json:"updated_at"`
}

// ========== 任务相关 ==========

// TaskCreateRequest 创建任务请求
type TaskCreateRequest struct {
	TaskName string `json:"task_name" binding:"required,max=20"`
	TaskNote string `json:"task_note" binding:"omitempty,max=200"`
	Duration int    `json:"duration" binding:"required,gt=0"`
}

// TaskUpdateRequest 更新任务请求
type TaskUpdateRequest struct {
	TaskID     int64  `json:"task_id,string"`
	TaskId     int64  `json:"taskId,string"` // 兼容前端 camelCase
	TaskName   string `json:"task_name" binding:"omitempty,max=20"`
	TaskNote   string `json:"task_note" binding:"omitempty,max=200"`
	Duration   int    `json:"duration" binding:"omitempty,gt=0"`
	Status     string `json:"status" binding:"omitempty"`
	TaskStatus string `json:"taskStatus" binding:"omitempty"` // 兼容前端
}

// TaskDeleteRequest 删除任务请求
type TaskDeleteRequest struct {
	TaskID int64 `json:"task_id,string"`
	TaskId int64 `json:"taskId,string"` // 兼容前端 camelCase
}

// TaskResponse 任务响应
type TaskResponse struct {
	TaskID         int64      `json:"task_id,string"`
	UserID         int64      `json:"user_id,string"`
	TaskName       string     `json:"task_name"`
	TaskNote       string     `json:"task_note"`
	Duration       int        `json:"duration"`
	ActualDuration int        `json:"actual_duration"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	StartTime      *time.Time `json:"start_time"`
	EndTime        *time.Time `json:"end_time"`
}

// ========== 专注相关 ==========

// StartFocusRequest 开始专注请求
type StartFocusRequest struct {
	TaskID      int64  `json:"task_id,string" form:"task_id"`
	TaskId      int64  `json:"taskId,string" form:"taskId"` // 兼容前端
	TaskName    string `json:"task_name" form:"task_name"`
	TaskName2   string `json:"taskName" form:"taskName"` // 兼容前端
	RoomID      int64  `json:"room_id,string" form:"room_id"`
	RoomId      int64  `json:"roomId,string" form:"roomId"` // 兼容前端
	SessionType string `json:"session_type" form:"session_type"`
	Duration    int    `json:"duration" form:"duration"`
}

// FocusResponse 专注响应
type FocusResponse struct {
	SessionID   int64     `json:"session_id,string"`
	UserID      int64     `json:"user_id,string"`
	SessionType string    `json:"session_type"`
	Duration    int64     `json:"duration"`
	StartTime   time.Time `json:"start_time"`
	Status      string    `json:"status"`
}

// StopFocusResponse 结束专注响应
type StopFocusResponse struct {
	SessionID      int64     `json:"session_id,string"`
	ActualDuration int64     `json:"actual_duration"`
	EndTime        time.Time `json:"end_time"`
	Status         string    `json:"status"`
}

// FocusRecordResponse 专注记录响应
type FocusRecordResponse struct {
	SessionID   int64      `json:"session_id,string"`
	SessionType string     `json:"session_type"`
	Duration    int64      `json:"duration"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	Status      string     `json:"status"`
}

// StudyReportResponse 学习报告响应
type StudyReportResponse struct {
	UserID           int64            `json:"user_id,string"`
	ReportType       string           `json:"report_type"`
	ReportDate       time.Time        `json:"report_date"`
	TotalDuration    int64            `json:"total_duration"`
	SessionCount     int64            `json:"session_count"`
	CompletedTasks   int64            `json:"completed_tasks"`
	AverageDuration  int64            `json:"average_duration"`
	SessionBreakdown map[string]int64 `json:"session_breakdown"`
	Content          string           `json:"content"` // AI 生成的报告内容
}

// ========== 房间相关 ==========

// RoomCreateRequest 创建房间请求
type RoomCreateRequest struct {
	RoomName    string `json:"room_name" binding:"omitempty,max=100"`
	RoomName2   string `json:"roomName" binding:"omitempty"` // 兼容前端
	MaxMembers  int    `json:"max_members" binding:"omitempty,gt=0"`
	MaxMembers2 int    `json:"maxMembers" binding:"omitempty"` // 兼容前端
	EndTime     *int64 `json:"end_time" binding:"omitempty"`
	MusicName   string `json:"music_name" binding:"omitempty"`
	MusicName2  string `json:"musicName" binding:"omitempty"` // 兼容前端
}

// RoomUpdateRequest 更新房间请求
type RoomUpdateRequest struct {
	RoomName    string `json:"room_name" binding:"omitempty,max=100"`
	RoomName2   string `json:"roomName" binding:"omitempty"` // 兼容前端
	MaxMembers  int    `json:"max_members" binding:"omitempty,gt=0"`
	MaxMembers2 int    `json:"maxMembers" binding:"omitempty"` // 兼容前端
	EndTime     *int64 `json:"end_time" binding:"omitempty"`
	MusicName   string `json:"music_name" binding:"omitempty"`
	MusicName2  string `json:"musicName" binding:"omitempty"` // 兼容前端
}

// RoomResponseDTO 房间响应
type RoomResponseDTO struct {
	ID             int64     `json:"id,string"`
	RoomID         int64     `json:"room_id,string"`
	RoomName       string    `json:"room_name"`
	CreatePerson   int64     `json:"create_person,string"`
	MaxMembers     int       `json:"max_members"`
	EndTime        *int64    `json:"end_time"`
	MusicID        *int64    `json:"music_id,string"`
	MusicName      string    `json:"music_name"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	CurrentMembers int       `json:"current_members"`
}

// RoomMemberResponse 房间成员响应
type RoomMemberResponse struct {
	UserID               int64      `json:"user_id,string"`
	Username             string     `json:"username"`
	Role                 string     `json:"role"`
	Status               string     `json:"status"`
	SessionFocusDuration int        `json:"session_focus_duration"`
	JoinedAt             *time.Time `json:"joined_at"`
	FocusStartTime       *time.Time `json:"focus_start_time"` // 兼容前端计算专注时间
}

// RoomMemberStatusUpdateRequest 房间成员状态更新请求
type RoomMemberStatusUpdateRequest struct {
	UserID         int64  `json:"userId,string"`
	Status         string `json:"status"`
	IsFocusing     bool   `json:"isFocusing"`
	FocusStartTime string `json:"focusStartTime"`
}

// ========== 好友相关 ==========

// FriendRequestRequest 好友申请请求
type FriendRequestRequest struct {
	UserName string `json:"user_name" binding:"required"`
	Message  string `json:"message" binding:"omitempty,max=255"`
}

// FriendRequestResponse 好友申请响应
type FriendRequestResponse struct {
	ID           int64  `json:"id,string"`
	FromUserID   int64  `json:"from_user_id,string"`
	FromUserName string `json:"from_user_name"`
	ToUserID     int64  `json:"to_user_id,string"`
	ToUserName   string `json:"to_user_name"`
	Status       string `json:"status"`
	Message      string `json:"message"`
}

// ProcessFriendRequestRequest 处理好友申请请求
type ProcessFriendRequestRequest struct {
	FromUserID int64  `json:"from_user_id,string"`
	Action     string `json:"action" binding:"required,oneof=accept reject"`
}

// FriendResponse 好友响应
type FriendResponse struct {
	FriendID     int64  `json:"friend_id,string"`
	FriendName   string `json:"friend_name"`
	FriendStatus string `json:"friend_status"`
}

// DeleteFriendRequest 删除好友请求
type DeleteFriendRequest struct {
	FriendName string `json:"friend_name" binding:"required"`
}

// ========== 背景音乐相关 ==========

// BackgroundMusicResponse 背景音乐响应
type BackgroundMusicResponse struct {
	ID        int64     `json:"id,string"`
	MusicName string    `json:"music_name"`
	AudioURL  string    `json:"audio_url"`
	Price     float64   `json:"price"`
	IsFree    bool      `json:"is_free"`
	Duration  *int      `json:"duration"`
	CreatedAt time.Time `json:"created_at"`
}
