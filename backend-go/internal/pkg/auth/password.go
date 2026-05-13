package auth

import (
	"github.com/tomato/backend/internal/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// PasswordManager 密码管理器
type PasswordManager struct {
	cost int
}

// NewPasswordManager 创建新的密码管理器
func NewPasswordManager() *PasswordManager {
	return &PasswordManager{
		cost: bcrypt.DefaultCost,
	}
}

// HashPassword 对密码进行哈希处理
func (pm *PasswordManager) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), pm.cost)
	if err != nil {
		return "", errors.NewWithError(
			errors.CodeInternalError,
			"密码加密失败",
			err,
		)
	}
	return string(hash), nil
}

// VerifyPassword 验证密码
func (pm *PasswordManager) VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// ValidatePassword 验证密码强度
func (pm *PasswordManager) ValidatePassword(password string) error {
	if len(password) < 6 {
		return errors.New(
			errors.CodeValidationError,
			"密码长度不能少于6个字符",
		)
	}

	if len(password) > 128 {
		return errors.New(
			errors.CodeValidationError,
			"密码长度不能超过128个字符",
		)
	}

	return nil
}
