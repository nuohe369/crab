package logger

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// Level represents log level
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
	DEBUG: "\033[36m", // Cyan
	INFO:  "\033[32m", // Green
	WARN:  "\033[33m", // Yellow
	ERROR: "\033[31m", // Red
}

// Logger is a generic logger
type Logger[T any] struct {
	module string
	color  string
	writer *Writer
}

// New creates module logger, automatically gets module name through generics
func New[T any]() *Logger[T] {
	name := reflect.TypeOf((*T)(nil)).Elem().Name()
	return &Logger[T]{
		module: name,
		color:  getModuleColor(name),
		writer: NewWriter(name),
	}
}

// NewWithName creates logger with specified module name
func NewWithName[T any](name string) *Logger[T] {
	return &Logger[T]{
		module: name,
		color:  getModuleColor(name),
		writer: NewWriter(name),
	}
}

// System represents system-level logger (for boot/common and other base layers)
type System struct {
	*Logger[struct{}]
}

// NewSystem creates system logger
func NewSystem(name string) *System {
	fullName := "system:" + name
	return &System{
		Logger: &Logger[struct{}]{
			module: fullName,
			color:  "\033[93m", // Bright yellow, fixed color for system level
			writer: NewWriter("system"),
		},
	}
}

// Extract traceId from ctx
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

	// Console colored output
	console := fmt.Sprintf("%s[%s]%s %s[%s]%s %s%s%s%s %s\n",
		"\033[90m", now.Format("2006-01-02 15:04:05"), "\033[0m",
		l.color, l.module, "\033[0m",
		levelColors[level], levelNames[level], "\033[0m",
		traceStr,
		text,
	)
	fmt.Print(console)

	// File output (no color)
	file := fmt.Sprintf("[%s] [%s] [%s]%s %s\n",
		now.Format("2006-01-02 15:04:05"),
		l.module,
		levelNames[level],
		traceFileStr,
		text,
	)
	l.writer.Write(file)
}

// Log methods with ctx
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

// Log methods without ctx (for compatibility)
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
func (l *Logger[T]) Close() error {
	return l.writer.Close()
}
