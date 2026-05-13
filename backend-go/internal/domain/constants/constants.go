package constants

// 用户状态常量
const (
	UserStatusOnline   = "在线"
	UserStatusOffline  = "离线"
	UserStatusFocusing = "专注中"
)

// 任务状态常量
const (
	TaskStatusUnfinished = "未完成"
	TaskStatusProcessing = "进行中"
	TaskStatusCompleted  = "已完成"
)

// 房间成员角色常量
const (
	RoomRoleOwner  = "房主"
	RoomRoleMember = "成员"
)

// 房间成员状态常量
const (
	RoomMemberStatusFocusing = "专注中"
	RoomMemberStatusResting  = "休息中"
)

// 好友请求状态常量
const (
	FriendRequestStatusPending  = "待处理"
	FriendRequestStatusAccepted = "已同意"
	FriendRequestStatusRejected = "已拒绝"
)

// 好友状态常量
const (
	FriendStatusActive = "正常"
)

// 知识库文件状态常量
const (
	KnowledgeStatusActive  = "active"
	KnowledgeStatusParsing = "parsing"
	KnowledgeStatusFailed  = "failed"
)

// 隐私级别常量
const (
	PrivacyLevelPublic  = "public"
	PrivacyLevelFriends = "friends"
	PrivacyLevelPrivate = "private"
)

// ReportType 报告类型
const (
	ReportTypeDaily   = "daily"
	ReportTypeWeekly  = "weekly"
	ReportTypeMonthly = "monthly"
)

// 专注会话类型常量
const (
	SessionTypeFocus     = "专注学习"
	SessionTypeShortRest = "短休息"
	SessionTypeLongRest  = "长休息"
)

// 专注会话状态常量
const (
	SessionStatusProcessing = "进行中"
	SessionStatusCompleted  = "已完成"
	SessionStatusCancelled  = "已取消"
)


// Tomato奖励常量
const (
	TomatoRewardOnTaskComplete = 1 // 完成任务的番茄奖励数
)

// 默认值常量
const (
	DefaultMaxRoomMembers = 20
	DefaultCoinsReward    = 0 // 默认硬币奖励
)

// 业务常量
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// 错误信息常量
const (
	ErrMsgUsernameExists      = "用户名已存在"
	ErrMsgEmailExists         = "邮箱已存在"
	ErrMsgPhoneExists         = "手机号已存在"
	ErrMsgUserNotFound        = "用户不存在"
	ErrMsgInvalidPassword     = "密码错误"
	ErrMsgInvalidToken        = "token无效或过期"
	ErrMsgUnauthorized        = "未授权"
	ErrMsgForbidden           = "禁止访问"
	ErrMsgTaskNotFound        = "任务不存在"
	ErrMsgRoomNotFound        = "房间不存在"
	ErrMsgFriendNotFound      = "好友不存在"
	ErrMsgAlreadyFriend       = "已是好友"
	ErrMsgFriendRequestExists = "好友申请已存在"
	ErrMsgRoomMemberNotFound  = "房间成员不存在"
	ErrMsgRoomFull            = "房间已满"
	ErrMsgInvalidRequest      = "请求参数错误"
	ErrMsgInternalError       = "内部服务器错误"
)

// 分页常量
const (
	DefaultPageNum = 1
	MinPageNum     = 1
	MinPageSize    = 1
)
