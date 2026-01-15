package util

import (
	"regexp"
	"testing"
)

func TestRandomString(t *testing.T) {
	tests := []int{0, 1, 10, 32, 100}

	for _, length := range tests {
		result := RandomString(length)
		if len(result) != length {
			t.Errorf("RandomString(%d) length = %d, want %d", length, len(result), length)
		}

		// Check that it only contains alphanumeric characters
		if length > 0 {
			matched, _ := regexp.MatchString("^[a-zA-Z0-9]+$", result)
			if !matched {
				t.Errorf("RandomString(%d) = %q contains invalid characters", length, result)
			}
		}
	}
}

func TestRandomStringUnique(t *testing.T) {
	seen := make(map[string]bool)
	count := 100

	for i := 0; i < count; i++ {
		s := RandomString(16)
		if seen[s] {
			t.Errorf("Duplicate random string: %s", s)
		}
		seen[s] = true
	}
}

func TestRandomDigits(t *testing.T) {
	tests := []int{0, 1, 6, 10, 20}

	for _, length := range tests {
		result := RandomDigits(length)
		if len(result) != length {
			t.Errorf("RandomDigits(%d) length = %d, want %d", length, len(result), length)
		}

		// Check that it only contains digits
		if length > 0 {
			matched, _ := regexp.MatchString("^[0-9]+$", result)
			if !matched {
				t.Errorf("RandomDigits(%d) = %q contains non-digit characters", length, result)
			}
		}
	}
}

func TestRandomInt(t *testing.T) {
	tests := []struct {
		min, max int64
	}{
		{0, 10},
		{1, 100},
		{-10, 10},
		{100, 1000},
	}

	for _, tt := range tests {
		for i := 0; i < 100; i++ {
			result := RandomInt(tt.min, tt.max)
			if result < tt.min || result >= tt.max {
				t.Errorf("RandomInt(%d, %d) = %d, out of range", tt.min, tt.max, result)
			}
		}
	}
}

func TestRandomIntDistribution(t *testing.T) {
	// Simple distribution test
	counts := make(map[int64]int)
	iterations := 10000
	min, max := int64(0), int64(10)

	for i := 0; i < iterations; i++ {
		result := RandomInt(min, max)
		counts[result]++
	}

	// Each value should appear roughly iterations/range times
	expected := iterations / int(max-min)
	tolerance := expected / 2 // 50% tolerance

	for i := min; i < max; i++ {
		if counts[i] < expected-tolerance || counts[i] > expected+tolerance {
			t.Logf("RandomInt distribution for %d: %d (expected ~%d)", i, counts[i], expected)
		}
	}
}

func BenchmarkRandomString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandomString(32)
	}
}

func BenchmarkRandomDigits(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandomDigits(6)
	}
}

func BenchmarkRandomInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandomInt(0, 1000000)
	}
}
