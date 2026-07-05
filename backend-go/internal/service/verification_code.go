package service

import (
	"context"
	"crypto/rand"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/tomato/backend/internal/domain/model"
	"github.com/tomato/backend/internal/pkg/errors"
	"go.uber.org/zap"
)

const (
	verificationCodeTTL      = 10 * time.Minute
	verificationCodeCooldown = 60 * time.Second
)

type emailVerificationCode struct {
	Code      string
	ExpiresAt time.Time
	SentAt    time.Time
	Pending   bool
}

var verificationCodes = struct {
	sync.Mutex
	items map[string]emailVerificationCode
}{
	items: make(map[string]emailVerificationCode),
}

func (s *authService) SendVerificationCode(ctx context.Context, req *model.SendVerificationCodeRequest) error {
	email := normalizeEmail(req.Email)
	if email == "" {
		return errors.New(errors.CodeValidationError, "邮箱不能为空")
	}

	if user, _ := s.userRepo.FindByEmail(ctx, email); user != nil {
		return errors.New(errors.CodeEmailExists, "邮箱已被注册")
	}

	now := time.Now()
	verificationCodes.Lock()
	pruneExpiredVerificationCodesLocked(now)
	if existing, ok := verificationCodes.items[email]; ok && now.Sub(existing.SentAt) < verificationCodeCooldown {
		verificationCodes.Unlock()
		return errors.New(errors.CodeOperationFailed, "验证码发送过于频繁，请稍后再试")
	}
	verificationCodes.items[email] = emailVerificationCode{
		ExpiresAt: now.Add(verificationCodeTTL),
		SentAt:    now,
		Pending:   true,
	}
	verificationCodes.Unlock()

	code, err := generateNumericCode(6)
	if err != nil {
		rollbackPendingVerificationCode(email, now)
		s.logger.Error("生成邮箱验证码失败", zap.Error(err))
		return errors.New(errors.CodeInternalError, "发送验证码失败")
	}

	if s.emailSender == nil {
		rollbackPendingVerificationCode(email, now)
		return errors.New(errors.CodeInternalError, "邮件服务未配置")
	}
	if err := s.emailSender.SendVerificationCode(ctx, email, code, verificationCodeTTL); err != nil {
		rollbackPendingVerificationCode(email, now)
		return err
	}

	verificationCodes.Lock()
	verificationCodes.items[email] = emailVerificationCode{
		Code:      code,
		ExpiresAt: now.Add(verificationCodeTTL),
		SentAt:    now,
	}
	verificationCodes.Unlock()

	s.logger.Info("邮箱验证码已发送",
		zap.String("email", email),
		zap.Duration("expires_in", verificationCodeTTL),
	)

	return nil
}

func (s *authService) verifyEmailCode(email, code string) error {
	email = normalizeEmail(email)
	code = strings.TrimSpace(code)
	if email == "" || code == "" {
		return errors.New(errors.CodeValidationError, "请输入邮箱验证码")
	}

	verificationCodes.Lock()
	defer verificationCodes.Unlock()

	item, ok := verificationCodes.items[email]
	if !ok {
		return errors.New(errors.CodeValidationError, "请先发送验证码")
	}
	if item.Pending {
		return errors.New(errors.CodeValidationError, "验证码正在发送中，请稍后再试")
	}
	if time.Now().After(item.ExpiresAt) {
		delete(verificationCodes.items, email)
		return errors.New(errors.CodeValidationError, "验证码已过期，请重新发送")
	}
	if item.Code != code {
		return errors.New(errors.CodeValidationError, "验证码错误")
	}

	delete(verificationCodes.items, email)
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func rollbackPendingVerificationCode(email string, sentAt time.Time) {
	verificationCodes.Lock()
	defer verificationCodes.Unlock()

	item, ok := verificationCodes.items[email]
	if ok && item.Pending && item.SentAt.Equal(sentAt) {
		delete(verificationCodes.items, email)
	}
}

func pruneExpiredVerificationCodesLocked(now time.Time) {
	for email, item := range verificationCodes.items {
		if now.After(item.ExpiresAt) {
			delete(verificationCodes.items, email)
		}
	}
}

func generateNumericCode(length int) (string, error) {
	if length <= 0 {
		return "", errors.New(errors.CodeInvalidRequest, "验证码长度无效")
	}

	max := big.NewInt(10)
	digits := make([]byte, length)
	for i := range digits {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		digits[i] = byte('0' + n.Int64())
	}

	return string(digits), nil
}
