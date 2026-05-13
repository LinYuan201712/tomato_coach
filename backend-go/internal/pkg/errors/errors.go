package errors

import (
	"fmt"
)

// ErrorCode 错误码类型
type ErrorCode int

// 业务错误码定义
const (
	// 成功
	CodeSuccess ErrorCode = 0

	// 认证错误 (1000-1099)
	CodeAuthError    ErrorCode = 1000
	CodeInvalidToken ErrorCode = 1001
	CodeTokenExpired ErrorCode = 1002
	CodeUnauthorized ErrorCode = 1003
	CodeForbidden    ErrorCode = 1004

	// 用户错误 (1100-1199)
	CodeUserNotFound    ErrorCode = 1100
	CodeUsernameExists  ErrorCode = 1101
	CodeEmailExists     ErrorCode = 1102
	CodePhoneExists     ErrorCode = 1103
	CodeInvalidPassword ErrorCode = 1104
	CodeInvalidUser     ErrorCode = 1105

	// 任务错误 (1200-1299)
	CodeTaskNotFound ErrorCode = 1200
	CodeTaskExists   ErrorCode = 1201
	CodeInvalidTask  ErrorCode = 1202

	// 房间错误 (1300-1399)
	CodeRoomNotFound  ErrorCode = 1300
	CodeRoomExists    ErrorCode = 1301
	CodeRoomFull      ErrorCode = 1302
	CodeInvalidRoom   ErrorCode = 1303
	CodeNotRoomOwner  ErrorCode = 1304
	CodeAlreadyInRoom ErrorCode = 1305
	CodeNotInRoom     ErrorCode = 1306

	// 好友错误 (1400-1499)
	CodeFriendNotFound        ErrorCode = 1400
	CodeAlreadyFriend         ErrorCode = 1401
	CodeFriendRequestExists   ErrorCode = 1402
	CodeFriendRequestNotFound ErrorCode = 1403
	CodeInvalidFriendRequest  ErrorCode = 1404

	// 验证错误 (1500-1599)
	CodeValidationError ErrorCode = 1500
	CodeInvalidRequest  ErrorCode = 1501
	CodeMissingParam    ErrorCode = 1502

	// 业务逻辑错误 (1600-1699)
	CodeBusinessError   ErrorCode = 1600
	CodeOperationFailed ErrorCode = 1601

	// 系统错误 (5000-5999)
	CodeInternalError ErrorCode = 5000
	CodeDatabaseError ErrorCode = 5001
	CodeNotFound      ErrorCode = 5002
)

// BusinessError 业务错误结构体
type BusinessError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
	err     error     `json:"-"`
}

// Error 实现error接口
func (e *BusinessError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("code=%d, message=%s, details=%s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("code=%d, message=%s", e.Code, e.Message)
}

// Unwrap 用于错误链
func (e *BusinessError) Unwrap() error {
	return e.err
}

// New 创建新的业务错误
func New(code ErrorCode, message string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
	}
}

// NewWithDetails 创建包含详情的业务错误
func NewWithDetails(code ErrorCode, message, details string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// NewWithError 从error创建业务错误
func NewWithError(code ErrorCode, message string, err error) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
		err:     err,
	}
}

// GetErrorMessage 根据错误码获取错误信息
func GetErrorMessage(code ErrorCode) string {
	messages := map[ErrorCode]string{
		CodeSuccess:               "操作成功",
		CodeAuthError:             "认证错误",
		CodeInvalidToken:          "Token无效",
		CodeTokenExpired:          "Token已过期",
		CodeUnauthorized:          "未授权",
		CodeForbidden:             "禁止访问",
		CodeUserNotFound:          "用户不存在",
		CodeUsernameExists:        "用户名已存在",
		CodeEmailExists:           "邮箱已存在",
		CodePhoneExists:           "手机号已存在",
		CodeInvalidPassword:       "密码错误",
		CodeInvalidUser:           "无效的用户",
		CodeTaskNotFound:          "任务不存在",
		CodeTaskExists:            "任务已存在",
		CodeInvalidTask:           "无效的任务",
		CodeRoomNotFound:          "房间不存在",
		CodeRoomExists:            "房间已存在",
		CodeRoomFull:              "房间已满",
		CodeInvalidRoom:           "无效的房间",
		CodeNotRoomOwner:          "不是房主",
		CodeAlreadyInRoom:         "已在房间中",
		CodeNotInRoom:             "不在房间中",
		CodeFriendNotFound:        "好友不存在",
		CodeAlreadyFriend:         "已是好友",
		CodeFriendRequestExists:   "好友申请已存在",
		CodeFriendRequestNotFound: "好友申请不存在",
		CodeInvalidFriendRequest:  "无效的好友申请",
		CodeValidationError:       "验证失败",
		CodeInvalidRequest:        "无效的请求",
		CodeMissingParam:          "缺少必要参数",
		CodeBusinessError:         "业务错误",
		CodeOperationFailed:       "操作失败",
		CodeInternalError:         "服务器内部错误",
		CodeDatabaseError:         "数据库错误",
		CodeNotFound:              "资源不存在",
	}

	if msg, ok := messages[code]; ok {
		return msg
	}
	return "未知错误"
}

// IsBusinessError 判断是否是业务错误
func IsBusinessError(err error) bool {
	_, ok := err.(*BusinessError)
	return ok
}

// AsBusinessError 转换为业务错误
func AsBusinessError(err error) *BusinessError {
	if be, ok := err.(*BusinessError); ok {
		return be
	}
	return nil
}
