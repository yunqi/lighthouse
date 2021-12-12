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
	"github.com/yunqi/lighthouse/internal/code"
	"github.com/yunqi/lighthouse/internal/xerror"
	"testing"
)

//
//func TestReadConnackPacket_V311(t *testing.T) {
//	connackPacketBytes := bytes.NewBuffer([]byte{32, 2,
//		0,
//		1,
//	})
//	packet, err := NewReader(connackPacketBytes).Read()
//	assert.Nil(t, err)
//	if cp, ok := packet.(*Connack); ok {
//		assert.False(t, cp.SessionPresent)
//		assert.EqualValues(t, 1, cp.Code)
//	} else {
//		t.Fatalf("Packet type error,want %v,got %v", reflect.TypeOf(&Connack{}), reflect.TypeOf(packet))
//	}
//}
func TestNewConnack(t *testing.T) {
	t.Run("correct test1", func(t *testing.T) {
		fixedHeader := &FixedHeader{
			PacketType:   CONNACK,
			Flags:        FixedHeaderFlagPubrel,
			RemainLength: 0,
		}
		connackBytes := bytes.NewBuffer([]byte{0,
			1,
		})
		connack, err := NewConnack(fixedHeader, Version311, connackBytes)
		assert.ErrorIs(t, err, xerror.ErrMalformed)
		assert.Nil(t, connack)

		fixedHeader = &FixedHeader{
			PacketType:   CONNACK,
			Flags:        FixedHeaderFlagReserved,
			RemainLength: 2,
		}
		connack, err = NewConnack(fixedHeader, Version311, connackBytes)
		assert.NoError(t, err)
		assert.NotNil(t, connack)
	})

	t.Run("error test1", func(t *testing.T) {
		connackBytes := bytes.NewBuffer([]byte{2,
			1,
		})
		fixedHeader := &FixedHeader{
			PacketType:   CONNACK,
			Flags:        FixedHeaderFlagReserved,
			RemainLength: 2,
		}
		connack, err := NewConnack(fixedHeader, Version311, connackBytes)
		assert.ErrorIs(t, err, xerror.ErrMalformed)
		assert.Nil(t, connack)
	})

	t.Run("error test2", func(t *testing.T) {
		connackBytes := bytes.NewBuffer([]byte{1})
		fixedHeader := &FixedHeader{
			PacketType:   CONNACK,
			Flags:        FixedHeaderFlagReserved,
			RemainLength: 2,
		}
		connack, err := NewConnack(fixedHeader, Version311, connackBytes)
		assert.ErrorIs(t, err, xerror.ErrMalformed)
		assert.Nil(t, connack)
	})

}

func TestConnack_Encode(t *testing.T) {

	t.Run("correct test1", func(t *testing.T) {
		// SessionPresent true
		// code 2
		connackBytes := bytes.NewBuffer([]byte{1, 2})
		fixedHeader := &FixedHeader{
			PacketType:   CONNACK,
			Flags:        FixedHeaderFlagReserved,
			RemainLength: 2,
		}

		connack, err := NewConnack(fixedHeader, Version311, connackBytes)
		assert.NoError(t, err)
		assert.NotNil(t, connack)
		buffer := &bytes.Buffer{}
		err = connack.Encode(buffer)
		assert.NoError(t, err)
		assert.EqualValues(t, []byte{CONNACK<<4 | FixedHeaderFlagReserved, 2, 1, 2}, buffer.Bytes())
	})

	t.Run("correct test2", func(t *testing.T) {
		// SessionPresent false
		// code 2
		connackBytes := bytes.NewBuffer([]byte{0, 2})
		fixedHeader := &FixedHeader{
			PacketType:   CONNACK,
			Flags:        FixedHeaderFlagReserved,
			RemainLength: 2,
		}

		connack, err := NewConnack(fixedHeader, Version311, connackBytes)
		assert.NoError(t, err)
		assert.NotNil(t, connack)
		buffer := &bytes.Buffer{}
		err = connack.Encode(buffer)
		assert.NoError(t, err)
		assert.EqualValues(t, []byte{CONNACK<<4 | FixedHeaderFlagReserved, 2, 0, 2}, buffer.Bytes())
	})
}

func TestConnack_String(t *testing.T) {

	fixedHeader := &FixedHeader{
		PacketType:   CONNACK,
		Flags:        FixedHeaderFlagReserved,
		RemainLength: 2,
	}
	type fields struct {
		Version        Version
		FixedHeader    *FixedHeader
		SessionPresent bool
		Code           code.Code
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"1",
			fields{
				Version:        Version311,
				FixedHeader:    fixedHeader,
				SessionPresent: false,
				Code:           0,
			},
			"Connack - Version: MQTT3.1.1, SessionPresent:false, Code:0",
		}, {
			"2",
			fields{
				Version:        Version311,
				FixedHeader:    fixedHeader,
				SessionPresent: true,
				Code:           1,
			},
			"Connack - Version: MQTT3.1.1, SessionPresent:true, Code:1",
		}, {
			"3",
			fields{
				Version:        Version311,
				FixedHeader:    fixedHeader,
				SessionPresent: true,
				Code:           11,
			},
			"Connack - Version: MQTT3.1.1, SessionPresent:true, Code:11",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Connack{
				Version:        tt.fields.Version,
				FixedHeader:    tt.fields.FixedHeader,
				SessionPresent: tt.fields.SessionPresent,
				Code:           tt.fields.Code,
			}
			if got := c.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
