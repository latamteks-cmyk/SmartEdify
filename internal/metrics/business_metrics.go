package metrics

import (
    "context"
    "time"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // Métricas de negocio específicas
    TenantActivityGauge = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "tenant_active_users_current",
            Help: "Current number of active users per tenant",
        },
        []string{"tenant_id", "user_type"},
    )

    AssemblyParticipationRate = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "assembly_participation_rate",
            Help: "Participation rate in assemblies by tenant",
        },
        []string{"tenant_id", "assembly_type"},
    )

    PaymentProcessingTime = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "payment_processing_duration_seconds",
            Help:    "Time taken to process payments",
            Buckets: []float64{1, 5, 10, 30, 60, 120},
        },
        []string{"tenant_id", "payment_method"},
    )

    DocumentSigningRate = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "documents_signed_total",
            Help: "Total number of documents signed",
        },
        []string{"tenant_id", "document_type"},
    )
)

type BusinessMetricsCollector struct {
    tenantActivityTracker map[string]map[string]int
    lastUpdateTime        time.Time
}

func NewBusinessMetricsCollector() *BusinessMetricsCollector {
    return &BusinessMetricsCollector{
        tenantActivityTracker: make(map[string]map[string]int),
        lastUpdateTime:        time.Now(),
    }
}

func (bmc *BusinessMetricsCollector) RecordUserActivity(tenantID, userType string) {
    if bmc.tenantActivityTracker[tenantID] == nil {
        bmc.tenantActivityTracker[tenantID] = make(map[string]int)
    }
    bmc.tenantActivityTracker[tenantID][userType]++
    
    TenantActivityGauge.WithLabelValues(tenantID, userType).Inc()
}

func (bmc *BusinessMetricsCollector) RecordAssemblyParticipation(tenantID, assemblyType string, participationRate float64) {
    AssemblyParticipationRate.WithLabelValues(tenantID, assemblyType).Set(participationRate)
}

func (bmc *BusinessMetricsCollector) RecordPaymentProcessing(tenantID, paymentMethod string, duration time.Duration) {
    PaymentProcessingTime.WithLabelValues(tenantID, paymentMethod).Observe(duration.Seconds())
}

func (bmc *BusinessMetricsCollector) RecordDocumentSigning(tenantID, documentType string) {
    DocumentSigningRate.WithLabelValues(tenantID, documentType).Inc()
}

// Función para limpiar métricas periódicamente
func (bmc *BusinessMetricsCollector) CleanupStaleMetrics(ctx context.Context) {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            // Limpiar métricas de tenants inactivos por más de 24h
            cutoff := time.Now().Add(-24 * time.Hour)
            if bmc.lastUpdateTime.Before(cutoff) {
                // Reset counters for inactive tenants
                for tenantID := range bmc.tenantActivityTracker {
                    for userType := range bmc.tenantActivityTracker[tenantID] {
                        TenantActivityGauge.WithLabelValues(tenantID, userType).Set(0)
                    }
                }
                bmc.tenantActivityTracker = make(map[string]map[string]int)
            }
            bmc.lastUpdateTime = time.Now()
        }
    }
}