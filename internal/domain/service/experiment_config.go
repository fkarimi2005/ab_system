package service

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/domain/service/input"
	"AB_system/pkg/errs"
	"context"
	"math"

	"github.com/google/uuid"
)

const (
	maxMetricsPerExperiment = 20
	minGuardrailWindow      = 60        // секунд
	maxGuardrailWindow      = 7 * 86400 // неделя
)

// metricLookup — всё, что нужно сервису экспериментов от каталога метрик.
type metricLookup interface {
	GetMetricsByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Metric, error)
}

// buildConfig проверяет метрики и guardrails эксперимента и превращает их в модели.
func (s *ExperimentService) buildConfig(
	ctx context.Context, metrics []input.MetricRefInput, guardrails []input.GuardrailInput,
) ([]models.ExperimentMetric, []models.ExperimentGuardrail, error) {
	if len(metrics) > maxMetricsPerExperiment || len(guardrails) > maxMetricsPerExperiment {
		return nil, nil, invalid(errs.ErrInvalidExperimentConfig, "не более %d метрик и %d guardrails", maxMetricsPerExperiment, maxMetricsPerExperiment)
	}

	var ids []uuid.UUID
	seenMetric := map[uuid.UUID]bool{}
	resMetrics := make([]models.ExperimentMetric, 0, len(metrics))
	for _, m := range metrics {
		role := models.MetricRole(m.Role)
		if role != models.MetricRoleTarget && role != models.MetricRoleDiagnostic {
			return nil, nil, invalid(errs.ErrInvalidExperimentConfig, "роль метрики должна быть target или diagnostic, получено %q", m.Role)
		}
		if seenMetric[m.MetricID] {
			return nil, nil, invalid(errs.ErrInvalidExperimentConfig, "метрика %s указана дважды", m.MetricID)
		}
		seenMetric[m.MetricID] = true
		ids = append(ids, m.MetricID)
		resMetrics = append(resMetrics, models.ExperimentMetric{MetricID: m.MetricID, Role: role})
	}

	seenGuardrail := map[uuid.UUID]bool{}
	resGuardrails := make([]models.ExperimentGuardrail, 0, len(guardrails))
	for _, g := range guardrails {
		cmp := models.GuardrailComparison(g.Comparison)
		if cmp != models.GuardrailAbove && cmp != models.GuardrailBelow {
			return nil, nil, invalid(errs.ErrInvalidExperimentConfig, "comparison должен быть gt или lt, получено %q", g.Comparison)
		}
		action := models.GuardrailAction(g.Action)
		if action != models.GuardrailPause && action != models.GuardrailRevertToControl {
			return nil, nil, invalid(errs.ErrInvalidExperimentConfig, "action должен быть pause или revert_to_control, получено %q", g.Action)
		}
		if g.WindowSeconds < minGuardrailWindow || g.WindowSeconds > maxGuardrailWindow {
			return nil, nil, invalid(errs.ErrInvalidExperimentConfig, "window_seconds должен быть от %d до %d", minGuardrailWindow, maxGuardrailWindow)
		}
		if math.IsNaN(g.Threshold) || math.IsInf(g.Threshold, 0) {
			return nil, nil, invalid(errs.ErrInvalidExperimentConfig, "threshold должен быть конечным числом")
		}
		if seenGuardrail[g.MetricID] {
			return nil, nil, invalid(errs.ErrInvalidExperimentConfig, "guardrail для метрики %s указан дважды", g.MetricID)
		}
		seenGuardrail[g.MetricID] = true
		if !seenMetric[g.MetricID] {
			ids = append(ids, g.MetricID)
		}
		resGuardrails = append(resGuardrails, models.ExperimentGuardrail{
			MetricID: g.MetricID, Threshold: g.Threshold, Comparison: cmp,
			WindowSeconds: g.WindowSeconds, Action: action,
		})
	}

	if err := s.requireActiveMetrics(ctx, ids); err != nil {
		return nil, nil, err
	}
	return resMetrics, resGuardrails, nil
}

// requireActiveMetrics проверяет, что все метрики есть в каталоге и не в архиве.
func (s *ExperimentService) requireActiveMetrics(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	found, err := s.metricRepo.GetMetricsByIDs(ctx, ids)
	if err != nil {
		return err
	}
	byID := make(map[uuid.UUID]models.Metric, len(found))
	for _, m := range found {
		byID[m.ID] = m
	}
	for _, id := range ids {
		m, ok := byID[id]
		if !ok {
			return invalid(errs.ErrInvalidExperimentConfig, "метрика %s не найдена в каталоге", id)
		}
		if m.IsArchived() {
			return invalid(errs.ErrInvalidExperimentConfig, "метрика %q в архиве", m.Key)
		}
	}
	return nil
}

// checkConfigForReview: на ревью уходит эксперимент хотя бы с одной целевой метрикой,
// и все его метрики на этот момент должны быть активны (их могли архивировать после привязки).
func (s *ExperimentService) checkConfigForReview(ctx context.Context, e models.Experiment) error {
	hasTarget := false
	var ids []uuid.UUID
	seen := map[uuid.UUID]bool{}
	for _, m := range e.Metrics {
		if m.Role == models.MetricRoleTarget {
			hasTarget = true
		}
		seen[m.MetricID] = true
		ids = append(ids, m.MetricID)
	}
	for _, g := range e.Guardrails {
		if !seen[g.MetricID] {
			ids = append(ids, g.MetricID)
		}
	}
	if !hasTarget {
		return errs.ErrNoTargetMetric
	}
	return s.requireActiveMetrics(ctx, ids)
}
