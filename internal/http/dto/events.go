package dto

import (
	"AB_system/internal/domain/service"
	"time"

	"github.com/google/uuid"
)

// EventRequest разбирается построчно в сервисе, поэтому поля здесь без жёстких проверок:
// одно плохое событие не должно отклонять весь пакет.
type EventRequest struct {
	EventID    string         `json:"event_id"`
	Type       string         `json:"type"`
	DecisionID string         `json:"decision_id"`
	SubjectID  string         `json:"subject_id"`
	Timestamp  string         `json:"timestamp"`
	Properties map[string]any `json:"properties"`
}

type IngestEventsRequest struct {
	Events []EventRequest `json:"events" binding:"required,min=1,max=1000"`
}

func (r IngestEventsRequest) ToInput() []service.EventInput {
	res := make([]service.EventInput, 0, len(r.Events))
	for _, e := range r.Events {
		res = append(res, service.EventInput{
			EventID: e.EventID, Type: e.Type, DecisionID: e.DecisionID,
			SubjectID: e.SubjectID, Timestamp: e.Timestamp, Properties: e.Properties,
		})
	}
	return res
}

type RejectionResponse struct {
	EventID string `json:"event_id"`
	Reason  string `json:"reason"`
}

type IngestEventsResponse struct {
	Accepted   int                 `json:"accepted"`
	Duplicates int                 `json:"duplicates"`
	Rejected   int                 `json:"rejected"`
	Details    []RejectionResponse `json:"rejected_details"`
}

func NewIngestEventsResponse(r service.IngestResult) IngestEventsResponse {
	res := IngestEventsResponse{
		Accepted: r.Accepted, Duplicates: r.Duplicates, Rejected: len(r.Rejected),
		Details: make([]RejectionResponse, 0, len(r.Rejected)),
	}
	for _, x := range r.Rejected {
		res.Details = append(res.Details, RejectionResponse{EventID: x.EventID, Reason: x.Reason})
	}
	return res
}

type AttributedEventResponse struct {
	EventID       string    `json:"event_id"`
	Type          string    `json:"type"`
	EventTS       time.Time `json:"timestamp"`
	Attributed    bool      `json:"attributed"`
	NotAttributed string    `json:"not_attributed_reason,omitempty"`
}

type AttributionResponse struct {
	DecisionID   uuid.UUID                 `json:"decision_id"`
	ExperimentID uuid.UUID                 `json:"experiment_id"`
	FlagKey      string                    `json:"flag_key"`
	SubjectID    string                    `json:"subject_id"`
	Variant      string                    `json:"variant"`
	ExposureAt   *time.Time                `json:"exposure_at,omitempty"`
	Events       []AttributedEventResponse `json:"events"`
}

func NewAttributionResponse(a service.DecisionAttribution) AttributionResponse {
	res := AttributionResponse{
		DecisionID: a.Decision.DecisionID, ExperimentID: a.Decision.ExperimentID,
		FlagKey: a.Decision.FlagKey, SubjectID: a.Decision.SubjectID, Variant: a.Decision.Variant,
		ExposureAt: a.ExposureAt, Events: make([]AttributedEventResponse, 0, len(a.Events)),
	}
	for _, e := range a.Events {
		res.Events = append(res.Events, AttributedEventResponse{
			EventID: e.EventID, Type: e.Type, EventTS: e.EventTS,
			Attributed: e.Attributed, NotAttributed: e.NotAttributed,
		})
	}
	return res
}
