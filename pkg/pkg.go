package pkg

import (
	"context"
	"log"

	"github.com/nuohe369/crab/pkg/cache"
	"github.com/nuohe369/crab/pkg/cron"
	"github.com/nuohe369/crab/pkg/jwt"
	"github.com/nuohe369/crab/pkg/logger"
	"github.com/nuohe369/crab/pkg/metrics"
	"github.com/nuohe369/crab/pkg/mq"
	"github.com/nuohe369/crab/pkg/pgsql"
	"github.com/nuohe369/crab/pkg/redis"
	"github.com/nuohe369/crab/pkg/snowflake"
	"github.com/nuohe369/crab/pkg/storage"
	"github.com/nuohe369/crab/pkg/trace"
)

var traceShutdown func(context.Context) error

// Config holds all infrastructure configuration
type Config struct {
	SnowflakeMachineID int64
	Databases          map[string]pgsql.Config
	Redis              map[string]redis.Config
	MQ                 mq.Config
	JWT                jwt.Config
	Metrics            metrics.Config
	Storage            storage.Config
	Trace              trace.Config
}

// Init initializes the infrastructure layer with provided configuration.
func Init(cfg Config) {
	log.Println("Initializing pkg infrastructure...")

	// Initialize Snowflake ID generator
	machineID := cfg.SnowflakeMachineID
	if machineID == 0 {
		machineID = 1 // Default to 1 if not configured
	}
	if err := snowflake.Init(machineID); err != nil {
		log.Fatalf("Snowflake initialization failed: %v", err)
	}
	log.Println("  ✓ Snowflake initialized")

	// Initialize databases (required)
	if len(cfg.Databases) == 0 {
		log.Fatal("No database configured")
	}

	// Initialize default database first (name: "default" or "")
	var defaultInitialized bool
	var defaultName string
	for _, name := range []string{"default", ""} {
		if cfg, ok := cfg.Databases[name]; ok {
			pgsql.MustInit(cfg)
			if name == "" {
				log.Printf("  ✓ PostgreSQL initialized (default)")
			} else {
				log.Printf("  ✓ PostgreSQL initialized (default: %s)", name)
			}
			defaultInitialized = true
			defaultName = name
			break
		}
	}

	// If no default found, use first database as default
	if !defaultInitialized {
		for name, dbCfg := range cfg.Databases {
			pgsql.MustInit(dbCfg)
			log.Printf("  ✓ PostgreSQL initialized (default: %s)", name)
			defaultName = name
			break
		}
	}

	// Register all databases by name (including default)
	for name, dbCfg := range cfg.Databases {
		if err := pgsql.InitNamed(name, dbCfg); err != nil {
			log.Fatalf("PostgreSQL initialization failed (%s): %v", name, err)
		}
		// Don't log the default database again
		if name != defaultName {
			log.Printf("  ✓ PostgreSQL initialized (%s)", name)
		}
	}

	// Set logger for all databases after initialization | 初始化完成后为所有数据库设置日志器
	pgsql.SetLogger(logger.NewWithName("sql"))

	// SnowflakeID converter is automatically recognized by XORM through FromDB/ToDB methods
	// SnowflakeID 转换器通过 FromDB/ToDB 方法被 XORM 自动识别
	log.Println("  ✓ SnowflakeID converter enabled (automatic)")

	// Initialize Redis (required)
	if len(cfg.Redis) == 0 {
		log.Fatal("No Redis configured")
	}

	// Initialize all Redis instances
	for name, redisCfg := range cfg.Redis {
		if err := redis.InitNamed(name, redisCfg); err != nil {
			log.Fatalf("Redis initialization failed (%s): %v", name, err)
		}
		if name == "default" || name == "" {
			log.Printf("  ✓ Redis initialized (default: %s)", name)
		} else {
			log.Printf("  ✓ Redis initialized (%s)", name)
		}
	}

	// Initialize cache (optional, depends on Redis)
	cache.Init(redis.Get())
	log.Println("  ✓ Cache initialized")

	// Initialize cron scheduler (optional)
	cron.Init(redis.Get())
	log.Println("  ✓ Cron initialized")

	// Initialize message queue (optional)
	if cfg.MQ.Driver != "" {
		if err := mq.Init(cfg.MQ); err != nil {
			log.Printf("  ⚠ Message queue initialization failed: %v", err)
		} else {
			log.Println("  ✓ Message queue initialized")
		}
	} else {
		log.Println("  - Message queue not configured, skipping")
	}

	// Initialize JWT (optional)
	if cfg.JWT.Secret != "" {
		jwt.Init(cfg.JWT)
		log.Println("  ✓ JWT initialized")
	} else {
		log.Println("  - JWT not configured, skipping")
	}

	// Initialize metrics (optional)
	if cfg.Metrics.Enabled {
		metrics.Init(cfg.Metrics)
		log.Println("  ✓ Metrics initialized")
	} else {
		log.Println("  - Metrics not enabled, skipping")
	}

	// Initialize storage (optional)
	if cfg.Storage.Driver != "" {
		if err := storage.Init(cfg.Storage); err != nil {
			log.Printf("  ⚠ Storage initialization failed: %v", err)
		} else {
			log.Println("  ✓ Storage initialized")
		}
	} else {
		log.Println("  - Storage not configured, skipping")
	}

	// Initialize distributed tracing (optional)
	if cfg.Trace.Endpoint != "" {
		shutdown, err := trace.Init(cfg.Trace)
		if err != nil {
			log.Printf("  ⚠ Trace initialization failed: %v", err)
		} else {
			traceShutdown = shutdown
			// Inject xorm hook to all databases (avoid duplicate injection)
			hook := &trace.XormHook{}
			injected := make(map[*pgsql.Client]bool)

			// Inject to all named databases
			for name := range cfg.Databases {
				if db := pgsql.Get(name); db != nil && !injected[db] {
					db.AddHook(hook)
					injected[db] = true
				}
			}

			// Ensure default database is also injected (if not already)
			if defaultDB := pgsql.Get(); defaultDB != nil && !injected[defaultDB] {
				defaultDB.AddHook(hook)
			}

			log.Println("  ✓ Trace initialized")
		}
	} else {
		log.Println("  - Trace not configured, skipping")
	}

	log.Println("Infrastructure initialization completed")
}

// Close shuts down the infrastructure.
func Close() {
	if traceShutdown != nil {
		traceShutdown(context.Background())
	}
	pgsql.Close()
	redis.Close()
	mq.Close()
}
