package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // SLI: Disponibilidad del servicio
    AuthServiceAvailability = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "auth_service_requests_total",
            Help: "Total number of requests to auth service",
        },
        []string{"method", "endpoint", "status_code"},
    )

    // SLI: Latencia de autenticación
    AuthLatency = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "auth_request_duration_seconds",
            Help:    "Duration of authentication requests",
            Buckets: []float64{0.1, 0.2, 0.5, 1.0, 2.0, 5.0},
        },
        []string{"method", "endpoint"},
    )

    // SLI: Tasa de éxito de login
    LoginSuccessRate = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "login_attempts_total",
            Help: "Total login attempts",
        },
        []string{"method", "result", "tenant_id"},
    )

    // SLI: Tiempo de respuesta de MFA
    MFAResponseTime = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "mfa_verification_duration_seconds",
            Help:    "Duration of MFA verification",
            Buckets: []float64{0.5, 1.0, 2.0, 5.0, 10.0},
        },
        []string{"method", "provider"},
    )

    // SLI: Errores de sistema críticos
    CriticalErrors = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "critical_errors_total",
            Help: "Total critical system errors",
        },
        []string{"component", "error_type"},
    )
)

// SLOs definidos
const (
    AvailabilitySLO     = 99.9  // 99.9% uptime
    LatencySLO         = 200   // <200ms P95
    LoginSuccessSLO    = 99.5  // 99.5% success rate
    MFAResponseSLO     = 5000  // <5s P99
    ErrorRateSLO       = 0.1   // <0.1% error rate
)