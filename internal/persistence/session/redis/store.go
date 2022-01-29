package redis

import (
	"bytes"
	"context"
	"github.com/yunqi/lighthouse/config"
	"github.com/yunqi/lighthouse/internal/persistence"
	"github.com/yunqi/lighthouse/internal/persistence/message/encoding"
	"github.com/yunqi/lighthouse/internal/persistence/session"
	red "github.com/yunqi/lighthouse/internal/redis"
	sess "github.com/yunqi/lighthouse/internal/session"
	"strconv"
	"sync"
	"time"
)

const (
	sessPrefix = "lighthouse:session:"
)

var _ session.Store = (*Store)(nil)

func init() {
	persistence.RegisterSessionStore("redis", New())
}

type Store struct {
	mu sync.RWMutex
	r  *red.Redis
}

func New() session.NewStore {
	return func(config *config.StoreType) (session.Store, error) {
		var opts []red.Option
		switch config.Type {
		case red.NodeType:
			opts = append(opts, red.WithNodeType())
		case red.ClusterType:
			opts = append(opts, red.WithClusterType())
		}
		return &Store{
			r: red.New(config.Redis.Addr, opts...),
		}, nil
	}
}

func getKey(clientID string) string {
	return sessPrefix + clientID
}

func (s *Store) Set(ctx context.Context, session *sess.Session) error {
	s.mu.Lock()

	b := &bytes.Buffer{}
	encoding.EncodeMessage(session.Will, b)

	err := s.r.Hmset(ctx, getKey(session.ClientId), map[string]interface{}{
		"client_id":           session.ClientId,
		"will":                b.Bytes(),
		"will_delay_interval": session.WillDelayInterval,
		"connected_at":        session.ConnectedAt.Unix(),
		"expiry_interval":     session.ExpiryInterval,
	})

	s.mu.Unlock()

	return err
}

func (s *Store) Remove(ctx context.Context, clientID string) error {
	s.mu.Lock()
	_, err := s.r.Del(ctx, getKey(clientID))
	s.mu.Unlock()

	return err
}

func (s *Store) Get(ctx context.Context, clientID string) (*sess.Session, error) {
	s.mu.RLock()
	_session, err := s.getSessionLocked(ctx, getKey(clientID))
	s.mu.RUnlock()
	return _session, err
}

func (s *Store) getSessionLocked(ctx context.Context, key string) (*sess.Session, error) {
	m, err := s.r.Hmget(ctx, key, "client_id", "will", "will_delay_interval", "connected_at", "expiry_interval")
	if err != nil {
		return nil, err
	}

	_sess := &sess.Session{}
	if m[0] != nil {
		_sess.ClientId = m[0].(string)
	}
	if m[1] != nil {
		_sess.Will, err = encoding.DecodeMessageFromBytes([]byte(m[1].(string)))
	}
	if m[2] != nil {
		parseUint, err := strconv.ParseUint(m[2].(string), 10, 32)
		if err != nil {
			return nil, err
		}
		_sess.WillDelayInterval = uint32(parseUint)
	}
	if m[3] != nil {
		parseInt, err := strconv.ParseInt(m[3].(string), 10, 64)
		if err != nil {
			return nil, err
		}
		_sess.ConnectedAt = time.Unix(parseInt, 0)
	}
	if m[4] != nil {
		parseUint, err := strconv.ParseUint(m[4].(string), 10, 32)
		if err != nil {
			return nil, err
		}
		_sess.ExpiryInterval = uint32(parseUint)
	}

	return _sess, nil
}

func (s *Store) SetSessionExpiry(ctx context.Context, clientID string, expiry uint32) error {
	s.mu.Lock()

	err := s.r.Hmset(ctx, getKey(clientID), map[string]interface{}{
		"expiry_interval": expiry,
	})

	s.mu.Unlock()

	return err
}

func (s *Store) Iterate(ctx context.Context, fn session.IterateFn) error {
	s.mu.RLock()

	keys, _, err := s.r.Scan(ctx, 0, sessPrefix+"*", 0)
	if err != nil {
		return err
	}

	for _, key := range keys {
		_sess, err := s.getSessionLocked(ctx, key)
		if err != nil {
			return err
		}

		if !fn(_sess) {
			return nil
		}
	}

	s.mu.RUnlock()

	return nil
}
