package pgsql

import (
	"context"
	"time"

	"xorm.io/xorm/contexts"
)

// SlowQueryThreshold defines the threshold for slow query detection (default: 1 second)
// SlowQueryThreshold 定义慢查询检测阈值（默认：1秒）
var SlowQueryThreshold = time.Second

// SlowQueryHook monitors slow queries
// SlowQueryHook 监控慢查询
type SlowQueryHook struct {
	threshold time.Duration
	logger    Logger
}

// NewSlowQueryHook creates a slow query monitoring hook
// NewSlowQueryHook 创建慢查询监控钩子
func NewSlowQueryHook(threshold time.Duration, logger Logger) *SlowQueryHook {
	if threshold <= 0 {
		threshold = SlowQueryThreshold
	}
	if logger == nil {
		logger = &defaultLogger{prefix: "slowquery"}
	}
	return &SlowQueryHook{
		threshold: threshold,
		logger:    logger,
	}
}

// BeforeProcess records start time
// BeforeProcess 记录开始时间
func (h *SlowQueryHook) BeforeProcess(c *contexts.ContextHook) (context.Context, error) {
	c.Ctx = context.WithValue(c.Ctx, "start_time", time.Now())
	return c.Ctx, nil
}

// AfterProcess checks query duration and logs slow queries
// AfterProcess 检查查询耗时并记录慢查询
func (h *SlowQueryHook) AfterProcess(c *contexts.ContextHook) error {
	startTime, ok := c.Ctx.Value("start_time").(time.Time)
	if !ok {
		return nil
	}

	duration := time.Since(startTime)
	if duration >= h.threshold {
		h.logger.Warn("Slow query detected: duration=%v, sql=%s, args=%v",
			duration, c.SQL, c.Args)
	}

	return nil
}
