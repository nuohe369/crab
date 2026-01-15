package jwt

import (
	"errors"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

// Claims represents JWT payload, stores only ID and Platform
type Claims struct {
	ID   int64  `json:"id"`   // User ID
	Plat string `json:"plat"` // Platform: admin/frontend
	jwt.RegisteredClaims
}

// Config represents JWT configuration
type Config struct {
	Secret string `toml:"secret"`
	Expire string `toml:"expire"` // e.g. "24h"
}

// GetExpire parses expiration duration
func (c Config) GetExpire() time.Duration {
	d, _ := time.ParseDuration(c.Expire)
	if d == 0 {
		d = 24 * time.Hour
	}
	return d
}

// Manager manages JWT operations
type Manager struct {
	secret []byte
	expire time.Duration
}

var defaultMgr *Manager

// Init initializes default manager
func Init(cfg Config) {
	defaultMgr = New(cfg)
}

// Get returns default manager
func Get() *Manager {
	return defaultMgr
}

// MustInit initializes and panics on error
func MustInit(cfg Config) {
	if cfg.Secret == "" {
		log.Fatal("jwt secret cannot be empty")
	}
	Init(cfg)
}

// New creates JWT manager
func New(cfg Config) *Manager {
	return &Manager{
		secret: []byte(cfg.Secret),
		expire: cfg.GetExpire(),
	}
}

// Generate generates token
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

// Parse parses token
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

// Refresh refreshes token
func (m *Manager) Refresh(tokenStr string) (string, error) {
	claims, err := m.Parse(tokenStr)
	if err != nil && !errors.Is(err, ErrExpiredToken) {
		return "", err
	}
	return m.Generate(claims.ID, claims.Plat)
}
