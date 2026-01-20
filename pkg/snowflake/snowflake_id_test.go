package snowflake

import (
	"encoding/json"
	"testing"
)

func TestSnowflakeID_String(t *testing.T) {
	id := SnowflakeID(1234567890123456789)
	expected := "1234567890123456789"
	if id.String() != expected {
		t.Errorf("String() = %s, want %s", id.String(), expected)
	}
}

func TestSnowflakeID_Int64(t *testing.T) {
	id := SnowflakeID(1234567890123456789)
	expected := int64(1234567890123456789)
	if id.Int64() != expected {
		t.Errorf("Int64() = %d, want %d", id.Int64(), expected)
	}
}

func TestSnowflakeID_IsZero(t *testing.T) {
	tests := []struct {
		name string
		id   SnowflakeID
		want bool
	}{
		{"zero value", SnowflakeID(0), true},
		{"non-zero value", SnowflakeID(123), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.IsZero(); got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSnowflakeID_Valid(t *testing.T) {
	tests := []struct {
		name string
		id   SnowflakeID
		want bool
	}{
		{"zero value", SnowflakeID(0), false},
		{"non-zero value", SnowflakeID(123), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.Valid(); got != tt.want {
				t.Errorf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSnowflakeID_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		id      SnowflakeID
		want    string
		wantErr bool
	}{
		{
			name:    "normal ID",
			id:      SnowflakeID(1234567890123456789),
			want:    `"1234567890123456789"`,
			wantErr: false,
		},
		{
			name:    "zero ID",
			id:      SnowflakeID(0),
			want:    `"0"`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.id.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != tt.want {
				t.Errorf("MarshalJSON() = %s, want %s", string(got), tt.want)
			}
		})
	}
}

func TestSnowflakeID_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    SnowflakeID
		wantErr bool
	}{
		{
			name:    "string format",
			data:    `"1234567890123456789"`,
			want:    SnowflakeID(1234567890123456789),
			wantErr: false,
		},
		{
			name:    "number format",
			data:    `1234567890123456789`,
			want:    SnowflakeID(1234567890123456789),
			wantErr: false,
		},
		{
			name:    "zero string",
			data:    `"0"`,
			want:    SnowflakeID(0),
			wantErr: false,
		},
		{
			name:    "zero number",
			data:    `0`,
			want:    SnowflakeID(0),
			wantErr: false,
		},
		{
			name:    "invalid string",
			data:    `"invalid"`,
			want:    SnowflakeID(0),
			wantErr: true,
		},
		{
			name:    "invalid format",
			data:    `true`,
			want:    SnowflakeID(0),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var id SnowflakeID
			err := id.UnmarshalJSON([]byte(tt.data))
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && id != tt.want {
				t.Errorf("UnmarshalJSON() = %d, want %d", id, tt.want)
			}
		})
	}
}

func TestSnowflakeID_FromDB(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    SnowflakeID
		wantErr bool
	}{
		{
			name:    "valid ID",
			data:    []byte("1234567890123456789"),
			want:    SnowflakeID(1234567890123456789),
			wantErr: false,
		},
		{
			name:    "zero ID",
			data:    []byte("0"),
			want:    SnowflakeID(0),
			wantErr: false,
		},
		{
			name:    "nil data",
			data:    nil,
			want:    SnowflakeID(0),
			wantErr: false,
		},
		{
			name:    "empty data",
			data:    []byte(""),
			want:    SnowflakeID(0),
			wantErr: false,
		},
		{
			name:    "invalid data",
			data:    []byte("invalid"),
			want:    SnowflakeID(0),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var id SnowflakeID
			err := id.FromDB(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && id != tt.want {
				t.Errorf("FromDB() = %d, want %d", id, tt.want)
			}
		})
	}
}

func TestSnowflakeID_ToDB(t *testing.T) {
	tests := []struct {
		name    string
		id      SnowflakeID
		want    int64
		wantErr bool
	}{
		{
			name:    "normal ID",
			id:      SnowflakeID(1234567890123456789),
			want:    1234567890123456789,
			wantErr: false,
		},
		{
			name:    "zero ID",
			id:      SnowflakeID(0),
			want:    0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.id.ToDB()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ToDB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSnowflakeID_JSONRoundTrip(t *testing.T) {
	type TestStruct struct {
		ID       SnowflakeID `json:"id"`
		UserID   SnowflakeID `json:"user_id"`
		Username string      `json:"username"`
	}

	original := TestStruct{
		ID:       SnowflakeID(1234567890123456789),
		UserID:   SnowflakeID(8876543210987654321),
		Username: "test_user",
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Check JSON format (IDs should be strings)
	expected := `{"id":"1234567890123456789","user_id":"8876543210987654321","username":"test_user"}`
	if string(data) != expected {
		t.Errorf("Marshal() = %s, want %s", string(data), expected)
	}

	// Unmarshal from JSON
	var result TestStruct
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify values
	if result.ID != original.ID {
		t.Errorf("ID = %d, want %d", result.ID, original.ID)
	}
	if result.UserID != original.UserID {
		t.Errorf("UserID = %d, want %d", result.UserID, original.UserID)
	}
	if result.Username != original.Username {
		t.Errorf("Username = %s, want %s", result.Username, original.Username)
	}
}
