package boot

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/nuohe369/crab/common"
	"github.com/nuohe369/crab/common/config"
	"github.com/nuohe369/crab/common/middleware"
	"github.com/nuohe369/crab/pkg"
	"github.com/nuohe369/crab/pkg/cron"
	"github.com/nuohe369/crab/pkg/json"
	"github.com/nuohe369/crab/pkg/metrics"
	"github.com/nuohe369/crab/pkg/pgsql"
)

// Module defines the interface that all modules must implement.
type Module interface {
	Name() string
	Models() []any // database models required by this module
	Init(ctx *ModuleContext) error
	Start() error
	Stop() error
}

var (
	modules []Module
	app     *fiber.App
)

// Register registers a module to the application.
func Register(m Module) {
	modules = append(modules, m)
}

// App returns the Fiber application instance.
func App() *fiber.App {
	return app
}

// GetModule retrieves a module by its name.
func GetModule(name string) Module {
	for _, m := range modules {
		if m.Name() == name {
			return m
		}
	}
	return nil
}

// GetAllModuleNames returns the names of all registered modules.
func GetAllModuleNames() []string {
	names := make([]string, len(modules))
	for i, m := range modules {
		names[i] = m.Name()
	}
	return names
}

// initBase initializes the base infrastructure including configuration and database.
func initBase() {
	if secretKey != "" {
		config.SetDecryptKey(secretKey)
	}
	config.MustLoad("config.toml")

	pkg.Init()
	common.Init()
}

// migrateModels performs database migration for the specified modules.
func migrateModels(targetModules []Module) {
	if pgsql.Get() == nil {
		return
	}

	// Collect all models from modules (deduplicated)
	seen := make(map[string]bool)
	var models []any
	for _, m := range targetModules {
		for _, md := range m.Models() {
			key := fmt.Sprintf("%T", md)
			if !seen[key] {
				seen[key] = true
				models = append(models, md)
			}
		}
	}

	if len(models) == 0 {
		log.Println("No models to migrate")
		return
	}

	// Execute migration
	if err := pgsql.Get().Engine().Sync2(models...); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}
	log.Printf("Database migration completed, %d models migrated", len(models))
}

// initAfterMigrate performs initialization after database migration.
func initAfterMigrate() {
	log.Println("Initialization completed")
}

// Run starts all registered modules.
func Run(addr string) {
	RunModules(nil, addr)
}

// RunModules starts the specified modules. If moduleNames is nil or empty, all modules are started.
func RunModules(moduleNames []string, addr string) {
	app = fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	initBase()

	// Register global middleware
	middleware.Setup(app)

	// Register metrics middleware and routes
	if metrics.Enabled() {
		app.Use(metrics.Middleware())
		app.Get(metrics.Path(), metrics.Handler())
	}

	// Determine which modules to start
	var targetModules []Module
	if len(moduleNames) == 0 {
		targetModules = modules
	} else {
		nameSet := make(map[string]bool)
		for _, name := range moduleNames {
			nameSet[name] = true
		}
		for _, m := range modules {
			if nameSet[m.Name()] {
				targetModules = append(targetModules, m)
			}
		}
	}

	if len(targetModules) == 0 {
		log.Fatal("No modules found to start")
	}

	// Migrate models declared by modules
	migrateModels(targetModules)

	// Post-migration initialization
	initAfterMigrate()

	// Initialize modules
	for _, m := range targetModules {
		group := app.Group("/" + m.Name())
		ctx := NewModuleContext(group, nil)
		if err := m.Init(ctx); err != nil {
			log.Fatalf("Module %s initialization failed: %v", m.Name(), err)
		}
		log.Printf("Module %s initialized", m.Name())
	}

	// Start modules
	for _, m := range targetModules {
		if err := m.Start(); err != nil {
			log.Fatalf("Module %s start failed: %v", m.Name(), err)
		}
		log.Printf("Module %s started", m.Name())
	}

	// Start cron scheduler
	cron.Start()

	// Start HTTP server
	go app.Listen(addr)
	log.Printf("Server started: %s (modules: %s)", addr, strings.Join(getModuleNames(targetModules), ", "))

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	cron.Stop()
	for i := len(targetModules) - 1; i >= 0; i-- {
		if err := targetModules[i].Stop(); err != nil {
			log.Printf("Module %s shutdown failed: %v", targetModules[i].Name(), err)
		}
	}
	app.Shutdown()
	log.Println("Server stopped")
}

// RunService starts a service by name as defined in the configuration file.
func RunService(serviceName string, addrOverride string) {
	// Load configuration first
	if secretKey != "" {
		config.SetDecryptKey(secretKey)
	}
	config.MustLoad("config.toml")

	svc := config.GetService(serviceName)
	if svc == nil {
		log.Fatalf("Service %s is not defined in configuration file", serviceName)
	}

	addr := svc.Addr
	if addrOverride != "" {
		addr = addrOverride
	}

	log.Printf("Starting service: %s", serviceName)

	RunModules(svc.Modules, addr)
}

func getModuleNames(mods []Module) []string {
	names := make([]string, len(mods))
	for i, m := range mods {
		names[i] = m.Name()
	}
	return names
}
