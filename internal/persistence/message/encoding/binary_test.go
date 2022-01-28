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
	"github.com/stretchr/testify/assert"
	"github.com/yunqi/lighthouse/internal/packet"
	"github.com/yunqi/lighthouse/internal/persistence/message"
	"testing"
)

func TestEncodeMessage(t *testing.T) {
	m := &message.Message{
		Dup:                    false,
		QoS:                    packet.QoS0,
		Retained:               false,
		Topic:                  "test",
		Payload:                []byte("payload"),
		PacketId:               packet.Id(11),
		ContentType:            "context/json",
		CorrelationData:        []byte("1"),
		MessageExpiry:          1,
		PayloadFormat:          packet.PayloadFormatBytes,
		ResponseTopic:          "",
		SubscriptionIdentifier: []uint32{1, 2},
	}
	buffer := &bytes.Buffer{}
	EncodeMessage(m, buffer)
	decodeMessage, err := DecodeMessage(bytes.NewReader(buffer.Bytes()))
	assert.NoError(t, err)
	assert.EqualValues(t, m, decodeMessage)
	decodeMessage, err = DecodeMessageFromBytes(buffer.Bytes())
	assert.NoError(t, err)
	assert.EqualValues(t, m, decodeMessage)
}
func TestDecodeMessageFromBytes(t *testing.T) {
	decodeMessage, err := DecodeMessageFromBytes([]byte{})
	assert.NoError(t, err)
	assert.Nil(t, decodeMessage)
}
