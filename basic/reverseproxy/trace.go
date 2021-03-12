package reverseproxy

import (
	"context"
	"time"
)

type traceContextKey string

const traceKey = traceContextKey("trace")

type TraceContext struct {
	RequestTime time.Time
	RequestId   string
}

func WithTraceContext(ctx context.Context, traceCtx *TraceContext) context.Context {
	return context.WithValue(ctx, traceKey, traceCtx)
}

func FromTraceContext(ctx context.Context) (*TraceContext, bool) {
	tctx, ok := ctx.Value(traceKey).(*TraceContext)
	return tctx, ok
}
