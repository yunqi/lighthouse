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

func TestFixedHeader_Encode(t *testing.T) {
	fixedHeader := FixedHeader{
		PacketType:   CONNECT,
		Flags:        FixedHeaderFlagReserved,
		RemainLength: 1111,
	}
	buf := bytes.Buffer{}
	err := fixedHeader.Encode(&buf)
	assert.NoError(t, err)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x10, 0xd7, 0x8}, buf.Bytes())

	buf.Reset()
	fixedHeader = FixedHeader{
		PacketType:   CONNECT,
		Flags:        FixedHeaderFlagReserved,
		RemainLength: RemainLength4ByteMax + 1,
	}
	err = fixedHeader.Encode(&buf)
	assert.ErrorIs(t, err, xerror.ErrMalformed)

}
