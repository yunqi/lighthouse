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

package binary

import (
	"bytes"
	"github.com/chenquan/go-pkg/xbinary"
	"github.com/yunqi/lighthouse/internal/packet"
	"github.com/yunqi/lighthouse/internal/persistence/message"
	"io"
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
	_ = xbinary.WriteBool(w, msg.Dup)
	_ = w.WriteByte(msg.QoS)
	_ = xbinary.WriteBool(w, msg.Retained)
	_ = xbinary.WriteBytes(w, []byte(msg.Topic))
	_ = xbinary.WriteBytes(w, msg.Payload)
	_ = xbinary.WriteUint16(w, msg.PacketId)
	if len(msg.ContentType) != 0 {
		_ = w.WriteByte(packet.PropContentType)
		_ = xbinary.WriteBytes(w, []byte(msg.ContentType))
	}
	if len(msg.CorrelationData) != 0 {
		_ = w.WriteByte(packet.PropCorrelationData)
		_ = xbinary.WriteBytes(w, msg.CorrelationData)
	}
	if msg.MessageExpiry != 0 {
		_ = w.WriteByte(packet.PropMessageExpiry)
		_ = xbinary.WriteUint32(w, msg.MessageExpiry)
	}
	_ = w.WriteByte(packet.PropPayloadFormat)
	_ = w.WriteByte(msg.PayloadFormat)

	if len(msg.ResponseTopic) != 0 {
		_ = w.WriteByte(packet.PropResponseTopic)
		_ = xbinary.WriteBytes(w, []byte(msg.ResponseTopic))
	}
	for _, v := range msg.SubscriptionIdentifier {
		_ = w.WriteByte(packet.PropSubscriptionIdentifier)
		l, _ := packet.EncodeRemainLength(int(v))
		_, _ = w.Write(l)
	}
}

func DecodeMessage(r *bytes.Reader) (*message.Message, error) {
	msg := &message.Message{}

	var err error
	msg.Dup, err = xbinary.ReadBool(r)
	if err != nil {
		return nil, err
	}
	msg.QoS, err = r.ReadByte()
	if err != nil {
		return nil, err
	}
	msg.Retained, err = xbinary.ReadBool(r)
	if err != nil {
		return nil, err
	}
	topic, err := xbinary.ReadBytes(r)
	if err != nil {
		return nil, err
	}
	msg.Topic = string(topic)
	msg.Payload, err = xbinary.ReadBytes(r)
	if err != nil {
		return nil, err
	}
	msg.PacketId, err = xbinary.ReadUint16(r)
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
			v, err := xbinary.ReadBytes(r)
			if err != nil {
				return nil, err
			}
			msg.ContentType = string(v)
		case packet.PropCorrelationData:
			msg.CorrelationData, err = xbinary.ReadBytes(r)
			if err != nil {
				return nil, err
			}
		case packet.PropMessageExpiry:
			msg.MessageExpiry, err = xbinary.ReadUint32(r)
			if err != nil {
				return nil, err
			}
		case packet.PropPayloadFormat:
			msg.PayloadFormat, err = r.ReadByte()
			if err != nil {
				return nil, err
			}
		case packet.PropResponseTopic:
			v, err := xbinary.ReadBytes(r)
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
