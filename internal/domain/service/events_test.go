package service

import (
	"AB_system/internal/domain/models"
	"AB_system/pkg/errs"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type memEventTypes struct{ byKey map[string]models.EventType }

func (m memEventTypes) CreateEventType(context.Context, *models.EventType) (*models.EventType, error) {
	return nil, nil
}
func (m memEventTypes) GetEventTypeByID(context.Context, uuid.UUID) (models.EventType, error) {
	return models.EventType{}, errs.ErrRecordNotFound
}
func (m memEventTypes) ListEventTypes(context.Context, bool) ([]models.EventType, error) {
	return nil, nil
}
func (m memEventTypes) UpdateEventType(context.Context, *models.EventType) error { return nil }
func (m memEventTypes) ArchiveEventType(context.Context, uuid.UUID) error        { return nil }
func (m memEventTypes) ActiveExposureTypeExists(context.Context) (bool, error)   { return false, nil }
func (m memEventTypes) GetEventTypesByKeys(_ context.Context, keys []string) ([]models.EventType, error) {
	var res []models.EventType
	for _, k := range keys {
		if t, ok := m.byKey[k]; ok {
			res = append(res, t)
		}
	}
	return res, nil
}

type memDecisions map[uuid.UUID]models.Decision

func (m memDecisions) RecordDecisions(context.Context, []models.Decision) error { return nil }
func (m memDecisions) GetDecisionsByIDs(_ context.Context, ids []uuid.UUID) ([]models.Decision, error) {
	var res []models.Decision
	for _, id := range ids {
		if d, ok := m[id]; ok {
			res = append(res, d)
		}
	}
	return res, nil
}
func (m memDecisions) GetDecision(_ context.Context, id uuid.UUID) (models.Decision, error) {
	if d, ok := m[id]; ok {
		return d, nil
	}
	return models.Decision{}, errs.ErrRecordNotFound
}

type memEvents struct {
	seen map[string]bool
	rows []models.Event
}

func (m *memEvents) InsertEventsIdempotent(_ context.Context, evs []models.Event) ([]bool, error) {
	out := make([]bool, len(evs))
	for i, e := range evs {
		if m.seen[e.EventID] {
			continue
		}
		m.seen[e.EventID] = true
		m.rows = append(m.rows, e)
		out[i] = true
	}
	return out, nil
}
func (m *memEvents) ListDecisionEvents(context.Context, uuid.UUID) ([]models.EventRow, error) {
	return nil, nil
}

var now0 = time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC)

type eventFixture struct {
	svc      *EventService
	events   *memEvents
	decision models.Decision
}

func newEventFixture() eventFixture {
	exposure := models.EventType{ID: uuid.New(), Key: "exposure", IsExposure: true}
	purchase := models.EventType{ID: uuid.New(), Key: "purchase", RequiresExposure: true}
	pageview := models.EventType{ID: uuid.New(), Key: "page_view"}
	archived := models.EventType{ID: uuid.New(), Key: "old_event", ArchivedAt: &now0}
	d := models.Decision{DecisionID: uuid.New(), ExperimentID: uuid.New(), SubjectID: "user-1", Variant: "red", FlagKey: "btn"}

	events := &memEvents{seen: map[string]bool{}}
	svc := NewEventService(
		memEventTypes{byKey: map[string]models.EventType{"exposure": exposure, "purchase": purchase, "page_view": pageview, "old_event": archived}},
		memDecisions{d.DecisionID: d},
		events,
	)
	svc.now = func() time.Time { return now0 }
	return eventFixture{svc: svc, events: events, decision: d}
}

func ts(d time.Duration) string { return now0.Add(d).Format(time.RFC3339) }

func (f eventFixture) ev(id, typ string, at time.Duration) EventInput {
	return EventInput{EventID: id, Type: typ, DecisionID: f.decision.DecisionID.String(), Timestamp: ts(at)}
}

func reasons(r IngestResult) map[string]string {
	m := map[string]string{}
	for _, x := range r.Rejected {
		m[x.EventID] = x.Reason
	}
	return m
}

func TestIngestAcceptsAndDeduplicates(t *testing.T) {
	f := newEventFixture()
	batch := []EventInput{f.ev("e1", "exposure", -time.Minute), f.ev("e2", "purchase", -30*time.Second), f.ev("e1", "exposure", -time.Minute)}
	res, err := f.svc.Ingest(context.Background(), batch)
	if err != nil {
		t.Fatal(err)
	}
	if res.Accepted != 2 || res.Duplicates != 1 || len(res.Rejected) != 0 {
		t.Fatalf("первая отправка: %+v", res)
	}
	res, _ = f.svc.Ingest(context.Background(), batch) // повтор всего пакета
	if res.Accepted != 0 || res.Duplicates != 3 {
		t.Fatalf("повтор пакета: %+v", res)
	}
	if len(f.events.rows) != 2 {
		t.Errorf("в хранилище должно быть 2 события, а их %d", len(f.events.rows))
	}
}

func TestIngestRejectionReasons(t *testing.T) {
	f := newEventFixture()
	noDecision := EventInput{EventID: "no-dec", Type: "purchase", Timestamp: ts(-time.Minute)}
	unknownDecision := f.ev("unk-dec", "purchase", -time.Minute)
	unknownDecision.DecisionID = uuid.NewString()
	wrongSubject := f.ev("subj", "purchase", -time.Minute)
	wrongSubject.SubjectID = "someone-else"
	badTime := f.ev("badtime", "purchase", 0)
	badTime.Timestamp = "вчера"
	bigProps := f.ev("big", "purchase", -time.Minute)
	bigProps.Properties = map[string]any{"blob": strings.Repeat("x", 9000)}

	batch := []EventInput{
		f.ev("ok", "exposure", -time.Minute),
		f.ev("unk-type", "nope", -time.Minute),
		f.ev("archived", "old_event", -time.Minute),
		noDecision, unknownDecision, wrongSubject, badTime, bigProps,
		{EventID: "", Type: "page_view", SubjectID: "u", Timestamp: ts(0)},
		f.ev("future", "purchase", 10*time.Minute),
	}
	res, err := f.svc.Ingest(context.Background(), batch)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"unk-type": ReasonUnknownType, "archived": ReasonArchivedType, "no-dec": ReasonMissingDecision,
		"unk-dec": ReasonUnknownDecision, "subj": ReasonSubjectMismatch, "badtime": ReasonInvalidEvent,
		"big": ReasonPropertiesTooBig, "": ReasonInvalidEvent, "future": ReasonFutureTimestamp,
	}
	got := reasons(res)
	for id, reason := range want {
		if got[id] != reason {
			t.Errorf("%q: причина %q, ожидалась %q", id, got[id], reason)
		}
	}
	if res.Accepted != 1 {
		t.Errorf("принято %d, ожидалось 1 (одно хорошее событие в пакете не должно пострадать)", res.Accepted)
	}
}

func TestIngestAgeBoundary(t *testing.T) {
	f := newEventFixture()
	res, _ := f.svc.Ingest(context.Background(), []EventInput{
		f.ev("exactly7d", "exposure", -MaxEventAge),
		f.ev("7d+1s", "exposure", -MaxEventAge-time.Second),
		f.ev("future-ok", "exposure", MaxEventFutureSkew),
		f.ev("future-bad", "exposure", MaxEventFutureSkew+time.Second),
	})
	got := reasons(res)
	if got["exactly7d"] != "" || got["future-ok"] != "" {
		t.Errorf("граничные значения должны приниматься: %v", got)
	}
	if got["7d+1s"] != ReasonTooOld || got["future-bad"] != ReasonFutureTimestamp {
		t.Errorf("за границей должны отклоняться: %v", got)
	}
}

func TestIngestSubjectTakenFromDecision(t *testing.T) {
	f := newEventFixture()
	if _, err := f.svc.Ingest(context.Background(), []EventInput{f.ev("e1", "exposure", -time.Minute)}); err != nil {
		t.Fatal(err)
	}
	if got := f.events.rows[0].SubjectID; got != "user-1" {
		t.Errorf("subject должен браться из решения, got %q", got)
	}
}

func TestIngestEventWithoutDecision(t *testing.T) {
	f := newEventFixture()
	res, _ := f.svc.Ingest(context.Background(), []EventInput{
		{EventID: "pv1", Type: "page_view", SubjectID: "u9", Timestamp: ts(-time.Minute)},
		{EventID: "pv2", Type: "page_view", Timestamp: ts(-time.Minute)}, // нет ни решения, ни пользователя
	})
	if res.Accepted != 1 || reasons(res)["pv2"] != ReasonInvalidEvent {
		t.Errorf("%+v", res)
	}
}

func TestIngestBatchSizeLimits(t *testing.T) {
	f := newEventFixture()
	if _, err := f.svc.Ingest(context.Background(), nil); err == nil {
		t.Error("пустой пакет должен отклоняться")
	}
	big := make([]EventInput, MaxEventBatch+1)
	if _, err := f.svc.Ingest(context.Background(), big); err == nil {
		t.Error("слишком большой пакет должен отклоняться")
	}
}

func TestAttribute(t *testing.T) {
	d := models.Decision{DecisionID: uuid.New()}
	row := func(id, typ string, exposure, requires bool, at time.Duration) models.EventRow {
		return models.EventRow{
			Event:   models.Event{EventID: id, EventTS: now0.Add(at)},
			TypeKey: typ, IsExposure: exposure, RequiresExposure: requires,
		}
	}
	byID := func(a DecisionAttribution) map[string]AttributedEvent {
		m := map[string]AttributedEvent{}
		for _, e := range a.Events {
			m[e.EventID] = e
		}
		return m
	}

	t.Run("конверсия после показа засчитывается", func(t *testing.T) {
		a := attribute(d, []models.EventRow{row("x", "exposure", true, false, 0), row("c", "purchase", false, true, time.Second)})
		if !byID(a)["c"].Attributed || a.ExposureAt == nil {
			t.Errorf("%+v", a)
		}
	})
	t.Run("показ и конверсия в одну секунду: засчитывается (<=)", func(t *testing.T) {
		a := attribute(d, []models.EventRow{row("x", "exposure", true, false, 0), row("c", "purchase", false, true, 0)})
		if !byID(a)["c"].Attributed {
			t.Errorf("%+v", a)
		}
	})
	t.Run("без показа не засчитывается", func(t *testing.T) {
		a := attribute(d, []models.EventRow{row("c", "purchase", false, true, 0)})
		if e := byID(a)["c"]; e.Attributed || e.NotAttributed != NotAttributedNoExposure {
			t.Errorf("%+v", e)
		}
	})
	t.Run("конверсия раньше показа не засчитывается", func(t *testing.T) {
		a := attribute(d, []models.EventRow{row("c", "purchase", false, true, -time.Minute), row("x", "exposure", true, false, 0)})
		if e := byID(a)["c"]; e.Attributed || e.NotAttributed != NotAttributedBeforeExposure {
			t.Errorf("%+v", e)
		}
	})
	t.Run("событие без требования показа засчитывается всегда", func(t *testing.T) {
		a := attribute(d, []models.EventRow{row("v", "page_view", false, false, 0)})
		if !byID(a)["v"].Attributed {
			t.Errorf("%+v", a)
		}
	})
	t.Run("берётся самый ранний показ", func(t *testing.T) {
		a := attribute(d, []models.EventRow{row("x2", "exposure", true, false, time.Minute), row("x1", "exposure", true, false, 0), row("c", "purchase", false, true, 30*time.Second)})
		if !byID(a)["c"].Attributed || !a.ExposureAt.Equal(now0) {
			t.Errorf("%+v", a)
		}
	})
}
