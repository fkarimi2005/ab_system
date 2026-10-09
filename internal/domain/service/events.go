package service

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/domain/repository"
	"AB_system/pkg/errs"
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	MaxEventBatch      = 1000
	MaxEventAge        = 7 * 24 * time.Hour // события старше отклоняем
	MaxEventFutureSkew = 5 * time.Minute    // допуск на расхождение часов клиента
	maxEventIDLen      = 128
	maxPropertiesBytes = 8 << 10
)

// Причины отклонения события.
const (
	ReasonInvalidEvent     = "invalid_event"
	ReasonUnknownType      = "unknown_type"
	ReasonArchivedType     = "archived_type"
	ReasonTooOld           = "too_old"
	ReasonFutureTimestamp  = "future_timestamp"
	ReasonMissingDecision  = "missing_decision"
	ReasonUnknownDecision  = "unknown_decision"
	ReasonSubjectMismatch  = "subject_mismatch"
	ReasonPropertiesTooBig = "properties_too_large"
)

// Причины, по которым событие не засчитано в атрибуции.
const (
	NotAttributedNoExposure     = "no_exposure"
	NotAttributedBeforeExposure = "before_exposure"
	NotAttributedNoDecision     = "no_decision"
)

type EventService struct {
	eventTypes repository.EventTypeRepository
	decisions  repository.DecisionRepository
	events     repository.EventRepository
	now        func() time.Time
}

func NewEventService(
	eventTypes repository.EventTypeRepository,
	decisions repository.DecisionRepository,
	events repository.EventRepository,
) *EventService {
	return &EventService{eventTypes: eventTypes, decisions: decisions, events: events, now: time.Now}
}

// EventInput — событие как оно пришло от клиента; разбор и проверка делаются здесь,
// чтобы одно плохое событие не отклоняло весь пакет.
type EventInput struct {
	EventID    string
	Type       string
	DecisionID string
	SubjectID  string
	Timestamp  string
	Properties map[string]any
}

type Rejection struct {
	EventID string
	Reason  string
}

type IngestResult struct {
	Accepted   int
	Duplicates int
	Rejected   []Rejection
}

// Ingest принимает пакет событий: возвращает, сколько принято, сколько дублей и что отклонено.
func (s *EventService) Ingest(ctx context.Context, batch []EventInput) (IngestResult, error) {
	if len(batch) == 0 || len(batch) > MaxEventBatch {
		return IngestResult{}, invalid(errs.ErrInvalidField, "в пакете должно быть от 1 до %d событий", MaxEventBatch)
	}
	now := s.now()

	// справочники одним запросом на пакет
	keys := make([]string, 0, len(batch))
	var decisionIDs []uuid.UUID
	for _, in := range batch {
		keys = append(keys, in.Type)
		if id, err := uuid.Parse(in.DecisionID); err == nil {
			decisionIDs = append(decisionIDs, id)
		}
	}
	types, err := s.eventTypes.GetEventTypesByKeys(ctx, keys)
	if err != nil {
		return IngestResult{}, err
	}
	typeByKey := make(map[string]models.EventType, len(types))
	for _, t := range types {
		typeByKey[t.Key] = t
	}
	found, err := s.decisions.GetDecisionsByIDs(ctx, decisionIDs)
	if err != nil {
		return IngestResult{}, err
	}
	decisionByID := make(map[uuid.UUID]models.Decision, len(found))
	for _, d := range found {
		decisionByID[d.DecisionID] = d
	}

	var res IngestResult
	var valid []models.Event
	for _, in := range batch {
		ev, reason := s.validate(in, now, typeByKey, decisionByID)
		if reason != "" {
			res.Rejected = append(res.Rejected, Rejection{EventID: in.EventID, Reason: reason})
			continue
		}
		valid = append(valid, ev)
	}

	if len(valid) > 0 {
		inserted, err := s.events.InsertEventsIdempotent(ctx, valid)
		if err != nil {
			return IngestResult{}, err
		}
		for _, isNew := range inserted {
			if isNew {
				res.Accepted++
			} else {
				res.Duplicates++
			}
		}
	}
	return res, nil
}

// validate проверяет одно событие. Возвращает событие к вставке или причину отклонения.
func (s *EventService) validate(
	in EventInput, now time.Time,
	typeByKey map[string]models.EventType, decisionByID map[uuid.UUID]models.Decision,
) (models.Event, string) {
	if in.EventID == "" || len(in.EventID) > maxEventIDLen {
		return models.Event{}, ReasonInvalidEvent
	}
	ts, err := time.Parse(time.RFC3339, in.Timestamp)
	if err != nil {
		return models.Event{}, ReasonInvalidEvent
	}
	t, ok := typeByKey[in.Type]
	if !ok {
		return models.Event{}, ReasonUnknownType
	}
	if t.IsArchived() {
		return models.Event{}, ReasonArchivedType
	}
	if now.Sub(ts) > MaxEventAge {
		return models.Event{}, ReasonTooOld
	}
	if ts.Sub(now) > MaxEventFutureSkew {
		return models.Event{}, ReasonFutureTimestamp
	}

	var props json.RawMessage
	if len(in.Properties) > 0 {
		b, err := json.Marshal(in.Properties)
		if err != nil {
			return models.Event{}, ReasonInvalidEvent
		}
		if len(b) > maxPropertiesBytes {
			return models.Event{}, ReasonPropertiesTooBig
		}
		props = b
	}

	ev := models.Event{
		EventID: in.EventID, EventTypeID: t.ID, SubjectID: in.SubjectID,
		EventTS: ts.UTC(), ReceivedAt: now.UTC(), Properties: props,
	}

	needsDecision := t.RequiresExposure || t.IsExposure
	if in.DecisionID == "" {
		if needsDecision {
			return models.Event{}, ReasonMissingDecision
		}
		if in.SubjectID == "" {
			return models.Event{}, ReasonInvalidEvent // без решения и без пользователя событие бесполезно
		}
		return ev, ""
	}
	id, err := uuid.Parse(in.DecisionID)
	if err != nil {
		return models.Event{}, ReasonInvalidEvent
	}
	d, ok := decisionByID[id]
	if !ok {
		return models.Event{}, ReasonUnknownDecision
	}
	if in.SubjectID != "" && in.SubjectID != d.SubjectID {
		return models.Event{}, ReasonSubjectMismatch
	}
	ev.DecisionID = &id
	ev.SubjectID = d.SubjectID // субъект берём из решения: источник истины
	return ev, ""
}

// ---------- атрибуция ----------

type AttributedEvent struct {
	EventID       string
	Type          string
	EventTS       time.Time
	Attributed    bool
	NotAttributed string // причина, если Attributed == false
}

type DecisionAttribution struct {
	Decision   models.Decision
	ExposureAt *time.Time
	Events     []AttributedEvent
}

// Attribution показывает, какие события решения засчитываются.
//
// Правило: событие типа с requires_exposure засчитывается, только если у этого decision_id
// есть событие показа не позже самого события (<=). Показ засчитывается всегда. Порядок
// прихода событий роли не играет — сравниваются только их собственные времена.
func (s *EventService) Attribution(ctx context.Context, decisionID uuid.UUID) (DecisionAttribution, error) {
	d, err := s.decisions.GetDecision(ctx, decisionID)
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			return DecisionAttribution{}, errs.ErrDecisionNotFound
		}
		return DecisionAttribution{}, err
	}
	rows, err := s.events.ListDecisionEvents(ctx, decisionID)
	if err != nil {
		return DecisionAttribution{}, err
	}
	return attribute(d, rows), nil
}

// attribute — чистая функция атрибуции (вынесена для тестов).
func attribute(d models.Decision, rows []models.EventRow) DecisionAttribution {
	res := DecisionAttribution{Decision: d, Events: make([]AttributedEvent, 0, len(rows))}
	for _, r := range rows {
		if r.IsExposure && (res.ExposureAt == nil || r.EventTS.Before(*res.ExposureAt)) {
			ts := r.EventTS
			res.ExposureAt = &ts
		}
	}
	for _, r := range rows {
		ev := AttributedEvent{EventID: r.EventID, Type: r.TypeKey, EventTS: r.EventTS, Attributed: true}
		switch {
		case r.IsExposure || !r.RequiresExposure:
			// показ и события без требования показа засчитываются
		case res.ExposureAt == nil:
			ev.Attributed, ev.NotAttributed = false, NotAttributedNoExposure
		case r.EventTS.Before(*res.ExposureAt):
			ev.Attributed, ev.NotAttributed = false, NotAttributedBeforeExposure
		}
		res.Events = append(res.Events, ev)
	}
	return res
}
