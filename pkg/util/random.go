package util

import (
	"crypto/rand"
	"math/big"
)

const (
	letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits  = "0123456789"
	chars   = letters + digits
)

// RandomString generates a random string
func RandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[num.Int64()]
	}
	return string(b)
}

// RandomDigits generates random digits
func RandomDigits(n int) string {
	b := make([]byte, n)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		b[i] = digits[num.Int64()]
	}
	return string(b)
}

// RandomInt generates a random integer [min, max)
func RandomInt(min, max int64) int64 {
	n, _ := rand.Int(rand.Reader, big.NewInt(max-min))
	return n.Int64() + min
}
