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
	"io"
)

type (
	Pubrec struct {
		BasePub
	}
)

func NewPubrec(fixedHeader *FixedHeader, version Version, r io.Reader) (*Pubrec, error) {
	p := &Pubrec{BasePub{
		FixedHeader: fixedHeader,
		Version:     version,
	}}
	err := p.Decode(r)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Pubrec) Encode(w io.Writer) (err error) {
	p.FixedHeader = &FixedHeader{PacketType: PUBREC, Flags: FixedHeaderFlagReserved}
	buf := &bytes.Buffer{}
	writeUint16(buf, p.PacketId)
	return encode(p.FixedHeader, buf, w)
}

func (p *Pubrec) Decode(r io.Reader) (err error) {
	return p.decode(r)
}

func (p *Pubrec) String() string {
	return p.string()
}

// CreateNewPubrel returns the Pubrel struct related to the Pubrec struct in QoS 2.
func (p *Pubrec) CreateNewPubrel() *Pubrel {
	pub := &Pubrel{
		BasePub{
			Version:     p.Version,
			FixedHeader: &FixedHeader{PacketType: PUBREL, Flags: FixedHeaderFlagPubrel},
			PacketId:    p.PacketId,
		},
	}
	return pub
}
