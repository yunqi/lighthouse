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
	"github.com/yunqi/lighthouse/internal/xerror"
	"io"
)

var (
	pubackDefaultFixedHeader = &FixedHeader{PacketType: PUBACK, Flags: FixedHeaderFlagReserved}
)

type (
	Puback struct {
		Version     Version
		FixedHeader *FixedHeader
		PacketId    Id
	}
)

// NewPuback returns a Puback instance by the given FixHeader and io.Reader
func NewPuback(fixedHeader *FixedHeader, version Version, r io.Reader) (*Puback, error) {
	p := &Puback{
		FixedHeader: fixedHeader,
		Version:     version,
	}
	err := p.Decode(r)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (bp *Puback) Encode(w io.Writer) (err error) {
	bp.FixedHeader = pubackDefaultFixedHeader
	buf := &bytes.Buffer{}
	writeUint16(buf, bp.PacketId)
	return encode(bp.FixedHeader, buf, w)
}

func (bp *Puback) Decode(r io.Reader) (err error) {
	b := make([]byte, bp.FixedHeader.RemainLength)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return xerror.ErrMalformed
	}
	buf := bytes.NewBuffer(b)
	bp.PacketId, err = readUint16(buf)
	if err != nil {
		return
	}
	return
}

func (bp *Puback) String() string {
	return fmt.Sprintf("Puback - Version: %s, PacketId: %d", bp.Version, bp.PacketId)
}
