package redis

import (
	"context"
	"crypto/tls"
	"github.com/chenquan/go-pkg/xsync"
	"github.com/go-redis/redis/v8"
	"io"
)

const (
	defaultDatabase          = 0
	redisExecuteStartTimeKye = "$redisExecuteStartTime"
	redisSpan                = "$redisSpan"
)

var (
	clusterManager = xsync.NewResourceManager()
	clientManager  = xsync.NewResourceManager()
)

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
		ctx := context.Background()

		if r.option.Timeout != 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, r.option.Timeout)
			defer cancel()
		}

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
