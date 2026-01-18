package boot

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/nuohe369/crab/common"
	"github.com/nuohe369/crab/common/config"
	"github.com/nuohe369/crab/common/middleware"
	"github.com/nuohe369/crab/pkg"
	"github.com/nuohe369/crab/pkg/cron"
	"github.com/nuohe369/crab/pkg/json"
	"github.com/nuohe369/crab/pkg/logger"
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

	// Group models by database engine
	type dbNamer interface {
		DBName() string
	}

	dbGroups := make(map[*pgsql.Client][]any)
	var skippedModels []string

	for _, md := range models {
		var db *pgsql.Client

		// Check if model specifies a database
		if namer, ok := md.(dbNamer); ok {
			dbName := namer.DBName()
			if dbName != "" {
				db = pgsql.Get(dbName)
				if db == nil {
					skippedModels = append(skippedModels, fmt.Sprintf("%T (database '%s' not configured)", md, dbName))
					continue
				}
			} else {
				// Model has DBName() but returns empty, use default
				db = pgsql.Get()
			}
		} else {
			// Model doesn't have DBName(), use default
			db = pgsql.Get()
		}

		if db == nil {
			skippedModels = append(skippedModels, fmt.Sprintf("%T (no default database)", md))
			continue
		}

		dbGroups[db] = append(dbGroups[db], md)
	}

	// Report skipped models
	if len(skippedModels) > 0 {
		log.Printf("⚠ Skipped %d models:", len(skippedModels))
		for _, model := range skippedModels {
			log.Printf("  - %s", model)
		}
	}

	// Execute migration for each database
	totalMigrated := 0
	for db, mds := range dbGroups {
		if err := db.Engine().Sync2(mds...); err != nil {
			log.Fatalf("Database migration failed: %v", err)
		}
		totalMigrated += len(mds)
	}

	log.Printf("Migrated %d models across %d databases", totalMigrated, len(dbGroups))
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
		DisableStartupMessage: true, // Disable Fiber's default startup message | 禁用 Fiber 的默认启动消息
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
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

	// Validate module dependencies and filter out modules with missing dependencies
	// 验证模块依赖并过滤掉缺少依赖的模块
	strictMode := config.IsStrictDependencyCheck()
	targetModules = ValidateAndFilterModules(targetModules, strictMode)

	if len(targetModules) == 0 {
		log.Fatal("❌ No modules available to start after dependency validation")
	}

	// Migrate models declared by modules
	databases := config.GetDatabases()
	// Check if any database has auto_migrate enabled
	autoMigrate := false
	for _, dbCfg := range databases {
		if dbCfg.AutoMigrate {
			autoMigrate = true
			break
		}
	}

	if autoMigrate {
		migrateModels(targetModules)
	} else {
		log.Println("Database auto migration is disabled")
	}

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

	// Print startup information | 打印启动信息
	printStartupInfo(addr, targetModules)

	// Start HTTP server (blocking)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
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

// printStartupInfo prints server startup information using our logger
// printStartupInfo 使用我们的日志器打印服务器启动信息
func printStartupInfo(addr string, modules []Module) {
	serverLog := logger.NewSystem("server")

	appCfg := config.GetApp()

	serverLog.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	serverLog.Info("🚀 %s v%s", appCfg.Name, appCfg.Version)
	serverLog.Info("📍 Address: %s", addr)
	serverLog.Info("📦 Modules: %s", strings.Join(getModuleNames(modules), ", "))
	serverLog.Info("🔧 Handlers: %d", app.HandlersCount())
	serverLog.Info("🆔 PID: %d", os.Getpid())
	serverLog.Info("🌍 Environment: %s", appCfg.Env)

	// Show strict mode status | 显示严格模式状态
	if config.IsStrictDependencyCheck() {
		serverLog.Info("🛡️  Strict dependency check: enabled")
	} else {
		serverLog.Info("🔧 Strict dependency check: disabled")
	}

	serverLog.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	serverLog.Info("Server is ready to accept connections")
}
