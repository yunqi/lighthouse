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
	"github.com/chenquan/lighthouse/internal/xerror"
	"io"
)

type (
	Subscribe struct {
		Version     Version
		FixedHeader *FixedHeader
		PacketId    PacketId
		Topics      []*Topic //suback响应之前填充
	}
)

// NewSubscribe returns a Subscribe instance by the given FixHeader and io.Reader.
func NewSubscribe(fh *FixedHeader, version Version, r io.Reader) (*Subscribe, error) {
	p := &Subscribe{FixedHeader: fh, Version: version}
	//判断 标志位 flags 是否合法[MQTT-3.8.1-1]
	if fh.Flags != FixedHeaderFlagSubscribe {
		return nil, xerror.ErrMalformed
	}
	err := p.Decode(r)
	if err != nil {
		return nil, err
	}
	return p, err
}
func (s *Subscribe) Encode(w io.Writer) (err error) {
	s.FixedHeader = &FixedHeader{PacketType: SUBSCRIBE, Flags: FixedHeaderFlagSubscribe}
	buf := &bytes.Buffer{}
	writeUint16(buf, s.PacketId)

	// payload
	for _, t := range s.Topics {
		writeBinary(buf, []byte(t.Name))
		buf.WriteByte(t.QoS)
	}
	return encode(s.FixedHeader, buf, w)
}

func (s *Subscribe) Decode(r io.Reader) (err error) {
	b := make([]byte, s.FixedHeader.RemainLength)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return xerror.ErrMalformed
	}
	bufr := bytes.NewBuffer(b)

	s.PacketId, err = readUint16(bufr)
	if err != nil {
		return err
	}
	// topics
	for bufr.Len() != 0 {
		topicFilter, err := UTF8DecodedStrings(true, bufr)
		if err != nil {
			return err
		}
		if !ValidTopicFilter(true, topicFilter) {
			return xerror.ErrMalformed
		}
		topicOpts, err := bufr.ReadByte()
		if err != nil {
			return xerror.ErrMalformed
		}
		topic := &Topic{
			Name: string(topicFilter),
		}
		topic.QoS = topicOpts
		if topic.QoS > QoS2 {
			return xerror.ErrProtocol
		}
		s.Topics = append(s.Topics, topic)

	}
	return
}

func (s *Subscribe) String() string {
	b, _ := json.Marshal(s)
	return string(b)
}
