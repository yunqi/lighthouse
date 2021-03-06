package memory

import (
	"context"
	"github.com/chenquan/go-pkg/xsync"
	"github.com/yunqi/lighthouse/config"
	"github.com/yunqi/lighthouse/internal/persistence"

	"github.com/yunqi/lighthouse/internal/persistence/session"
	sess "github.com/yunqi/lighthouse/internal/session"
)

var _ session.Store = (*Store)(nil)

func init() {
	persistence.RegisterSessionStore("memory", New())
}
func New() session.NewStore {
	return func(config *config.StoreType) (session.Store, error) {
		return &Store{
			m: xsync.NewSharedMap(xsync.WithShardBlockSize(64)),
		}, nil
	}
}

type Store struct {
	m *xsync.SharedMap
}

func (s *Store) Set(ctx context.Context, session *sess.Session) error {

	s.m.Store(session.ClientId, session)
	return nil
}

func (s *Store) Remove(ctx context.Context, clientID string) error {
	s.m.Delete(clientID)
	return nil
}

func (s *Store) Get(ctx context.Context, clientID string) (*sess.Session, error) {
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

func (s *Store) SetSessionExpiry(ctx context.Context, clientID string, expiry uint32) error {

	if s, ok := s.m.Load(clientID); ok {
		s.(*sess.Session).ExpiryInterval = expiry
	}

	return nil
}

func (s *Store) Iterate(ctx context.Context, fn session.IterateFn) error {
	s.m.Range(func(key, value interface{}) bool {
		return fn(value.(*sess.Session))
	})
	return nil
}
