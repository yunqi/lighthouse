package trie

import (
	"github.com/yunqi/lighthouse/internal/persistence/message"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrieDB_ClearAll(t *testing.T) {
	a := assert.New(t)
	s := NewStore()
	s.AddOrReplace(&message.Message{
		Topic: "a/b/c",
	})
	s.AddOrReplace(&message.Message{
		Topic:   "a/b/c/d",
		Payload: []byte{1, 2, 3},
	})
	s.ClearAll()
	a.Nil(s.GetRetainedMessage("a/b/c"))
	a.Nil(s.GetRetainedMessage("a/b/c/d"))
}

func TestTrieDB_GetRetainedMessage(t *testing.T) {
	a := assert.New(t)
	s := NewStore()
	tt := []*message.Message{
		{
			Topic:   "a/b/c/d",
			Payload: []byte{1, 2, 3},
		},
		{
			Topic:   "a/b/c/",
			Payload: []byte{1, 2, 3, 4},
		},
		{
			Topic:   "a/",
			Payload: []byte{1, 2, 3},
		},
	}
	for _, v := range tt {
		s.AddOrReplace(v)
	}
	for _, v := range tt {
		rs := s.GetRetainedMessage(v.Topic)
		a.Equal(v.Topic, rs.Topic)
		a.Equal(v.Payload, rs.Payload)
	}
	a.Nil(s.GetRetainedMessage("a/b"))
}

func TestTrieDB_GetMatchedMessages(t *testing.T) {
	a := assert.New(t)
	s := NewStore()
	msgs := []*message.Message{
		{
			Topic:   "a/b/c/d",
			Payload: []byte{1, 2, 3},
		},
		{
			Topic:   "a/b/c/",
			Payload: []byte{1, 2, 3, 4},
		},
		{
			Topic:   "a/",
			Payload: []byte{1, 2, 3},
		},
		{
			Topic:   "a/b",
			Payload: []byte{1, 2, 3},
		},
		{
			Topic:   "b/a",
			Payload: []byte{1, 2, 3},
		},
		{
			Topic:   "a",
			Payload: []byte{1, 2, 3},
		},
	}
	tt := []struct {
		TopicFilter string
		expected    map[string]*message.Message
	}{
		{
			TopicFilter: "a/+/c/",
			expected: map[string]*message.Message{
				"a/b/c/": {
					Payload: []byte{1, 2, 3, 4},
				},
			},
		},
		{
			TopicFilter: "a/+",
			expected: map[string]*message.Message{
				"a/": {
					Payload: []byte{1, 2, 3},
				},
				"a/b": {
					Payload: []byte{1, 2, 3},
				},
			},
		},
		{
			TopicFilter: "#",
			expected: map[string]*message.Message{
				"a/b/c/d": {
					Payload: []byte{1, 2, 3},
				},
				"a/b/c/": {
					Payload: []byte{1, 2, 3, 4},
				},
				"a/": {
					Payload: []byte{1, 2, 3},
				},
				"a/b": {
					Payload: []byte{1, 2, 3},
				},
				"b/a": {
					Payload: []byte{1, 2, 3},
				},
				"a": {
					Payload: []byte{1, 2, 3},
				},
			},
		},
		{
			TopicFilter: "a/#",
			expected: map[string]*message.Message{
				"a/b/c/d": {
					Payload: []byte{1, 2, 3},
				},
				"a/b/c/": {
					Payload: []byte{1, 2, 3, 4},
				},
				"a/": {
					Payload: []byte{1, 2, 3},
				},
				"a/b": {
					Payload: []byte{1, 2, 3},
				},
				"a": {
					Payload: []byte{1, 2, 3},
				},
			},
		},
		{
			TopicFilter: "a/b/c/d",
			expected: map[string]*message.Message{
				"a/b/c/d": {
					Payload: []byte{1, 2, 3},
				},
			},
		},
	}
	for _, v := range msgs {
		s.AddOrReplace(v)
	}
	for _, v := range tt {
		t.Run(v.TopicFilter, func(t *testing.T) {
			rs := s.GetMatchedMessages(v.TopicFilter)
			a.Equal(len(v.expected), len(rs))
			got := make(map[string]*message.Message)
			for _, v := range rs {
				got[v.Topic] = v
			}
			for k, v := range v.expected {
				a.Equal(v.Payload, got[k].Payload)
			}
		})
	}
}

func TestTrieDB_Remove(t *testing.T) {
	a := assert.New(t)
	s := NewStore()
	s.AddOrReplace(&message.Message{
		Topic: "a/b/c",
	})
	s.AddOrReplace(&message.Message{
		Topic:   "a/b/c/d",
		Payload: []byte{1, 2, 3},
	})
	a.NotNil(s.GetRetainedMessage("a/b/c"))
	s.Remove("a/b/c")
	a.Nil(s.GetRetainedMessage("a/b/c"))
}

func TestTrieDB_Iterate(t *testing.T) {
	a := assert.New(t)
	s := NewStore()
	msgs := []*message.Message{
		{
			Topic:   "a/b/c/d",
			Payload: []byte{1, 2, 3},
		},
		{
			Topic:   "a/b/c/",
			Payload: []byte{1, 2, 3, 4},
		},
		{
			Topic:   "a/",
			Payload: []byte{1, 2, 3},
		},
		{
			Topic:   "a/b",
			Payload: []byte{1, 2, 3},
		},
		{
			Topic:   "a",
			Payload: []byte{1, 2, 3},
		},
		{
			Topic:   "$SYS/a/b",
			Payload: []byte{1, 2, 3},
		},
	}

	for _, v := range msgs {
		s.AddOrReplace(v)
	}
	var rs []*message.Message
	s.Iterate(func(message *message.Message) bool {
		rs = append(rs, message)
		return true
	})
	a.ElementsMatch(msgs, rs)
}

func TestTrieDB_Iterate_Cancel(t *testing.T) {
	a := assert.New(t)
	s := NewStore()
	msgs := []*message.Message{
		{
			Topic:   "a/b/c/d",
			Payload: []byte{1, 2, 3},
		},
		{
			Topic:   "a/b/c/",
			Payload: []byte{1, 2, 3, 4},
		},
		{
			Topic:   "a/",
			Payload: []byte{1, 2, 3},
		},
		{
			Topic:   "a/b",
			Payload: []byte{1, 2, 3},
		},
		{
			Topic:   "a",
			Payload: []byte{1, 2, 3},
		},
	}

	for _, v := range msgs {
		s.AddOrReplace(v)
	}
	var i int
	var rs []*message.Message
	s.Iterate(func(message *message.Message) bool {
		if i == 2 {
			return false
		}
		rs = append(rs, message)
		i++
		return true
	})
	a.Len(rs, 2)

}
