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
	"encoding/json"
	"github.com/yunqi/lighthouse/internal/xerror"
	"io"
)

type BasePub struct {
	Version     Version
	FixedHeader *FixedHeader
	PacketId    PacketId
}

func (bp *BasePub) decode(r io.Reader) (err error) {
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

func (bp *BasePub) string() string {
	b, _ := json.Marshal(bp)
	return string(b)
}
