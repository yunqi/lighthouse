package redis

import (
	"context"
	"fmt"
	red "github.com/go-redis/redis/v8"
	"github.com/yunqi/lighthouse/internal/breaker"
	"io"
	"time"
)

const (
	NodeType    = "ClusterType"
	ClusterType = "Type"
)

type (
	Type  string
	Redis struct {
		addr   string
		option *option
		brk    breaker.Breaker
	}
	option struct {
		Type         Type
		Pass         string
		Tls          bool
		MaxRetries   int
		MinIdleConns int
		Timeout      time.Duration
		SlowTime     time.Duration
	}

	Option func(opt *option)

	// Client interface represents a redis node.
	Client interface {
		red.Cmdable
		io.Closer
	}
)

func (r Type) String() string {
	return string(r)
}

func WithNodeType() Option {
	return func(opt *option) {
		opt.Type = NodeType
	}
}

func WithClusterType() Option {
	return func(opt *option) {
		opt.Type = ClusterType
	}
}

// New returns a Redis with given options.
func New(addr string, opts ...Option) *Redis {
	_option := new(option)
	_option.Type = NodeType
	_option.Timeout = time.Second * 20
	_option.SlowTime = time.Second * 3
	for _, opt := range opts {
		opt(_option)
	}

	r := &Redis{
		addr:   addr,
		option: _option,
		brk:    breaker.NewBreaker(addr),
	}

	return r
}

func (r *Redis) getRedis() (Client, error) {
	switch r.option.Type {
	case ClusterType:
		return getCluster(r)
	case NodeType:
		return getClient(r)
	default:
		return nil, fmt.Errorf("redis type '%s' is not supported", r.option.Type)
	}
}

// Hmget is the implementation of redis hmget command.
func (r *Redis) Hmget(key string, fields ...string) (val []interface{}, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		conn, err := r.getRedis()
		if err != nil {
			return err
		}
		ctx, cancelFunc := r.getContext()
		defer cancelFunc()

		v, err := conn.HMGet(ctx, key, fields...).Result()
		if err != nil {
			return err
		}
		val = v

		return nil
	}, acceptable)

	return
}

// Hset is the implementation of redis hset command.
func (r *Redis) Hset(key, field string, value interface{}) error {
	return r.brk.DoWithAcceptable(func() error {
		conn, err := r.getRedis()
		if err != nil {
			return err
		}
		ctx, cancelFunc := r.getContext()
		defer cancelFunc()

		return conn.HSet(ctx, key, field, value).Err()
	}, acceptable)
}

// Hdel is the implementation of redis hdel command.
func (r *Redis) Hdel(key string, fields ...string) (val bool, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		conn, err := r.getRedis()
		if err != nil {
			return err
		}

		ctx, cancelFunc := r.getContext()
		defer cancelFunc()
		v, err := conn.HDel(ctx, key, fields...).Result()
		if err != nil {
			return err
		}

		val = v == 1
		return nil
	}, acceptable)

	return
}

// Hmset is the implementation of redis hmset command.
func (r *Redis) Hmset(key string, fieldsAndValues map[string]interface{}) error {
	return r.brk.DoWithAcceptable(func() error {
		conn, err := r.getRedis()
		if err != nil {
			return err
		}
		ctx, cancelFunc := r.getContext()
		defer cancelFunc()
		return conn.HMSet(ctx, key, fieldsAndValues).Err()
	}, acceptable)
}

// Hscan is the implementation of redis hscan command.
func (r *Redis) Hscan(key string, cursor uint64, match string, count int64) (keys []string, cur uint64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		conn, err := r.getRedis()
		if err != nil {
			return err
		}
		ctx, cancelFunc := r.getContext()
		defer cancelFunc()
		keys, cur, err = conn.HScan(ctx, key, cursor, match, count).Result()
		return err
	}, acceptable)

	return
}

// Scan is the implementation of redis scan command.
func (r *Redis) Scan(cursor uint64, match string, count int64) (keys []string, cur uint64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		conn, err := r.getRedis()
		if err != nil {
			return err
		}
		ctx, cancelFunc := r.getContext()
		defer cancelFunc()
		keys, cur, err = conn.Scan(ctx, cursor, match, count).Result()
		return err
	}, acceptable)

	return
}

func acceptable(err error) bool {
	return err == nil || err == red.Nil
}

// Del deletes keys.
func (r *Redis) Del(keys ...string) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		conn, err := r.getRedis()
		if err != nil {
			return err
		}
		ctx, cancelFunc := r.getContext()
		defer cancelFunc()
		v, err := conn.Del(ctx, keys...).Result()
		if err != nil {
			return err
		}

		val = int(v)
		return nil
	}, acceptable)

	return
}

func (r *Redis) Ping() error {
	err := r.brk.DoWithAcceptable(func() error {
		conn, err := r.getRedis()
		if err != nil {
			return err
		}
		ctx, cancelFunc := r.getContext()
		defer cancelFunc()
		err = conn.Ping(ctx).Err()
		return err
	}, acceptable)

	return err
}

// Llen is the implementation of redis llen command.
func (r *Redis) Llen(key string) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		conn, err := r.getRedis()
		if err != nil {
			return err
		}
		ctx, cancelFunc := r.getContext()
		defer cancelFunc()
		v, err := conn.LLen(ctx, key).Result()
		if err != nil {
			return err
		}

		val = int(v)
		return nil
	}, acceptable)

	return
}

// Hgetall is the implementation of redis hgetall command.
func (r *Redis) Hgetall(key string) (val map[string]string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		conn, err := r.getRedis()
		if err != nil {
			return err
		}
		ctx, cancelFunc := r.getContext()
		defer cancelFunc()
		val, err = conn.HGetAll(ctx, key).Result()
		return err
	}, acceptable)

	return
}

// Lrem is the implementation of redis lrem command.
func (r *Redis) Lrem(key string, count int, value interface{}) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		conn, err := r.getRedis()
		if err != nil {
			return err
		}
		ctx, cancelFunc := r.getContext()
		defer cancelFunc()
		v, err := conn.LRem(ctx, key, int64(count), value).Result()
		if err != nil {
			return err
		}

		val = int(v)
		return nil
	}, acceptable)

	return
}

// Rpush is the implementation of redis rpush command.
func (r *Redis) Rpush(key string, values ...interface{}) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		conn, err := r.getRedis()
		if err != nil {
			return err
		}
		ctx, cancelFunc := r.getContext()
		defer cancelFunc()
		v, err := conn.RPush(ctx, key, values...).Result()
		if err != nil {
			return err
		}

		val = int(v)
		return nil
	}, acceptable)

	return
}

// Lrange is the implementation of redis lrange command.
func (r *Redis) Lrange(key string, start, stop int) (val []string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		conn, err := r.getRedis()
		if err != nil {
			return err
		}
		ctx, cancelFunc := r.getContext()
		defer cancelFunc()
		val, err = conn.LRange(ctx, key, int64(start), int64(stop)).Result()
		return err
	}, acceptable)

	return
}

// Lset is the implementation of redis lset command.
func (r *Redis) Lset(key string, index int64, value interface{}) (val string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		conn, err := r.getRedis()
		if err != nil {
			return err
		}
		ctx, cancelFunc := r.getContext()
		defer cancelFunc()
		val, err = conn.LSet(ctx, key, index, value).Result()
		return err
	}, acceptable)

	return
}
func (r *Redis) getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), r.option.Timeout)
}

func (r *Redis) Close() error {
	conn, err := r.getRedis()
	if err != nil {
		return err
	}
	return conn.Close()
}
