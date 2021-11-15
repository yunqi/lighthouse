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
	"errors"
	"github.com/yunqi/lighthouse/internal/packet"
	"io"
	"time"
)

//go:generate mockgen -destination ./queue_mock.go -package queue -source queue.go
var (
	ErrClosed                   = errors.New("queue has been closed")
	ErrDropExceedsMaxPacketSize = errors.New("maximum packet size exceeded")
	ErrDropQueueFull            = errors.New("the message queue is full")
	ErrDropExpired              = errors.New("the message is expired")
	ErrDropExpiredInflight      = errors.New("the inflight message is expired")
)

type (
	Message interface {
		Id() packet.PacketId
		SetId(id packet.PacketId)
	}

	Element struct {
		// At represents the entry time.
		At time.Time
		// Expiry represents the expiry time.
		// Empty means never expire.
		Expiry time.Time
		Message
	}
	Queue interface {
		io.Closer
		Add(elem *Element) error
		// Replace replaces the PUBLISH with the PUBREL with the same packet id.
		Replace(elem *Element) (replaced bool, err error)
		Read(packetIds []packet.PacketId) ([]*Element, error)
		// Remove removes the elem for a given id.
		Remove(pid packet.PacketId) error
		Clean() error
		ReadInflight(maxSize uint) (elems []*Element, err error)
	}

	Notifier interface {
		// Dropped will be called when the element in the queue is dropped.
		// The err indicates the reason of why it is dropped.
		// The MessageWithID field in elem param can be queue.Pubrel or queue.Publish.
		Dropped(elem *Element, err error)
		InflightAdded(delta int)
		MsgQueueAdded(delta int)
	}
)
