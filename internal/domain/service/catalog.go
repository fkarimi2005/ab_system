package service

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/domain/repository"
	"AB_system/pkg/errs"
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/google/uuid"
)

var (
	keyPattern        = regexp.MustCompile(`^[a-z][a-z0-9_]{1,63}$`)
	valueFieldPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)
	allowedPercentile = []int{50, 90, 95, 99}
)

func invalid(base error, format string, args ...any) error {
	return fmt.Errorf("%w: %s", base, fmt.Sprintf(format, args...))
}

// ---------- каталог типов событий ----------

type EventTypeService struct {
	repo repository.EventTypeRepository
}

func NewEventTypeService(repo repository.EventTypeRepository) *EventTypeService {
	return &EventTypeService{repo: repo}
}

type CreateEventTypeInput struct {
	Key              string
	Name             string
	Description      string
	RequiresExposure bool
	IsExposure       bool
}

func (s *EventTypeService) Create(ctx context.Context, in CreateEventTypeInput) (*models.EventType, error) {
	in.Name = strings.TrimSpace(in.Name)
	if !keyPattern.MatchString(in.Key) {
		return nil, invalid(errs.ErrInvalidEventType, "key должен соответствовать %s", keyPattern)
	}
	if in.Name == "" {
		return nil, invalid(errs.ErrInvalidEventType, "name не может быть пустым")
	}
	if in.IsExposure && in.RequiresExposure {
		return nil, invalid(errs.ErrInvalidEventType, "тип показа не может требовать показа")
	}
	if in.IsExposure {
		exists, err := s.repo.ActiveExposureTypeExists(ctx)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errs.ErrExposureTypeExists
		}
	}
	return s.repo.CreateEventType(ctx, &models.EventType{
		Key: in.Key, Name: in.Name, Description: in.Description,
		RequiresExposure: in.RequiresExposure, IsExposure: in.IsExposure,
	})
}

func (s *EventTypeService) Get(ctx context.Context, id uuid.UUID) (models.EventType, error) {
	e, err := s.repo.GetEventTypeByID(ctx, id)
	if errors.Is(err, errs.ErrRecordNotFound) {
		return models.EventType{}, errs.ErrEventTypeNotFound
	}
	return e, err
}

func (s *EventTypeService) List(ctx context.Context, includeArchived bool) ([]models.EventType, error) {
	return s.repo.ListEventTypes(ctx, includeArchived)
}

type UpdateEventTypeInput struct {
	Name             string
	Description      string
	RequiresExposure bool
}

// Update меняет имя, описание и правило атрибуции. Ключ и признак показа неизменны.
func (s *EventTypeService) Update(ctx context.Context, id uuid.UUID, in UpdateEventTypeInput) (models.EventType, error) {
	e, err := s.Get(ctx, id)
	if err != nil {
		return models.EventType{}, err
	}
	if e.IsArchived() {
		return models.EventType{}, invalid(errs.ErrInvalidEventType, "тип события в архиве")
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return models.EventType{}, invalid(errs.ErrInvalidEventType, "name не может быть пустым")
	}
	if e.IsExposure && in.RequiresExposure {
		return models.EventType{}, invalid(errs.ErrInvalidEventType, "тип показа не может требовать показа")
	}
	e.Name, e.Description, e.RequiresExposure = in.Name, in.Description, in.RequiresExposure
	if err := s.repo.UpdateEventType(ctx, &e); err != nil {
		return models.EventType{}, err
	}
	return s.Get(ctx, id)
}

func (s *EventTypeService) Archive(ctx context.Context, id uuid.UUID) (models.EventType, error) {
	e, err := s.Get(ctx, id)
	if err != nil {
		return models.EventType{}, err
	}
	if e.IsExposure {
		return models.EventType{}, invalid(errs.ErrInvalidEventType, "тип показа нельзя архивировать: на нём держится атрибуция")
	}
	if err := s.repo.ArchiveEventType(ctx, id); err != nil {
		return models.EventType{}, err
	}
	return s.Get(ctx, id)
}

// ---------- каталог метрик ----------

type MetricService struct {
	metrics    repository.MetricRepository
	eventTypes repository.EventTypeRepository
}

func NewMetricService(metrics repository.MetricRepository, eventTypes repository.EventTypeRepository) *MetricService {
	return &MetricService{metrics: metrics, eventTypes: eventTypes}
}

type CreateMetricInput struct {
	Key                    string
	Name                   string
	Kind                   models.MetricKind
	EventTypeID            uuid.UUID
	DenominatorEventTypeID *uuid.UUID
	Percentile             *int
	ValueField             string
}

// validateMetricShape проверяет согласованность полей метрики по её виду (без обращения к БД).
func validateMetricShape(in CreateMetricInput) error {
	if !keyPattern.MatchString(in.Key) {
		return invalid(errs.ErrInvalidMetric, "key должен соответствовать %s", keyPattern)
	}
	if strings.TrimSpace(in.Name) == "" {
		return invalid(errs.ErrInvalidMetric, "name не может быть пустым")
	}
	if in.EventTypeID == uuid.Nil {
		return invalid(errs.ErrInvalidMetric, "event_type_id обязателен")
	}
	switch in.Kind {
	case models.MetricKindCount:
		if in.DenominatorEventTypeID != nil || in.Percentile != nil || in.ValueField != "" {
			return invalid(errs.ErrInvalidMetric, "для count не задаются denominator, percentile и value_field")
		}
	case models.MetricKindConversion, models.MetricKindErrorRate:
		if in.Percentile != nil || in.ValueField != "" {
			return invalid(errs.ErrInvalidMetric, "для %s не задаются percentile и value_field", in.Kind)
		}
	case models.MetricKindLatencyPercentile:
		if in.DenominatorEventTypeID != nil {
			return invalid(errs.ErrInvalidMetric, "для latency_percentile не задаётся denominator")
		}
		if in.Percentile == nil || !slices.Contains(allowedPercentile, *in.Percentile) {
			return invalid(errs.ErrInvalidMetric, "percentile должен быть одним из %v", allowedPercentile)
		}
		if !valueFieldPattern.MatchString(in.ValueField) {
			return invalid(errs.ErrInvalidMetric, "value_field обязателен и должен соответствовать %s", valueFieldPattern)
		}
	default:
		return invalid(errs.ErrInvalidMetric, "неизвестный вид метрики %q", in.Kind)
	}
	return nil
}

func (s *MetricService) activeEventType(ctx context.Context, id uuid.UUID, what string) error {
	e, err := s.eventTypes.GetEventTypeByID(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			return invalid(errs.ErrInvalidMetric, "%s: тип события не найден", what)
		}
		return err
	}
	if e.IsArchived() {
		return invalid(errs.ErrInvalidMetric, "%s: тип события в архиве", what)
	}
	return nil
}

func (s *MetricService) Create(ctx context.Context, in CreateMetricInput) (*models.Metric, error) {
	in.Name = strings.TrimSpace(in.Name)
	if err := validateMetricShape(in); err != nil {
		return nil, err
	}
	if err := s.activeEventType(ctx, in.EventTypeID, "event_type_id"); err != nil {
		return nil, err
	}
	if in.DenominatorEventTypeID != nil {
		if err := s.activeEventType(ctx, *in.DenominatorEventTypeID, "denominator_event_type_id"); err != nil {
			return nil, err
		}
	}
	return s.metrics.CreateMetric(ctx, &models.Metric{
		Key: in.Key, Name: in.Name, Kind: in.Kind, EventTypeID: in.EventTypeID,
		DenominatorEventTypeID: in.DenominatorEventTypeID, Percentile: in.Percentile, ValueField: in.ValueField,
	})
}

func (s *MetricService) Get(ctx context.Context, id uuid.UUID) (models.Metric, error) {
	m, err := s.metrics.GetMetricByID(ctx, id)
	if errors.Is(err, errs.ErrRecordNotFound) {
		return models.Metric{}, errs.ErrMetricNotFound
	}
	return m, err
}

func (s *MetricService) List(ctx context.Context, includeArchived bool) ([]models.Metric, error) {
	return s.metrics.ListMetrics(ctx, includeArchived)
}

// Rename меняет только название: остальное определяет смысл метрики, и менять его у
// метрики, на которую уже ссылаются эксперименты, нельзя — нужна новая метрика.
func (s *MetricService) Rename(ctx context.Context, id uuid.UUID, name string) (models.Metric, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.Metric{}, invalid(errs.ErrInvalidMetric, "name не может быть пустым")
	}
	m, err := s.Get(ctx, id)
	if err != nil {
		return models.Metric{}, err
	}
	if m.IsArchived() {
		return models.Metric{}, invalid(errs.ErrInvalidMetric, "метрика в архиве")
	}
	if err := s.metrics.UpdateMetricName(ctx, id, name); err != nil {
		return models.Metric{}, err
	}
	return s.Get(ctx, id)
}

func (s *MetricService) Archive(ctx context.Context, id uuid.UUID) (models.Metric, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return models.Metric{}, err
	}
	if err := s.metrics.ArchiveMetric(ctx, id); err != nil {
		return models.Metric{}, err
	}
	return s.Get(ctx, id)
}
