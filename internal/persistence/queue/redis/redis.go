package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/yunqi/lighthouse/internal/code"
	"github.com/yunqi/lighthouse/internal/packet"
	"github.com/yunqi/lighthouse/internal/persistence/queue"
	red "github.com/yunqi/lighthouse/internal/redis"
	"github.com/yunqi/lighthouse/internal/xerror"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	queuePrefix = "lighthouse:queue:"
)

var _ queue.Queue = (*Queue)(nil)

func getKey(clientID string) string {
	return queuePrefix + clientID
}

type Options struct {
	MaxQueuedMsg    int
	ClientID        string
	InflightExpiry  time.Duration
	DefaultNotifier queue.Notifier
	Redis           *red.Redis
}

type Queue struct {
	cond           *sync.Cond
	clientID       string
	key            string
	version        packet.Version
	readBytesLimit uint32
	// max is the maximum queue length
	max int
	// len is the length of the list
	len             int
	r               *red.Redis
	closed          bool
	inflightDrained bool
	// current is the current read index of Queue list.
	current        int
	readCache      map[packet.Id][]byte
	err            error
	log            *zap.Logger
	inflightExpiry time.Duration
	notifier       queue.Notifier
	redisClient    *redis.Client
}

func New(opts Options) (*Queue, error) {
	return &Queue{
		cond:            sync.NewCond(&sync.Mutex{}),
		clientID:        opts.ClientID,
		key:             getKey(opts.ClientID),
		max:             opts.MaxQueuedMsg,
		len:             0,
		r:               opts.Redis,
		closed:          false,
		inflightDrained: false,
		current:         0,
		inflightExpiry:  opts.InflightExpiry,
		notifier:        opts.DefaultNotifier,
		//log:             server.LoggerWithField(zap.String("queue", "redis")),
	}, nil
}

func wrapError(err error) *xerror.Error {
	return &xerror.Error{
		Code: code.UnspecifiedError,
		ErrorDetails: xerror.ErrorDetails{
			ReasonString:   []byte(err.Error()),
			UserProperties: nil,
		},
	}
}

func (q *Queue) Close() error {
	q.cond.L.Lock()
	defer func() {
		q.cond.L.Unlock()
		q.cond.Signal()
	}()
	q.closed = true
	return nil
}

func (q *Queue) setLen(ctx context.Context) error {
	llen, err := q.r.Llen(ctx, q.key)
	if err != nil {
		return err
	}
	q.len = llen
	return nil
}

func (q *Queue) Init(ctx context.Context, opts *queue.InitOptions) error {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	if opts.CleanStart {
		_, err := q.r.Del(ctx, q.key)
		if err != nil {
			return wrapError(err)
		}
	}
	err := q.setLen(ctx)
	if err != nil {
		return err
	}
	q.version = opts.Version
	q.readBytesLimit = opts.ReadBytesLimit
	q.closed = false
	q.inflightDrained = false
	q.current = 0
	q.readCache = make(map[packet.Id][]byte)
	q.notifier = opts.Notifier
	q.cond.Signal()
	return nil
}

func (q *Queue) Clean(ctx context.Context) error {
	_, err := q.r.Del(ctx, q.key)
	return err
}

func (q *Queue) Add(ctx context.Context, elem *queue.Element) (err error) {
	now := time.Now()
	q.cond.L.Lock()
	var dropErr error
	var dropBytes []byte
	var dropElem *queue.Element
	var drop bool
	defer func() {
		q.cond.L.Unlock()
		q.cond.Signal()
	}()

	defer func() {
		if drop {
			if dropErr == queue.ErrDropExpiredInflight {
				q.notifier.NotifyInflightAdded(-1)
				q.current--
			}
			if dropBytes == nil {
				q.notifier.NotifyDropped(elem, dropErr)
				return
			} else {
				_, err = q.r.Lrem(ctx, q.key, 1, dropBytes)
			}
			q.notifier.NotifyDropped(dropElem, dropErr)
		} else {
			q.notifier.NotifyMsgQueueAdded(1)
			q.len++
		}
		_, err = q.r.Rpush(ctx, q.key, elem.Encode())
	}()
	if q.len >= q.max {
		// set default drop error
		dropErr = queue.ErrDropQueueFull
		drop = true
		var rs []string
		// drop expired inflight message
		rs, err = q.r.Lrange(ctx, q.key, 0, q.len)

		if err != nil {
			return
		}
		var frontBytes []byte
		var frontElem *queue.Element
		for i := 0; i < len(rs); i++ {
			b := []byte(rs[i])
			e := &queue.Element{}
			err = e.Decode(b)
			if err != nil {
				return
			}
			// inflight message
			if i < q.current && queue.ElemExpiry(now, e) {
				dropBytes = b
				dropElem = e
				dropErr = queue.ErrDropExpiredInflight
				return
			}
			// non-inflight message
			if i >= q.current {
				if i == q.current {
					frontBytes = b
					frontElem = e
				}
				// drop qos0 message in the queue
				pub := e.Message.(*queue.Publish)
				// drop expired non-inflight message
				if pub.Id() == 0 && queue.ElemExpiry(now, e) {
					dropBytes = b
					dropElem = e
					dropErr = queue.ErrDropExpired
					return
				}
				if pub.Id() == 0 && pub.QoS == packet.QoS0 && dropElem == nil {
					dropBytes = b
					dropElem = e
				}
			}
		}
		// drop the current elem if there is no more non-inflight messages.
		if q.inflightDrained && q.current >= q.len {
			return
		}
		rs, err = q.r.Lrange(ctx, q.key, q.current, q.len)

		if err != nil {
			return err
		}
		if dropElem != nil {
			return
		}
		if elem.Message.(*queue.Publish).QoS == packet.QoS0 {
			return
		}
		if frontElem != nil {
			// drop the front message
			dropBytes = frontBytes
			dropElem = frontElem
		}
		// the the messages in the queue are all inflight messages, drop the current elem
		return
	}
	return nil
}

func (q *Queue) Replace(ctx context.Context, elem *queue.Element) (replaced bool, err error) {
	//conn := q.pool.Get()
	q.cond.L.Lock()
	defer func() {
		//conn.Close()
		q.cond.L.Unlock()
	}()
	id := elem.Id()
	eb := elem.Encode()
	stop := q.current - 1
	if stop < 0 {
		stop = 0
	}
	rs, err := q.r.Lrange(ctx, q.key, 0, stop)
	if err != nil {
		return false, err
	}
	for k, v := range rs {
		b := []byte(v)
		e := &queue.Element{}
		err = e.Decode(b)
		if err != nil {
			return false, err
		}
		if e.Id() == elem.Id() {
			_, err = q.r.Lset(ctx, q.key, int64(k), eb)
			if err != nil {
				return false, err
			}
			q.readCache[id] = eb
			return true, nil
		}
	}

	return false, nil
}

func (q *Queue) Read(ctx context.Context, pids []packet.Id) (elems []*queue.Element, err error) {
	now := time.Now()
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	if !q.inflightDrained {
		panic("must call ReadInflight to drain all inflight messages before Read")
	}
	for q.current >= q.len && !q.closed {
		q.cond.Wait()
	}
	if q.closed {
		return nil, queue.ErrClosed
	}
	rs, err := q.r.Lrange(ctx, q.key, q.current, q.current+len(pids)-1)
	if err != nil {
		return nil, wrapError(err)
	}
	var msgQueueDelta, inflightDelta int
	var pflag int
	for i := 0; i < len(rs); i++ {
		b := []byte(rs[i])
		e := &queue.Element{}
		err := e.Decode(b)
		if err != nil {
			return nil, err
		}
		// remove expired message
		if queue.ElemExpiry(now, e) {
			_, err = q.r.Lrem(ctx, q.key, 1, b)
			q.len--
			if err != nil {
				return nil, err
			}
			q.notifier.NotifyDropped(e, queue.ErrDropExpired)
			msgQueueDelta--
			continue
		}

		// remove message which exceeds maximum packet size
		pub := e.Message.(*queue.Publish)
		if size := pub.TotalBytes(q.version); size > q.readBytesLimit {
			_, err = q.r.Lrem(ctx, q.key, 1, b)
			q.len--
			if err != nil {
				return nil, err
			}
			q.notifier.NotifyDropped(e, queue.ErrDropExceedsMaxPacketSize)
			msgQueueDelta--
			continue
		}

		if e.Message.(*queue.Publish).QoS == 0 {
			_, err = q.r.Lrem(ctx, q.key, 1, b)

			q.len--
			msgQueueDelta--
			if err != nil {
				return nil, err
			}
		} else {
			e.Message.SetId(pids[pflag])
			if q.inflightExpiry != 0 {
				e.Expiry = now.Add(q.inflightExpiry)
			}
			pflag++
			nb := e.Encode()
			_, err = q.r.Lset(ctx, q.key, int64(q.current), nb)
			q.current++
			inflightDelta++
			q.readCache[e.Message.Id()] = nb
		}
		elems = append(elems, e)
	}
	q.notifier.NotifyMsgQueueAdded(msgQueueDelta)
	q.notifier.NotifyInflightAdded(inflightDelta)
	return
}

func (q *Queue) ReadInflight(ctx context.Context, maxSize uint) (elems []*queue.Element, err error) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	rs, err := q.r.Lrange(ctx, q.key, q.current, q.current+int(maxSize)-1)
	if len(rs) == 0 {
		q.inflightDrained = true
		return
	}
	if err != nil {
		return nil, wrapError(err)
	}
	beginIndex := q.current
	for index, v := range rs {
		b := []byte(v)
		e := &queue.Element{}
		err := e.Decode(b)
		if err != nil {
			return nil, err
		}
		id := e.Message.Id()
		if id != 0 {
			if q.inflightExpiry != 0 {
				e.Expiry = time.Now().Add(q.inflightExpiry)
				b = e.Encode()
				_, err = q.r.Lset(ctx, q.key, int64(beginIndex+index), b)
				if err != nil {
					return nil, err
				}
			}
			elems = append(elems, e)
			q.readCache[id] = b
			q.current++
		} else {
			q.inflightDrained = true
			return elems, nil
		}
	}
	return
}

func (q *Queue) Remove(ctx context.Context, pid packet.Id) error {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	if b, ok := q.readCache[pid]; ok {
		_, err := q.r.Lrem(ctx, q.key, 1, b)
		if err != nil {
			return err
		}
		q.notifier.NotifyMsgQueueAdded(-1)
		q.notifier.NotifyInflightAdded(-1)
		delete(q.readCache, pid)
		q.len--
		q.current--
	}
	return nil
}
