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
	"bufio"
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewReader(t *testing.T) {
	buf := bytes.NewBufferString("test")
	reader := NewReader(buf)
	assert.NotNil(t, reader)
	assert.EqualValues(t, 2048, reader.buf.Size())

	newReader := bufio.NewReader(buf)
	reader = NewReader(newReader)

	assert.NotNil(t, reader)
	assert.EqualValues(t, 4096, reader.buf.Size())
}
func TestReader_Read(t *testing.T) {
	t.Run("correct test", func(t *testing.T) {
		packetBytes := []byte{
			0x10,                           // PacketType,Flags
			13,                             // RemainLength
			0x00, 0x04, 'M', 'Q', 'T', 'T', // Protocol name
			0x04,      // Protocol Level
			0x0,       // Connect Flags
			0x0, 0x02, // Keep Alive
			0x00, 0x01, 't', // Client Identifier
		}
		reader := bytes.NewReader(packetBytes)
		newReader := NewReader(reader)
		packet, err := newReader.Read()
		assert.NoError(t, err)
		assert.Equal(t, "Connect - Version: MQTT3.1.1,ProtocolLevel: 4, UsernameFlag: false, PasswordFlag: false, ProtocolName: MQTT, CleanSession: false, KeepAlive: 2, ClientId: t, Username: , Password: , WillFlag: false, WillRetain: false, WillQos: 0, WillTopic: , WillMessage: ", packet.String())
		buffer := &bytes.Buffer{}
		assert.NoError(t, packet.Encode(buffer))
		assert.EqualValues(t, packetBytes, buffer.Bytes())
	})

	t.Run("error test1", func(t *testing.T) {
		var packetBytes []byte
		reader := bytes.NewReader(packetBytes)
		newReader := NewReader(reader)
		packet, err := newReader.Read()
		assert.Error(t, err)
		assert.Nil(t, packet)

	})
	t.Run("error test2", func(t *testing.T) {
		packetBytes := []byte{
			0x10, // PacketType,Flags
		}
		reader := bytes.NewReader(packetBytes)
		newReader := NewReader(reader)
		packet, err := newReader.Read()
		assert.Error(t, err)
		assert.Nil(t, packet)

	})

}

func TestNewWriter(t *testing.T) {
	buf := bytes.NewBufferString("test")
	reader := NewWriter(buf)
	assert.NotNil(t, reader)
	assert.EqualValues(t, 2048, reader.buf.Size())

	newReader := bufio.NewWriter(buf)
	reader = NewWriter(newReader)

	assert.NotNil(t, reader)
	assert.EqualValues(t, 4096, reader.buf.Size())

}
func TestWriter_WritePacketAndFlush(t *testing.T) {
	t.Run("correct test", func(t *testing.T) {
		connect := &Connect{
			Version: Version311,
			FixedHeader: &FixedHeader{
				PacketType:   CONNECT,
				Flags:        FixedHeaderFlagReserved,
				RemainLength: 13,
			},
			ProtocolName:  []byte("MQTT"),
			ProtocolLevel: 4,
			ConnectFlags: ConnectFlags{
				CleanSession: false,
				WillFlag:     false,
				WillQoS:      0,
				WillRetain:   false,
				PasswordFlag: false,
				UsernameFlag: false,
			},
			KeepAlive:   2,
			WillTopic:   nil,
			WillMessage: nil,
			ClientId:    []byte("t"),
			Username:    nil,
			Password:    nil,
		}
		buffer := &bytes.Buffer{}
		writer := NewWriter(buffer)
		err := writer.WritePacketAndFlush(connect)
		assert.NoError(t, err)
		assert.EqualValues(t, []byte{
			0x10,                           // PacketType,Flags
			13,                             // RemainLength
			0x00, 0x04, 'M', 'Q', 'T', 'T', // Protocol name
			0x04,      // Protocol Level
			0x0,       // Connect Flags
			0x0, 0x02, // Keep Alive
			0x00, 0x01, 't', // Client Identifier
		}, buffer.Bytes())
	})

}
func TestWriter_WriteAndFlush(t *testing.T) {
	packetBytes := []byte{
		0x10,                           // PacketType,Flags
		13,                             // RemainLength
		0x00, 0x04, 'M', 'Q', 'T', 'T', // Protocol name
		0x04,      // Protocol Level
		0x0,       // Connect Flags
		0x0, 0x02, // Keep Alive
		0x00, 0x01, 't', // Client Identifier
	}
	buffer := &bytes.Buffer{}
	writer := NewWriter(buffer)
	err := writer.Write(packetBytes)
	assert.NoError(t, err)
	err = writer.Flush()
	assert.NoError(t, err)
	assert.EqualValues(t, []byte{
		0x10,                           // PacketType,Flags
		13,                             // RemainLength
		0x00, 0x04, 'M', 'Q', 'T', 'T', // Protocol name
		0x04,      // Protocol Level
		0x0,       // Connect Flags
		0x0, 0x02, // Keep Alive
		0x00, 0x01, 't', // Client Identifier
	}, buffer.Bytes())

}
