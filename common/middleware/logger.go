package middleware

import (
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nuohe369/crab/pkg/logger"
	"github.com/nuohe369/crab/pkg/trace"
)

// loggers caches logger instances for each module | loggers 缓存各模块的日志器
var (
	loggers   = make(map[string]*logger.Logger[struct{}])
	loggersMu sync.RWMutex
)

// getLogger gets or creates a logger for the specified module
// getLogger 获取或创建指定模块的日志器
func getLogger(module string) *logger.Logger[struct{}] {
	// Fast path: read lock check | 快速路径: 读锁检查
	loggersMu.RLock()
	if l, ok := loggers[module]; ok {
		loggersMu.RUnlock()
		return l
	}
	loggersMu.RUnlock()

	// Slow path: write lock create | 慢速路径: 写锁创建
	loggersMu.Lock()
	defer loggersMu.Unlock()

	// Double check to avoid duplicate creation | 双重检查，避免重复创建
	if l, ok := loggers[module]; ok {
		return l
	}

	l := logger.NewWithName[struct{}](module)
	loggers[module] = l
	return l
}

// Logger returns a request logging middleware for the specified module
// Logger 返回指定模块的请求日志中间件
func Logger(module string) fiber.Handler {
	log := getLogger(module)
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		latency := time.Since(start)
		status := c.Response().StatusCode()
		method := c.Method()
		path := c.Path()
		traceID := trace.TraceID(c.UserContext())

		if traceID != "" {
			log.InfoCtx(c.UserContext(), "%s %s %d %v", method, path, status, latency)
		} else {
			log.Info("%s %s %d %v", method, path, status, latency)
		}

		return err
	}
}

// SmartLogger returns a smart logging middleware that auto-detects module name from request path
// SmartLogger 返回智能日志中间件，从请求路径自动检测模块名称
//
// Examples | 示例:
//
//	/testapi/user -> logs to testapi module | 记录到 testapi 模块
//	/ws/basic -> logs to ws module | 记录到 ws 模块
//	/health -> logs to http module | 记录到 http 模块
func SmartLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Auto-detect module name from path
		// 从路径自动检测模块名称
		moduleName := extractModuleName(c.Path())
		log := getLogger(moduleName)

		err := c.Next()

		latency := time.Since(start)
		status := c.Response().StatusCode()
		method := c.Method()
		path := c.Path()
		traceID := trace.TraceID(c.UserContext())

		// Log with appropriate level based on status code
		// 根据状态码使用适当的日志级别
		if status >= 500 {
			// Server errors | 服务器错误
			if traceID != "" {
				log.ErrorCtx(c.UserContext(), "%s %s %d %v", method, path, status, latency)
			} else {
				log.Error("%s %s %d %v", method, path, status, latency)
			}
		} else if status >= 400 {
			// Client errors | 客户端错误
			if traceID != "" {
				log.WarnCtx(c.UserContext(), "%s %s %d %v", method, path, status, latency)
			} else {
				log.Warn("%s %s %d %v", method, path, status, latency)
			}
		} else {
			// Success | 成功
			if traceID != "" {
				log.InfoCtx(c.UserContext(), "%s %s %d %v", method, path, status, latency)
			} else {
				log.Info("%s %s %d %v", method, path, status, latency)
			}
		}

		return err
	}
}

// extractModuleName extracts module name from request path
// extractModuleName 从请求路径提取模块名称
//
// Examples | 示例:
//
//	/testapi/user -> testapi
//	/ws/basic -> ws
//	/health -> http (fallback)
//	/ -> http (fallback)
func extractModuleName(path string) string {
	if path == "" || path == "/" {
		return "http"
	}

	// Remove leading slash | 移除开头的斜杠
	if path[0] == '/' {
		path = path[1:]
	}

	// Get first segment | 获取第一段
	parts := strings.Split(path, "/")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}

	return "http"
}
