package redis

import (
	"context"
	"crypto/tls"
	"github.com/chenquan/go-pkg/xsync"
	"github.com/go-redis/redis/v8"
	"github.com/yunqi/lighthouse/internal/xlog"
	"go.uber.org/zap"
	"io"
	"time"
)

const (
	defaultDatabase          = 0
	redisExecuteStartTimeKye = "$redisExecuteStartTime"
)

var (
	clusterManager = xsync.NewResourceManager()
	clientManager  = xsync.NewResourceManager()
)

type Hook struct {
	log      *xlog.Log
	slowTime time.Duration
}

func newHook(slowTime time.Duration) *Hook {
	return &Hook{
		slowTime: slowTime,
		log:      xlog.LoggerModule("redis"),
	}
}

func (h *Hook) BeforeProcess(ctx context.Context, _ redis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, redisExecuteStartTimeKye, time.Now()), nil
}

func (h *Hook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	t := ctx.Value(redisExecuteStartTimeKye)
	if v, ok := t.(time.Time); ok {
		d := time.Now().Sub(v)
		fields := []zap.Field{
			zap.Duration("time", d),
			zap.Any("cmd", cmd.Args()),
		}
		if d > h.slowTime {
			h.log.Warn("redis execute slow time", fields...)
		} else {
			h.log.Debug("redis execute time", fields...)
		}
	}

	return nil
}

func (h *Hook) BeforeProcessPipeline(ctx context.Context, _ []redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (h *Hook) AfterProcessPipeline(_ context.Context, _ []redis.Cmder) error {
	return nil
}

func getCluster(r *Redis) (*redis.ClusterClient, error) {

	val, err := clusterManager.Get(r.addr, func() (io.Closer, error) {
		var tlsConfig *tls.Config
		if r.option.Tls {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		store := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        []string{r.addr},
			Password:     r.option.Pass,
			MaxRetries:   r.option.MaxRetries,
			MinIdleConns: r.option.MinIdleConns,
			TLSConfig:    tlsConfig,
		})

		store.AddHook(newHook(r.option.SlowTime))

		ctx, cancel := context.WithTimeout(context.Background(), r.option.Timeout)
		defer cancel()
		err := store.Ping(ctx).Err()
		if err != nil {
			return nil, err
		}

		return store, nil
	})
	if err != nil {
		return nil, err
	}
	return val.(*redis.ClusterClient), nil
}

func getClient(r *Redis) (*redis.Client, error) {
	val, err := clientManager.Get(r.addr, func() (io.Closer, error) {
		var tlsConfig *tls.Config
		if r.option.Tls {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		store := redis.NewClient(&redis.Options{
			Addr:         r.addr,
			Password:     r.option.Pass,
			MaxRetries:   r.option.MaxRetries,
			MinIdleConns: r.option.MinIdleConns,
			DB:           defaultDatabase,
			TLSConfig:    tlsConfig,
		})

		store.AddHook(newHook(r.option.SlowTime))

		ctx, cancel := context.WithTimeout(context.Background(), r.option.Timeout)
		defer cancel()
		err := store.Ping(ctx).Err()
		if err != nil {
			return nil, err
		}

		return store, nil
	})

	if err != nil {
		return nil, err
	}

	return val.(*redis.Client), nil
}
