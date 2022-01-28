package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/yunqi/lighthouse/internal/xlog"
	"github.com/yunqi/lighthouse/internal/xtrace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"time"
)

type Hook struct {
	log      *xlog.Log
	slowTime time.Duration
	tracer   trace.Tracer
}

func newHook(slowTime time.Duration) *Hook {
	return &Hook{
		slowTime: slowTime,
		log:      xlog.LoggerModule("redis"),
		tracer:   otel.GetTracerProvider().Tracer(xtrace.Name),
	}
}

func (h *Hook) BeforeProcess(ctx context.Context, _ redis.Cmder) (context.Context, error) {
	ctx, span := h.tracer.Start(ctx, "redis")
	return context.WithValue(context.WithValue(ctx, redisExecuteStartTimeKye, time.Now()), redisSpan, span), nil
}

func (h *Hook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {

	logger := h.log.WithContext(ctx)
	t := ctx.Value(redisExecuteStartTimeKye)
	if span, ok := ctx.Value(redisSpan).(trace.Span); ok {
		defer span.End()
	}

	if v, ok := t.(time.Time); ok {
		duration := time.Since(v)

		fields := []zap.Field{
			zap.Duration("duration", duration),
			zap.Any("cmd", cmd.Args()),
		}

		err := cmd.Err()
		if err != nil {
			logger.Error("redis", append(fields, zap.Error(err))...)
			return err
		}

		if duration > h.slowTime {
			logger.Warn("redis execute slow time", fields...)
		} else {
			logger.Debug("redis execute time", fields...)
		}
	}

	return cmd.Err()
}

func (h *Hook) BeforeProcessPipeline(ctx context.Context, _ []redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (h *Hook) AfterProcessPipeline(_ context.Context, _ []redis.Cmder) error {
	return nil
}
