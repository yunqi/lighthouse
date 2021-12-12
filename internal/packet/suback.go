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
	"fmt"
	"github.com/yunqi/lighthouse/internal/code"
	"github.com/yunqi/lighthouse/internal/xerror"
	"io"
)

var subackDefaultFixedHeader = &FixedHeader{PacketType: SUBACK, Flags: FixedHeaderFlagReserved}

type (
	Suback struct {
		Version     Version
		FixedHeader *FixedHeader
		PacketId    PacketId
		Payload     []code.Code
	}
)

// NewSuback returns a Suback instance by the given FixHeader and io.Reader.
func NewSuback(fixedHeader *FixedHeader, version Version, r io.Reader) (*Suback, error) {
	p := &Suback{FixedHeader: fixedHeader, Version: version}
	//判断 标志位 flags 是否合法[MQTT-3.8.1-1]
	if fixedHeader.Flags != FixedHeaderFlagReserved {
		return nil, xerror.ErrMalformed
	}
	err := p.Decode(r)
	return p, err
}
func (s *Suback) Encode(w io.Writer) (err error) {
	s.FixedHeader = subackDefaultFixedHeader
	bufw := &bytes.Buffer{}
	writeUint16(bufw, s.PacketId)

	bufw.Write(s.Payload)
	return encode(s.FixedHeader, bufw, w)
}

func (s *Suback) Decode(r io.Reader) (err error) {
	b := make([]byte, s.FixedHeader.RemainLength)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return xerror.ErrMalformed
	}
	buf := bytes.NewBuffer(b)

	s.PacketId, err = readUint16(buf)
	if err != nil {
		return xerror.ErrMalformed
	}

	for buf.Len() != 0 {
		b, err := buf.ReadByte()
		if err != nil {
			return xerror.ErrMalformed
		}
		s.Payload = append(s.Payload, b)
	}
	return
}

func (s *Suback) String() string {
	return fmt.Sprintf("Suback - Versoin: %s, PacketId: %d", s.Version, s.PacketId)
}
