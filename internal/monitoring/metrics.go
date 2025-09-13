package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	AuthRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_requests_total",
			Help: "Total de peticiones al servicio de autenticación",
		},
		[]string{"endpoint", "status"},
	)

	AuthRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_request_duration_seconds",
			Help:    "Duración de las peticiones de autenticación",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)
)
