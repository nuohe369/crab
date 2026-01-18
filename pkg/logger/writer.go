package logger

import (
	"os"
	"path/filepath"
	"sync"
	"time"
)

const logDir = "logs"

// Writer writes log files with daily rotation support
type Writer struct {
	module string
	file   *os.File
	date   string
	mu     sync.Mutex
}

// NewWriter creates writer
func NewWriter(module string) *Writer {
	w := &Writer{module: module}
	w.rotate()
	return w
}

// Write writes log message
func (w *Writer) Write(msg string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Check date change and auto-rotate
	today := time.Now().Format("2006-01-02")
	if today != w.date {
		w.rotate()
	}

	if w.file != nil {
		w.file.WriteString(msg)
	}
}

// rotate rotates log file
func (w *Writer) rotate() {
	if w.file != nil {
		w.file.Close()
	}

	w.date = time.Now().Format("2006-01-02")
	dir := filepath.Join(logDir, w.module)
	os.MkdirAll(dir, 0755)

	path := filepath.Join(dir, w.date+".log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	w.file = f
}

// Close closes file
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}
