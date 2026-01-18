package pkg

import (
	"context"
	"log"

	"github.com/nuohe369/crab/common/config"
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

// Init initializes the infrastructure layer.
func Init() {
	log.Println("Initializing pkg infrastructure...")

	// Initialize Snowflake ID generator
	snowflakeCfg := config.GetSnowflake()
	machineID := snowflakeCfg.MachineID
	if machineID == 0 {
		machineID = 1 // Default to 1 if not configured
	}
	if err := snowflake.Init(machineID); err != nil {
		log.Fatalf("Snowflake initialization failed: %v", err)
	}
	log.Println("  ✓ Snowflake initialized")

	// Initialize databases (required)
	databases := config.GetDatabases()
	if len(databases) == 0 {
		log.Fatal("No database configured")
	}

	// Initialize default database first (name: "default" or "")
	var defaultInitialized bool
	var defaultName string
	for _, name := range []string{"default", ""} {
		if cfg, ok := databases[name]; ok {
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
		for name, cfg := range databases {
			pgsql.MustInit(cfg)
			log.Printf("  ✓ PostgreSQL initialized (default: %s)", name)
			defaultName = name
			break
		}
	}

	// Register all databases by name (including default)
	for name, cfg := range databases {
		if err := pgsql.InitNamed(name, cfg); err != nil {
			log.Fatalf("PostgreSQL initialization failed (%s): %v", name, err)
		}
		// Don't log the default database again
		if name != defaultName {
			log.Printf("  ✓ PostgreSQL initialized (%s)", name)
		}
	}

	// Set logger for all databases after initialization | 初始化完成后为所有数据库设置日志器
	pgsql.SetLogger(logger.NewWithName[struct{}]("sql"))

	// Initialize Redis (required)
	redisInstances := config.GetRedisInstances()
	if len(redisInstances) == 0 {
		log.Fatal("No Redis configured")
	}

	// Initialize all Redis instances
	for name, cfg := range redisInstances {
		if err := redis.InitNamed(name, cfg); err != nil {
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
	mqCfg := config.GetMQ()
	if mqCfg.Driver != "" {
		if err := mq.Init(mqCfg); err != nil {
			log.Printf("  ⚠ Message queue initialization failed: %v", err)
		} else {
			log.Println("  ✓ Message queue initialized")
		}
	} else {
		log.Println("  - Message queue not configured, skipping")
	}

	// Initialize JWT (optional)
	jwtCfg := config.GetJWT()
	if jwtCfg.Secret != "" {
		jwt.Init(jwtCfg)
		log.Println("  ✓ JWT initialized")
	} else {
		log.Println("  - JWT not configured, skipping")
	}

	// Initialize metrics (optional)
	metricsCfg := config.GetMetrics()
	if metricsCfg.Enabled {
		metrics.Init(metricsCfg)
		log.Println("  ✓ Metrics initialized")
	} else {
		log.Println("  - Metrics not enabled, skipping")
	}

	// Initialize storage (optional)
	storageCfg := config.GetStorage()
	if storageCfg.Driver != "" {
		if err := storage.Init(storageCfg); err != nil {
			log.Printf("  ⚠ Storage initialization failed: %v", err)
		} else {
			log.Println("  ✓ Storage initialized")
		}
	} else {
		log.Println("  - Storage not configured, skipping")
	}

	// Initialize distributed tracing (optional)
	traceCfg := config.GetTrace()
	if traceCfg.Endpoint != "" {
		shutdown, err := trace.Init(traceCfg)
		if err != nil {
			log.Printf("  ⚠ Trace initialization failed: %v", err)
		} else {
			traceShutdown = shutdown
			// Inject xorm hook to all databases (avoid duplicate injection)
			hook := &trace.XormHook{}
			injected := make(map[*pgsql.Client]bool)

			// Inject to all named databases
			for name := range databases {
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
