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
	"github.com/chenquan/lighthouse/internal/xerror"
	"io"
)

type (
	Pingresp struct {
		FixedHeader *FixedHeader
	}
)

// NewPingresp returns a Pingresp instance by the given FixHeader and io.Reader
func NewPingresp(fixedHeader *FixedHeader, r io.Reader) (*Pingresp, error) {
	if fixedHeader.Flags != FixedHeaderFlagReserved {
		return nil, xerror.ErrMalformed
	}
	p := &Pingresp{FixedHeader: fixedHeader}
	err := p.Decode(r)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Pingresp) Encode(w io.Writer) (err error) {
	p.FixedHeader = &FixedHeader{PacketType: PINGRESP}
	return p.FixedHeader.Encode(w)
}

func (p *Pingresp) Decode(r io.Reader) (err error) {
	if p.FixedHeader.RemainLength != 0 {
		return xerror.ErrMalformed
	}
	return nil
}

func (p *Pingresp) String() string {
	return "PINGRESP"
}
