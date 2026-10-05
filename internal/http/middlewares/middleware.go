package middlewares

import (
	"AB_system/internal/http/observability"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log/slog"
	"time"
)

// TraceID берёт X-Trace-Id из запроса или создаёт новый и кладёт его в контекст.
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(observability.TraceHeader)
		if id == "" {
			id = uuid.NewString()
		}
		c.Header(observability.TraceHeader, id)
		c.Request = c.Request.WithContext(
			observability.WithTraceID(c.Request.Context(), id),
		)
		c.Next()
	}
}

// RequestLogger пишет одну JSON-строку на каждый запрос.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		ctx := c.Request.Context()
		slog.InfoContext(ctx, "http request",
			"method", c.Request.Method,
			"path", c.FullPath(),
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"trace_id", observability.GetTraceID(ctx),
		)
	}
}
