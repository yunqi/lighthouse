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
	"testing"
)

func TestNewConnect(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		fixedHeader := &FixedHeader{
			PacketType:   CONNECT,
			Flags:        FixedHeaderFlagReserved,
			RemainLength: 13,
		}
		connectBytes := bytes.NewBuffer([]byte{0x00, 0x04, 'M', 'Q', 'T', 'T',
			0x04,      // Protocol Level
			0x0,       // Connect Flags
			0x0, 0x02, // Keep Alive
			0x00, 0x01, 't', // Client Identifier
		})
		connect, err := NewConnect(fixedHeader, Version311, connectBytes)
		assert.NoError(t, err)
		assert.False(t, connect.ConnectFlags.UsernameFlag)
		assert.False(t, connect.ConnectFlags.PasswordFlag)
		assert.False(t, connect.ConnectFlags.WillFlag)
		assert.False(t, connect.ConnectFlags.CleanSession)
		assert.EqualValues(t, 0, connect.ConnectFlags.WillQoS)
		assert.EqualValues(t, 2, connect.KeepAlive)
		assert.EqualValues(t, []byte{'t'}, connect.ClientId)
		assert.Nil(t, connect.Username)
		assert.Nil(t, connect.Password)
		assert.Nil(t, connect.WillTopic)
		assert.EqualValues(t, Version311, connect.ProtocolLevel)
		assert.EqualValues(t, []byte("MQTT"), connect.ProtocolName)

		_, _ = connect, err
	})
}
