package unack

import (
	"context"
	"github.com/yunqi/lighthouse/internal/packet"
)

// Store represents a unack store for one client.
// Unack store is used to persist the unacknowledged qos2 messages.
type Store interface {
	// Init will be called when the client connect.
	// If cleanStart set to true, the implementation should remove any associated data in backend store.
	// If it set to false, the implementation should retrieve the associated data from backend store.
	Init(ctx context.Context, cleanStart bool) error
	// Set sets the given id into store.
	// The return boolean indicates whether the id exist.
	Set(ctx context.Context, id packet.Id) (bool, error)
	// Remove removes the given id from store.
	Remove(ctx context.Context, id packet.Id) error
}
