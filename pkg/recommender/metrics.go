package recommender

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Recommendation generation metrics
	recommendGenerateDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "eidos_recommend_generation_duration_seconds",
			Help:    "Time taken to generate a complete configuration recommendation",
			Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
		},
	)

	recommendGenerateTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "eidos_recommend_generation_total",
			Help: "Total number of recommendation generation attempts",
		},
		[]string{"status"}, // success or error
	)
)
