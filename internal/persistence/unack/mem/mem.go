package mem

import (
	"github.com/yunqi/lighthouse/internal/packet"
	"github.com/yunqi/lighthouse/internal/persistence/unack"
)

var _ unack.Store = (*Store)(nil)

type Store struct {
	clientID     string
	unackpublish map[packet.PacketId]struct{}
}

type Options struct {
	ClientID string
}

func New(opts Options) *Store {
	return &Store{
		clientID:     opts.ClientID,
		unackpublish: make(map[packet.PacketId]struct{}),
	}
}

func (s *Store) Init(cleanStart bool) error {
	if cleanStart {
		s.unackpublish = make(map[packet.PacketId]struct{})
	}
	return nil
}

func (s *Store) Set(id packet.PacketId) (bool, error) {
	if _, ok := s.unackpublish[id]; ok {
		return true, nil
	}
	s.unackpublish[id] = struct{}{}
	return false, nil
}

func (s *Store) Remove(id packet.PacketId) error {
	delete(s.unackpublish, id)
	return nil
}
