package redis

import (
	"context"
	"github.com/yunqi/lighthouse/internal/packet"
	"github.com/yunqi/lighthouse/internal/persistence/unack"
	red "github.com/yunqi/lighthouse/internal/redis"
	"strconv"
	"time"
)

const (
	unackPrefix = "lighthouse:unack:"
)

var _ unack.Store = (*Store)(nil)

type Store struct {
	key          string
	clientID     string
	r            *red.Redis
	unackpublish map[packet.PacketId]struct{}
	timeout      time.Duration
}

type Options struct {
	ClientID string
	R        *red.Redis
}

func New(opts Options) *Store {
	return &Store{
		clientID:     opts.ClientID,
		key:          getKey(opts.ClientID),
		r:            opts.R,
		unackpublish: make(map[packet.PacketId]struct{}),
	}
}

func getKey(clientID string) string {
	return unackPrefix + clientID
}
func (s *Store) Init(cleanStart bool) error {
	if cleanStart {
		s.unackpublish = make(map[packet.PacketId]struct{})

		_, err := s.r.Del(s.key)
		if err != nil {
			return err
		}
	}
	return nil
}
func (s *Store) getContext() (context.Context, context.CancelFunc) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), s.timeout)
	return ctx, cancelFunc
}

func (s *Store) Set(id packet.PacketId) (bool, error) {
	// from cache
	if _, ok := s.unackpublish[id]; ok {
		return true, nil
	}
	err := s.r.Hset(s.key, strconv.FormatUint(uint64(id), 10), "1")
	if err != nil {
		return false, err
	}
	s.unackpublish[id] = struct{}{}
	return false, nil
}

func (s *Store) Remove(id packet.PacketId) error {
	_, err := s.r.Hdel(s.key, strconv.FormatUint(uint64(id), 10))
	if err != nil {
		return err
	}
	delete(s.unackpublish, id)
	return nil
}
