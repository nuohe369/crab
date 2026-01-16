package server

import (
	"github.com/gofiber/fiber/v2"
)

// Server Fiber HTTP server wrapper
type Server struct {
	app *fiber.App
}

// New creates a Fiber server
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
func (s *Server) App() *fiber.App {
	return s.app
}

// Start starts the server
func (s *Server) Start(addr string) error {
	go s.app.Listen(addr)
	return nil
}

// Stop stops the server
func (s *Server) Stop() error {
	return s.app.Shutdown()
}
