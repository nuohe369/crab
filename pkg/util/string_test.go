package util

import "testing"

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"", true},
		{" ", true},
		{"  ", true},
		{"\t", true},
		{"\n", true},
		{" \t\n ", true},
		{"a", false},
		{" a ", false},
		{"hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := IsEmpty(tt.input); got != tt.expected {
				t.Errorf("IsEmpty(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsNotEmpty(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"", false},
		{" ", false},
		{"a", true},
		{"hello", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := IsNotEmpty(tt.input); got != tt.expected {
				t.Errorf("IsNotEmpty(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestCamelToSnake(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"a", "a"},
		{"A", "a"},
		{"camelCase", "camel_case"},
		{"CamelCase", "camel_case"},
		{"getHTTPResponse", "get_h_t_t_p_response"},
		{"userID", "user_i_d"},
		{"already_snake", "already_snake"},
		{"ABC", "a_b_c"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := CamelToSnake(tt.input); got != tt.expected {
				t.Errorf("CamelToSnake(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSnakeToCamel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"a", "A"},
		{"snake_case", "SnakeCase"},
		{"user_id", "UserId"},
		{"get_http_response", "GetHttpResponse"},
		{"already", "Already"},
		{"_leading", "Leading"},
		{"trailing_", "Trailing"},
		{"double__underscore", "DoubleUnderscore"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := SnakeToCamel(tt.input); got != tt.expected {
				t.Errorf("SnakeToCamel(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello", 3, "hel"},
		{"hello", 0, ""},
		{"", 5, ""},
		{"你好世界", 2, "你好"},
		{"hello世界", 6, "hello世"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := Truncate(tt.input, tt.maxLen); got != tt.expected {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.expected)
			}
		})
	}
}
