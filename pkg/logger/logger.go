package logger

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// Level represents log level
// Level 表示日志级别
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

var levelNames = map[Level]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
}

var levelColors = map[Level]string{
	DEBUG: "\033[36m", // Cyan | 青色
	INFO:  "\033[32m", // Green | 绿色
	WARN:  "\033[33m", // Yellow | 黄色
	ERROR: "\033[31m", // Red | 红色
}

// Logger is a generic logger
// Logger 是一个泛型日志器
type Logger[T any] struct {
	module string
	color  string
	writer *Writer
}

// New creates module logger, automatically gets module name through generics
// New 创建模块日志器，通过泛型自动获取模块名
func New[T any]() *Logger[T] {
	name := reflect.TypeOf((*T)(nil)).Elem().Name()
	return &Logger[T]{
		module: name,
		color:  getModuleColor(name),
		writer: NewWriter(name),
	}
}

// NewWithName creates logger with specified module name
// NewWithName 创建指定模块名的日志器
func NewWithName[T any](name string) *Logger[T] {
	return &Logger[T]{
		module: name,
		color:  getModuleColor(name),
		writer: NewWriter(name),
	}
}

// System represents system-level logger (for boot/common and other base layers)
// System 表示系统级日志器（用于 boot/common 等基础层）
type System struct {
	*Logger[struct{}]
}

// NewSystem creates system logger
// NewSystem 创建系统日志器
func NewSystem(name string) *System {
	fullName := "system:" + name
	return &System{
		Logger: &Logger[struct{}]{
			module: fullName,
			color:  "\033[93m", // Bright yellow, fixed color for system level | 亮黄色，系统级固定颜色
			writer: NewWriter("system"),
		},
	}
}

// Extract traceId from ctx
// 从 ctx 中提取 traceId
func getTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

func (l *Logger[T]) log(ctx context.Context, level Level, msg string, args ...any) {
	now := time.Now()
	text := msg
	if len(args) > 0 {
		text = fmt.Sprintf(msg, args...)
	}

	traceID := getTraceID(ctx)
	traceStr := ""
	traceFileStr := ""
	if traceID != "" {
		traceStr = fmt.Sprintf(" %s[%s]%s", "\033[90m", traceID[:16], "\033[0m")
		traceFileStr = fmt.Sprintf(" [%s]", traceID[:16])
	}

	// Console colored output | 控制台彩色输出
	console := fmt.Sprintf("%s[%s]%s %s[%s]%s %s%s%s%s %s\n",
		"\033[90m", now.Format("2006-01-02 15:04:05"), "\033[0m",
		l.color, l.module, "\033[0m",
		levelColors[level], levelNames[level], "\033[0m",
		traceStr,
		text,
	)
	fmt.Print(console)

	// File output (no color) | 文件输出（无颜色）
	file := fmt.Sprintf("[%s] [%s] [%s]%s %s\n",
		now.Format("2006-01-02 15:04:05"),
		l.module,
		levelNames[level],
		traceFileStr,
		text,
	)
	l.writer.Write(file)
}

// Log methods with ctx | 带 ctx 的日志方法
func (l *Logger[T]) DebugCtx(ctx context.Context, msg string, args ...any) {
	l.log(ctx, DEBUG, msg, args...)
}

func (l *Logger[T]) InfoCtx(ctx context.Context, msg string, args ...any) {
	l.log(ctx, INFO, msg, args...)
}

func (l *Logger[T]) WarnCtx(ctx context.Context, msg string, args ...any) {
	l.log(ctx, WARN, msg, args...)
}

func (l *Logger[T]) ErrorCtx(ctx context.Context, msg string, args ...any) {
	l.log(ctx, ERROR, msg, args...)
}

// Log methods without ctx (for compatibility) | 不带 ctx 的日志方法（兼容性）
func (l *Logger[T]) Debug(msg string, args ...any) {
	l.log(nil, DEBUG, msg, args...)
}

func (l *Logger[T]) Info(msg string, args ...any) {
	l.log(nil, INFO, msg, args...)
}

func (l *Logger[T]) Warn(msg string, args ...any) {
	l.log(nil, WARN, msg, args...)
}

func (l *Logger[T]) Error(msg string, args ...any) {
	l.log(nil, ERROR, msg, args...)
}

// Close closes logger
// Close 关闭日志器
func (l *Logger[T]) Close() error {
	return l.writer.Close()
}
