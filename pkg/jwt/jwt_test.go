package jwt

import (
	"testing"
	"time"
)

func TestJWTGenerateAndParse(t *testing.T) {
	mgr := New(Config{
		Secret: "test-secret-key",
		Expire: "1h",
	})

	tests := []struct {
		name   string
		userID int64
		plat   string
	}{
		{"admin user", 1, "admin"},
		{"frontend user", 12345, "frontend"},
		{"zero id", 0, "api"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := mgr.Generate(tt.userID, tt.plat)
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			if token == "" {
				t.Error("Generated token is empty")
			}

			claims, err := mgr.Parse(token)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if claims.ID != tt.userID {
				t.Errorf("ID mismatch: got %d, want %d", claims.ID, tt.userID)
			}

			if claims.Plat != tt.plat {
				t.Errorf("Plat mismatch: got %s, want %s", claims.Plat, tt.plat)
			}
		})
	}
}

func TestJWTExpired(t *testing.T) {
	mgr := New(Config{
		Secret: "test-secret",
		Expire: "1ms", // Very short expiration
	})

	token, err := mgr.Generate(1, "test")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	_, err = mgr.Parse(token)
	if err != ErrExpiredToken {
		t.Errorf("Expected ErrExpiredToken, got %v", err)
	}
}

func TestJWTInvalidToken(t *testing.T) {
	mgr := New(Config{
		Secret: "test-secret",
		Expire: "1h",
	})

	tests := []struct {
		name  string
		token string
	}{
		{"empty", ""},
		{"garbage", "not-a-valid-token"},
		{"malformed", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mgr.Parse(tt.token)
			if err == nil {
				t.Error("Expected error for invalid token")
			}
		})
	}
}

func TestJWTWrongSecret(t *testing.T) {
	mgr1 := New(Config{Secret: "secret-1", Expire: "1h"})
	mgr2 := New(Config{Secret: "secret-2", Expire: "1h"})

	token, _ := mgr1.Generate(1, "test")

	_, err := mgr2.Parse(token)
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}
}

func TestJWTRefresh(t *testing.T) {
	mgr := New(Config{
		Secret: "test-secret",
		Expire: "1h",
	})

	originalToken, _ := mgr.Generate(123, "admin")

	// Wait a bit to ensure different IssuedAt
	time.Sleep(10 * time.Millisecond)

	newToken, err := mgr.Refresh(originalToken)
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}

	// Tokens might be same if generated in same second, that's OK
	// The important thing is the new token is valid
	claims, err := mgr.Parse(newToken)
	if err != nil {
		t.Fatalf("Parse refreshed token failed: %v", err)
	}

	if claims.ID != 123 || claims.Plat != "admin" {
		t.Error("Refreshed token should preserve claims")
	}
}

func TestConfigGetExpire(t *testing.T) {
	tests := []struct {
		expire   string
		expected time.Duration
	}{
		{"24h", 24 * time.Hour},
		{"1h", time.Hour},
		{"30m", 30 * time.Minute},
		{"", 24 * time.Hour},        // default
		{"invalid", 24 * time.Hour}, // default on parse error
	}

	for _, tt := range tests {
		t.Run(tt.expire, func(t *testing.T) {
			cfg := Config{Expire: tt.expire}
			if got := cfg.GetExpire(); got != tt.expected {
				t.Errorf("GetExpire() = %v, want %v", got, tt.expected)
			}
		})
	}
}
