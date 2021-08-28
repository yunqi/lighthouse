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
	"encoding/json"
	"github.com/yunqi/lighthouse/internal/xerror"
	"io"
)

type (
	Disconnect struct {
		Version     Version
		FixedHeader *FixedHeader
	}
)

// NewDisConnect returns a Disconnect instance by the given FixHeader and io.Reader
func NewDisConnect(fixedHeader *FixedHeader, version Version, r io.Reader) (*Disconnect, error) {
	if fixedHeader.Flags != 0 {
		return nil, xerror.ErrMalformed
	}
	p := &Disconnect{FixedHeader: fixedHeader, Version: version}
	err := p.Decode(r)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (d *Disconnect) Encode(w io.Writer) (err error) {
	d.FixedHeader = &FixedHeader{PacketType: DISCONNECT, Flags: FixedHeaderFlagReserved}
	d.FixedHeader.RemainLength = 0
	return d.FixedHeader.Encode(w)
}

func (d *Disconnect) Decode(r io.Reader) (err error) {
	b := make([]byte, d.FixedHeader.RemainLength)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return xerror.ErrMalformed
	}
	return
}

func (d *Disconnect) String() string {
	b, _ := json.Marshal(d)
	return string(b)
}
