package store

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/yunqi/lighthouse/internal/packet"
	"github.com/yunqi/lighthouse/internal/persistence/queue"
	"testing"
	"time"
)

func TestNewMemory(t *testing.T) {
	memory := NewMemory()
	assert.NotNil(t, memory)
}

func TestMemory_Add(t *testing.T) {
	memory := NewMemory()
	now := time.Now()
	memory.Add(&queue.Element{
		At:     now,
		Expiry: now,
	})
	front := memory.l.Front()
	element := front.Value.(*queue.Element)
	assert.Equal(t, now, element.Expiry)
	assert.Equal(t, now, element.At)

	time1 := time.UnixMilli(1)
	memory.Add(&queue.Element{
		At:     time1,
		Expiry: time1,
	})

	front = memory.l.Front()
	next := front.Value.(*queue.Element)
	assert.Equal(t, time1, next.Expiry)
	assert.Equal(t, time1, next.At)

}

func TestMemory_Len(t *testing.T) {
	memory := NewMemory()
	now := time.Now()
	memory.Add(&queue.Element{
		At:     now,
		Expiry: now,
	})
	assert.Equal(t, 1, memory.Len())

	memory.Add(&queue.Element{
		At:     now,
		Expiry: now,
	})
	assert.Equal(t, 2, memory.Len())
}
func TestMemory_Front(t *testing.T) {
	memory := NewMemory()
	now := time.Now()
	memory.Add(&queue.Element{
		At:     now,
		Expiry: now,
	})
	assert.Equal(t, memory.l.Front().Value.(*queue.Element), memory.Front())
}

func TestMemory_Remove(t *testing.T) {
	memory := NewMemory()
	now := time.Now()
	e := &queue.Element{
		At:     now,
		Expiry: now,
	}
	memory.Add(e)
	assert.Equal(t, e, memory.l.Front().Value.(*queue.Element))

	assert.Equal(t, 1, memory.l.Len())
	memory.Remove(e)
	assert.Equal(t, 0, memory.l.Len())

}

func TestMemory_Replace(t *testing.T) {
	memory := NewMemory()
	now := time.Now()
	controller := gomock.NewController(t)
	message := queue.NewMockMessage(controller)
	message.EXPECT().Id().AnyTimes().Return(packet.PacketId(1))
	e1 := &queue.Element{
		At:      now,
		Expiry:  now,
		Message: message,
	}
	memory.Add(e1)

	time1 := time.UnixMilli(1)
	e2 := &queue.Element{
		At:      time1,
		Expiry:  time1,
		Message: message,
	}
	memory.Add(e2)

	replaced, err := memory.Replace(e2)
	assert.True(t, replaced)
	assert.NoError(t, err)
	assert.Equal(t, time1, memory.l.Front().Value.(*queue.Element).At)

}
