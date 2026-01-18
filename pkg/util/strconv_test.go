package util

import (
	"strconv"
	"testing"
)

func TestInt64ToString(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{100, "100"},
		{9999, "9999"},
		{10000, "10000"},
		{10001, "10001"},   // Outside cache range
		{-1, "-1"},         // Negative number
		{999999, "999999"}, // Large number
	}

	for _, tt := range tests {
		result := Int64ToString(tt.input)
		if result != tt.expected {
			t.Errorf("Int64ToString(%d) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestInt64ToStringCache(t *testing.T) {
	// Test that cached values return the same string instance
	str1 := Int64ToString(100)
	str2 := Int64ToString(100)

	// Both should be the same cached string
	if str1 != str2 {
		t.Error("Cached strings should be identical")
	}
}

func TestStringToInt64(t *testing.T) {
	tests := []struct {
		input       string
		expected    int64
		shouldError bool
	}{
		{"0", 0, false},
		{"123", 123, false},
		{"9999", 9999, false},
		{"-100", -100, false},
		{"invalid", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		result, err := StringToInt64(tt.input)
		if tt.shouldError {
			if err == nil {
				t.Errorf("StringToInt64(%s) should return error", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("StringToInt64(%s) unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("StringToInt64(%s) = %d, want %d", tt.input, result, tt.expected)
			}
		}
	}
}

func TestMustStringToInt64(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"123", 123},
		{"0", 0},
		{"invalid", 0}, // Returns 0 on error
		{"", 0},        // Returns 0 on error
	}

	for _, tt := range tests {
		result := MustStringToInt64(tt.input)
		if result != tt.expected {
			t.Errorf("MustStringToInt64(%s) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestInt64ToStringBatch(t *testing.T) {
	input := []int64{1, 2, 3, 100, 10001}
	expected := []string{"1", "2", "3", "100", "10001"}

	result := Int64ToStringBatch(input)

	if len(result) != len(expected) {
		t.Fatalf("Expected length %d, got %d", len(expected), len(result))
	}

	for i := range result {
		if result[i] != expected[i] {
			t.Errorf("Index %d: expected %s, got %s", i, expected[i], result[i])
		}
	}
}

func TestInt64ToStringBatchEmpty(t *testing.T) {
	result := Int64ToStringBatch([]int64{})
	if len(result) != 0 {
		t.Errorf("Expected empty slice, got length %d", len(result))
	}
}

// Benchmark tests to verify cache performance improvement

func BenchmarkInt64ToStringCached(b *testing.B) {
	// Benchmark cached values (within cache range)
	for i := 0; i < b.N; i++ {
		Int64ToString(100)
	}
}

func BenchmarkInt64ToStringUncached(b *testing.B) {
	// Benchmark uncached values (outside cache range)
	for i := 0; i < b.N; i++ {
		Int64ToString(100000)
	}
}

func BenchmarkStdlibFormatInt(b *testing.B) {
	// Benchmark standard library for comparison
	for i := 0; i < b.N; i++ {
		strconv.FormatInt(100, 10)
	}
}

func BenchmarkInt64ToStringBatch(b *testing.B) {
	nums := []int64{1, 10, 100, 1000, 10000}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Int64ToStringBatch(nums)
	}
}
