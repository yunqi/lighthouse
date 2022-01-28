package session

import (
	"context"
	"github.com/yunqi/lighthouse/config"
	"github.com/yunqi/lighthouse/internal/session"
)

// IterateFn is the callback function used by Iterate()
// Return false means to stop the iteration.
type IterateFn func(session *session.Session) bool
type NewStore func(config *config.StoreType) (Store, error)
type Store interface {
	Set(ctx context.Context, session *session.Session) error
	Remove(ctx context.Context, clientID string) error
	Get(ctx context.Context, clientID string) (*session.Session, error)
	Iterate(ctx context.Context, fn IterateFn) error
	SetSessionExpiry(ctx context.Context, clientID string, expiry uint32) error
}
