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

func TestNewDisConnect(t *testing.T) {
	t.Run("correct test", func(t *testing.T) {
		fixedHeader := &FixedHeader{
			PacketType:   DISCONNECT,
			Flags:        FixedHeaderFlagReserved,
			RemainLength: 0,
		}

		disconnect, err := NewDisconnect(fixedHeader, Version311, &bytes.Buffer{})
		assert.NoError(t, err)
		assert.NotNil(t, disconnect)
	})

	t.Run("Flags test", func(t *testing.T) {
		fixedHeader := &FixedHeader{
			PacketType:   DISCONNECT,
			Flags:        FixedHeaderFlagPubrel,
			RemainLength: 0,
		}

		disconnect, err := NewDisconnect(fixedHeader, Version311, &bytes.Buffer{})
		assert.ErrorIs(t, err, xerror.ErrMalformed)
		assert.Nil(t, disconnect)
	})
}
func TestDisconnect_Encode(t *testing.T) {
	t.Run("", func(t *testing.T) {
		fixedHeader := &FixedHeader{
			PacketType:   DISCONNECT,
			Flags:        FixedHeaderFlagReserved,
			RemainLength: 0,
		}

		disconnect, err := NewDisconnect(fixedHeader, Version311, &bytes.Buffer{})
		assert.NoError(t, err)
		assert.NotNil(t, disconnect)
		buffer := &bytes.Buffer{}
		err = disconnect.Encode(buffer)
		assert.NoError(t, err)
		assert.Equal(t, []byte{0xe0, 0x0}, buffer.Bytes())
	})
}
func TestDisconnect_String(t *testing.T) {
	fixedHeader := &FixedHeader{
		PacketType:   DISCONNECT,
		Flags:        FixedHeaderFlagReserved,
		RemainLength: 0,
	}

	disconnect, err := NewDisconnect(fixedHeader, Version311, &bytes.Buffer{})
	assert.NoError(t, err)
	assert.NotNil(t, disconnect)
	assert.Equal(t, "Disconnect - Version: MQTT3.1.1", disconnect.String())
}
