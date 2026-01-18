package pgsql

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	xormlog "xorm.io/xorm/log"
)

// Logger defines the logger interface for pgsql package
// Logger 定义 pgsql 包的日志器接口
// This allows pgsql to be independent of pkg/logger
// 这使得 pgsql 可以独立于 pkg/logger
type Logger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Warn(format string, args ...any)
	Error(format string, args ...any)
}

// defaultLogger is a simple logger using standard log package
// defaultLogger 是使用标准 log 包的简单日志器
type defaultLogger struct {
	prefix string
}

func (l *defaultLogger) Debug(format string, args ...any) {
	log.Printf("[DEBUG] [%s] %s", l.prefix, fmt.Sprintf(format, args...))
}

func (l *defaultLogger) Info(format string, args ...any) {
	log.Printf("[INFO] [%s] %s", l.prefix, fmt.Sprintf(format, args...))
}

func (l *defaultLogger) Warn(format string, args ...any) {
	log.Printf("[WARN] [%s] %s", l.prefix, fmt.Sprintf(format, args...))
}

func (l *defaultLogger) Error(format string, args ...any) {
	log.Printf("[ERROR] [%s] %s", l.prefix, fmt.Sprintf(format, args...))
}

// XormLogger xorm log adapter
// XormLogger xorm 日志适配器
type XormLogger struct {
	log       Logger
	level     xormlog.LogLevel
	showSQL   bool
	showDebug bool
}

// NewXormLogger creates xorm log adapter with default logger
// NewXormLogger 创建带默认日志器的 xorm 日志适配器
func NewXormLogger(showSQL bool) *XormLogger {
	return &XormLogger{
		log:       &defaultLogger{prefix: "xorm"},
		level:     xormlog.LOG_INFO,
		showSQL:   showSQL,
		showDebug: false,
	}
}

// SetLogger sets the logger instance
// SetLogger 设置日志器实例
// Call this after logger is initialized to inject the real logger
// 在日志器初始化后调用此方法以注入真实的日志器
func (l *XormLogger) SetLogger(logger Logger) {
	if logger != nil {
		l.log = logger
	}
}

// Level returns log level
// Level 返回日志级别
func (l *XormLogger) Level() xormlog.LogLevel {
	return l.level
}

// SetLevel sets log level
// SetLevel 设置日志级别
func (l *XormLogger) SetLevel(level xormlog.LogLevel) {
	l.level = level
}

// ShowSQL whether to show SQL
// ShowSQL 是否显示 SQL
func (l *XormLogger) ShowSQL(show ...bool) {
	if len(show) > 0 {
		l.showSQL = show[0]
	}
}

// IsShowSQL whether to show SQL
func (l *XormLogger) IsShowSQL() bool {
	return l.showSQL
}

// Debug debug log
func (l *XormLogger) Debug(v ...interface{}) {
	if l.level <= xormlog.LOG_DEBUG && l.showDebug {
		l.log.Debug("[%s] %s", getCaller(), fmt.Sprint(v...))
	}
}

// Debugf debug log (formatted)
func (l *XormLogger) Debugf(format string, v ...interface{}) {
	if l.level <= xormlog.LOG_DEBUG && l.showDebug {
		l.log.Debug("[%s] %s", getCaller(), fmt.Sprintf(format, v...))
	}
}

// Info info log
func (l *XormLogger) Info(v ...interface{}) {
	if l.level <= xormlog.LOG_INFO {
		l.log.Info("[%s] %s", getCaller(), fmt.Sprint(v...))
	}
}

// Infof info log (formatted)
func (l *XormLogger) Infof(format string, v ...interface{}) {
	if l.level <= xormlog.LOG_INFO {
		l.log.Info("[%s] %s", getCaller(), fmt.Sprintf(format, v...))
	}
}

// Warn warning log
func (l *XormLogger) Warn(v ...interface{}) {
	if l.level <= xormlog.LOG_WARNING {
		l.log.Warn("[%s] %s", getCaller(), fmt.Sprint(v...))
	}
}

// Warnf warning log (formatted)
func (l *XormLogger) Warnf(format string, v ...interface{}) {
	if l.level <= xormlog.LOG_WARNING {
		l.log.Warn("[%s] %s", getCaller(), fmt.Sprintf(format, v...))
	}
}

// Error error log
func (l *XormLogger) Error(v ...interface{}) {
	if l.level <= xormlog.LOG_ERR {
		l.log.Error("[%s] %s", getCaller(), fmt.Sprint(v...))
	}
}

// Errorf error log (formatted)
func (l *XormLogger) Errorf(format string, v ...interface{}) {
	if l.level <= xormlog.LOG_ERR {
		l.log.Error("[%s] %s", getCaller(), fmt.Sprintf(format, v...))
	}
}

// getCaller returns the caller information (file:line)
// Extracts module name from call stack for better log tracing
func getCaller() string {
	for i := 3; i < 10; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		// Skip xorm internal files
		if strings.Contains(file, "xorm.io/xorm") {
			continue
		}

		// Skip pgsql package files
		if strings.Contains(file, "/pkg/pgsql/") {
			continue
		}

		// Extract meaningful path (module/xxx/...)
		if idx := strings.Index(file, "/module/"); idx != -1 {
			file = file[idx+1:] // Remove prefix, keep "module/xxx/..."
		} else if idx := strings.Index(file, "/common/"); idx != -1 {
			file = file[idx+1:] // Keep "common/xxx/..."
		} else {
			// Fallback: get last 2 segments
			parts := strings.Split(file, "/")
			if len(parts) >= 2 {
				file = strings.Join(parts[len(parts)-2:], "/")
			}
		}

		return fmt.Sprintf("%s:%d", file, line)
	}
	return "unknown"
}
