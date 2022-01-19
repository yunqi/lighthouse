package session

import (
	"github.com/yunqi/lighthouse/internal/session"
)

// IterateFn is the callback function used by Iterate()
// Return false means to stop the iteration.
type IterateFn func(session *session.Session) bool

type Store interface {
	Set(session *session.Session) error
	Remove(clientID string) error
	Get(clientID string) (*session.Session, error)
	Iterate(fn IterateFn) error
	SetSessionExpiry(clientID string, expiry uint32) error
}
