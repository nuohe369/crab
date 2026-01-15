package util

import (
	"strings"
	"unicode"
)

// IsEmpty checks if string is empty
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsNotEmpty checks if string is not empty
func IsNotEmpty(s string) bool {
	return !IsEmpty(s)
}

// CamelToSnake converts camelCase to snake_case
func CamelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// SnakeToCamel converts snake_case to CamelCase
func SnakeToCamel(s string) string {
	var result strings.Builder
	upper := true
	for _, r := range s {
		if r == '_' {
			upper = true
		} else {
			if upper {
				result.WriteRune(unicode.ToUpper(r))
				upper = false
			} else {
				result.WriteRune(r)
			}
		}
	}
	return result.String()
}

// Truncate truncates a string
func Truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen])
}
