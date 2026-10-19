package service

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/domain/repository"
	"AB_system/pkg/errs"
	"context"
	"errors"
	"github.com/google/uuid"
	"log/slog"
	"time"
)

type GuardrailService struct {
	eventRepo      repository.EventRepository
	experimentRepo repository.ExperimentRepository
	metricRepo     repository.MetricRepository
	guardrailRepo  repository.GuardrailRepository
}

func NewGuardrailService(
	eventRepo repository.EventRepository,
	experimentRepo repository.ExperimentRepository,
	metricRepo repository.MetricRepository,
	guardRepo repository.GuardrailRepository) *GuardrailService {
	return &GuardrailService{
		eventRepo:      eventRepo,
		experimentRepo: experimentRepo,
		metricRepo:     metricRepo,
		guardrailRepo:  guardRepo,
	}
}

func (s *GuardrailService) computeMetric(ctx context.Context, experimentID, variantID, metricID uuid.UUID, from, to time.Time) (float64, bool, error) {
	metric, err := s.metricRepo.GetMetricByID(ctx, metricID)
	if err != nil {
		return 0, false, err
	}
	clickCount, err := s.eventRepo.CountEventsByEventTypeForGuard(ctx, "click", experimentID, variantID, from, to)
	if err != nil {
		return 0, false, err
	}
	purchaseCount, err := s.eventRepo.CountEventsByEventTypeForGuard(ctx, "purchase", experimentID, variantID, from, to)
	if err != nil {
		return 0, false, err
	}
	showCount, err := s.eventRepo.CountEventsByEventTypeForGuard(ctx, "shown", experimentID, variantID, from, to)
	if err != nil {
		return 0, false, err
	}
	requestCount, err := s.eventRepo.CountEventsByEventTypeForGuard(ctx, "api_request", experimentID, variantID, from, to)
	if err != nil {
		return 0, false, err
	}
	errorCount, err := s.eventRepo.CountEventsByEventTypeForGuard(ctx, "error", experimentID, variantID, from, to)
	if err != nil {
		return 0, false, err
	}

	switch metric.Key {
	case "error_rate":
		if requestCount < 100 {
			return 0, false, nil
		}
		return float64(errorCount) / float64(requestCount), true, nil
	case "purchase_rate":
		if clickCount < 100 {
			return 0, false, nil
		}
		return float64(purchaseCount) / float64(clickCount), true, nil
	case "ctr":
		if showCount < 100 {
			return 0, false, nil
		}
		return float64(clickCount) / float64(showCount), true, nil
	}
	return 0, false, nil
}

func (s *GuardrailService) Check(ctx context.Context) error {
	experiments, err := s.experimentRepo.ListRunningExperiments(ctx)
	if err != nil {
		return err
	}
	to := time.Now().UTC()
experimentLoop:
	for _, e := range experiments {
		if e.RevertedToControl {
			continue
		}
		for _, v := range e.Variants {
			if v.IsControl {
				continue
			}
			for _, g := range e.Guardrails {
				from := to.Add(-time.Duration(g.WindowSeconds) * time.Second)
				value, ok, err := s.computeMetric(ctx, e.ID, v.ID, g.MetricID, from, to)
				if err != nil {
					slog.Error("guardrail compute metric failed", "experiment_id", e.ID, "err", err)
					continue
				}
				if !ok {
					continue
				}
				violated := (g.Comparison == models.GuardrailAbove && value > g.Threshold) ||
					(g.Comparison == models.GuardrailBelow && value < g.Threshold)
				if !violated {
					continue
				}

				switch g.Action {
				case models.GuardrailRevertToControl:
					err = s.experimentRepo.SetRevertedToControl(ctx, e.ID)
					if err != nil {
						slog.Error("guardrail revert failed", "experiment_id", e.ID, "err", err)
						continue experimentLoop
					}
				case models.GuardrailPause:
					err = s.experimentRepo.TransitionStatus(ctx, e.ID,
						models.ExperimentStatusRunning, models.ExperimentStatusPaused)
					if err != nil && !errors.Is(err, errs.ErrInvalidTransition) {
						slog.Error("guardrail pause failed", "experiment_id", e.ID, "err", err)
						continue experimentLoop
					}
				}

				slog.Info("guardrail triggered",
					"experiment_id", e.ID, "variant", v.Name, "metric_id", g.MetricID,
					"value", value, "threshold", g.Threshold, "action", g.Action)

				if err := s.guardrailRepo.SaveTrigger(ctx, &models.GuardrailTrigger{
					ExperimentID:  e.ID,
					VariantID:     v.ID,
					MetricID:      g.MetricID,
					Value:         value,
					Threshold:     g.Threshold,
					Comparison:    g.Comparison,
					WindowSeconds: g.WindowSeconds,
					Action:        g.Action,
					TriggeredAt:   time.Now().UTC(),
				}); err != nil {
					slog.Error("guardrail save trigger failed", "experiment_id", e.ID, "err", err)
				}
				continue experimentLoop
			}
		}
	}
	return nil
}

func (s *GuardrailService) Get(ctx context.Context, experimentID uuid.UUID) (models.GuardrailTrigger, error) {
	if experimentID == uuid.Nil {
		return models.GuardrailTrigger{}, errs.ErrIdIsEmpty
	}
	return s.guardrailRepo.Get(ctx, experimentID)
}
