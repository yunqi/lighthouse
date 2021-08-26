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

const (
	DupTure    byte = 1 << 3
	QoS3       byte = 1 << 3
	RetainTure byte = 1
)

type (
	Publish struct {
		Version     Version
		FixedHeader *FixedHeader
		Dup         bool   //是否重发 [MQTT-3.3.1.-1]
		QoS         uint8  //qos等级
		Retain      bool   //是否保留消息
		TopicName   []byte //主题名
		PacketId           //报文标识符
		Payload     []byte
	}
)

// NewPublish returns a Publishing instance by the given FixHeader and io.Reader.
func NewPublish(fixedHeader *FixedHeader, version Version, r io.Reader) (*Publish, error) {
	p := &Publish{FixedHeader: fixedHeader, Version: version}
	p.Dup = (1 & (fixedHeader.Flags >> 3)) > 0
	p.QoS = (fixedHeader.Flags >> 1) & 3
	if p.QoS == 0 && p.Dup { //[MQTT-3.3.1-2]、 [MQTT-4.3.1-1]
		return nil, xerror.ErrMalformed
	}
	if p.QoS > QoS2 {
		return nil, xerror.ErrMalformed
	}
	if fixedHeader.Flags&1 == 1 { //保留标志
		p.Retain = true
	}
	err := p.Decode(r)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Publish) Encode(w io.Writer) (err error) {
	p.FixedHeader = &FixedHeader{PacketType: PUBLISH}
	buf := &bytes.Buffer{}
	var dup, retain byte
	if p.Dup {
		dup = DupTure
	}
	if p.Retain {
		retain = RetainTure
	}
	p.FixedHeader.Flags = dup | retain | (p.QoS << 1)
	// variable header
	writeBinary(buf, p.TopicName)
	if p.QoS == QoS1 || p.QoS == QoS2 {
		writeUint16(buf, p.PacketId)
	}
	buf.Write(p.Payload)

	// 写入
	return encode(p.FixedHeader, buf, w)
}

func (p *Publish) Decode(r io.Reader) (err error) {
	restBuffer := make([]byte, p.FixedHeader.RemainLength)
	_, err = io.ReadFull(r, restBuffer)
	if err != nil {
		return xerror.ErrMalformed
	}
	buf := bytes.NewBuffer(restBuffer)
	p.TopicName, err = UTF8DecodedStrings(true, buf)
	if err != nil {
		return
	}
	if !ValidTopicName(true, p.TopicName) {
		return xerror.ErrMalformed
	}

	if p.QoS > QoS0 {
		// The Packet Identifier field is only present in PUBLISH Packets where the QoS level is 1 or 2.
		p.PacketId, err = readUint16(buf)
		if err != nil {
			return
		}
	}
	p.Payload = buf.Next(buf.Len())
	return nil
}

func (p *Publish) String() string {
	b, _ := json.Marshal(p)
	return string(b)
}
