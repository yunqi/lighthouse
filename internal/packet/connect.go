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

type (
	// Connect represents the MQTT Connect packet.
	Connect struct {
		Version     Version
		FixedHeader *FixedHeader

		ProtocolName  []byte
		ProtocolLevel byte
		// The ConnectFlags byte contains a number of parameters specifying the behavior of the MQTT connection.
		// It also indicates the presence or absence of fields in the payload.
		ConnectFlags *ConnectFlags
		// The KeepAlive is a time interval measured in seconds.
		// Expressed as a 16-bit word, it is the maximum time interval that is permitted
		// to elapse between the point at which the Client finishes transmitting one Control Packet
		// and the point it starts sending the next.
		KeepAlive uint16

		WillTopic []byte
		WillMsg   []byte

		//auth
		ClientId []byte
		Username []byte
		Password []byte
	}
	ConnectFlags struct {

		// CleanSession: bit 1 of the ConnectFlags byte.
		// This bit specifies the handling of the Session state.
		CleanSession bool
		// WillFlag: bit 2 of the ConnectFlags.
		WillFlag bool
		// WillQoS bits 4 and 3 of the ConnectFlags.
		// These two bits specify the QoS level to be used when publishing the Will Message.
		WillQoS byte
		// WillRetain:bit 5 of the ConnectFlags.
		WillRetain bool
		// PasswordFlag:bit 7 of the ConnectFlags.
		PasswordFlag bool
		// UsernameFlag
		UsernameFlag bool
	}
)

// NewConnect returns a Connect instance by the given FixHeader and io.Reader
func NewConnect(fixedHeader *FixedHeader, version Version, r io.Reader) (*Connect, error) {
	//b1 := buffer[0] //一定是16
	p := &Connect{FixedHeader: fixedHeader, Version: version}
	//判断 标志位 flags 是否合法[MQTT-2.2.2-2]
	if fixedHeader.Flags != FixedHeaderFlagReserved {
		return nil, xerror.ErrMalformed
	}
	err := p.Decode(r)
	if err != nil {
		return nil, err
	}
	return p, err
}

var (
	ProtocolNamePrefix = []byte{0x00, 0x04}
)

const (
	_ = 1 << iota
	CleanSessionTure
	willFlagTure
	willQos1
	WillQos2
	willRetainTrue
	passwordFlagTrue
	usernameFlagTrue
)

func (c *Connect) Encode(w io.Writer) (err error) {
	c.FixedHeader = &FixedHeader{PacketType: CONNECT, Flags: FixedHeaderFlagReserved}
	buf := &bytes.Buffer{}
	// 协议头
	buf.Write(ProtocolNamePrefix)
	buf.Write(c.ProtocolName[:])
	buf.WriteByte(c.ProtocolLevel)
	// connect flags
	var (
		usernameFlag byte = 0
		passwordFlag byte = 0
		willRetain   byte = 0
		willFlag     byte = 0
		willQos      byte = 0
		CleanSession byte = 0
		reserved     byte = 0
	)
	if c.ConnectFlags.UsernameFlag {
		usernameFlag = usernameFlagTrue
	}
	if c.ConnectFlags.PasswordFlag {
		passwordFlag = passwordFlagTrue
	}
	if c.ConnectFlags.WillRetain {
		willRetain = willRetainTrue
	}
	if c.ConnectFlags.WillQoS == 1 {
		willQos = willQos1
	} else if c.ConnectFlags.WillQoS == 2 {
		willQos = WillQos2
	}
	if c.ConnectFlags.WillFlag {
		willFlag = willFlagTure
	}
	if c.ConnectFlags.CleanSession {
		CleanSession = CleanSessionTure
	}
	connectFlags := usernameFlag | passwordFlag | willRetain | willFlag | willQos | CleanSession | reserved
	buf.Write([]byte{connectFlags})
	writeUint16(buf, c.KeepAlive)

	// client identifier
	clientIdBytes, _, err := UTF8EncodedStrings(c.ClientId)
	if err != nil {
		return err
	}
	buf.Write(clientIdBytes)
	if c.ConnectFlags.WillFlag {
		// will topic
		willTopicBytes, _, err := UTF8EncodedStrings(c.WillTopic)
		if err != nil {
			return err
		}
		buf.Write(willTopicBytes)

		// Will Message
		willMsgBytes, _, err := UTF8EncodedStrings(c.WillMsg)
		if err != nil {
			return err
		}
		buf.Write(willMsgBytes)
	}
	if c.ConnectFlags.UsernameFlag {
		usernameBytes, _, err := UTF8EncodedStrings(c.Username)
		if err != nil {
			return err
		}
		buf.Write(usernameBytes)
	}
	if c.ConnectFlags.PasswordFlag {
		passwordBytes, _, err := UTF8EncodedStrings(c.Password)
		if err != nil {
			return err
		}
		buf.Write(passwordBytes)
	}
	return encode(c.FixedHeader, buf, w)
}

// Decode 解码可变报头的长度（10字节）加上有效载荷
func (c *Connect) Decode(r io.Reader) (err error) {
	restBuffer := make([]byte, c.FixedHeader.RemainLength)
	_, err = io.ReadFull(r, restBuffer)
	if err != nil {
		return
	}
	buf := bytes.NewBuffer(restBuffer)
	protocolName, err := UTF8DecodedStrings(true, buf)
	if err != nil {
		return err
	}

	c.ProtocolName = protocolName

	c.ProtocolLevel, err = buf.ReadByte()
	if err != nil {
		return xerror.ErrMalformed
	}
	c.Version = Version(c.ProtocolLevel)
	if _, ok := version2protocolName[c.Version]; !ok {
		return xerror.NewError(code.V3UnacceptableProtocolVersion)
	}
	connectFlags, err := buf.ReadByte()
	if err != nil {
		return xerror.ErrMalformed
	}
	reserved := 1 & connectFlags
	if reserved != 0 { //[MQTT-3.1.2-3]
		return xerror.ErrMalformed
	}
	c.ConnectFlags = &ConnectFlags{}
	c.ConnectFlags.CleanSession = (1 & (connectFlags >> 1)) > 0
	c.ConnectFlags.WillFlag = (1 & (connectFlags >> 2)) > 0
	c.ConnectFlags.WillQoS = 3 & (connectFlags >> 3)
	if !c.ConnectFlags.WillFlag && c.ConnectFlags.WillQoS != 0 { //[MQTT-3.1.2-11]
		return xerror.ErrMalformed
	}
	c.ConnectFlags.WillRetain = (1 & (connectFlags >> 5)) > 0
	if !c.ConnectFlags.WillFlag && c.ConnectFlags.WillRetain { //[MQTT-3.1.2-11]
		return xerror.ErrMalformed
	}
	c.ConnectFlags.PasswordFlag = (1 & (connectFlags >> 6)) > 0
	c.ConnectFlags.UsernameFlag = (1 & (connectFlags >> 7)) > 0
	c.KeepAlive, err = readUint16(buf)
	if err != nil {
		return err
	}
	return c.unpackPayload(buf)
}

func (c *Connect) String() string {
	return fmt.Sprintf(
		"Connect - Version: %s,ProtocolLevel: %v, UsernameFlag: %v, PasswordFlag: %v, ProtocolName: %s, CleanSession: %v, KeepAlive: %v, ClientId: %s, Username: %s, Password: %s, WillFlag: %v, WillRetain: %v, WillQos: %v, WillTopic: %s, WillMsg: %s",
		c.Version, c.ProtocolLevel, c.ConnectFlags.UsernameFlag, c.ConnectFlags.PasswordFlag, c.ProtocolName, c.ConnectFlags.CleanSession, c.KeepAlive, c.ClientId, c.Username, c.Password, c.ConnectFlags.WillFlag, c.ConnectFlags.WillRetain, c.ConnectFlags.WillQoS, c.WillTopic, c.WillMsg)
}

func (c *Connect) unpackPayload(buf *bytes.Buffer) error {
	var err error
	c.ClientId, err = UTF8DecodedStrings(true, buf)
	if err != nil {
		return err
	}

	if IsVersion3(c.Version) && len(c.ClientId) == 0 && !c.ConnectFlags.CleanSession { // v311 [MQTT-3.1.3-7]
		return xerror.NewError(code.V3IdentifierRejected) // v311 //[MQTT-3.1.3-8]
	}
	if c.ConnectFlags.WillFlag {
		c.WillTopic, err = UTF8DecodedStrings(true, buf)
		if err != nil {
			return err
		}
		c.WillMsg, err = UTF8DecodedStrings(true, buf)
		if err != nil {
			return err
		}
	}

	if c.ConnectFlags.UsernameFlag {
		c.Username, err = UTF8DecodedStrings(true, buf)
		if err != nil {
			return err
		}
	}
	if c.ConnectFlags.PasswordFlag {
		c.Password, err = UTF8DecodedStrings(true, buf)
		if err != nil {
			return err
		}
	}
	return nil
}
