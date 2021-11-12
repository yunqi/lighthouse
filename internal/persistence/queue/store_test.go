/*
 *    Copyright 2021 chenquan
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package queue

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/yunqi/lighthouse/internal/packet"
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
	memory.Add(&Element{
		At:     now,
		Expiry: now,
	})
	front := memory.l.Front()
	element := front.Value.(*Element)
	assert.Equal(t, now, element.Expiry)
	assert.Equal(t, now, element.At)

	time1 := time.UnixMilli(1)
	memory.Add(&Element{
		At:     time1,
		Expiry: time1,
	})

	front = memory.l.Front()
	next := front.Value.(*Element)
	assert.Equal(t, time1, next.Expiry)
	assert.Equal(t, time1, next.At)

}

func TestMemory_Len(t *testing.T) {
	memory := NewMemory()
	now := time.Now()
	memory.Add(&Element{
		At:     now,
		Expiry: now,
	})
	assert.Equal(t, 1, memory.Len())

	memory.Add(&Element{
		At:     now,
		Expiry: now,
	})
	assert.Equal(t, 2, memory.Len())
}
func TestMemory_Front(t *testing.T) {
	memory := NewMemory()
	now := time.Now()
	memory.Add(&Element{
		At:     now,
		Expiry: now,
	})
	assert.Equal(t, memory.l.Front().Value.(*Element), memory.Front())
}

func TestMemory_Remove(t *testing.T) {
	memory := NewMemory()
	now := time.Now()
	e := &Element{
		At:     now,
		Expiry: now,
	}
	memory.Add(e)
	assert.Equal(t, e, memory.l.Front().Value.(*Element))

	assert.Equal(t, 1, memory.l.Len())
	memory.Remove(e)
	assert.Equal(t, 0, memory.l.Len())

}

func TestMemory_Replace(t *testing.T) {
	memory := NewMemory()
	now := time.Now()
	controller := gomock.NewController(t)
	message2 := NewMockMessage(controller)
	message2.EXPECT().Id().AnyTimes().Return(packet.PacketId(1))
	e1 := &Element{
		At:      now,
		Expiry:  now,
		Message: message2,
	}
	memory.Add(e1)

	time2 := time.UnixMilli(1)
	e2 := &Element{
		At:      time2,
		Expiry:  time2,
		Message: message2,
	}
	memory.Add(e2)
	replaced, err := memory.Replace(e2)
	assert.True(t, replaced)
	assert.NoError(t, err)
	assert.Equal(t, time2, memory.l.Front().Value.(*Element).At)

	message3 := NewMockMessage(controller)
	message3.EXPECT().Id().AnyTimes().Return(packet.PacketId(2))
	time3 := time.UnixMilli(1)
	e3 := &Element{
		At:      time3,
		Expiry:  time3,
		Message: message3,
	}
	replaced, err = memory.Replace(e3)
	assert.False(t, replaced)
	assert.NoError(t, err)

}

func TestMemory_Iterator(t *testing.T) {
	memory := NewMemory()
	n := int64(5)
	elems := make([]*Element, 5)
	for i := int64(0); i < n; i++ {
		e1 := &Element{
			At:     time.UnixMilli(i),
			Expiry: time.UnixMilli(i),
		}
		elems[i] = e1
		memory.Add(e1)
	}
	i := n - 1
	for itr := memory.Iterator(); itr.HasNext(); {
		elem, _ := itr.Next()
		_ = itr.Remove()
		assert.Equal(t, elems[i].At, elem.At)
		i--
	}
	assert.EqualValues(t, -1, i)
	assert.EqualValues(t, 0, memory.l.Len())

}
func TestMemory_Reset(t *testing.T) {
	memory := NewMemory()
	n := int64(5)
	elems := make([]*Element, 5)
	for i := int64(0); i < n; i++ {
		e1 := &Element{
			At:     time.UnixMilli(i),
			Expiry: time.UnixMilli(i),
		}
		elems[i] = e1
		memory.Add(e1)
	}
	assert.EqualValues(t, n, memory.l.Len())

	memory.Reset()
	assert.EqualValues(t, 0, memory.l.Len())

}
