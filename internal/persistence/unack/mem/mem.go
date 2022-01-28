package mem

import (
	"context"
	"github.com/yunqi/lighthouse/internal/packet"
	"github.com/yunqi/lighthouse/internal/persistence/unack"
)

var _ unack.Store = (*Store)(nil)

type Store struct {
	clientID     string
	unackpublish map[packet.Id]struct{}
}

type Options struct {
	ClientID string
}

func New(opts Options) *Store {
	return &Store{
		clientID:     opts.ClientID,
		unackpublish: make(map[packet.Id]struct{}),
	}
}

func (s *Store) Init(_ context.Context, cleanStart bool) error {
	if cleanStart {
		s.unackpublish = make(map[packet.Id]struct{})
	}
	return nil
}

func (s *Store) Set(_ context.Context, id packet.Id) (bool, error) {
	if _, ok := s.unackpublish[id]; ok {
		return true, nil
	}
	s.unackpublish[id] = struct{}{}
	return false, nil
}

func (s *Store) Remove(_ context.Context, id packet.Id) error {
	delete(s.unackpublish, id)
	return nil
}
