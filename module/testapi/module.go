package testapi

import (
	"github.com/nuohe369/crab/boot"
	"github.com/nuohe369/crab/module/testapi/internal/handler"
)

func init() {
	boot.Register(&Module{})
}

type Module struct{}

func (m *Module) Name() string {
	return "testapi"
}

func (m *Module) Models() []any {
	return nil
}

func (m *Module) Init(ctx *boot.ModuleContext) error {
	handler.Setup(ctx.Router)
	return nil
}

func (m *Module) Start() error {
	return nil
}

func (m *Module) Stop() error {
	return nil
}
