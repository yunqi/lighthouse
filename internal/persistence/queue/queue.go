package queue

import (
	"github.com/yunqi/lighthouse/internal/packet"
	"time"
)

type (
	Message interface {
		Id() packet.PacketId
		SetId(id packet.PacketId)
	}

	Elem struct {
		// At represents the entry time.
		At time.Time
		// Expiry represents the expiry time.
		// Empty means never expire.
		Expiry time.Time
		Message
	}
	Queue interface {
		Add(elem Elem)
	}
)
