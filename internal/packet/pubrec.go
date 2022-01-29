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
	pubrecDefaultFixedHeader = &FixedHeader{PacketType: PUBREC, Flags: FixedHeaderFlagReserved}
)

type (
	Pubrec struct {
		Version     Version
		FixedHeader *FixedHeader
		PacketId    Id
	}
)

func NewPubrec(fixedHeader *FixedHeader, version Version, r io.Reader) (*Pubrec, error) {
	p := &Pubrec{
		FixedHeader: fixedHeader,
		Version:     version,
	}
	err := p.Decode(r)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Pubrec) Encode(w io.Writer) (err error) {
	p.FixedHeader = pubrecDefaultFixedHeader
	buf := &bytes.Buffer{}
	writeUint16(buf, p.PacketId)
	return encode(p.FixedHeader, buf, w)
}

func (p *Pubrec) Decode(r io.Reader) (err error) {
	b := make([]byte, p.FixedHeader.RemainLength)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return xerror.ErrMalformed
	}
	buf := bytes.NewBuffer(b)
	p.PacketId, err = readUint16(buf)
	if err != nil {
		return
	}
	return

}

func (p *Pubrec) String() string {
	return fmt.Sprintf("Pubrec - Version: %s, PacketId: %d", p.Version, p.PacketId)
}

// CreateNewPubrel returns the Pubrel struct related to the Pubrec struct in QoS 2.
func (p *Pubrec) CreateNewPubrel() *Pubrel {
	pub := &Pubrel{
		FixedHeader: &FixedHeader{PacketType: PUBREL, Flags: FixedHeaderFlagPubrel, RemainLength: 2},
		PacketId:    p.PacketId,
	}
	return pub
}
