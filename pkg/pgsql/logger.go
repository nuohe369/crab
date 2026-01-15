package pgsql

import (
	"fmt"
	"log"

	xormlog "xorm.io/xorm/log"
)

// Logger defines the logger interface for pgsql package.
// This allows pgsql to be independent of pkg/logger.
type Logger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Warn(format string, args ...any)
	Error(format string, args ...any)
}

// defaultLogger is a simple logger using standard log package
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
type XormLogger struct {
	log       Logger
	level     xormlog.LogLevel
	showSQL   bool
	showDebug bool
}

// NewXormLogger creates xorm log adapter with default logger
func NewXormLogger(showSQL bool) *XormLogger {
	return &XormLogger{
		log:       &defaultLogger{prefix: "xorm"},
		level:     xormlog.LOG_INFO,
		showSQL:   showSQL,
		showDebug: false,
	}
}

// SetLogger sets the logger instance.
// Call this after logger is initialized to inject the real logger.
func (l *XormLogger) SetLogger(logger Logger) {
	if logger != nil {
		l.log = logger
	}
}

// Level returns log level
func (l *XormLogger) Level() xormlog.LogLevel {
	return l.level
}

// SetLevel sets log level
func (l *XormLogger) SetLevel(level xormlog.LogLevel) {
	l.level = level
}

// ShowSQL whether to show SQL
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
		l.log.Debug("%s", fmt.Sprint(v...))
	}
}

// Debugf debug log (formatted)
func (l *XormLogger) Debugf(format string, v ...interface{}) {
	if l.level <= xormlog.LOG_DEBUG && l.showDebug {
		l.log.Debug(format, v...)
	}
}

// Info info log
func (l *XormLogger) Info(v ...interface{}) {
	if l.level <= xormlog.LOG_INFO {
		l.log.Info("%s", fmt.Sprint(v...))
	}
}

// Infof info log (formatted)
func (l *XormLogger) Infof(format string, v ...interface{}) {
	if l.level <= xormlog.LOG_INFO {
		l.log.Info(format, v...)
	}
}

// Warn warning log
func (l *XormLogger) Warn(v ...interface{}) {
	if l.level <= xormlog.LOG_WARNING {
		l.log.Warn("%s", fmt.Sprint(v...))
	}
}

// Warnf warning log (formatted)
func (l *XormLogger) Warnf(format string, v ...interface{}) {
	if l.level <= xormlog.LOG_WARNING {
		l.log.Warn(format, v...)
	}
}

// Error error log
func (l *XormLogger) Error(v ...interface{}) {
	if l.level <= xormlog.LOG_ERR {
		l.log.Error("%s", fmt.Sprint(v...))
	}
}

// Errorf error log (formatted)
func (l *XormLogger) Errorf(format string, v ...interface{}) {
	if l.level <= xormlog.LOG_ERR {
		l.log.Error(format, v...)
	}
}
