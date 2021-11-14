/*
 *    Copyright 2021 chenquan
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package encoding

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/yunqi/lighthouse/internal/packet"
	"github.com/yunqi/lighthouse/internal/persistence/message"
	"github.com/yunqi/lighthouse/internal/xio"
	"io"
)

var (
	read2BytesPool = xio.GetNBytePool(2)
	read4BytesPool = xio.GetNBytePool(4)
)

func DecodeMessageFromBytes(b []byte) (*message.Message, error) {
	if len(b) == 0 {
		return nil, nil
	}
	return DecodeMessage(bytes.NewReader(b))
}

// EncodeMessage Encoding the msg is written into w.
func EncodeMessage(msg *message.Message, w *bytes.Buffer) {
	if msg == nil {
		return
	}
	WriteBool(w, msg.Dup)
	_ = w.WriteByte(msg.QoS)
	WriteBool(w, msg.Retained)
	WriteBytes(w, []byte(msg.Topic))
	WriteBytes(w, msg.Payload)
	WriteUint16(w, msg.PacketId)
	if len(msg.ContentType) != 0 {
		_ = w.WriteByte(packet.PropContentType)
		WriteBytes(w, []byte(msg.ContentType))
	}
	if len(msg.CorrelationData) != 0 {
		_ = w.WriteByte(packet.PropCorrelationData)
		WriteBytes(w, msg.CorrelationData)
	}
	if msg.MessageExpiry != 0 {
		_ = w.WriteByte(packet.PropMessageExpiry)
		WriteUint32(w, msg.MessageExpiry)
	}
	_ = w.WriteByte(packet.PropPayloadFormat)
	_ = w.WriteByte(msg.PayloadFormat)

	if len(msg.ResponseTopic) != 0 {
		_ = w.WriteByte(packet.PropResponseTopic)
		WriteBytes(w, []byte(msg.ResponseTopic))
	}
	for _, v := range msg.SubscriptionIdentifier {
		_ = w.WriteByte(packet.PropSubscriptionIdentifier)
		l, _ := packet.EncodeRemainLength(int(v))
		_, _ = w.Write(l)
	}
}

func WriteBool(w *bytes.Buffer, b bool) {
	if b {
		_ = w.WriteByte(1)
	} else {
		_ = w.WriteByte(0)
	}
}
func DecodeMessage(r *bytes.Reader) (*message.Message, error) {
	msg := &message.Message{}

	var err error
	msg.Dup, err = ReadBool(r)
	if err != nil {
		return nil, err
	}
	msg.QoS, err = r.ReadByte()
	if err != nil {
		return nil, err
	}
	msg.Retained, err = ReadBool(r)
	if err != nil {
		return nil, err
	}
	topic, err := ReadBytes(r)
	if err != nil {
		return nil, err
	}
	msg.Topic = string(topic)
	msg.Payload, err = ReadBytes(r)
	if err != nil {
		return nil, err
	}
	msg.PacketId, err = ReadUint16(r)
	if err != nil {
		return nil, err
	}
	for {
		pt, err := r.ReadByte()
		if err == io.EOF {
			return msg, nil
		}
		if err != nil {
			return nil, err
		}
		switch pt {
		case packet.PropContentType:
			v, err := ReadBytes(r)
			if err != nil {
				return nil, err
			}
			msg.ContentType = string(v)
		case packet.PropCorrelationData:
			msg.CorrelationData, err = ReadBytes(r)
			if err != nil {
				return nil, err
			}
		case packet.PropMessageExpiry:
			msg.MessageExpiry, err = ReadUint32(r)
			if err != nil {
				return nil, err
			}
		case packet.PropPayloadFormat:
			msg.PayloadFormat, err = r.ReadByte()
			if err != nil {
				return nil, err
			}
		case packet.PropResponseTopic:
			v, err := ReadBytes(r)
			if err != nil {
				return nil, err
			}
			msg.ResponseTopic = string(v)
		case packet.PropSubscriptionIdentifier:
			si, err := packet.DecodeRemainLength(r)
			if err != nil {
				return nil, err
			}
			msg.SubscriptionIdentifier = append(msg.SubscriptionIdentifier, uint32(si))
		}
	}

}
func ReadBool(r *bytes.Reader) (bool, error) {
	b, err := r.ReadByte()
	if err != nil {
		return false, err
	}
	if b == 0 {
		return false, nil
	}
	return true, nil
}

func ReadBytes(r *bytes.Reader) (b []byte, err error) {
	l := read2BytesPool.Get().([]byte)
	defer read2BytesPool.Put(l)
	_, err = io.ReadFull(r, l)
	if err != nil {
		return nil, err
	}

	length := int(binary.BigEndian.Uint16(l))
	if length == 0 {
		return nil, nil
	}
	payload := make([]byte, length)

	_, err = io.ReadFull(r, payload)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func WriteBytes(w *bytes.Buffer, s []byte) {
	WriteUint16(w, uint16(len(s)))
	_, _ = w.Write(s)
}
func WriteUint16(w *bytes.Buffer, i uint16) {
	_ = w.WriteByte(byte(i >> 8))
	_ = w.WriteByte(byte(i))
}

func ReadUint16(r *bytes.Reader) (uint16, error) {
	r.Size()
	if r.Size() < 2 {
		return 0, errors.New("invalid length")
	}
	uint16Data := read2BytesPool.Get().([]byte)
	defer read2BytesPool.Put(uint16Data)
	_, err := r.Read(uint16Data)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(uint16Data), nil
}

func ReadUint32(r *bytes.Reader) (uint32, error) {
	if r.Size() < 4 {
		return 0, errors.New("invalid length")
	}
	uint32Data := read4BytesPool.Get().([]byte)
	defer read2BytesPool.Put(uint32Data)
	_, err := r.Read(uint32Data)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(uint32Data), nil
}

func WriteUint32(w *bytes.Buffer, i uint32) {
	w.WriteByte(byte(i >> 24))
	w.WriteByte(byte(i >> 16))
	w.WriteByte(byte(i >> 8))
	w.WriteByte(byte(i))
}
