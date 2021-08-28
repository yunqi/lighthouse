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
	"encoding/json"
	"github.com/yunqi/lighthouse/internal/code"
	"github.com/yunqi/lighthouse/internal/xerror"
	"io"
)

type (
	Unsuback struct {
		Version     Version
		FixedHeader *FixedHeader
		PacketId    PacketId
		Payload     []code.Code
	}
)

func (u *Unsuback) Encode(w io.Writer) (err error) {
	u.FixedHeader = &FixedHeader{PacketType: UNSUBACK, Flags: FixedHeaderFlagReserved}
	buf := &bytes.Buffer{}
	writeUint16(buf, u.PacketId)

	// payload
	buf.Write(u.Payload)

	return encode(u.FixedHeader, buf, w)
}

func (u *Unsuback) Decode(r io.Reader) (err error) {
	b := make([]byte, u.FixedHeader.RemainLength)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return xerror.ErrMalformed
	}
	buf := bytes.NewBuffer(b)
	u.PacketId, err = readUint16(buf)
	if err != nil {
		return
	}
	return nil
}

func (u *Unsuback) String() string {
	b, _ := json.Marshal(u)
	return string(b)
}
