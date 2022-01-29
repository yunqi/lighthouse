package config

import "time"

type (
	Persistence struct {
		Session      StoreType `yaml:"session"`
		Subscription StoreType `yaml:"subscription"`
		Queue        StoreType `yaml:"queue"`
	}

	StoreType struct {
		Type  string         `yaml:"type"` // memory|redis
		Redis RedisStoreType `yaml:"redis"`
	}

	RedisStoreType struct {
		Type string `yaml:"nodeType"`
		// Addr is the redis server address.
		// If empty, use "127.0.0.1:6379" as default.
		Addr string `yaml:"addr"`
		// Password is the redis password.
		Password string `yaml:"password"`
		// Database is the number of the redis database to be connected.
		Database uint `yaml:"database"`
		// MaxIdle is the maximum number of idle connections in the pool.
		// If nil, use 1000 as default.
		// This value will pass to redis.Pool.MaxIde.
		MaxIdle *uint `yaml:"maxIdle"`
		// MaxActive is the maximum number of connections allocated by the pool at a given time.
		// If nil, use 0 as default.
		// If zero, there is no limit on the number of connections in the pool.
		// This value will pass to redis.Pool.MaxActive.
		MaxActive *uint `yaml:"maxActive"`
		// Close connections after remaining idle for this duration. If the value
		// is zero, then idle connections are not closed. Applications should set
		// the timeout to a value less than the server's timeout.
		// Ff zero, use 240 * time.Second as default.
		// This value will pass to redis.Pool.IdleTimeout.
		IdleTimeout time.Duration `yaml:"idleTimeout"`
		Timeout     time.Duration `yaml:"timeout"`
	}
)
