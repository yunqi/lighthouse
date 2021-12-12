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

type (
	Pubcomp struct {
		Version     Version
		FixedHeader *FixedHeader
		PacketId    PacketId
	}
)

// NewPubcomp returns a Pubcomp instance by the given FixHeader and io.Reader
func NewPubcomp(fixedHeader *FixedHeader, version Version, r io.Reader) (*Pubcomp, error) {
	p := &Pubcomp{
		FixedHeader: fixedHeader, Version: version,
	}
	err := p.Decode(r)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (pb *Pubcomp) Encode(w io.Writer) (err error) {
	pb.FixedHeader = &FixedHeader{PacketType: PUBCOMP, Flags: FixedHeaderFlagReserved}
	buf := &bytes.Buffer{}
	writeUint16(buf, pb.PacketId)
	return encode(pb.FixedHeader, buf, w)
}

func (pb *Pubcomp) Decode(r io.Reader) (err error) {
	b := make([]byte, pb.FixedHeader.RemainLength)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return xerror.ErrMalformed
	}
	buf := bytes.NewBuffer(b)
	pb.PacketId, err = readUint16(buf)
	if err != nil {
		return
	}
	return
}

// String returns string.
func (pb *Pubcomp) String() string {
	return fmt.Sprintf("Pubcomp - Version: %s, PacketId: %d", pb.Version, pb.PacketId)
}
