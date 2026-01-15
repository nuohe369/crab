package boot

import (
	"github.com/gofiber/fiber/v2"
)

// ModuleContext provides context for module initialization.
type ModuleContext struct {
	Router fiber.Router // route group for the module
	Config any          // optional module-specific configuration
}

// NewModuleContext creates a new module context.
func NewModuleContext(router fiber.Router, config any) *ModuleContext {
	return &ModuleContext{
		Router: router,
		Config: config,
	}
}
