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
	"github.com/yunqi/lighthouse/internal/code"
	"github.com/yunqi/lighthouse/internal/xerror"
	"io"
)

type Connack struct {
	Version        Version
	FixedHeader    *FixedHeader
	SessionPresent bool
	Code           code.Code
}

// NewConnack returns a Connack instance by the given FixHeader and io.Reader
func NewConnack(fixedHeader *FixedHeader, version Version, r io.Reader) (*Connack, error) {
	connack := &Connack{FixedHeader: fixedHeader, Version: version}
	if fixedHeader.Flags != FixedHeaderFlagReserved {
		return nil, xerror.ErrMalformed
	}
	err := connack.Decode(r)
	if err != nil {
		return nil, err
	}
	return connack, err
}

// Encode the packet struct into bytes and writes it into io.Writer.
func (c *Connack) Encode(w io.Writer) (err error) {
	c.FixedHeader = &FixedHeader{PacketType: CONNACK, Flags: FixedHeaderFlagReserved}
	buf := &bytes.Buffer{}
	if c.SessionPresent {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}
	// Connect Return code
	buf.WriteByte(c.Code)

	return encode(c.FixedHeader, buf, w)
}

// Decode 解码r中可变报头
func (c *Connack) Decode(r io.Reader) (err error) {
	restBuffer := make([]byte, c.FixedHeader.RemainLength)
	_, err = io.ReadFull(r, restBuffer)
	if err != nil {
		return xerror.ErrMalformed
	}
	buf := bytes.NewBuffer(restBuffer)
	// 当前会话
	sessionPresentByte, err := buf.ReadByte()
	if (127 & (sessionPresentByte >> 1)) > 0 {
		return xerror.ErrMalformed
	}
	c.SessionPresent = sessionPresentByte == 1
	// 连接返回码
	codeByte, err := buf.ReadByte()
	if err != nil {
		return xerror.ErrMalformed
	}
	c.Code = codeByte
	return
}

func (c *Connack) String() string {
	return fmt.Sprintf("Connack - Version:%s, SessionPresent:%v, Code:%v",
		c.Version, c.SessionPresent, c.Code)
}
