package memory

import (
	"github.com/chenquan/go-pkg/xsync"

	"github.com/yunqi/lighthouse/internal/persistence/session"
	sess "github.com/yunqi/lighthouse/internal/session"
)

var _ session.Store = (*Store)(nil)

func New() *Store {
	return &Store{
		m: xsync.NewSharedMap(xsync.WithShardBlockSize(64)),
	}
}

type Store struct {
	m *xsync.SharedMap
}

func (s *Store) Set(session *sess.Session) error {

	s.m.Store(session.ClientID, session)
	return nil
}

func (s *Store) Remove(clientID string) error {
	s.m.Delete(clientID)
	return nil
}

func (s *Store) Get(clientID string) (*sess.Session, error) {
	if val, b := s.m.Load(clientID); b {
		return val.(*sess.Session), nil
	} else {
		return nil, nil
	}

}

func (s *Store) GetAll() ([]*sess.Session, error) {
	sessions := make([]*sess.Session, 0)

	s.m.Range(func(key, value interface{}) bool {
		sessions = append(sessions, value.(*sess.Session))
		return true
	})

	return sessions, nil
}

func (s *Store) SetSessionExpiry(clientID string, expiry uint32) error {

	if s, ok := s.m.Load(clientID); ok {
		s.(*sess.Session).ExpiryInterval = expiry
	}

	return nil
}

func (s *Store) Iterate(fn session.IterateFn) error {
	s.m.Range(func(key, value interface{}) bool {
		return fn(value.(*sess.Session))
	})
	return nil
}
