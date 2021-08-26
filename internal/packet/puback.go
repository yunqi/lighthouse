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
	Puback struct {
		BasePub
	}
)

// NewPuback returns a Puback instance by the given FixHeader and io.Reader
func NewPuback(fixedHeader *FixedHeader, version Version, r io.Reader) (*Puback, error) {
	p := &Puback{BasePub{
		FixedHeader: fixedHeader,
		Version:     version,
	}}
	err := p.Decode(r)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (bp *Puback) Encode(w io.Writer) (err error) {
	bp.FixedHeader = &FixedHeader{PacketType: PUBACK, Flags: FixedHeaderFlagReserved}
	buf := &bytes.Buffer{}
	writeUint16(buf, bp.PacketId)
	return encode(bp.FixedHeader, buf, w)
}

func (bp *Puback) Decode(r io.Reader) (err error) {
	return bp.decode(r)

}

func (bp *Puback) String() string {
	return bp.string()
}
