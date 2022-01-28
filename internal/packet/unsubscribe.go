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
	Unsubscribe struct {
		Version     Version
		FixedHeader *FixedHeader
		PacketId    Id
		Topics      []string
	}
)

// NewUnsubscribe returns a Unsubscribe instance by the given FixHeader and io.Reader.
func NewUnsubscribe(fixedHeader *FixedHeader, version Version, r io.Reader) (*Unsubscribe, error) {
	p := &Unsubscribe{FixedHeader: fixedHeader, Version: version}
	//判断 标志位 flags 是否合法[MQTT-3.10.1-1]
	if fixedHeader.Flags != FixedHeaderFlagUnsubscribe {
		return nil, xerror.ErrMalformed
	}
	err := p.Decode(r)
	if err != nil {
		return nil, err
	}
	return p, err
}

func (u *Unsubscribe) Encode(w io.Writer) (err error) {
	u.FixedHeader = &FixedHeader{PacketType: UNSUBSCRIBE, Flags: FixedHeaderFlagUnsubscribe}
	buf := &bytes.Buffer{}
	writeUint16(buf, u.PacketId)
	for _, topic := range u.Topics {
		writeBinary(buf, []byte(topic))
	}
	return encode(u.FixedHeader, buf, w)
}

func (u *Unsubscribe) Decode(r io.Reader) (err error) {
	b := make([]byte, u.FixedHeader.RemainLength)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return xerror.ErrMalformed
	}
	bufr := bytes.NewBuffer(b)
	u.PacketId, err = readUint16(bufr)
	if err != nil {
		return
	}
	// topics
	for bufr.Len() != 0 {
		topicFilter, err := UTF8DecodedStrings(true, bufr)
		if err != nil {
			return err
		}
		if !ValidTopicFilter(true, topicFilter) {
			return xerror.ErrProtocol
		}
		u.Topics = append(u.Topics, string(topicFilter))
	}
	return
}

func (u *Unsubscribe) String() string {
	return fmt.Sprintf("Unsubscribe - Version: %s, PacketId: %d, Topics: %v", u.Version, u.PacketId, u.Topics)
}
