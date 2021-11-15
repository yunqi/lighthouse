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

package store

import (
	"errors"
	"github.com/yunqi/lighthouse/internal/packet"
	"github.com/yunqi/lighthouse/internal/persistence/queue"
	"sync"
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

// IsExpired return whether the elem is expired
func IsExpired(now time.Time, elem *queue.Element) bool {
	if !elem.Expiry.IsZero() {
		return now.After(elem.Expiry)
	}
	return false
}

var _ queue.Queue = (*Queue)(nil)

type (
	// InitOptions is used to pass some required client information to the queue.Init()
	InitOptions struct {
		// CleanStart is the cleanStart field in the connect packet.
		CleanStart bool
		// Version is the client MQTT protocol version.
		Version packet.Version
		// ReadBytesLimit indicates the maximum publish size that is allow to read.
		ReadBytesLimit uint32
		Notifier       queue.Notifier
	}
	Options struct {
		MaxQueuedMsg   int
		InflightExpiry time.Duration
		ClientId       string
	}

	Queue struct {
		clientId       string
		version        packet.Version
		maxSize        int
		readBytesLimit uint32

		inflightDrained bool
		inflightExpiry  time.Duration
		closed          bool

		notifier queue.Notifier
		cond     *sync.Cond
		store    queue.Store
	}
)

func New(store queue.Store, options *Options) *Queue {
	return &Queue{
		clientId:       options.ClientId,
		inflightExpiry: options.InflightExpiry,
		maxSize:        options.MaxQueuedMsg,
		cond:           sync.NewCond(&sync.Mutex{}),
		store:          store,
	}
}
func (q *Queue) Init(opts *InitOptions) error {
	q.cond.L.Lock()
	defer func() {
		q.cond.L.Unlock()
		q.cond.Signal()
	}()
	q.closed = false
	q.inflightDrained = false
	if opts.CleanStart {
		q.store.Reset()
	}
	q.readBytesLimit = opts.ReadBytesLimit
	q.version = opts.Version
	q.notifier = opts.Notifier

	return nil
}

func (q *Queue) ReadInflight(maxSize uint) (elems []*queue.Element, err error) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	length := q.store.Len()
	if q.store.Len() == 0 {
		q.inflightDrained = true
		return nil, nil
	}
	if int(maxSize) < length {
		length = int(maxSize)
	}
	i := 0
	for itr := q.store.Iterator(); itr.HasNext() && i < length; i++ {
		element, _ := itr.Next()
		if element.Id() != 0 {
			if q.inflightExpiry != 0 {
				element.Expiry = time.Now().Add(q.inflightExpiry)
			}
			elems = append(elems, element)
		} else {
			q.inflightDrained = true
			break
		}
	}
	return
}

func (q *Queue) Replace(elem *queue.Element) (replaced bool, err error) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	return q.store.Replace(elem)
}

func (q *Queue) Read(packetIds []packet.PacketId) (elements []*queue.Element, err error) {
	nowTime := time.Now()
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	if !q.inflightDrained {
		panic("must call ReadInflight to drain all inflight messages before Read")
	}
	for q.store.Len() == 0 && !q.closed {
		// wait notify
		q.cond.Wait()
	}
	if q.closed {
		return nil, ErrClosed
	}
	length := q.store.Len()
	packetIdsLen := len(packetIds)
	if packetIdsLen < length {
		length = packetIdsLen
	}
	var msgQueueDelta, inflightDelta int
	var pflag int

	i := 0
	for itr := q.store.Iterator(); itr.HasNext() && i < length; i++ {
		current, _ := itr.Next()
		if IsExpired(nowTime, current) {
			q.notifier.Dropped(current, ErrDropExpired)
			_ = itr.Remove()
			msgQueueDelta--
			continue
		}

		pubMsg := current.Message.(*queue.Publish)

		if size := pubMsg.TotalBytes(q.version); size > q.readBytesLimit {
			q.notifier.Dropped(current, ErrDropExceedsMaxPacketSize)
			_ = itr.Remove()
			msgQueueDelta--
			continue
		}
		// remove qos 0 message after read
		if pubMsg.QoS == 0 {

			_ = itr.Remove()
			msgQueueDelta--
		} else {
			pubMsg.SetId(packetIds[pflag])
			// When the message becomes inflight message, update the expiry time.
			if q.inflightExpiry != 0 {
				current.Expiry = nowTime.Add(q.inflightExpiry)
			}
			pflag++
			inflightDelta++

		}
		elements = append(elements, current)
	}
	q.notifier.MsgQueueAdded(msgQueueDelta)
	q.notifier.InflightAdded(inflightDelta)
	return
}

func (q *Queue) Remove(pid packet.PacketId) error {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	for itr := q.store.Iterator(); itr.HasNext(); {
		element, _ := itr.Next()
		if element.Id() == pid {
			_ = itr.Remove()
			q.notifier.MsgQueueAdded(-1)
			q.notifier.InflightAdded(-1)
			return nil
		}
	}
	return nil
}

func (q *Queue) Clean() error {
	return nil
}

func (q *Queue) Close() error {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	q.closed = true
	q.cond.Signal()
	return nil
}

func (q *Queue) Add(elem *queue.Element) (err error) {
	nowTime := time.Now()
	q.cond.L.Lock()
	defer func() {
		q.cond.L.Unlock()
		// wake
		q.cond.Signal()
	}()
	var drop bool
	var dropElem *queue.Element
	var dropErr error
	defer func() {
		if drop {
			if dropElem == nil {
				// Discard current data
				q.notifier.Dropped(elem, dropErr)
				return
			}
			if err == ErrDropExpiredInflight {
				q.notifier.InflightAdded(-1)
			}

			q.store.Remove(dropElem)
			q.notifier.Dropped(dropElem, dropErr)
		} else {
			q.notifier.MsgQueueAdded(1)
		}
		q.store.Add(elem)
	}()

	if q.store.Len() >= q.maxSize {
		err = ErrDropQueueFull
		drop = true

		if v := q.store.Front(); v != nil &&
			IsExpired(nowTime, v) {
			dropElem = v
			dropErr = ErrDropExpiredInflight
			return
		}

		// drop the current elem if there is no more non-inflight messages.
		if q.inflightDrained && q.store.Front() == nil {
			return
		}

		for itr := q.store.Iterator(); itr.HasNext(); {
			e, _ := itr.Next()
			pubMsg := e.Message.(*queue.Publish)
			if pubMsg.Id() == 0 {

				if IsExpired(nowTime, e) {
					dropElem = e
					dropErr = ErrDropExpired
					return
				}
				// drop qos0 message in the queue
				if pubMsg.QoS == packet.QoS0 && dropElem == nil {
					dropElem = e
				}
			}
		}
		if dropElem != nil {
			return
		}
		if elem.Message.(*queue.Publish).QoS == packet.QoS0 {
			return
		}
		if q.inflightDrained {
			// drop the front message
			dropElem = q.store.Front()
			return
		}

	}
	return nil
}
