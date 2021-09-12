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
	"io"
)

type (
	Pubrel struct {
		BasePub
	}
)

// NewPubrel returns a Pubrel instance by the given FixHeader and io.Reader.
func NewPubrel(fixedHeader *FixedHeader, r io.Reader) (*Pubrel, error) {
	p := &Pubrel{BasePub{
		FixedHeader: fixedHeader,
	}}
	err := p.Decode(r)
	if err != nil {
		return nil, err
	}
	return p, nil
}
func (p *Pubrel) Encode(w io.Writer) (err error) {
	p.FixedHeader = &FixedHeader{PacketType: PUBREL, Flags: FixedHeaderFlagReserved}
	buf := &bytes.Buffer{}
	writeUint16(buf, p.PacketId)
	return encode(p.FixedHeader, buf, w)
}

func (p *Pubrel) Decode(r io.Reader) (err error) {
	return p.decode(r)
}

func (p *Pubrel) String() string {
	return fmt.Sprintf("Pubrel - Version:%s, PacketId:%d", p.Version, p.PacketId)
}

// CreateNewPubcomp returns the Pubcomp struct related to the Pubrel struct in QoS 2.
func (p *Pubrel) CreateNewPubcomp() *Pubcomp {
	pub := &Pubcomp{BasePub{
		Version:     p.Version,
		FixedHeader: &FixedHeader{PacketType: PUBCOMP, Flags: FixedHeaderFlagReserved, RemainLength: 2},
		PacketId:    p.PacketId,
	}}
	return pub
}
