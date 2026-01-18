// Package config provides configuration file loading and hot-reload functionality
// Package config 提供配置文件加载和热重载功能
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
	decryptKey string // Decryption key for encrypted configuration values | 加密配置值的解密密钥
	// ErrEncryptedNoKey indicates configuration contains encrypted values but no decryption key was provided
	// ErrEncryptedNoKey 表示配置包含加密值但未提供解密密钥
	ErrEncryptedNoKey = errors.New("configuration contains encrypted values ENC(), please provide decryption key with -k parameter")
)

// SetDecryptKey sets the decryption key for encrypted configuration values.
// SetDecryptKey 设置加密配置值的解密密钥
func SetDecryptKey(key string) {
	decryptKey = key
}

// Load loads a TOML configuration file into the target structure.
// Load 将 TOML 配置文件加载到目标结构中
func Load(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Check if configuration contains encrypted values but no key was provided | 检查配置是否包含加密值但未提供密钥
	if decryptKey == "" && strings.Contains(string(data), "ENC(") {
		return ErrEncryptedNoKey
	}

	if err := toml.Unmarshal(data, target); err != nil {
		return err
	}
	// Decrypt encrypted fields | 解密加密字段
	if decryptKey != "" {
		if err := decryptFields(reflect.ValueOf(target), decryptKey); err != nil {
			return err
		}
	}
	return nil
}

// MustLoad loads configuration, exits on failure.
// MustLoad 加载配置，失败时退出
func MustLoad(path string, target any) {
	if err := Load(path, target); err != nil {
		log.Fatalf("Configuration loading failed: %v", err)
	}
}

// decryptFields recursively decrypts encrypted fields in a struct.
// decryptFields 递归解密结构体中的加密字段
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
// Watcher 提供配置文件的热重载功能
type Watcher struct {
	path     string            // Configuration file path | 配置文件路径
	target   any               // Configuration target structure | 配置目标结构
	mu       sync.RWMutex      // Mutex for concurrent access | 并发访问互斥锁
	watcher  *fsnotify.Watcher // File system watcher | 文件系统监视器
	onChange func()            // Change callback | 变更回调
}

// NewWatcher creates a configuration watcher.
// NewWatcher 创建配置监视器
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
// OnChange 设置配置变更的回调函数
func (w *Watcher) OnChange(fn func()) {
	w.onChange = fn
}

// Start begins watching for configuration changes.
// Start 开始监视配置变更
func (w *Watcher) Start() error {
	if err := w.watcher.Add(w.path); err != nil {
		return err
	}

	go w.watch()
	return nil
}

// Stop stops watching for configuration changes.
// Stop 停止监视配置变更
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
			// Handle write and create events (some editors delete then create) | 处理写入和创建事件（某些编辑器会先删除再创建）
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
// Get 返回配置（线程安全）
func (w *Watcher) Get() any {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.target
}

// DecryptValue decrypts a single value.
// DecryptValue 解密单个值
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
// EncryptValue 加密单个值
func EncryptValue(value, key string) (string, error) {
	return crypto.Encrypt(value, key)
}

// IsEncrypted checks if a value is encrypted.
// IsEncrypted 检查值是否已加密
func IsEncrypted(value string) bool {
	return strings.HasPrefix(value, "ENC(") && strings.HasSuffix(value, ")")
}
