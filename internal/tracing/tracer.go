package tracing

import (
    "net/http"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type TracingConfig struct {
    ServiceName     string
    ServiceVersion  string
    JaegerEndpoint  string
    Environment     string
}

func InitTracer(config TracingConfig) (*trace.TracerProvider, error) {
    // Crear exportador Jaeger
    exp, err := jaeger.New(jaeger.WithCollectorEndpoint(
        jaeger.WithEndpoint(config.JaegerEndpoint),
    ))
    if err != nil {
        return nil, err
    }

    // Configurar resource
    resource := resource.NewWithAttributes(
        semconv.SchemaURL,
        semconv.ServiceNameKey.String(config.ServiceName),
        semconv.ServiceVersionKey.String(config.ServiceVersion),
        semconv.DeploymentEnvironmentKey.String(config.Environment),
    )

    // Crear TracerProvider
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exp),
        trace.WithResource(resource),
        trace.WithSampler(trace.AlwaysSample()),
    )

    otel.SetTracerProvider(tp)
    return tp, nil
}

func TraceMiddleware(serviceName string) func(http.Handler) http.Handler {
    tracer := otel.Tracer(serviceName)
    
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx, span := tracer.Start(r.Context(), r.Method+" "+r.URL.Path)
            defer span.End()
            
            // Agregar atributos del request
            span.SetAttributes(
                semconv.HTTPMethodKey.String(r.Method),
                semconv.HTTPURLKey.String(r.URL.String()),
                semconv.HTTPUserAgentKey.String(r.UserAgent()),
            )
            
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}