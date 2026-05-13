package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tomato/backend/internal/pkg/errors"
)

// CustomClaims JWT自定义声明
type CustomClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// TokenManager JWT令牌管理器
type TokenManager struct {
	secret     string
	expiration int64
}

// NewTokenManager 创建新的令牌管理器
func NewTokenManager(secret string, expiration int64) *TokenManager {
	return &TokenManager{
		secret:     secret,
		expiration: expiration,
	}
}

// GenerateToken 生成JWT令牌
func (tm *TokenManager) GenerateToken(userID int64, username string) (string, error) {
	now := time.Now()
	expirationTime := now.Add(time.Duration(tm.expiration) * time.Second)

	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    "tomato-study-room",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(tm.secret))
	if err != nil {
		return "", errors.NewWithError(
			errors.CodeAuthError,
			"生成token失败",
			err,
		)
	}

	return tokenString, nil
}

// ValidateToken 验证并解析JWT令牌
func (tm *TokenManager) ValidateToken(tokenString string) (*CustomClaims, error) {
	claims := &CustomClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(tm.secret), nil
	})

	if err != nil {
		return nil, errors.NewWithError(
			errors.CodeInvalidToken,
			"token验证失败",
			err,
		)
	}

	if !token.Valid {
		return nil, errors.New(
			errors.CodeInvalidToken,
			"token无效",
		)
	}

	// 检查token是否过期
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, errors.New(
			errors.CodeTokenExpired,
			"token已过期",
		)
	}

	return claims, nil
}

// RefreshToken 刷新令牌（可选实现）
func (tm *TokenManager) RefreshToken(oldToken string) (string, error) {
	claims, err := tm.ValidateToken(oldToken)
	if err != nil {
		return "", err
	}

	return tm.GenerateToken(claims.UserID, claims.Username)
}

// ExtractClaims 从token字符串中提取声明
func (tm *TokenManager) ExtractClaims(tokenString string) (*CustomClaims, error) {
	return tm.ValidateToken(tokenString)
}
