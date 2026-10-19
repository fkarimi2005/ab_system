package jobs

import (
	"AB_system/internal/domain/service"
	"context"
	"log/slog"
	"time"
)

func StartJob(ctx context.Context, reportSvc *service.ReportService, guardrailSvc *service.GuardrailService) {
	t := time.NewTicker(time.Minute)
	t.Stop()
	for {
		select {
		case <-t.C:
			if err := reportSvc.Update(ctx); err != nil {
				slog.Error("update reports", "err", err)
			}
			if err := guardrailSvc.Check(ctx); err != nil {
				slog.Error("check guardrails", "err", err)
			}
		case <-ctx.Done():
			return
		}
	}

}
