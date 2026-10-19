package middlewares

import (
	"AB_system/internal/http/observability"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log/slog"
	"time"
)

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

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		ctx := c.Request.Context()
		status := c.Writer.Status()

		level := slog.LevelInfo
		switch {
		case status >= 500:
			level = slog.LevelError
		case status >= 400:
			level = slog.LevelWarn
		}

		attrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", c.FullPath()),
			slog.Int("status", status),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
			slog.String("trace_id", observability.GetTraceID(ctx)),
		}
		if actor, ok := ActorFrom(c); ok {
			attrs = append(attrs, slog.String("user_id", actor.ID.String()))
		}
		if last := c.Errors.Last(); last != nil {
			attrs = append(attrs, slog.String("err", last.Err.Error()))
		}
		slog.LogAttrs(ctx, level, "http request", attrs...)
	}
}
