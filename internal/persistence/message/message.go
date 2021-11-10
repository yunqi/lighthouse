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

package message

import (
	"fmt"
	"github.com/yunqi/lighthouse/internal/packet"
)

type (
	Message struct {
		Dup      bool
		QoS      uint8
		Retained bool
		Topic    string
		Payload  []byte
		PacketId packet.PacketId

		ContentType            string
		CorrelationData        []byte
		MessageExpiry          uint32
		PayloadFormat          packet.PayloadFormat
		ResponseTopic          string
		SubscriptionIdentifier []uint32
	}
)

func (m *Message) String() string {
	return fmt.Sprintf("Message - Dup:%v, QoS:%d, Retained:%v, Topic:%s, Payload:%s,PacketId:%d, MessageExpiry:%d",
		m.Dup, m.QoS, m.Retained, m.Topic, m.Payload, m.PacketId, m.MessageExpiry,
	)
}

func FromPublish(publish *packet.Publish) *Message {
	return &Message{
		Dup:      publish.Dup,
		QoS:      publish.QoS,
		Retained: publish.Retain,
		Topic:    string(publish.TopicName),
		Payload:  publish.Payload,
	}
}

// TotalBytes return the publish packets total bytes.
func (m *Message) TotalBytes(version packet.Version) uint32 {
	remainLenght := len(m.Payload) + 2 + len(m.Topic)
	if m.QoS > packet.QoS0 {
		remainLenght += 2
	}
	if version == packet.Version5 {
		propertyLenght := 0
		if m.PayloadFormat == packet.PayloadFormatString {
			propertyLenght += 2
		}
		if l := len(m.ContentType); l != 0 {
			propertyLenght += 3 + l
		}
		if l := len(m.CorrelationData); l != 0 {
			propertyLenght += 3 + l
		}

		for _, v := range m.SubscriptionIdentifier {
			propertyLenght++
			propertyLenght += getVariableLength(int(v))
		}

		if m.MessageExpiry != 0 {
			propertyLenght += 5
		}
		if l := len(m.ResponseTopic); l != 0 {
			propertyLenght += 3 + l
		}

		remainLenght += propertyLenght + getVariableLength(propertyLenght)
	}
	if remainLenght <= packet.RemainLength1ByteMax {
		return 2 + uint32(remainLenght)
	} else if remainLenght <= packet.RemainLength2ByteMax {
		return 3 + uint32(remainLenght)
	} else if remainLenght <= packet.RemainLength3ByteMax {
		return 4 + uint32(remainLenght)
	}
	return 5 + uint32(remainLenght)
}
func getVariableLength(l int) int {
	if l <= packet.RemainLength1ByteMax {
		return 1
	} else if l <= packet.RemainLength2ByteMax {

		return 2
	} else if l <= packet.RemainLength3ByteMax {
		return 3
	} else if l <= packet.RemainLength4ByteMax {
		return 4
	}
	return 0
}
