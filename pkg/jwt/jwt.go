// Package jwt provides JSON Web Token generation and validation
// Package jwt 提供 JSON Web Token 生成和验证
package jwt

import (
	"errors"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token") // Invalid token error | 无效令牌错误
	ErrExpiredToken = errors.New("token expired") // Expired token error | 令牌过期错误
)

// Claims represents JWT payload, stores only ID and Platform
// Claims 表示 JWT 载荷，仅存储 ID 和平台
type Claims struct {
	ID   int64  `json:"id"`   // User ID | 用户 ID
	Plat string `json:"plat"` // Platform: admin/frontend | 平台：admin/frontend
	jwt.RegisteredClaims
}

// Config represents JWT configuration
// Config 表示 JWT 配置
type Config struct {
	Secret string `toml:"secret"` // Secret key | 密钥
	Expire string `toml:"expire"` // Expiration duration e.g. "24h" | 过期时间，例如 "24h"
}

// GetExpire parses expiration duration
// GetExpire 解析过期时间
func (c Config) GetExpire() time.Duration {
	d, _ := time.ParseDuration(c.Expire)
	if d == 0 {
		d = 24 * time.Hour
	}
	return d
}

// Manager manages JWT operations
// Manager 管理 JWT 操作
type Manager struct {
	secret []byte        // Secret key | 密钥
	expire time.Duration // Expiration duration | 过期时间
}

var defaultMgr *Manager // Default manager instance | 默认管理器实例

// Init initializes the default manager
// Init 初始化默认管理器
func Init(cfg Config) {
	defaultMgr = New(cfg)
}

// Get returns the default manager
// Get 返回默认管理器
func Get() *Manager {
	return defaultMgr
}

// MustInit initializes and panics on error
// MustInit 初始化，失败时 panic
func MustInit(cfg Config) {
	if cfg.Secret == "" {
		log.Fatal("jwt secret cannot be empty")
	}
	Init(cfg)
}

// New creates a JWT manager
// New 创建 JWT 管理器
func New(cfg Config) *Manager {
	return &Manager{
		secret: []byte(cfg.Secret),
		expire: cfg.GetExpire(),
	}
}

// Generate generates a JWT token
// Generate 生成 JWT 令牌
func (m *Manager) Generate(id int64, plat string) (string, error) {
	now := time.Now()
	claims := Claims{
		ID:   id,
		Plat: plat,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expire)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// Parse parses and validates a JWT token
// Parse 解析并验证 JWT 令牌
func (m *Manager) Parse(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		return m.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// Refresh refreshes an expired or valid token
// Refresh 刷新过期或有效的令牌
func (m *Manager) Refresh(tokenStr string) (string, error) {
	claims, err := m.Parse(tokenStr)
	if err != nil && !errors.Is(err, ErrExpiredToken) {
		return "", err
	}
	return m.Generate(claims.ID, claims.Plat)
}
