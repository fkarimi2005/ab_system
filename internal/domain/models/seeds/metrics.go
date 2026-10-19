package seeds

import (
	"AB_system/internal/domain/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedMetrics(db *gorm.DB) error {
	clickID, err := GetEventTypesIDByKey(db, "click")
	shownID, err := GetEventTypesIDByKey(db, "shown")
	purchaseID, err := GetEventTypesIDByKey(db, "purchase")
	errorID, err := GetEventTypesIDByKey(db, "error")
	latencyID, err := GetEventTypesIDByKey(db, "api_request")
	if err != nil {
		return err
	}

	metrics := []models.Metric{
		{
			Key:         "click_count",
			Name:        "Click Count",
			Kind:        models.MetricKindCount,
			EventTypeID: clickID,
		},
		{
			Key:                    "ctr",
			Name:                   "Click-Through Rate",
			Kind:                   models.MetricKindConversion,
			EventTypeID:            clickID,
			DenominatorEventTypeID: uuidPtr(shownID),
		},
		{
			Key:         "error_rate",
			Name:        "Error Rate",
			Kind:        models.MetricKindErrorRate,
			EventTypeID: errorID,
		},
		{
			Key:         "shown_count",
			Name:        "Shown Count",
			Kind:        models.MetricKindCount,
			EventTypeID: shownID,
		},
		{
			Key:         "purchase_count",
			Name:        "Purchase Count",
			Kind:        models.MetricKindCount,
			EventTypeID: purchaseID,
		},
		{
			Key:                    "purchase_rate",
			Name:                   "Purchase Rate",
			Kind:                   models.MetricKindConversion,
			EventTypeID:            purchaseID,
			DenominatorEventTypeID: uuidPtr(shownID),
		},
		{
			Key:         "api_latency_p95",
			Name:        "API Latency P95",
			Kind:        models.MetricKindLatencyPercentile,
			EventTypeID: latencyID,
			Percentile:  intPtr(95),
			ValueField:  "latency_ms",
		},
		{
			Key:         "api_request_count",
			Name:        "API Request Count",
			Kind:        models.MetricKindCount,
			EventTypeID: clickID,
		},
		{
			Key:         "error_count",
			Name:        "Error Count",
			Kind:        models.MetricKindCount,
			EventTypeID: errorID,
		},
	}

	for _, metric := range metrics {
		if err := db.Where("key = ?", metric.Key).
			Assign(metric).
			FirstOrCreate(&metric).Error; err != nil {
			return err
		}
	}

	return nil
}

func uuidPtr(id uuid.UUID) *uuid.UUID {
	return &id
}

func intPtr(n int) *int {
	return &n
}

type MetricType string

const (
	ClickCount      MetricType = "click_count"
	CTR             MetricType = "ctr"
	ErrorCount      MetricType = "error_count"
	ErrorRate       MetricType = "error_rate"
	ShownCount      MetricType = "shown_count"
	PurchaseCount   MetricType = "purchase_count"
	PurchaseRate    MetricType = "purchase_rate"
	APIRequestCount MetricType = "api_request_count"
	APILatencyP95   MetricType = "api_latency_p95"
)
