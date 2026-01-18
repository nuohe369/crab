// Package crypto provides encryption and decryption utilities using AES-GCM
// Package crypto 提供使用 AES-GCM 的加密和解密工具
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"strings"
)

const encPrefix = "ENC(" // Encrypted string prefix | 加密字符串前缀
const encSuffix = ")"    // Encrypted string suffix | 加密字符串后缀

// deriveKey derives 32-byte AES key from password
// deriveKey 从密码派生 32 字节 AES 密钥
func deriveKey(key string) []byte {
	hash := sha256.Sum256([]byte(key))
	return hash[:]
}

// Encrypt encrypts plaintext using AES-GCM
// Encrypt 使用 AES-GCM 加密明文
func Encrypt(plaintext, key string) (string, error) {
	block, err := aes.NewCipher(deriveKey(key))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return encPrefix + base64.StdEncoding.EncodeToString(ciphertext) + encSuffix, nil
}

// Decrypt decrypts ciphertext using AES-GCM
// Decrypt 使用 AES-GCM 解密密文
func Decrypt(ciphertext, key string) (string, error) {
	if !IsEncrypted(ciphertext) {
		return ciphertext, nil
	}
	ciphertext = ciphertext[len(encPrefix) : len(ciphertext)-len(encSuffix)]

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(deriveKey(key))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// IsEncrypted checks if string is encrypted
// IsEncrypted 检查字符串是否已加密
func IsEncrypted(s string) bool {
	return strings.HasPrefix(s, encPrefix) && strings.HasSuffix(s, encSuffix)
}
