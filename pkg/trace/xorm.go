package trace

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"xorm.io/xorm/contexts"
)

// XormHook xorm tracing hook
type XormHook struct{}

// BeforeProcess creates span before execution
func (h *XormHook) BeforeProcess(c *contexts.ContextHook) (context.Context, error) {
	tracer := otel.Tracer("xorm")
	ctx, span := tracer.Start(c.Ctx, "SQL",
		trace.WithSpanKind(trace.SpanKindClient),
	)

	// Record SQL statement
	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.statement", c.SQL),
	)

	// Record parameters (limit length to avoid oversized)
	if len(c.Args) > 0 {
		args := fmt.Sprintf("%v", c.Args)
		if len(args) > 500 {
			args = args[:500] + "..."
		}
		span.SetAttributes(attribute.String("db.args", args))
	}

	return ctx, nil
}

// AfterProcess ends span after execution
func (h *XormHook) AfterProcess(c *contexts.ContextHook) error {
	span := trace.SpanFromContext(c.Ctx)
	if span == nil {
		return nil
	}
	defer span.End()

	// Record execution time
	span.SetAttributes(
		attribute.Int64("db.execution_time_ms", c.ExecuteTime.Milliseconds()),
	)

	// Record error
	if c.Err != nil {
		span.RecordError(c.Err)
		span.SetStatus(codes.Error, c.Err.Error())
	}

	return nil
}
