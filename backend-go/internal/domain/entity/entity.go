package entity

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// BaseEntity 基础实体，包含公共字段
type BaseEntity struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// User 用户实体
type User struct {
	BaseEntity
	UserID       int64      `gorm:"column:user_id;uniqueIndex;not null" json:"user_id"`
	Username     string     `gorm:"column:username;uniqueIndex;size:50;not null" json:"username"`
	Status       string     `gorm:"column:status;size:10;default:'离线'" json:"status"`
	Email        string     `gorm:"column:email;uniqueIndex;size:100;not null" json:"email"`
	Phone        string     `gorm:"column:phone;uniqueIndex;size:20" json:"phone"`
	Sex          string     `gorm:"column:sex;size:10" json:"sex"`
	Birthday     *time.Time `gorm:"column:birthday;type:date" json:"birthday"`
	PasswordHash string     `gorm:"column:password_hash;size:255;not null" json:"password_hash"`
	Tomato       int        `gorm:"column:tomato;default:0" json:"tomato"`
	Province     string     `gorm:"column:province;size:50" json:"province"`
	Avatar       string     `gorm:"column:avatar;size:255" json:"avatar"`
	Goals          string     `gorm:"column:goals;type:text" json:"goals"`
	PreferredStyle string     `gorm:"column:preferred_style;size:255" json:"preferred_style"`
	ProfileLock    string     `gorm:"column:profile_lock;size:20;default:'soft'" json:"profile_lock"` // unlocked, soft, hard
	LockSuggestions string    `gorm:"column:lock_suggestions;type:text" json:"lock_suggestions"`      // JSON storage for pending facts
	Deleted        bool       `gorm:"column:deleted;default:false;index" json:"deleted"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// Task 任务实体
type Task struct {
	BaseEntity
	TaskID         int64      `gorm:"column:task_id;uniqueIndex;not null" json:"task_id"`
	UserID         int64      `gorm:"column:user_id;not null;index:idx_user_status" json:"user_id"`
	TaskName       string     `gorm:"column:task_name;size:20;not null" json:"task_name"`
	TaskNote       string     `gorm:"column:task_note;size:200" json:"task_note"`
	Duration       int        `gorm:"column:duration;not null" json:"duration"`
	ActualDuration int        `gorm:"column:actual_duration;default:0" json:"actual_duration"`
	Status         string     `gorm:"column:status;size:20;default:'未完成';index:idx_user_status" json:"status"`
	StartTime      *time.Time `gorm:"column:start_time" json:"start_time"`
	EndTime        *time.Time `gorm:"column:end_time" json:"end_time"`
	User           *User      `gorm:"foreignKey:UserID;references:ID" json:"-"`
}

// TableName 指定表名
func (Task) TableName() string {
	return "tasks"
}

// Room 自习室实体
type Room struct {
	BaseEntity
	RoomID       int64        `gorm:"column:room_id;uniqueIndex;not null" json:"room_id"`
	RoomName     string       `gorm:"column:room_name;size:100;not null" json:"room_name"`
	CreatePerson int64        `gorm:"column:create_person;not null;index" json:"create_person"`
	MaxMembers   int          `gorm:"column:max_members;default:20" json:"max_members"`
	EndTime      *int64       `gorm:"column:end_time" json:"end_time"`
	MusicID      *int64       `gorm:"column:music_id" json:"music_id"`
	MusicName    string       `gorm:"column:music_name;size:100" json:"music_name"`
	Creator      *User        `gorm:"foreignKey:CreatePerson;references:ID" json:"-"`
	Members      []RoomMember `gorm:"foreignKey:RoomID;references:RoomID" json:"-"`
}

// TableName 指定表名
func (Room) TableName() string {
	return "room"
}

// RoomMember 房间成员实体
type RoomMember struct {
	BaseEntity
	RoomID               int64      `gorm:"column:room_id;not null;uniqueIndex:uk_room_user" json:"room_id"`
	UserID               int64      `gorm:"column:user_id;not null;uniqueIndex:uk_room_user" json:"user_id"`
	Role                 string     `gorm:"column:role;size:10;default:'成员'" json:"role"`
	Status               string     `gorm:"column:status;size:20;default:'专注中'" json:"status"`
	JoinedAt             *time.Time `gorm:"column:joined_at;autoCreateTime" json:"joined_at"`
	SessionFocusDuration int        `gorm:"column:session_focus_duration;default:0" json:"session_focus_duration"`
	Room                 *Room      `gorm:"foreignKey:RoomID;references:RoomID" json:"-"`
	User                 *User      `gorm:"foreignKey:UserID;references:ID" json:"-"`
}

// TableName 指定表名
func (RoomMember) TableName() string {
	return "roommember"
}

// Friend 好友关系实体
type Friend struct {
	BaseEntity
	UserID       int64  `gorm:"column:user_id;not null;uniqueIndex:uk_user_friend" json:"user_id"`
	FriendID     int64  `gorm:"column:friend_id;not null;uniqueIndex:uk_user_friend" json:"friend_id"`
	FriendStatus string `gorm:"column:friend_status;size:20" json:"friend_status"`
	User         *User  `gorm:"foreignKey:UserID;references:ID" json:"-"`
	FriendUser   *User  `gorm:"foreignKey:FriendID;references:ID" json:"-"`
}

// TableName 指定表名
func (Friend) TableName() string {
	return "friend"
}

// FriendRequest 好友请求实体
type FriendRequest struct {
	BaseEntity
	FromUserID   int64  `gorm:"column:from_user_id;not null;index" json:"from_user_id"`
	FromUserName string `gorm:"column:from_user_name;size:255;not null" json:"from_user_name"`
	ToUserID     int64  `gorm:"column:to_user_id;not null;index" json:"to_user_id"`
	ToUserName   string `gorm:"column:to_user_name;size:255;not null" json:"to_user_name"`
	Status       string `gorm:"column:status;size:20;not null" json:"status"`
	Message      string `gorm:"column:message;size:255;not null" json:"message"`
	FromUser     *User  `gorm:"foreignKey:FromUserID;references:ID" json:"-"`
	ToUser       *User  `gorm:"foreignKey:ToUserID;references:ID" json:"-"`
}

// TableName 指定表名
func (FriendRequest) TableName() string {
	return "friendrequest"
}

// FocusSession 专注会话实体
type FocusSession struct {
	BaseEntity
	SessionID      int64      `gorm:"column:session_id;not null" json:"session_id"`
	UserID         int64      `gorm:"column:user_id;not null;index" json:"user_id"`
	RoomID         *int64     `gorm:"column:room_id" json:"room_id"`
	TaskID         *int64     `gorm:"column:task_id" json:"task_id"`
	SessionType    string     `gorm:"column:session_type;size:50;not null" json:"session_type"`
	Duration       int        `gorm:"column:duration;not null" json:"duration"`
	ActualDuration int        `gorm:"column:actual_duration;default:0" json:"actual_duration"`
	StartTime      time.Time  `gorm:"column:start_time;not null;index" json:"start_time"`
	EndTime        *time.Time `gorm:"column:end_time" json:"end_time"`
	Status         string     `gorm:"column:status;size:20;default:'进行中'" json:"status"`
	User           *User      `gorm:"foreignKey:UserID;references:ID" json:"-"`
	Room           *Room      `gorm:"foreignKey:RoomID;references:RoomID" json:"-"`
	Task           *Task      `gorm:"foreignKey:TaskID;references:TaskID" json:"-"`
}

// TableName 指定表名
func (FocusSession) TableName() string {
	return "focussession"
}

// BackgroundMusic 背景音乐实体
type BackgroundMusic struct {
	BaseEntity
	MusicName string  `gorm:"column:music_name;size:100;not null;index" json:"music_name"`
	AudioURL  string  `gorm:"column:audio_url;size:500;not null" json:"audio_url"`
	Price     float64 `gorm:"column:price;type:decimal(10,2);default:0" json:"price"`
	IsFree    bool    `gorm:"column:is_free;default:true" json:"is_free"`
	Duration  *int    `gorm:"column:duration" json:"duration"`
}

// TableName 指定表名
func (BackgroundMusic) TableName() string {
	return "backgroundmusic"
}

// StudyReport 学习报告实体
type StudyReport struct {
	BaseEntity
	ReportID         int64     `gorm:"column:report_id;not null" json:"report_id"`
	UserID           int64     `gorm:"column:user_id;not null;index;uniqueIndex:uk_user_date" json:"user_id"`
	ReportType       string    `gorm:"column:report_type;size:10;not null;uniqueIndex:uk_user_date" json:"report_type"`
	ReportDate       time.Time `gorm:"column:report_date;type:date;uniqueIndex:uk_user_date" json:"report_date"`
	TotalFocusTime   int       `gorm:"column:total_focus_time;default:0" json:"total_focus_time"`
	CompletedTasks   int       `gorm:"column:completed_tasks;default:0" json:"completed_tasks"`
	AvgDailyDuration float32   `gorm:"column:avg_daily_duration" json:"avg_daily_duration"`
	Content          string    `gorm:"column:content;type:text" json:"content"`       // AI 生成的报告内容
	MetaData         string    `gorm:"column:meta_data;type:text" json:"meta_data"` // 存储原始数据的快照 (JSON)
	User             *User     `gorm:"foreignKey:UserID;references:ID" json:"-"`
}

// TableName 指定表名
func (StudyReport) TableName() string {
	return "studyreport"
}

// UserPrivacy 用户隐私设置实体
type UserPrivacy struct {
	BaseEntity
	UserID             int64  `gorm:"column:user_id;not null;uniqueIndex" json:"user_id"`
	ShowBirthday       string `gorm:"column:show_birthday;size:20;default:'public'" json:"show_birthday"`
	ShowStudyTime      string `gorm:"column:show_study_time;size:20;default:'public'" json:"show_study_time"`
	ShowLocation       string `gorm:"column:show_location;size:20;default:'public'" json:"show_location"`
	AllowFriendRequest bool   `gorm:"column:allow_friend_request;default:true" json:"allow_friend_request"`
	Searchable         bool   `gorm:"column:searchable;default:true" json:"searchable"`
	User               *User  `gorm:"foreignKey:UserID;references:ID" json:"-"`
}

// TableName 指定表名
func (UserPrivacy) TableName() string {
	return "userprivacy"
}

// UserCurrency 用户货币实体
type UserCurrency struct {
	BaseEntity
	UserID   int64 `gorm:"column:user_id;not null;uniqueIndex" json:"user_id"`
	Coins    int   `gorm:"column:coins;default:0" json:"coins"`
	CheckDay int   `gorm:"column:check_day;default:0" json:"check_day"`
	User     *User `gorm:"foreignKey:UserID;references:ID" json:"-"`
}

// TableName 指定表名
func (UserCurrency) TableName() string {
	return "usercurrency"
}

// CheckinRecord 签到记录实体
type CheckinRecord struct {
	BaseEntity
	UserID      int64     `gorm:"column:user_id;not null;uniqueIndex:uk_user_checkin_date;index" json:"user_id"`
	CheckinDate time.Time `gorm:"column:checkin_date;type:date;uniqueIndex:uk_user_checkin_date" json:"checkin_date"`
	User        *User     `gorm:"foreignKey:UserID;references:ID" json:"-"`
}

// TableName 指定表名
func (CheckinRecord) TableName() string {
	return "checkin_records"
}

// SystemConfig 系统配置实体
type SystemConfig struct {
	BaseEntity
	ConfigKey   string `gorm:"column:config_key;size:100;uniqueIndex;not null" json:"config_key"`
	ConfigValue string `gorm:"column:config_value;size:500;not null" json:"config_value"`
	Description string `gorm:"column:description;size:255" json:"description"`
}

// TableName 指定表名
func (SystemConfig) TableName() string {
	return "system_config"
}

// StatusEnum 用户状态枚举
type StatusEnum string

const (
	StatusOnline   StatusEnum = "在线"
	StatusOffline  StatusEnum = "离线"
	StatusFocusing StatusEnum = "专注中"
)

// Value 实现driver.Valuer接口
func (s StatusEnum) Value() (driver.Value, error) {
	return string(s), nil
}

// Scan 实现sql.Scanner接口
func (s *StatusEnum) Scan(value interface{}) error {
	*s = StatusEnum(value.(string))
	return nil
}

// TaskStatusEnum 任务状态枚举
type TaskStatusEnum string

const (
	TaskStatusUnfinished TaskStatusEnum = "未完成"
	TaskStatusProcessing TaskStatusEnum = "进行中"
	TaskStatusCompleted  TaskStatusEnum = "已完成"
)

// FriendRequestStatusEnum 好友请求状态枚举
type FriendRequestStatusEnum string

const (
	FriendRequestStatusPending  FriendRequestStatusEnum = "待处理"
	FriendRequestStatusAccepted FriendRequestStatusEnum = "已同意"
	FriendRequestStatusRejected FriendRequestStatusEnum = "已拒绝"
)

// PrivacyLevelEnum 隐私级别枚举
type PrivacyLevelEnum string

const (
	PrivacyLevelPublic  PrivacyLevelEnum = "public"
	PrivacyLevelFriends PrivacyLevelEnum = "friends"
	PrivacyLevelPrivate PrivacyLevelEnum = "private"
)

// JSON 通用JSON字段
type JSON json.RawMessage

// Value 实现driver.Valuer接口
func (j JSON) Value() (driver.Value, error) {
	return string(j), nil
}

// Scan 实现sql.Scanner接口
func (j *JSON) Scan(value interface{}) error {
	*j = JSON(value.([]byte))
	return nil
}

// ChatSession 聊天会话实体
type ChatSession struct {
	BaseEntity
	SessionID string `gorm:"column:session_id;uniqueIndex;size:50;not null" json:"session_id"`
	UserID    int64  `gorm:"column:user_id;index;not null" json:"user_id"`
	Title     string `gorm:"column:title;size:255" json:"title"`
	Summary   string `gorm:"column:summary;type:text" json:"summary"`
	MsgCount  int    `gorm:"column:msg_count;default:0" json:"msg_count"`
}

func (ChatSession) TableName() string {
	return "chat_sessions"
}

// ChatMessage 聊天消息实体
type ChatMessage struct {
	BaseEntity
	SessionID string `gorm:"column:session_id;index;size:50;not null" json:"session_id"`
	UserID    int64  `gorm:"column:user_id;index;not null" json:"user_id"`
	Role      string `gorm:"column:role;size:20;not null" json:"role"`
	Content   string `gorm:"column:content;type:text;not null" json:"content"`
	Reasoning string `gorm:"column:reasoning;type:text" json:"reasoning"`
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}

// KnowledgeFolder 知识库文件夹实体
type KnowledgeFolder struct {
	BaseEntity
	UserID   int64  `gorm:"column:user_id;index;not null" json:"user_id"`
	Name     string `gorm:"column:name;size:255;not null" json:"name"`
	ParentID int64  `gorm:"column:parent_id;default:0" json:"parent_id"`
}

func (KnowledgeFolder) TableName() string {
	return "knowledge_folders"
}

// KnowledgeFile 知识库文件元数据实体
type KnowledgeFile struct {
	BaseEntity
	UserID      int64  `gorm:"column:user_id;index;not null" json:"user_id"`
	FolderID    int64  `gorm:"column:folder_id;default:0;index" json:"folder_id"`
	FileName    string `gorm:"column:file_name;size:255;not null" json:"file_name"`
	DisplayName string `gorm:"column:display_name;size:255;not null" json:"display_name"`
	FilePath    string `gorm:"column:file_path;size:500" json:"file_path"`
	FileSize    int64  `gorm:"column:file_size" json:"file_size"`
	DocID       string `gorm:"column:doc_id;size:100;index" json:"doc_id"` // 对应 RAG 中的 DocID
	Status      string `gorm:"column:status;size:20;default:'active'" json:"status"` // active, parsing, failed
}

func (KnowledgeFile) TableName() string {
	return "knowledge_files"
}
