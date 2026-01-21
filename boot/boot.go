package boot

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nuohe369/crab/common"
	"github.com/nuohe369/crab/common/config"
	bizErrors "github.com/nuohe369/crab/common/errors"
	"github.com/nuohe369/crab/common/middleware"
	"github.com/nuohe369/crab/common/response"
	"github.com/nuohe369/crab/pkg"
	"github.com/nuohe369/crab/pkg/cron"
	"github.com/nuohe369/crab/pkg/json"
	"github.com/nuohe369/crab/pkg/logger"
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

	// Initialize logger configuration | 初始化日志器配置
	logger.SetConfig(config.GetLogger())

	// Build pkg.Config from common/config
	snowflakeCfg := config.GetSnowflake()
	pkgCfg := pkg.Config{
		SnowflakeMachineID: snowflakeCfg.MachineID,
		Databases:          config.GetDatabases(),
		Redis:              config.GetRedisInstances(),
		JWT:                config.GetJWT(),
	}

	pkg.Init(pkgCfg)
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
		ErrorHandler:          customErrorHandler, // Custom error handler | 自定义错误处理器
		
		// High concurrency configuration | 高并发配置
		Prefork:               false,                // Multi-process mode (enable in production) | 多进程模式（生产环境可开启）
		ReadBufferSize:        8192,                 // Read buffer size | 读缓冲区大小
		WriteBufferSize:       8192,                 // Write buffer size | 写缓冲区大小
		ReadTimeout:           10 * time.Second,     // Read timeout | 读超时
		WriteTimeout:          10 * time.Second,     // Write timeout | 写超时
		IdleTimeout:           120 * time.Second,    // Idle timeout | 空闲超时
		BodyLimit:             4 * 1024 * 1024,      // 4MB body limit | 4MB 请求体限制
		Concurrency:           256 * 1024,           // Max concurrent connections | 最大并发连接数
		DisableKeepalive:      false,                // Keep-alive enabled | 启用长连接
		ReduceMemoryUsage:     false,                // High performance mode | 高性能模式
	})

	initBase()

	// Register global middleware
	middleware.Setup(app)

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

	// Setup graceful shutdown | 设置优雅关闭
	setupGracefulShutdown(targetModules)

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

// customErrorHandler handles errors in a unified way
// customErrorHandler 统一处理错误
func customErrorHandler(c *fiber.Ctx, err error) error {
	// Check if it's a BizError | 检查是否为业务错误
	var bizErr *bizErrors.BizError
	if errors.As(err, &bizErr) {
		// Get HTTP status code from business error code | 从业务错误码获取 HTTP 状态码
		statusCode := getHTTPStatusCode(bizErr.Code)
		c.Status(statusCode)
		
		return c.JSON(response.Response{
			Code: bizErr.Code,
			Msg:  bizErr.Msg,
		})
	}

	// Check if it's a Fiber error | 检查是否为 Fiber 错误
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		c.Status(fiberErr.Code)
		
		// Map common Fiber errors to business error codes | 映射常见 Fiber 错误到业务错误码
		code := mapFiberErrorCode(fiberErr.Code)
		return c.JSON(response.Response{
			Code: code,
			Msg:  fiberErr.Message,
		})
	}

	// Unknown error, return 500 | 未知错误，返回 500
	c.Status(fiber.StatusInternalServerError)
	
	// Log the error for debugging | 记录错误用于调试
	serverLog := logger.NewSystem("server")
	serverLog.Error("Unhandled error: %v", err)
	
	return c.JSON(response.Response{
		Code: response.CodeServerError,
		Msg:  response.CodeServerError.Msg(),
	})
}

// getHTTPStatusCode maps business error codes to HTTP status codes
// getHTTPStatusCode 将业务错误码映射到 HTTP 状态码
func getHTTPStatusCode(code response.Code) int {
	switch {
	// Success | 成功
	case code == response.CodeSuccess:
		return fiber.StatusOK

	// Authentication errors (401) | 认证错误 (401)
	case code == response.CodeUnauth,
		code == response.CodeTokenExpired,
		code == response.CodeTokenInvalid:
		return fiber.StatusUnauthorized

	// Authorization errors (403) | 授权错误 (403)
	case code == response.CodeForbid:
		return fiber.StatusForbidden

	// Not found errors (404) | 未找到错误 (404)
	case code == response.CodeNotFound,
		code == response.CodeUserNotFound:
		return fiber.StatusNotFound

	// Parameter errors (400) | 参数错误 (400)
	case code == response.CodeParamError,
		code == response.CodeParamMissing,
		code == response.CodeParamInvalid:
		return fiber.StatusBadRequest

	// Conflict errors (409) | 冲突错误 (409)
	case code == response.CodeDuplicate,
		code == response.CodeUserExists:
		return fiber.StatusConflict

	// Rate limiting (429) | 限流 (429)
	case code == response.CodeTooManyRequests:
		return fiber.StatusTooManyRequests

	// Server errors (500) | 服务器错误 (500)
	case code == response.CodeServerError,
		code == response.CodeDBError,
		code == response.CodeRedisError:
		return fiber.StatusInternalServerError

	// Business errors (422) | 业务错误 (422)
	case code >= 4000 && code < 5000:
		return fiber.StatusUnprocessableEntity

	// Default to 500 for unknown errors | 未知错误默认返回 500
	default:
		return fiber.StatusInternalServerError
	}
}

// mapFiberErrorCode maps Fiber HTTP status codes to business error codes
// mapFiberErrorCode 将 Fiber HTTP 状态码映射到业务错误码
func mapFiberErrorCode(statusCode int) response.Code {
	switch statusCode {
	case fiber.StatusBadRequest:
		return response.CodeParamError
	case fiber.StatusUnauthorized:
		return response.CodeUnauth
	case fiber.StatusForbidden:
		return response.CodeForbid
	case fiber.StatusNotFound:
		return response.CodeNotFound
	case fiber.StatusConflict:
		return response.CodeDuplicate
	case fiber.StatusTooManyRequests:
		return response.CodeTooManyRequests
	case fiber.StatusInternalServerError:
		return response.CodeServerError
	default:
		return response.CodeError
	}
}

// setupGracefulShutdown sets up graceful shutdown for the application
// setupGracefulShutdown 为应用程序设置优雅关闭
func setupGracefulShutdown(targetModules []Module) {
	// Create channel to listen for interrupt signals | 创建通道监听中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		// Wait for interrupt signal | 等待中断信号
		sig := <-quit
		serverLog := logger.NewSystem("server")
		serverLog.Info("Received shutdown signal: %v", sig)
		serverLog.Info("Starting graceful shutdown...")

		// Create shutdown context with timeout | 创建带超时的关闭上下文
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Shutdown HTTP server | 关闭 HTTP 服务器
		serverLog.Info("Shutting down HTTP server...")
		if err := app.ShutdownWithContext(ctx); err != nil {
			serverLog.Error("HTTP server shutdown error: %v", err)
		} else {
			serverLog.Info("HTTP server stopped")
		}

		// Stop cron scheduler | 停止定时任务调度器
		serverLog.Info("Stopping cron scheduler...")
		cron.Stop()
		serverLog.Info("Cron scheduler stopped")

		// Stop modules | 停止模块
		serverLog.Info("Stopping modules...")
		for _, m := range targetModules {
			if err := m.Stop(); err != nil {
				serverLog.Error("Module %s stop error: %v", m.Name(), err)
			} else {
				serverLog.Info("Module %s stopped", m.Name())
			}
		}

		// Close database connections | 关闭数据库连接
		serverLog.Info("Closing database connections...")
		pgsql.Close()
		serverLog.Info("Database connections closed")

		// Close pkg resources | 关闭 pkg 资源
		serverLog.Info("Closing pkg resources...")
		pkg.Close()
		serverLog.Info("Pkg resources closed")

		serverLog.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		serverLog.Info("✅ Graceful shutdown completed")
		serverLog.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		// Exit the application | 退出应用程序
		os.Exit(0)
	}()
}
