package breaker

import (
	"fmt"
	"github.com/chenquan/go-pkg/xmath"
	"strings"
	"sync"
	"time"
)

const (
	numHistoryReasons = 5
	timeFormat        = "15:04:05"
)

type (
	// Promise interface defines the callbacks that returned by Breaker.Allow.
	Promise interface {
		// Accept tells the Breaker that the call is successful.
		Accept()
		// Reject tells the Breaker that the call is failed.
		Reject(reason error)
	}
	Acceptable func(err error) bool
	Fallback   func(err error) error

	// Option defines the method to customize a Breaker.
	Option func(breaker *circuitBreaker)

	Breaker interface {

		// Name returns the name of the Breaker.
		Name() string

		// Allow checks if the request is allowed.
		// If allowed, a promise will be returned, the caller needs to call promise.Accept()
		// on success, or call promise.Reject() on failure.
		// If not allow, ErrServiceUnavailable will be returned.
		Allow() (Promise, error)
		Do(req func() error) error
		DoWithAcceptable(req func() error, acceptable Acceptable) error

		DoWithFallback(req func() error, fallback Fallback) error

		DoWithFallbackAcceptable(req func() error, fallback Fallback, acceptable Acceptable) error
	}
	circuitBreaker struct {
		name string
		throttle
	}

	throttle interface {
		allow() (Promise, error)
		doReq(req func() error, fallback Fallback, acceptable Acceptable) error
	}
	internalThrottle interface {
		allow() (internalPromise, error)
		doReq(req func() error, fallback Fallback, acceptable Acceptable) error
	}
	loggedThrottle struct {
		name string
		internalThrottle
		errWin *errorWindow
	}
)

// NewBreaker returns a Breaker object.
// opts can be used to customize the Breaker.
func NewBreaker(name string, opts ...Option) Breaker {
	var b circuitBreaker
	b.name = name
	for _, opt := range opts {
		opt(&b)
	}

	b.throttle = newLoggedThrottle(name, newGoogleBreaker())

	return &b
}

func (cb *circuitBreaker) Name() string {
	return cb.name
}

func (cb *circuitBreaker) Allow() (Promise, error) {
	return cb.throttle.allow()
}

func (cb *circuitBreaker) Do(req func() error) error {
	return cb.throttle.doReq(req, nil, defaultAcceptable)
}

func (cb *circuitBreaker) DoWithAcceptable(req func() error, acceptable Acceptable) error {
	return cb.throttle.doReq(req, nil, acceptable)
}

func (cb *circuitBreaker) DoWithFallback(req func() error, fallback Fallback) error {
	return cb.throttle.doReq(req, fallback, defaultAcceptable)
}

func (cb *circuitBreaker) DoWithFallbackAcceptable(req func() error, fallback Fallback,
	acceptable Acceptable) error {
	return cb.throttle.doReq(req, fallback, acceptable)
}

type errorWindow struct {
	reasons [numHistoryReasons]string
	index   int
	count   int
	lock    sync.Mutex
}

func (ew *errorWindow) add(reason error) {
	ew.lock.Lock()
	ew.reasons[ew.index] = fmt.Sprintf("%s %s", time.Now().Format(timeFormat), reason)
	ew.index = (ew.index + 1) % numHistoryReasons
	ew.count = xmath.MinInt(ew.count+1, numHistoryReasons)
	ew.lock.Unlock()
}

func (ew *errorWindow) String() string {
	var reasons []string

	ew.lock.Lock()
	// reverse order
	for i := ew.index - 1; i >= ew.index-ew.count; i-- {
		reasons = append(reasons, ew.reasons[(i+numHistoryReasons)%numHistoryReasons])
	}
	ew.lock.Unlock()

	return strings.Join(reasons, "\n")
}

func defaultAcceptable(err error) bool {
	return err == nil
}

func newLoggedThrottle(name string, t internalThrottle) loggedThrottle {
	return loggedThrottle{
		name:             name,
		internalThrottle: t,
		errWin:           new(errorWindow),
	}
}

func (lt loggedThrottle) allow() (Promise, error) {
	promise, err := lt.internalThrottle.allow()
	return &promiseWithReason{
		promise: promise,
		errWin:  lt.errWin,
	}, lt.logError(err)
}

func (lt loggedThrottle) doReq(req func() error, fallback Fallback, acceptable Acceptable) error {
	return lt.logError(lt.internalThrottle.doReq(req, fallback, func(err error) bool {
		accept := acceptable(err)
		if !accept {
			lt.errWin.add(err)
		}
		return accept
	}))
}

func (lt loggedThrottle) logError(err error) error {
	if err == ErrServiceUnavailable {
		// if circuit open, not possible to have empty error window
		//stat.Report(fmt.Sprintf(
		//	"proc(%s/%d), callee: %s, breaker is open and requests dropped\nlast errors:\n%s",
		//	proc.ProcessName(), proc.Pid(), lt.name, lt.errWin))
	}

	return err
}

type promiseWithReason struct {
	promise internalPromise
	errWin  *errorWindow
}

func (p *promiseWithReason) Accept() {
	p.promise.Accept()
}

func (p *promiseWithReason) Reject(reason error) {
	p.errWin.add(reason)
	p.promise.Reject()
}
