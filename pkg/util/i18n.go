package util

import (
	"database/sql/driver"
	"errors"

	"github.com/bytedance/sonic"
)

// Default language
const (
	LangZh = "zh"
	LangEn = "en"
)

// DefaultLang default language
var DefaultLang = LangZh

// I18nText internationalized text type, stored as JSON: {"zh":"中文","en":"English"}
type I18nText map[string]string

// NewI18nText creates internationalized text
func NewI18nText() I18nText {
	return make(I18nText)
}

// Set sets text for specified language
func (t I18nText) Set(lang, text string) I18nText {
	if t == nil {
		t = make(I18nText)
	}
	t[lang] = text
	return t
}

// Get returns text for specified language, falls back to default language if not found
func (t I18nText) Get(lang string) string {
	if t == nil {
		return ""
	}
	if text, ok := t[lang]; ok && text != "" {
		return text
	}
	// Fall back to default language
	if text, ok := t[DefaultLang]; ok {
		return text
	}
	// Return any non-empty language
	for _, text := range t {
		if text != "" {
			return text
		}
	}
	return ""
}

// GetDefault returns text for default language
func (t I18nText) GetDefault() string {
	return t.Get(DefaultLang)
}

// IsEmpty checks if empty
func (t I18nText) IsEmpty() bool {
	if t == nil || len(t) == 0 {
		return true
	}
	for _, text := range t {
		if text != "" {
			return false
		}
	}
	return true
}

// Value implements driver.Valuer interface for database write
func (t I18nText) Value() (driver.Value, error) {
	if t == nil || len(t) == 0 {
		return "{}", nil
	}
	return sonic.Marshal(t)
}

// Scan implements sql.Scanner interface for database read
func (t *I18nText) Scan(value interface{}) error {
	if value == nil {
		*t = make(I18nText)
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("invalid type for I18nText")
	}

	if len(bytes) == 0 {
		*t = make(I18nText)
		return nil
	}

	return sonic.Unmarshal(bytes, t)
}

// MarshalJSON implements sonic.Marshaler interface
func (t I18nText) MarshalJSON() ([]byte, error) {
	if t == nil {
		return []byte("{}"), nil
	}
	return sonic.Marshal(map[string]string(t))
}

// UnmarshalJSON implements sonic.Unmarshaler interface
func (t *I18nText) UnmarshalJSON(data []byte) error {
	if t == nil {
		return errors.New("I18nText: UnmarshalJSON on nil pointer")
	}
	var m map[string]string
	if err := sonic.Unmarshal(data, &m); err != nil {
		return err
	}
	*t = m
	return nil
}

// FromDB implements xorm FromDB interface
func (t *I18nText) FromDB(data []byte) error {
	return t.Scan(data)
}

// ToDB implements xorm ToDB interface
func (t I18nText) ToDB() ([]byte, error) {
	if t == nil || len(t) == 0 {
		return []byte("{}"), nil
	}
	return sonic.Marshal(t)
}
