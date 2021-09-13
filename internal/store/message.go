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

package store

import (
	"fmt"
	"github.com/yunqi/lighthouse/internal/packet"
)

type (
	Message struct {
		Dup           bool
		QoS           uint8
		Retained      bool
		Topic         string
		Payload       []byte
		PacketId      packet.PacketId
		MessageExpiry uint32
	}
)

func (m *Message) String() string {
	return fmt.Sprintf("Message - Dup:%v, QoS:%d, Retained:%v, Topic:%s, Payload:%s,PacketId:%d, MessageExpiry:%d",
		m.Dup, m.QoS, m.Retained, m.Topic, m.Payload, m.PacketId, m.MessageExpiry,
	)
}

func MessageFromPublish(publish *packet.Publish) *Message {
	return &Message{
		Dup:      publish.Dup,
		QoS:      publish.QoS,
		Retained: publish.Retain,
		Topic:    string(publish.TopicName),
		Payload:  publish.Payload,
	}
}
