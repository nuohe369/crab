// Package ws WebSocket example module
//
// Demonstrates various usages of pkg/ws, accessible via different paths after startup:
//
//   - /ws/basic       - Basic usage
//   - /ws/multiuser   - Multi-user targeted messaging
//   - /ws/callback    - Callback handling
//   - /ws/cluster     - Redis cluster mode
//
// Test: websocat ws://localhost:3000/ws/basic
package ws

import (
	"github.com/nuohe369/crab/boot"
	"github.com/nuohe369/crab/module/ws/example_01_basic"
	"github.com/nuohe369/crab/module/ws/example_02_multiuser"
	"github.com/nuohe369/crab/module/ws/example_03_callback"
	"github.com/nuohe369/crab/module/ws/example_04_cluster"
	"github.com/nuohe369/crab/module/ws/example_05_service"
)

func init() {
	boot.Register(&Module{})
}

type Module struct{}

func (m *Module) Name() string { return "ws" }

func (m *Module) Models() []any { return nil }

func (m *Module) Init(ctx *boot.ModuleContext) error {
	example_01_basic.Setup(ctx.Router)
	example_02_multiuser.Setup(ctx.Router)
	example_03_callback.Setup(ctx.Router)
	example_04_cluster.Setup(ctx.Router)
	example_05_service.Setup(ctx.Router)

	return nil
}

func (m *Module) Start() error { return nil }
func (m *Module) Stop() error  { return nil }
