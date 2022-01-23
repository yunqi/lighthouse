package redis

import (
	"bytes"
	"github.com/chenquan/go-pkg/xbinary"
	"github.com/yunqi/lighthouse/internal/persistence/subscription"
	"github.com/yunqi/lighthouse/internal/persistence/subscription/memory"
	red "github.com/yunqi/lighthouse/internal/redis"
	subsc "github.com/yunqi/lighthouse/internal/subscription"
	"strings"
	"sync"
)

const (
	subPrefix = "lighthouse:subscription:"
)

var _ subscription.Store = (*sub)(nil)

func EncodeSubscription(sub *subsc.Subscription) []byte {
	w := &bytes.Buffer{}
	_ = xbinary.WriteBytes(w, []byte(sub.ShareName))
	_ = xbinary.WriteBytes(w, []byte(sub.TopicFilter))
	_ = xbinary.WriteUint32(w, sub.ID)
	w.WriteByte(sub.QoS)
	_ = xbinary.WriteBool(w, sub.NoLocal)
	_ = xbinary.WriteBool(w, sub.RetainAsPublished)
	w.WriteByte(sub.RetainHandling)
	return w.Bytes()
}

func DecodeSubscription(b []byte) (*subsc.Subscription, error) {
	sub := &subsc.Subscription{}
	r := bytes.NewBuffer(b)
	share, err := xbinary.ReadBytes(r)
	if err != nil {
		return &subsc.Subscription{}, err
	}
	sub.ShareName = string(share)
	topic, err := xbinary.ReadBytes(r)
	if err != nil {
		return &subsc.Subscription{}, err
	}
	sub.TopicFilter = string(topic)
	sub.ID, err = xbinary.ReadUint32(r)
	if err != nil {
		return &subsc.Subscription{}, err
	}
	sub.QoS, err = r.ReadByte()
	if err != nil {
		return &subsc.Subscription{}, err
	}
	sub.NoLocal, err = xbinary.ReadBool(r)
	if err != nil {
		return &subsc.Subscription{}, err
	}
	sub.RetainAsPublished, err = xbinary.ReadBool(r)
	if err != nil {
		return &subsc.Subscription{}, err
	}
	sub.RetainHandling, err = r.ReadByte()
	if err != nil {
		return nil, err
	}
	return sub, nil
}

func New(r *red.Redis) *sub {
	return &sub{
		mu:       &sync.Mutex{},
		memStore: memory.NewStore(),
		r:        r,
	}
}

type sub struct {
	mu       *sync.Mutex
	memStore *memory.TrieDB
	r        *red.Redis
}

// Init loads the subscriptions of given clientIDs from backend into memory.
func (s *sub) Init(clientIDs []string) error {
	if len(clientIDs) == 0 {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, clientId := range clientIDs {
		rs, err := s.r.Hgetall(subPrefix + clientId)
		if err != nil {
			return err
		}

		for _, value := range rs {
			sub, err := DecodeSubscription([]byte(value))
			if err != nil {
				return err
			}
			s.memStore.SubscribeLocked(strings.TrimLeft(clientId, subPrefix), sub)
		}
	}
	return nil
}

func (s *sub) Close() error {
	_ = s.memStore.Close()
	return nil
}

func (s *sub) Subscribe(clientID string, subscriptions ...*subsc.Subscription) (rs subscription.SubscribeResult, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// hset sub:clientID topicFilter xxx
	m := map[string]interface{}{}
	for _, v := range subscriptions {
		m[subscription.GetFullTopicName(v.ShareName, v.TopicFilter)] = EncodeSubscription(v)
	}
	err = s.r.Hmset(subPrefix+clientID, m)
	if err != nil {
		return nil, err
	}
	rs = s.memStore.SubscribeLocked(clientID, subscriptions...)
	return rs, nil
}

func (s *sub) Unsubscribe(clientID string, topics ...string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.r.Hdel(subPrefix+clientID, topics...)
	if err != nil {
		return err
	}
	s.memStore.UnsubscribeLocked(clientID, topics...)
	return nil
}

func (s *sub) UnsubscribeAll(clientID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.r.Del(subPrefix + clientID)
	if err != nil {
		return err
	}
	s.memStore.UnsubscribeAllLocked(clientID)
	return nil
}

func (s *sub) Iterate(fn subscription.IterateFn, options subscription.IterationOptions) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.memStore.IterateLocked(fn, options)
}

func (s *sub) GetStats() subscription.Stats {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.memStore.GetStatusLocked()
}

func (s *sub) GetClientStats(clientID string) (subscription.Stats, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.memStore.GetClientStatsLocked(clientID)
}