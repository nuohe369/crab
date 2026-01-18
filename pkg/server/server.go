// Package server provides Fiber HTTP server wrapper
// Package server 提供 Fiber HTTP 服务器包装器
package server

import (
	"github.com/gofiber/fiber/v2"
)

// Server represents a Fiber HTTP server wrapper
// Server 表示 Fiber HTTP 服务器包装器
type Server struct {
	app *fiber.App // Fiber application instance | Fiber 应用实例
}

// New creates a Fiber server
// New 创建 Fiber 服务器
func New(config ...fiber.Config) *Server {
	var cfg fiber.Config
	if len(config) > 0 {
		cfg = config[0]
	}
	return &Server{
		app: fiber.New(cfg),
	}
}

// App returns the Fiber instance
// App 返回 Fiber 实例
func (s *Server) App() *fiber.App {
	return s.app
}

// Start starts the server
// Start 启动服务器
func (s *Server) Start(addr string) error {
	go s.app.Listen(addr)
	return nil
}

// Stop stops the server
// Stop 停止服务器
func (s *Server) Stop() error {
	return s.app.Shutdown()
}
