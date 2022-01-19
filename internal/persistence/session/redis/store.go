package redis

import (
	"bytes"
	"github.com/yunqi/lighthouse/internal/persistence/message/binary"
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

type Store struct {
	mu sync.RWMutex
	r  *red.Redis
}

func New(r *red.Redis) *Store {
	return &Store{
		r: r,
	}
}

func getKey(clientID string) string {
	return sessPrefix + clientID
}

func (s *Store) Set(session *sess.Session) error {
	s.mu.Lock()

	b := &bytes.Buffer{}
	binary.EncodeMessage(session.Will, b)

	err := s.r.Hmset(getKey(session.ClientID), map[string]interface{}{
		"client_id":           session.ClientID,
		"will":                b.Bytes(),
		"will_delay_interval": session.WillDelayInterval,
		"connected_at":        session.ConnectedAt.Unix(),
		"expiry_interval":     session.ExpiryInterval,
	})

	s.mu.Unlock()

	return err
}

func (s *Store) Remove(clientID string) error {
	s.mu.Lock()
	_, err := s.r.Del(getKey(clientID))
	s.mu.Unlock()

	return err
}

func (s *Store) Get(clientID string) (*sess.Session, error) {
	s.mu.RLock()
	_session, err := s.getSessionLocked(getKey(clientID))
	s.mu.RUnlock()
	return _session, err
}

func (s *Store) getSessionLocked(key string) (*sess.Session, error) {
	m, err := s.r.Hmget(key, "client_id", "will", "will_delay_interval", "connected_at", "expiry_interval")
	if err != nil {
		return nil, err
	}

	_sess := &sess.Session{}
	if m[0] != nil {
		_sess.ClientID = m[0].(string)
	}
	if m[1] != nil {
		_sess.Will, err = binary.DecodeMessageFromBytes([]byte(m[1].(string)))
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

func (s *Store) SetSessionExpiry(clientID string, expiry uint32) error {
	s.mu.Lock()

	err := s.r.Hmset(getKey(clientID), map[string]interface{}{
		"expiry_interval": expiry,
	})

	s.mu.Unlock()

	return err
}

func (s *Store) Iterate(fn session.IterateFn) error {
	s.mu.RLock()

	keys, _, err := s.r.Scan(0, sessPrefix+"*", 0)
	if err != nil {
		return err
	}

	for _, key := range keys {
		_sess, err := s.getSessionLocked(key)
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
