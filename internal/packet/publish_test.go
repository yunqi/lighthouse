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

package packet

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"reflect"
	"testing"
)

func TestReadWritePublishPacket_V311(t *testing.T) {
	a := assert.New(t)
	var tt = []struct {
		topicName []byte
		dup       bool
		retain    bool
		qos       uint8
		pid       uint16
		payload   []byte
	}{
		{
			topicName: []byte("abc"),
			dup:       true,
			retain:    false,
			qos:       QoS1, pid: 10,
			payload: []byte("a")},
		{
			topicName: []byte("test topic name2"),
			dup:       false,
			retain:    true,
			qos:       QoS0,
			payload:   []byte("test payload2")},
		{
			topicName: []byte("test topic name3"),
			dup:       false,
			retain:    true,
			qos:       QoS2,
			pid:       11,
			payload:   []byte("test payload3"),
		},
	}

	for _, v := range tt {
		b := make([]byte, 0, 2048)
		buf := bytes.NewBuffer(b)
		pub := &Publish{
			Version:   Version311,
			Dup:       v.dup,
			QoS:       v.qos,
			Retain:    v.retain,
			TopicName: v.topicName,
			PacketId:  v.pid,
			Payload:   v.payload,
		}
		a.Nil(NewWriter(buf).WritePacketAndFlush(pub))
		packet, err := NewReader(buf).Read()
		a.Nil(err, string(v.topicName))
		_, err = buf.ReadByte()
		a.Equal(io.EOF, err, string(v.topicName))
		if p, ok := packet.(*Publish); ok {
			a.Equal(p.TopicName, pub.TopicName)
			a.Equal(p.PacketId, pub.PacketId)
			a.Equal(p.Payload, pub.Payload)
			a.Equal(p.Retain, pub.Retain)
			a.Equal(p.QoS, pub.QoS)
			a.Equal(p.Dup, pub.Dup)
		} else {
			t.Fatalf("Packet type error,want %v,got %v", reflect.TypeOf(&Publish{}), reflect.TypeOf(packet))
		}

	}

}
