package repository

import (
	"AB_system/internal/http/observability"
	"context"
	"errors"
	"log/slog"

	"gorm.io/gorm"
)

func CheckError(ctx context.Context, op string, err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		slog.WarnContext(ctx, "record not found",
			"op", op,
			"trace_id", observability.GetTraceID(ctx),
		)
		return TranslateGormError(err)
	}

	slog.ErrorContext(ctx, "repository error",
		"op", op,
		"trace_id", observability.GetTraceID(ctx),
		"err", err,
	)
	return TranslateGormError(err)
}
