package queue

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/chenquan/go-pkg/xbinary"
	"github.com/yunqi/lighthouse/internal/packet"
	"github.com/yunqi/lighthouse/internal/persistence/message"
	"github.com/yunqi/lighthouse/internal/persistence/message/encoding"
	"time"
)

type Message interface {
	Id() packet.Id
	SetId(id packet.Id)
}

type Publish struct {
	*message.Message
}

func (p *Publish) Id() packet.Id {
	return p.PacketId
}
func (p *Publish) SetId(id packet.Id) {
	p.PacketId = id
}

type Pubrel struct {
	PacketID packet.Id
}

func (p *Pubrel) Id() packet.Id {
	return p.PacketID
}
func (p *Pubrel) SetId(id packet.Id) {
	p.PacketID = id
}

// Element represents the element store in the queue.
type Element struct {
	// At represents the entry time.
	At time.Time
	// Expiry represents the expiry time.
	// Empty means never expire.
	Expiry time.Time
	Message
}

// Encode encodes the publish structure into bytes and write it to the buffer
func (p *Publish) Encode(b *bytes.Buffer) {
	encoding.EncodeMessage(p.Message, b)
}

func (p *Publish) Decode(b *bytes.Buffer) (err error) {
	msg, err := encoding.DecodeMessage(bytes.NewReader(b.Bytes()))
	if err != nil {
		return err
	}
	p.Message = msg
	return nil
}

// Encode encode the pubrel structure into bytes.
func (p *Pubrel) Encode(b *bytes.Buffer) {
	_ = xbinary.WriteUint16(b, p.PacketID)
}

func (p *Pubrel) Decode(b *bytes.Buffer) (err error) {
	p.PacketID, err = xbinary.ReadUint16(b)
	return
}

// Encode encode the elem structure into bytes.
// Format: 8 byte timestamp | 1 byte identifier| data
func (e *Element) Encode() []byte {
	b := bytes.NewBuffer(make([]byte, 0, 100))
	rs := make([]byte, 19)
	binary.BigEndian.PutUint64(rs[0:9], uint64(e.At.Unix()))
	binary.BigEndian.PutUint64(rs[9:18], uint64(e.Expiry.Unix()))
	switch m := e.Message.(type) {
	case *Publish:
		rs[18] = 0
		b.Write(rs)
		m.Encode(b)
	case *Pubrel:
		rs[18] = 1
		b.Write(rs)
		m.Encode(b)
	}
	return b.Bytes()
}

func (e *Element) Decode(b []byte) (err error) {
	if len(b) < 19 {
		return errors.New("invalid input length")
	}
	e.At = time.Unix(int64(binary.BigEndian.Uint64(b[0:9])), 0)
	e.Expiry = time.Unix(int64(binary.BigEndian.Uint64(b[9:19])), 0)
	switch b[18] {
	case 0: // publish
		p := &Publish{}
		buf := bytes.NewBuffer(b[19:])
		err = p.Decode(buf)
		e.Message = p
	case 1: // pubrel
		p := &Pubrel{}
		buf := bytes.NewBuffer(b[19:])
		err = p.Decode(buf)
		e.Message = p
	default:
		return errors.New("invalid identifier")
	}
	return
}
