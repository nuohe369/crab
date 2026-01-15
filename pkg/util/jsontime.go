package util

import (
	"database/sql/driver"
	"time"
)

// DateTime custom time type, JSON serializes to "2006-01-02 15:04:05"
type DateTime time.Time

// MarshalJSON implements json.Marshaler interface
func (t DateTime) MarshalJSON() ([]byte, error) {
	tt := time.Time(t)
	if tt.IsZero() {
		return []byte(`""`), nil
	}
	return []byte(`"` + tt.Format(DateTimeFormat) + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler interface
func (t *DateTime) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == `""` || s == "null" {
		return nil
	}
	// Remove quotes
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	tt, err := time.ParseInLocation(DateTimeFormat, s, time.Local)
	if err != nil {
		return err
	}
	*t = DateTime(tt)
	return nil
}

// Time converts to time.Time
func (t DateTime) Time() time.Time {
	return time.Time(t)
}

// String implements Stringer interface
func (t DateTime) String() string {
	tt := time.Time(t)
	if tt.IsZero() {
		return ""
	}
	return tt.Format(DateTimeFormat)
}

// Value implements driver.Valuer interface (for database write)
func (t DateTime) Value() (driver.Value, error) {
	return time.Time(t), nil
}

// Scan implements sql.Scanner interface (for database read)
func (t *DateTime) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		*t = DateTime(v)
	}
	return nil
}

// NewDateTime creates DateTime from time.Time
func NewDateTime(t time.Time) DateTime {
	return DateTime(t)
}

// NowDateTime returns current time as DateTime
func NowDateTime() DateTime {
	return DateTime(time.Now())
}

// DateTimePtr creates *DateTime from *time.Time
func DateTimePtr(t *time.Time) *DateTime {
	if t == nil {
		return nil
	}
	dt := DateTime(*t)
	return &dt
}

// DT shorthand, convenient for use in handlers
func DT(t time.Time) DateTime {
	return DateTime(t)
}

// DTPtr shorthand, handles pointer type
func DTPtr(t *time.Time) *DateTime {
	return DateTimePtr(t)
}
