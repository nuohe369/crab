package ws

import (
	"testing"
)

func TestNewMessage(t *testing.T) {
	msg := NewMessage(123, "chat", map[string]string{"text": "hello"})

	if msg.UserID != 123 {
		t.Errorf("UserID = %d, want 123", msg.UserID)
	}
	if msg.Type != "chat" {
		t.Errorf("Type = %s, want chat", msg.Type)
	}
	if msg.Payload == nil {
		t.Error("Payload should not be nil")
	}
}

func TestNewBroadcast(t *testing.T) {
	msg := NewBroadcast("system", "hello everyone")

	if msg.UserID != 0 {
		t.Errorf("Broadcast UserID = %d, want 0", msg.UserID)
	}
	if msg.Type != "system" {
		t.Errorf("Type = %s, want system", msg.Type)
	}
}

func TestMessageBytes(t *testing.T) {
	msg := NewMessage(1, "test", map[string]int{"count": 42})
	data := msg.Bytes()

	if len(data) == 0 {
		t.Error("Bytes() returned empty data")
	}

	// Should be valid JSON
	parsed, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	if parsed.UserID != msg.UserID {
		t.Errorf("UserID mismatch: got %d, want %d", parsed.UserID, msg.UserID)
	}
	if parsed.Type != msg.Type {
		t.Errorf("Type mismatch: got %s, want %s", parsed.Type, msg.Type)
	}
}

func TestParseMessage(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid message",
			input:   `{"user_id":1,"type":"chat","payload":"hello"}`,
			wantErr: false,
		},
		{
			name:    "minimal message",
			input:   `{"type":"ping"}`,
			wantErr: false,
		},
		{
			name:    "with nested payload",
			input:   `{"type":"data","payload":{"key":"value","num":123}}`,
			wantErr: false,
		},
		{
			name:    "invalid json",
			input:   `{invalid}`,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   ``,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := ParseMessage([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && msg == nil {
				t.Error("ParseMessage() returned nil message")
			}
		})
	}
}

func TestMessageRoundTrip(t *testing.T) {
	original := &Message{
		UserID: 999,
		Type:   "notification",
		Payload: map[string]any{
			"title":   "Test",
			"content": "Hello World",
			"count":   float64(42), // JSON numbers become float64
		},
	}

	data := original.Bytes()
	parsed, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	if parsed.UserID != original.UserID {
		t.Errorf("UserID mismatch: got %d, want %d", parsed.UserID, original.UserID)
	}
	if parsed.Type != original.Type {
		t.Errorf("Type mismatch: got %s, want %s", parsed.Type, original.Type)
	}

	// Check payload
	payload, ok := parsed.Payload.(map[string]any)
	if !ok {
		t.Fatalf("Payload type mismatch: got %T", parsed.Payload)
	}
	if payload["title"] != "Test" {
		t.Errorf("Payload title mismatch: got %v", payload["title"])
	}
}
