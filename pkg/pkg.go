package pkg

import (
	"context"
	"log"

	"github.com/nuohe369/crab/pkg/cache"
	"github.com/nuohe369/crab/pkg/cron"
	"github.com/nuohe369/crab/pkg/jwt"
	"github.com/nuohe369/crab/pkg/logger"
	"github.com/nuohe369/crab/pkg/pgsql"
	"github.com/nuohe369/crab/pkg/redis"
	"github.com/nuohe369/crab/pkg/snowflake"
)

var _ func(context.Context) error // placeholder for removed traceShutdown

// Config holds all infrastructure configuration
type Config struct {
	SnowflakeMachineID int64
	Databases          map[string]pgsql.Config
	Redis              map[string]redis.Config
	JWT                jwt.Config
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

	// Initialize JWT (optional)
	if cfg.JWT.Secret != "" {
		jwt.Init(cfg.JWT)
		log.Println("  ✓ JWT initialized")
	} else {
		log.Println("  - JWT not configured, skipping")
	}

	log.Println("Infrastructure initialization completed")
}

// Close shuts down the infrastructure.
func Close() {
	pgsql.Close()
	redis.Close()
}
