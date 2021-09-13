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
	"github.com/yunqi/lighthouse/internal/xerror"
	"io"
)

type (
	Pingreq struct {
		FixedHeader *FixedHeader
	}
)

// NewPingreq returns a Pingreq instance by the given FixHeader and io.Reader
func NewPingreq(fixedHeader *FixedHeader, r io.Reader) (*Pingreq, error) {
	if fixedHeader.Flags != FixedHeaderFlagReserved {
		return nil, xerror.ErrMalformed
	}
	p := &Pingreq{FixedHeader: fixedHeader}
	err := p.Decode(r)
	if err != nil {
		return nil, err
	}
	return p, nil
}
func (p *Pingreq) Encode(w io.Writer) (err error) {
	p.FixedHeader = &FixedHeader{PacketType: PINGREQ}
	return p.FixedHeader.Encode(w)
}

func (p *Pingreq) Decode(_ io.Reader) (err error) {
	if p.FixedHeader.RemainLength != 0 {
		return xerror.ErrMalformed
	}
	return nil
}

func (p *Pingreq) String() string {
	return "PINGREQ"
}

func (p *Pingreq) CreatePingresp() *Pingresp {
	fixedHeader := pingrespDefaultFixedHeader
	return &Pingresp{FixedHeader: fixedHeader}
}
