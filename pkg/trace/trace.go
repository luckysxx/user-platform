package trace

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const (
	TraceIDKey    contextKey = "X-Trace-Id"
	HeaderTraceID string     = "x-trace-id"
)

// NewTraceID generates a fresh UUID for tracing requests.
func NewTraceID() string {
	return uuid.New().String()
}

// IntoContext injects the given trace ID into the context.
func IntoContext(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// FromContext extracts the trace ID from the context. Returns empty string if not found.
func FromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if val, ok := ctx.Value(TraceIDKey).(string); ok {
		return val
	}
	return ""
}
