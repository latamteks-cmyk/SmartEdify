package monitoring

import (
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
)

func InitJaeger(serviceName string) (opentracing.Tracer, func(), error) {
	cfg := config.Configuration{
		ServiceName: serviceName,
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: "127.0.0.1:6831",
		},
	}
	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		return nil, nil, err
	}
	opentracing.SetGlobalTracer(tracer)
	return tracer, func() { closer.Close() }, nil
}
