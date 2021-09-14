package queue

import (
	"github.com/yunqi/lighthouse/internal/packet"
	"github.com/yunqi/lighthouse/internal/persistence/message"
)

type (
	Publish struct {
		*message.Message
	}
	Pubrel struct {
		packet.PacketId
	}
)

func (p *Publish) Id() packet.PacketId {
	return p.PacketId
}
func (p *Publish) SetId(id packet.PacketId) {
	p.PacketId = id
}

//--------------------

func (p *Pubrel) Id() packet.PacketId {
	return p.PacketId
}

func (p *Pubrel) SetId(id packet.PacketId) {
	p.PacketId = id
}
