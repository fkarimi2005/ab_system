package service

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/domain/service/input"
	"AB_system/pkg/errs"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func intp(n int) *int { return &n }

func TestValidateMetricShape(t *testing.T) {
	et, den := uuid.New(), uuid.New()
	tests := []struct {
		name string
		in   CreateMetricInput
		ok   bool
	}{
		{"count ok", CreateMetricInput{Key: "clicks", Name: "Clicks", Kind: models.MetricKindCount, EventTypeID: et}, true},
		{"count с denominator", CreateMetricInput{Key: "clicks", Name: "n", Kind: models.MetricKindCount, EventTypeID: et, DenominatorEventTypeID: &den}, false},
		{"conversion ok", CreateMetricInput{Key: "cr", Name: "n", Kind: models.MetricKindConversion, EventTypeID: et}, true},
		{"conversion с denominator ok", CreateMetricInput{Key: "cr", Name: "n", Kind: models.MetricKindConversion, EventTypeID: et, DenominatorEventTypeID: &den}, true},
		{"conversion с percentile", CreateMetricInput{Key: "cr", Name: "n", Kind: models.MetricKindConversion, EventTypeID: et, Percentile: intp(95)}, false},
		{"error_rate ok", CreateMetricInput{Key: "errs", Name: "n", Kind: models.MetricKindErrorRate, EventTypeID: et}, true},
		{"latency ok", CreateMetricInput{Key: "p95", Name: "n", Kind: models.MetricKindLatencyPercentile, EventTypeID: et, Percentile: intp(95), ValueField: "latency_ms"}, true},
		{"latency без percentile", CreateMetricInput{Key: "p95", Name: "n", Kind: models.MetricKindLatencyPercentile, EventTypeID: et, ValueField: "latency_ms"}, false},
		{"latency percentile 42", CreateMetricInput{Key: "p95", Name: "n", Kind: models.MetricKindLatencyPercentile, EventTypeID: et, Percentile: intp(42), ValueField: "latency_ms"}, false},
		{"latency без value_field", CreateMetricInput{Key: "p95", Name: "n", Kind: models.MetricKindLatencyPercentile, EventTypeID: et, Percentile: intp(95)}, false},
		{"latency с denominator", CreateMetricInput{Key: "p95", Name: "n", Kind: models.MetricKindLatencyPercentile, EventTypeID: et, Percentile: intp(95), ValueField: "x", DenominatorEventTypeID: &den}, false},
		{"неизвестный kind", CreateMetricInput{Key: "x1", Name: "n", Kind: "avg", EventTypeID: et}, false},
		{"плохой key", CreateMetricInput{Key: "Bad Key", Name: "n", Kind: models.MetricKindCount, EventTypeID: et}, false},
		{"пустое имя", CreateMetricInput{Key: "clicks", Name: " ", Kind: models.MetricKindCount, EventTypeID: et}, false},
		{"нет event_type", CreateMetricInput{Key: "clicks", Name: "n", Kind: models.MetricKindCount}, false},
	}
	for _, tt := range tests {
		tt.in.Name = trimmed(tt.in.Name)
		err := validateMetricShape(tt.in)
		if (err == nil) != tt.ok {
			t.Errorf("%s: err=%v, ok=%v", tt.name, err, tt.ok)
		}
		if err != nil && !errors.Is(err, errs.ErrInvalidMetric) {
			t.Errorf("%s: ошибка должна оборачивать ErrInvalidMetric: %v", tt.name, err)
		}
	}
}

func trimmed(s string) string {
	for len(s) > 0 && s[0] == ' ' {
		s = s[1:]
	}
	return s
}

type fakeMetrics map[uuid.UUID]models.Metric

func (f fakeMetrics) GetMetricsByIDs(_ context.Context, ids []uuid.UUID) ([]models.Metric, error) {
	var res []models.Metric
	for _, id := range ids {
		if m, ok := f[id]; ok {
			res = append(res, m)
		}
	}
	return res, nil
}

func TestBuildConfig(t *testing.T) {
	active, other := uuid.New(), uuid.New()
	archivedID := uuid.New()
	now := time.Now()
	svc := &ExperimentService{metricRepo: fakeMetrics{
		active:     {ID: active, Key: "a"},
		other:      {ID: other, Key: "b"},
		archivedID: {ID: archivedID, Key: "old", ArchivedAt: &now},
	}}
	gr := func(m uuid.UUID) input.GuardrailInput {
		return input.GuardrailInput{MetricID: m, Threshold: 0.1, Comparison: "gt", WindowSeconds: 600, Action: "pause"}
	}
	tests := []struct {
		name string
		m    []input.MetricRefInput
		g    []input.GuardrailInput
		ok   bool
	}{
		{"пусто", nil, nil, true},
		{"target+diagnostic", []input.MetricRefInput{{MetricID: active, Role: "target"}, {MetricID: other, Role: "diagnostic"}}, []input.GuardrailInput{gr(other)}, true},
		{"guardrail на метрику вне списка метрик эксперимента", []input.MetricRefInput{{MetricID: active, Role: "target"}}, []input.GuardrailInput{gr(other)}, true},
		{"дубль метрики", []input.MetricRefInput{{MetricID: active, Role: "target"}, {MetricID: active, Role: "diagnostic"}}, nil, false},
		{"плохая роль", []input.MetricRefInput{{MetricID: active, Role: "main"}}, nil, false},
		{"метрики нет в каталоге", []input.MetricRefInput{{MetricID: uuid.New(), Role: "target"}}, nil, false},
		{"метрика в архиве", []input.MetricRefInput{{MetricID: archivedID, Role: "target"}}, nil, false},
		{"guardrail: плохой comparison", nil, []input.GuardrailInput{{MetricID: active, Comparison: "eq", WindowSeconds: 600, Action: "pause"}}, false},
		{"guardrail: плохой action", nil, []input.GuardrailInput{{MetricID: active, Comparison: "gt", WindowSeconds: 600, Action: "delete"}}, false},
		{"guardrail: окно слишком маленькое", nil, []input.GuardrailInput{{MetricID: active, Comparison: "gt", WindowSeconds: 59, Action: "pause"}}, false},
		{"guardrail: окно больше недели", nil, []input.GuardrailInput{{MetricID: active, Comparison: "gt", WindowSeconds: 7*86400 + 1, Action: "pause"}}, false},
		{"guardrail: граница окна 60", nil, []input.GuardrailInput{{MetricID: active, Comparison: "lt", WindowSeconds: 60, Action: "revert_to_control"}}, true},
		{"guardrail: дубль метрики", nil, []input.GuardrailInput{gr(active), gr(active)}, false},
		{"guardrail: метрика в архиве", nil, []input.GuardrailInput{gr(archivedID)}, false},
	}
	for _, tt := range tests {
		_, _, err := svc.buildConfig(context.Background(), tt.m, tt.g)
		if (err == nil) != tt.ok {
			t.Errorf("%s: err=%v, ok=%v", tt.name, err, tt.ok)
		}
		if err != nil && !errors.Is(err, errs.ErrInvalidExperimentConfig) {
			t.Errorf("%s: ошибка должна оборачивать ErrInvalidExperimentConfig: %v", tt.name, err)
		}
	}
}

func TestCheckConfigForReview(t *testing.T) {
	m1, m2 := uuid.New(), uuid.New()
	now := time.Now()
	svc := &ExperimentService{metricRepo: fakeMetrics{m1: {ID: m1}, m2: {ID: m2, ArchivedAt: &now}}}
	ctx := context.Background()

	if err := svc.checkConfigForReview(ctx, models.Experiment{}); !errors.Is(err, errs.ErrNoTargetMetric) {
		t.Errorf("без метрик: %v", err)
	}
	onlyDiag := models.Experiment{Metrics: []models.ExperimentMetric{{MetricID: m1, Role: models.MetricRoleDiagnostic}}}
	if err := svc.checkConfigForReview(ctx, onlyDiag); !errors.Is(err, errs.ErrNoTargetMetric) {
		t.Errorf("только diagnostic: %v", err)
	}
	good := models.Experiment{Metrics: []models.ExperimentMetric{{MetricID: m1, Role: models.MetricRoleTarget}}}
	if err := svc.checkConfigForReview(ctx, good); err != nil {
		t.Errorf("корректный эксперимент: %v", err)
	}
	archived := models.Experiment{Metrics: []models.ExperimentMetric{{MetricID: m1, Role: models.MetricRoleTarget}, {MetricID: m2, Role: models.MetricRoleDiagnostic}}}
	if err := svc.checkConfigForReview(ctx, archived); !errors.Is(err, errs.ErrInvalidExperimentConfig) {
		t.Errorf("метрику архивировали после привязки: %v", err)
	}
}

type fakeEventTypes struct {
	exposureExists bool
	created        []models.EventType
}

func (f *fakeEventTypes) CreateEventType(_ context.Context, e *models.EventType) (*models.EventType, error) {
	f.created = append(f.created, *e)
	return e, nil
}
func (f *fakeEventTypes) GetEventTypeByID(context.Context, uuid.UUID) (models.EventType, error) {
	return models.EventType{}, errs.ErrRecordNotFound
}
func (f *fakeEventTypes) ListEventTypes(context.Context, bool) ([]models.EventType, error) {
	return nil, nil
}
func (f *fakeEventTypes) UpdateEventType(context.Context, *models.EventType) error { return nil }
func (f *fakeEventTypes) ArchiveEventType(context.Context, uuid.UUID) error        { return nil }
func (f *fakeEventTypes) ActiveExposureTypeExists(context.Context) (bool, error) {
	return f.exposureExists, nil
}

func TestEventTypeCreateRules(t *testing.T) {
	ctx := context.Background()

	repo := &fakeEventTypes{}
	svc := NewEventTypeService(repo)
	if _, err := svc.Create(ctx, CreateEventTypeInput{Key: "exposure", Name: "Показ", IsExposure: true}); err != nil {
		t.Fatalf("первый тип показа: %v", err)
	}

	repo.exposureExists = true
	if _, err := svc.Create(ctx, CreateEventTypeInput{Key: "shown2", Name: "Показ 2", IsExposure: true}); !errors.Is(err, errs.ErrExposureTypeExists) {
		t.Errorf("второй тип показа: %v", err)
	}
	if _, err := svc.Create(ctx, CreateEventTypeInput{Key: "purchase", Name: "Покупка", RequiresExposure: true}); err != nil {
		t.Errorf("обычный тип при существующем показе: %v", err)
	}
	if _, err := svc.Create(ctx, CreateEventTypeInput{Key: "bad", Name: "x", IsExposure: true, RequiresExposure: true}); !errors.Is(err, errs.ErrInvalidEventType) {
		t.Errorf("тип показа, требующий показа: %v", err)
	}
	for _, key := range []string{"", "A", "1abc", "with space", "кириллица"} {
		if _, err := svc.Create(ctx, CreateEventTypeInput{Key: key, Name: "x"}); !errors.Is(err, errs.ErrInvalidEventType) {
			t.Errorf("key %q должен быть отклонён: %v", key, err)
		}
	}
}
