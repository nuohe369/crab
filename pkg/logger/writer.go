package logger

import (
	"bufio"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	logDir = "logs"
)

// Writer writes log files with daily rotation support
type Writer struct {
	module   string
	file     *os.File
	buffer   *bufio.Writer
	date     string
	mu       sync.Mutex
	stopChan chan struct{} // Signal to stop periodic flush | 停止定期刷新的信号
	wg       sync.WaitGroup
}

// NewWriter creates writer
func NewWriter(module string) *Writer {
	w := &Writer{
		module:   module,
		stopChan: make(chan struct{}),
	}
	w.rotate()
	
	// Start background flusher | 启动后台刷新器
	w.wg.Add(1)
	go w.periodicFlush()
	
	return w
}

// periodicFlush periodically flushes buffer to disk
// periodicFlush 定期将缓冲区刷新到磁盘
func (w *Writer) periodicFlush() {
	defer w.wg.Done()
	
	cfg := GetConfig()
	ticker := time.NewTicker(cfg.FlushInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			w.Flush()
		case <-w.stopChan:
			return
		}
	}
}

// Write writes log message
func (w *Writer) Write(msg string) {
	// Check if file logging is disabled | 检查是否禁用文件日志
	if !GetConfig().Enabled {
		return
	}
	
	w.mu.Lock()
	defer w.mu.Unlock()

	// Check date change and auto-rotate
	today := time.Now().Format("2006-01-02")
	if today != w.date {
		w.rotate()
	}

	if w.buffer != nil {
		w.buffer.WriteString(msg)
		// Auto flush if buffer is getting full (>75%) | 缓冲区超过 75% 时自动刷新
		cfg := GetConfig()
		if w.buffer.Buffered() > cfg.BufferSize*3/4 {
			w.buffer.Flush()
		}
	}
}

// Flush flushes buffered data to disk
// Flush 将缓冲数据刷新到磁盘
func (w *Writer) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.buffer != nil {
		return w.buffer.Flush()
	}
	return nil
}

// rotate rotates log file
func (w *Writer) rotate() {
	// Flush and close old file | 刷新并关闭旧文件
	if w.buffer != nil {
		w.buffer.Flush()
	}
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
	
	// Use configured buffer size | 使用配置的缓冲区大小
	cfg := GetConfig()
	w.buffer = bufio.NewWriterSize(f, cfg.BufferSize)
}

// Close closes file
func (w *Writer) Close() error {
	// Stop periodic flush goroutine | 停止定期刷新的 goroutine
	close(w.stopChan)
	w.wg.Wait()
	
	w.mu.Lock()
	defer w.mu.Unlock()
	
	// Flush buffer before closing | 关闭前刷新缓冲区
	if w.buffer != nil {
		if err := w.buffer.Flush(); err != nil {
			return err
		}
	}
	
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}
