package config

import (
	"errors"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/fsnotify/fsnotify"
	"github.com/nuohe369/crab/pkg/crypto"
)

var (
	decryptKey string
	// ErrEncryptedNoKey indicates configuration contains encrypted values but no decryption key was provided
	ErrEncryptedNoKey = errors.New("configuration contains encrypted values ENC(), please provide decryption key with -k parameter")
)

// SetDecryptKey sets the decryption key for encrypted configuration values.
func SetDecryptKey(key string) {
	decryptKey = key
}

// Load loads a TOML configuration file into the target structure.
func Load(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Check if configuration contains encrypted values but no key was provided
	if decryptKey == "" && strings.Contains(string(data), "ENC(") {
		return ErrEncryptedNoKey
	}

	if err := toml.Unmarshal(data, target); err != nil {
		return err
	}
	// Decrypt encrypted fields
	if decryptKey != "" {
		if err := decryptFields(reflect.ValueOf(target), decryptKey); err != nil {
			return err
		}
	}
	return nil
}

// MustLoad loads configuration, exits on failure.
func MustLoad(path string, target any) {
	if err := Load(path, target); err != nil {
		log.Fatalf("Configuration loading failed: %v", err)
	}
}

// decryptFields recursively decrypts encrypted fields in a struct.
func decryptFields(v reflect.Value, key string) error {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			s := field.String()
			if crypto.IsEncrypted(s) {
				decrypted, err := crypto.Decrypt(s, key)
				if err != nil {
					return errors.New("decryption failed: " + err.Error())
				}
				field.SetString(decrypted)
			}
		case reflect.Struct:
			if err := decryptFields(field, key); err != nil {
				return err
			}
		case reflect.Ptr:
			if !field.IsNil() {
				if err := decryptFields(field, key); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Watcher provides hot-reload functionality for configuration files.
type Watcher struct {
	path     string
	target   any
	mu       sync.RWMutex
	watcher  *fsnotify.Watcher
	onChange func()
}

// NewWatcher creates a configuration watcher.
func NewWatcher(path string, target any) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	cw := &Watcher{
		path:    path,
		target:  target,
		watcher: w,
	}

	if err := cw.reload(); err != nil {
		w.Close()
		return nil, err
	}

	return cw, nil
}

// OnChange sets the callback function for configuration changes.
func (w *Watcher) OnChange(fn func()) {
	w.onChange = fn
}

// Start begins watching for configuration changes.
func (w *Watcher) Start() error {
	if err := w.watcher.Add(w.path); err != nil {
		return err
	}

	go w.watch()
	return nil
}

// Stop stops watching for configuration changes.
func (w *Watcher) Stop() error {
	return w.watcher.Close()
}

func (w *Watcher) watch() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			// Handle write and create events (some editors delete then create)
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				w.reload()
				if w.onChange != nil {
					w.onChange()
				}
			}
		case _, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
		}
	}
}

func (w *Watcher) reload() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return Load(w.path, w.target)
}

// Get returns the configuration (thread-safe).
func (w *Watcher) Get() any {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.target
}

// DecryptValue decrypts a single value.
func DecryptValue(value string) string {
	if decryptKey == "" || !crypto.IsEncrypted(value) {
		return value
	}
	if decrypted, err := crypto.Decrypt(value, decryptKey); err == nil {
		return decrypted
	}
	return value
}

// EncryptValue encrypts a single value.
func EncryptValue(value, key string) (string, error) {
	return crypto.Encrypt(value, key)
}

// IsEncrypted checks if a value is encrypted.
func IsEncrypted(value string) bool {
	return strings.HasPrefix(value, "ENC(") && strings.HasSuffix(value, ")")
}
