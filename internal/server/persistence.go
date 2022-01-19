package server

import (
	"github.com/yunqi/lighthouse/config"
	"github.com/yunqi/lighthouse/internal/persistence/queue"
	"github.com/yunqi/lighthouse/internal/persistence/session"
	"github.com/yunqi/lighthouse/internal/persistence/subscription"
	"github.com/yunqi/lighthouse/internal/persistence/unack"
)

type (
	NewPersistence func(config config.Config) (Persistence, error)
	Persistence    interface {
		Open() error
		NewQueueStore(config config.Config, defaultNotifier queue.Notifier, clientID string) (queue.Store, error)
		NewSubscriptionStore(config config.Config) (subscription.Store, error)
		NewSessionStore(config config.Config) (session.Store, error)
		NewUnackStore(config config.Config, clientID string) (unack.Store, error)
		Close() error
	}
)
