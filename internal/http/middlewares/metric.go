package middlewares

import (
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var experimentRequests = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "experiment_http_requests_total",
		Help: "Количество запросов по экспериментам и вариантам",
	},
	[]string{"experiment_id", "variant_id"},
)

func RecordExperimentRequest(
	experimentID uuid.UUID,
	variantID uuid.UUID,
) {
	experimentRequests.WithLabelValues(
		experimentID.String(),
		variantID.String(),
	).Inc()
}
