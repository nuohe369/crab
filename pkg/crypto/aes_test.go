package crypto

import (
	"strings"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
		key       string
	}{
		{"simple", "hello world", "my-secret-key"},
		{"empty", "", "key"},
		{"chinese", "你好世界", "密钥"},
		{"special chars", "!@#$%^&*()_+-=[]{}|;':\",./<>?", "key123"},
		{"long text", strings.Repeat("a", 1000), "long-key-test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := Encrypt(tt.plaintext, tt.key)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			if !IsEncrypted(encrypted) {
				t.Errorf("IsEncrypted returned false for encrypted string")
			}

			decrypted, err := Decrypt(encrypted, tt.key)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			if decrypted != tt.plaintext {
				t.Errorf("Decrypt mismatch: got %q, want %q", decrypted, tt.plaintext)
			}
		})
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	plaintext := "secret message"
	key := "correct-key"
	wrongKey := "wrong-key"

	encrypted, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	_, err = Decrypt(encrypted, wrongKey)
	if err == nil {
		t.Error("Decrypt with wrong key should fail")
	}
}

func TestDecryptPlaintext(t *testing.T) {
	plaintext := "not encrypted"
	key := "any-key"

	result, err := Decrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Decrypt plaintext failed: %v", err)
	}

	if result != plaintext {
		t.Errorf("Decrypt plaintext mismatch: got %q, want %q", result, plaintext)
	}
}

func TestIsEncrypted(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"ENC(abc123)", true},
		{"ENC()", true},
		{"ENC(test", false},
		{"abc)", false},
		{"plaintext", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := IsEncrypted(tt.input); got != tt.expected {
				t.Errorf("IsEncrypted(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestEncryptProducesDifferentCiphertext(t *testing.T) {
	plaintext := "same message"
	key := "same-key"

	encrypted1, _ := Encrypt(plaintext, key)
	encrypted2, _ := Encrypt(plaintext, key)

	if encrypted1 == encrypted2 {
		t.Error("Same plaintext should produce different ciphertext (due to random nonce)")
	}

	// But both should decrypt to same plaintext
	decrypted1, _ := Decrypt(encrypted1, key)
	decrypted2, _ := Decrypt(encrypted2, key)

	if decrypted1 != decrypted2 {
		t.Error("Both ciphertexts should decrypt to same plaintext")
	}
}
