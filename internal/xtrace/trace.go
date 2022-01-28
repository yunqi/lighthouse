package xtrace

import (
	"errors"
	"fmt"
	"github.com/yunqi/lighthouse/config"
	"github.com/yunqi/lighthouse/internal/xlog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.uber.org/zap"
	"sync"
)

const Name = "lighthouse"

const (
	Jaeger = "jaeger"
	Zipkin = "zipkin"
)

var (
	ErrUnknownExporter = errors.New("unknown exporter error")
)

var (
	agents = make(map[string]struct{})
	lock   sync.Mutex
)

// StartAgent starts a opentelemetry agent.
func StartAgent(c *config.Trace) {
	lock.Lock()
	defer lock.Unlock()

	_, ok := agents[c.Endpoint]
	if ok {
		return
	}

	// if error happens, let later calls run.
	if err := startAgent(c); err != nil {
		return
	}

	agents[c.Endpoint] = struct{}{}
}

func startAgent(c *config.Trace) error {
	opts := []sdktrace.TracerProviderOption{
		// Set the sampling rate based on the parent span to 100%
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(c.Sampler))),
		// Record information about this application in a Resource.
		sdktrace.WithResource(resource.NewSchemaless(semconv.ServiceNameKey.String(c.Name))),
	}

	if len(c.Endpoint) > 0 {
		exp, err := createExporter(c)
		if err != nil {
			xlog.Panic("opentelemetry exporter err", zap.Error(err))
			return err
		}

		// Always be sure to batch in production.
		opts = append(opts, sdktrace.WithBatcher(exp))
	}

	tp := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))

	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		xlog.Error("[opentelemetry] error", zap.Error(err))
	}))

	return nil
}
func createExporter(c *config.Trace) (sdktrace.SpanExporter, error) {
	// Just support jaeger and zipkin now, more for later
	switch c.Batcher {
	case Jaeger:
		return jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(c.Endpoint)))
	case Zipkin:
		return zipkin.New(c.Endpoint)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownExporter, c.Batcher)
	}
}
