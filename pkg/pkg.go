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

	// Initialize Snowflake ID generator (machine ID defaults to 1)
	if err := snowflake.Init(1); err != nil {
		log.Fatalf("Snowflake initialization failed: %v", err)
	}
	log.Println("  ✓ Snowflake initialized")

	// Initialize database
	pgsql.MustInit(config.GetDatabase())
	// Inject logger to pgsql
	pgsql.SetLogger(logger.NewSystem("xorm"))
	log.Println("  ✓ PostgreSQL initialized")

	// Initialize Redis
	redis.MustInit(config.GetRedis())
	log.Println("  ✓ Redis initialized")

	// Initialize cache (two-level cache: local + Redis)
	cache.Init(redis.Get())
	log.Println("  ✓ Cache initialized")

	// Initialize cron scheduler
	cron.Init(redis.Get())
	log.Println("  ✓ Cron initialized")

	// Initialize message queue
	mq.MustInit(config.GetMQ())
	log.Println("  ✓ Message queue initialized")

	// Initialize JWT
	jwt.MustInit(config.GetJWT())
	log.Println("  ✓ JWT initialized")

	// Initialize metrics
	metrics.Init(config.GetMetrics())

	// Initialize storage
	storage.MustInit(config.GetStorage())

	// Initialize distributed tracing
	traceCfg := config.GetTrace()
	if traceCfg.Endpoint != "" {
		shutdown, err := trace.Init(traceCfg)
		if err != nil {
			log.Printf("  ⚠ Trace initialization failed: %v", err)
		} else {
			traceShutdown = shutdown
			// Inject xorm hook
			if pgsql.Get() != nil {
				pgsql.Get().AddHook(&trace.XormHook{})
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
