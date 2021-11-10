package memory

import (
	"github.com/yunqi/lighthouse/internal/packet"
	"github.com/yunqi/lighthouse/internal/persistence/queue"
	"github.com/yunqi/lighthouse/internal/persistence/store"
	"sync"
	"time"
)

var _ queue.Queue = (*Queue)(nil)

type (
	Options struct {
		MaxQueuedMsg    int
		InflightExpiry  time.Duration
		ClientId        string
		DefaultNotifier queue.Notifier
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
		store    store.Store
		len      int
	}
)

func New(store store.Store, options *Options) *Queue {
	return &Queue{
		clientId:       options.ClientId,
		inflightExpiry: options.InflightExpiry,
		notifier:       options.DefaultNotifier,
		maxSize:        options.MaxQueuedMsg,
		cond:           sync.NewCond(&sync.Mutex{}),
		store:          store,
	}
}
func (q *Queue) Init(opts *queue.InitOptions) error {
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
	for iterator := q.store.Iterator(); iterator.HasNext() && i < length; i++ {
		element, _ := iterator.Next()
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
		return nil, queue.ErrClosed
	}
	length := q.store.Len()
	packetIdsLen := len(packetIds)
	if packetIdsLen < length {
		length = packetIdsLen
	}
	var msgQueueDelta, inflightDelta int
	var pflag int

	i := 0
	for iterator := q.store.Iterator(); iterator.HasNext() && i < length; i++ {
		current, _ := iterator.Next()
		if queue.IsExpired(nowTime, current) {
			q.notifier.Dropped(current, queue.ErrDropExpired)
			_ = iterator.Remove()
			msgQueueDelta--
			continue
		}

		pubMsg := current.Message.(*queue.Publish)

		if size := pubMsg.TotalBytes(q.version); size > q.readBytesLimit {
			q.notifier.Dropped(current, queue.ErrDropExceedsMaxPacketSize)
			_ = iterator.Remove()
			msgQueueDelta--
			continue
		}
		// remove qos 0 message after read
		if pubMsg.QoS == 0 {

			_ = iterator.Remove()
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

	for iterator := q.store.Iterator(); iterator.HasNext(); {
		element, _ := iterator.Next()
		if element.Id() == pid {
			_ = iterator.Remove()
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
			if err == queue.ErrDropExpiredInflight {
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
		err = queue.ErrDropQueueFull
		drop = true

		if v := q.store.Front(); v != nil &&
			queue.IsExpired(nowTime, v) {
			dropElem = v
			dropErr = queue.ErrDropExpiredInflight
			return
		}

		// drop the current elem if there is no more non-inflight messages.
		if q.inflightDrained && q.store.Front() == nil {
			return
		}

		for iterator := q.store.Iterator(); iterator.HasNext(); {
			e, _ := iterator.Next()
			pubMsg := e.Message.(*queue.Publish)
			if pubMsg.Id() == 0 {

				if queue.IsExpired(nowTime, e) {
					dropElem = e
					dropErr = queue.ErrDropExpired
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
