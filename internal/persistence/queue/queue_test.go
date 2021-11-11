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
	"github.com/yunqi/lighthouse/internal/packet"
	"github.com/yunqi/lighthouse/internal/persistence/message"
	"testing"
	"time"
)

type testNotifier struct {
	dropElem    []*Element
	dropErr     error
	inflightLen int
	msgQueueLen int
}

func (t *testNotifier) Dropped(elem *Element, err error) {
	t.dropElem = append(t.dropElem, elem)
	t.dropErr = err
}

func (t *testNotifier) InflightAdded(delta int) {
	t.inflightLen += delta
	if t.inflightLen < 0 {
		t.inflightLen = 0
	}
}
func (t *testNotifier) MsgQueueAdded(delta int) {
	t.msgQueueLen += delta
	if t.msgQueueLen < 0 {
		t.msgQueueLen = 0
	}
}

func TestQueue_Add(t *testing.T) {
	// 2 inflight message + 3 new message
	var initElems = []*Element{
		{
			At:     time.Now(),
			Expiry: time.Time{},
			Message: &Publish{
				Message: &message.Message{
					QoS:      packet.QoS1,
					Retained: false,
					Topic:    "/topic1_qos1",
					Payload:  []byte("qos1"),
					PacketId: 1,
				},
			},
		}, {
			At:     time.Now(),
			Expiry: time.Time{},
			Message: &Publish{
				Message: &message.Message{
					QoS:      packet.QoS2,
					Retained: false,
					Topic:    "/topic1_qos2",
					Payload:  []byte("qos2"),
					PacketId: 2,
				},
			},
		}, {
			At:     time.Now(),
			Expiry: time.Time{},
			Message: &Publish{
				Message: &message.Message{
					QoS:      packet.QoS1,
					Retained: false,
					Topic:    "/topic1_qos1",
					Payload:  []byte("qos1"),
					PacketId: 0,
				},
			},
		},
		{
			At:     time.Now(),
			Expiry: time.Time{},
			Message: &Publish{
				Message: &message.Message{
					QoS:      packet.QoS0,
					Retained: false,
					Topic:    "/topic1_qos0",
					Payload:  []byte("qos0"),
					PacketId: 0,
				},
			},
		},
		{
			At:     time.Now(),
			Expiry: time.Time{},
			Message: &Publish{
				Message: &message.Message{
					QoS:      packet.QoS2,
					Retained: false,
					Topic:    "/topic1_qos2",
					Payload:  []byte("qos2"),
					PacketId: 0,
				},
			},
		},
	}

	opt := &Options{
		MaxQueuedMsg:   2,
		InflightExpiry: 0,
		ClientId:       "",
	}
	q := New(NewMemory(), opt)
	_ = q.Init(&InitOptions{
		CleanStart:     true,
		Version:        packet.Version5,
		ReadBytesLimit: 100,
		Notifier:       &testNotifier{},
	})
	for _, e := range initElems {
		q.Add(e)
	}
}
