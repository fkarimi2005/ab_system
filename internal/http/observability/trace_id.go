package observability

import "context"

type ctxKey struct{}

const TraceHeader = "X-Trace-Id"

func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

func GetTraceID(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}
