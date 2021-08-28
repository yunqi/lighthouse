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
	"github.com/yunqi/lighthouse/internal/xerror"
	"testing"
)

func TestNewConnect(t *testing.T) {
	t.Run("correct test1", func(t *testing.T) {
		fixedHeader := &FixedHeader{
			PacketType:   CONNECT,
			Flags:        FixedHeaderFlagReserved,
			RemainLength: 13,
		}
		connectBytes := bytes.NewBuffer([]byte{
			0x00, 0x04, 'M', 'Q', 'T', 'T', // Protocol name
			0x04,      // Protocol Level
			0x0,       // Connect Flags
			0x0, 0x02, // Keep Alive
			0x00, 0x01, 't', // Client Identifier
		})
		connect, err := NewConnect(fixedHeader, Version311, connectBytes)
		assert.NoError(t, err)
		assert.False(t, connect.UsernameFlag)
		assert.False(t, connect.PasswordFlag)
		assert.False(t, connect.WillFlag)
		assert.False(t, connect.CleanSession)
		assert.EqualValues(t, 0, connect.WillQoS)
		assert.EqualValues(t, 2, connect.KeepAlive)
		assert.EqualValues(t, []byte{'t'}, connect.ClientId)
		assert.Nil(t, connect.Username)
		assert.Nil(t, connect.Password)
		assert.Nil(t, connect.WillTopic)
		assert.EqualValues(t, Version311, connect.ProtocolLevel)
		assert.EqualValues(t, []byte("MQTT"), connect.ProtocolName)

		_, _ = connect, err
	})

	t.Run("FixedHeaderFlagReserved error", func(t *testing.T) {
		fixedHeader := &FixedHeader{
			PacketType:   CONNECT,
			Flags:        FixedHeaderFlagPubrel,
			RemainLength: 13,
		}
		connectBytes := bytes.NewBuffer([]byte{
			0x00, 0x03, 'M', 'Q', 'T', 'T', // Protocol name
			0x04,      // Protocol Level
			0x0,       // Connect Flags
			0x0, 0x02, // Keep Alive
			0x00, 0x01, 't', // Client Identifier
		})
		connect, err := NewConnect(fixedHeader, Version311, connectBytes)
		assert.ErrorIs(t, err, xerror.ErrMalformed)
		assert.Nil(t, connect)
	})
	t.Run("protocol name error", func(t *testing.T) {
		fixedHeader := &FixedHeader{
			PacketType:   CONNECT,
			Flags:        FixedHeaderFlagReserved,
			RemainLength: 13,
		}
		connectBytes := bytes.NewBuffer([]byte{
			0x00, 0x03, 'M', 'Q', 'T', 'T', // Protocol name
			0x04,      // Protocol Level
			0x0,       // Connect Flags
			0x0, 0x02, // Keep Alive
			0x00, 0x01, 't', // Client Identifier
		})
		connect, err := NewConnect(fixedHeader, Version311, connectBytes)
		assert.ErrorIs(t, err, xerror.ErrV3UnacceptableProtocolVersion)
		assert.Nil(t, connect)
	})

	t.Run("remainLength error", func(t *testing.T) {
		fixedHeader := &FixedHeader{
			PacketType:   CONNECT,
			Flags:        FixedHeaderFlagReserved,
			RemainLength: 12,
		}
		connectBytes := bytes.NewBuffer([]byte{
			0x00, 0x04, 'M', 'Q', 'T', 'T', // Protocol name
			0x04,      // Protocol Level
			0x0,       // Connect Flags
			0x0, 0x02, // Keep Alive
			0x00, 0x01, 't', // Client Identifier
		})
		connect, err := NewConnect(fixedHeader, Version311, connectBytes)
		assert.ErrorIs(t, err, xerror.ErrMalformed)
		assert.Nil(t, connect)
	})

	t.Run("protocol Level error", func(t *testing.T) {
		fixedHeader := &FixedHeader{
			PacketType:   CONNECT,
			Flags:        FixedHeaderFlagReserved,
			RemainLength: 13,
		}
		connectBytes := bytes.NewBuffer([]byte{
			0x00, 0x04, 'M', 'Q', 'T', 'T', // Protocol name
			0x02,      // Protocol Level
			0x0,       // Connect Flags
			0x0, 0x02, // Keep Alive
			0x00, 0x01, 't', // Client Identifier
		})
		connect, err := NewConnect(fixedHeader, Version311, connectBytes)
		assert.ErrorIs(t, err, xerror.ErrV3UnacceptableProtocolVersion)
		assert.Nil(t, connect)
	})

	t.Run(" Connect Flags error", func(t *testing.T) {
		fixedHeader := &FixedHeader{
			PacketType:   CONNECT,
			Flags:        FixedHeaderFlagReserved,
			RemainLength: 13,
		}
		connectBytes := bytes.NewBuffer([]byte{
			0x00, 0x04, 'M', 'Q', 'T', 'T', // Protocol name
			0x04,      // Protocol Level
			0x01,      // Connect Flags
			0x0, 0x02, // Keep Alive
			0x00, 0x01, 't', // Client Identifier
		})
		connect, err := NewConnect(fixedHeader, Version311, connectBytes)
		assert.ErrorIs(t, err, xerror.ErrMalformed)
		assert.Nil(t, connect)
	})
	t.Run("will QoS error", func(t *testing.T) {
		fixedHeader := &FixedHeader{
			PacketType:   CONNECT,
			Flags:        FixedHeaderFlagReserved,
			RemainLength: 13,
		}
		connectBytes := bytes.NewBuffer([]byte{
			0x00, 0x04, 'M', 'Q', 'T', 'T', // Protocol name
			0x04,      // Protocol Level
			0x18,      // Connect Flags
			0x0, 0x02, // Keep Alive
			0x00, 0x01, 't', // Client Identifier
		})
		connect, err := NewConnect(fixedHeader, Version311, connectBytes)
		assert.ErrorIs(t, err, xerror.ErrMalformed)
		assert.Nil(t, connect)
	})

	t.Run("Will Retain error", func(t *testing.T) {
		fixedHeader := &FixedHeader{
			PacketType:   CONNECT,
			Flags:        FixedHeaderFlagReserved,
			RemainLength: 13,
		}
		connectBytes := bytes.NewBuffer([]byte{
			0x00, 0x04, 'M', 'Q', 'T', 'T', // Protocol name
			0x04,      // Protocol Level
			0x20,      // Connect Flags
			0x0, 0x02, // Keep Alive
			0x00, 0x01, 't', // Client Identifier
		})
		connect, err := NewConnect(fixedHeader, Version311, connectBytes)
		assert.ErrorIs(t, err, xerror.ErrMalformed)
		assert.Nil(t, connect)
	})

	t.Run("KeepAlive error", func(t *testing.T) {
		fixedHeader := &FixedHeader{
			PacketType:   CONNECT,
			Flags:        FixedHeaderFlagReserved,
			RemainLength: 9,
		}
		connectBytes := bytes.NewBuffer([]byte{
			0x00, 0x04, 'M', 'Q', 'T', 'T', // Protocol name
			0x04,      // Protocol Level
			0x0,       // Connect Flags
			0x0, 0x02, // Keep Alive
			0x00, 0x01, 't', // Client Identifier
		})
		connect, err := NewConnect(fixedHeader, Version311, connectBytes)
		assert.ErrorIs(t, err, xerror.ErrMalformed)
		assert.Nil(t, connect)
	})
	t.Run("ClientId error", func(t *testing.T) {
		fixedHeader := &FixedHeader{
			PacketType:   CONNECT,
			Flags:        FixedHeaderFlagReserved,
			RemainLength: 12,
		}
		connectBytes := bytes.NewBuffer([]byte{
			0x00, 0x04, 'M', 'Q', 'T', 'T', // Protocol name
			0x04,      // Protocol Level
			0x0,       // Connect Flags
			0x0, 0x02, // Keep Alive
			0x00, 0x00, // Client Identifier
		})
		connect, err := NewConnect(fixedHeader, Version311, connectBytes)
		assert.ErrorIs(t, err, xerror.ErrV3IdentifierRejected)
		assert.Nil(t, connect)
	})
	t.Run("WillTopic error", func(t *testing.T) {
		fixedHeader := &FixedHeader{
			PacketType:   CONNECT,
			Flags:        FixedHeaderFlagReserved,
			RemainLength: 13,
		}
		connectBytes := bytes.NewBuffer([]byte{
			0x00, 0x04, 'M', 'Q', 'T', 'T', // Protocol name
			0x04,      // Protocol Level
			0x4,       // Connect Flags
			0x0, 0x02, // Keep Alive
			0x00, 0x01, 't', // Client Identifier
		})
		connect, err := NewConnect(fixedHeader, Version311, connectBytes)
		assert.ErrorIs(t, err, xerror.ErrMalformed)
		assert.Nil(t, connect)
	})

	t.Run("WillMessage error", func(t *testing.T) {
		fixedHeader := &FixedHeader{
			PacketType:   CONNECT,
			Flags:        FixedHeaderFlagReserved,
			RemainLength: 17,
		}
		connectBytes := bytes.NewBuffer([]byte{
			0x00, 0x04, 'M', 'Q', 'T', 'T', // Protocol name
			0x04,      // Protocol Level
			0x4,       // Connect Flags
			0x0, 0x02, // Keep Alive
			0x00, 0x01, 't', // Client Identifier
			0x00, 0x02, 't', '2', // Will Topic
		})
		connect, err := NewConnect(fixedHeader, Version311, connectBytes)
		assert.ErrorIs(t, err, xerror.ErrMalformed)
		assert.Nil(t, connect)
	})

	t.Run("Username error", func(t *testing.T) {
		fixedHeader := &FixedHeader{
			PacketType:   CONNECT,
			Flags:        FixedHeaderFlagReserved,
			RemainLength: 17,
		}
		connectBytes := bytes.NewBuffer([]byte{
			0x00, 0x04, 'M', 'Q', 'T', 'T', // Protocol name
			0x04,      // Protocol Level
			0x80,      // Connect Flags
			0x0, 0x02, // Keep Alive
			0x00, 0x01, 't', // Client Identifier
			0x00, 0x03, 't', '2', // User Name
		})
		connect, err := NewConnect(fixedHeader, Version311, connectBytes)
		assert.ErrorIs(t, err, xerror.ErrMalformed)
		assert.Nil(t, connect)
	})
	t.Run("Password error", func(t *testing.T) {
		fixedHeader := &FixedHeader{
			PacketType:   CONNECT,
			Flags:        FixedHeaderFlagReserved,
			RemainLength: 17,
		}
		connectBytes := bytes.NewBuffer([]byte{
			0x00, 0x04, 'M', 'Q', 'T', 'T', // Protocol name
			0x04,      // Protocol Level
			0xc0,      // Connect Flags
			0x0, 0x02, // Keep Alive
			0x00, 0x01, 't', // Client Identifier
			0x00, 0x02, 't', '2', // User Name
			0x00, 0x03, 't', '2', // Password
		})
		connect, err := NewConnect(fixedHeader, Version311, connectBytes)
		assert.ErrorIs(t, err, xerror.ErrMalformed)
		assert.Nil(t, connect)
	})

}

func TestConnect_Encode(t *testing.T) {
	fixedHeader := &FixedHeader{
		PacketType:   CONNECT,
		Flags:        FixedHeaderFlagReserved,
		RemainLength: 25,
	}
	connectBytes := bytes.NewBuffer([]byte{
		0x00, 0x04, 'M', 'Q', 'T', 'T', // Protocol name
		0x04,      // Protocol Level
		0xfe,      // Connect Flags
		0x0, 0x02, // Keep Alive
		0x00, 0x01, 't', // Client Identifier
		0x00, 0x01, 't', // Will Topic
		0x00, 0x01, 't', // Will Message
		0x00, 0x01, 't', // User Name
		0x00, 0x01, 't', // Password
	})
	connect, err := NewConnect(fixedHeader, Version311, connectBytes)
	buffer := &bytes.Buffer{}
	err = connect.Encode(buffer)
	assert.NoError(t, err)
	assert.NotNil(t, buffer)
}

func TestConnect_String(t *testing.T) {
	type fields struct {
		Version       Version
		FixedHeader   *FixedHeader
		ProtocolName  []byte
		ProtocolLevel byte
		ConnectFlags  ConnectFlags
		KeepAlive     uint16
		WillTopic     []byte
		WillMessage   []byte
		ClientId      []byte
		Username      []byte
		Password      []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"1",
			fields{
				Version:       Version311,
				FixedHeader:   &FixedHeader{},
				ProtocolName:  []byte("MQTT"),
				ProtocolLevel: 4,
				ConnectFlags: ConnectFlags{
					CleanSession: true,
					WillFlag:     true,
					WillQoS:      0,
					WillRetain:   true,
					PasswordFlag: true,
					UsernameFlag: true,
				},
				KeepAlive:   0,
				WillTopic:   []byte("t"),
				WillMessage: []byte("t"),
				ClientId:    []byte("t"),
				Username:    []byte("t"),
				Password:    []byte("t"),
			},
			"Connect - Version: MQTT3.1.1,ProtocolLevel: 4, UsernameFlag: true, PasswordFlag: true, ProtocolName: MQTT, CleanSession: true, KeepAlive: 0, ClientId: t, Username: t, Password: t, WillFlag: true, WillRetain: true, WillQos: 0, WillTopic: t, WillMessage: t",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Connect{
				Version:       tt.fields.Version,
				FixedHeader:   tt.fields.FixedHeader,
				ProtocolName:  tt.fields.ProtocolName,
				ProtocolLevel: tt.fields.ProtocolLevel,
				ConnectFlags:  tt.fields.ConnectFlags,
				KeepAlive:     tt.fields.KeepAlive,
				WillTopic:     tt.fields.WillTopic,
				WillMessage:   tt.fields.WillMessage,
				ClientId:      tt.fields.ClientId,
				Username:      tt.fields.Username,
				Password:      tt.fields.Password,
			}
			if got := c.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
