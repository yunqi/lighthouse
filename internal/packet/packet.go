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
	"encoding/binary"
	"github.com/yunqi/lighthouse/internal/xerror"
	"io"
	"unicode/utf8"
)

const (
	Version31  Version = 0x03
	Version311 Version = 0x04
	Version5   Version = 0x05
	// MaximumSize The maximum packet size of a MQTT packet
	MaximumSize = 268435456

	// SubscribeFailure 订阅失败
	SubscribeFailure = 0x80

	MaxPacketID Id = 65535
	MinPacketID Id = 1
	// UTF8EncodedStringsMaxLen There is a limit on the size of a string that can be passed in one of these UTF-8 encoded string components; you cannot use a string that would encode to more than 65535 bytes.
	// http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.html#_Table_2.2_-
	// 2 byte = uint16
	UTF8EncodedStringsMaxLen = 1<<16 - 1
	// http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.html#_Table_2.4_Size
	RemainLength1ByteMin = 0
	RemainLength1ByteMax = 1<<7 - 1
	RemainLength2ByteMin = 1 << 7
	RemainLength2ByteMax = 1<<14 - 1
	RemainLength3ByteMin = 1 << 14
	RemainLength3ByteMax = 1<<21 - 1
	RemainLength4ByteMin = 1 << 21
	RemainLength4ByteMax = 1<<28 - 1
)

//Packet type
const (
	// RESERVED Forbidden
	RESERVED Type = iota
	// CONNECT Client request to connect to Server
	CONNECT
	// CONNACK Connect acknowledgment
	CONNACK
	// PUBLISH message
	PUBLISH
	// PUBACK Publish acknowledgment
	PUBACK
	// PUBREC Publish received (assured delivery part 1)
	PUBREC
	// PUBREL Publish release (assured delivery part 2)
	PUBREL
	// PUBCOMP Publish complete (assured delivery part 3)
	PUBCOMP
	// SUBSCRIBE Client subscribe request
	SUBSCRIBE
	// SUBACK Subscribe acknowledgment
	SUBACK
	// UNSUBSCRIBE Unsubscribe request
	UNSUBSCRIBE
	// UNSUBACK Unsubscribe acknowledgment
	UNSUBACK
	// PINGREQ PING request
	PINGREQ
	// PINGRESP PING response
	PINGRESP
	// DISCONNECT Client is disconnecting
	DISCONNECT
	// AUTHReserved Forbidden
	AUTHReserved

	// Flag in the FixHeader

	FixedHeaderFlagReserved    = 0
	FixedHeaderFlagSubscribe   = 2
	FixedHeaderFlagUnsubscribe = 2
	FixedHeaderFlagPubrel      = 2
)

// QoS levels
const (
	QoS0 QoS = 0x00
	QoS1 QoS = 0x01
	QoS2 QoS = 0x02
)
const (
	PropPayloadFormat          byte = 1
	PropMessageExpiry          byte = 2
	PropContentType            byte = 3
	PropResponseTopic          byte = 8
	PropCorrelationData        byte = 9
	PropSubscriptionIdentifier byte = 11
	PropSessionExpiryInterval  byte = 17
	PropAssignedClientID       byte = 18
	PropServerKeepAlive        byte = 19
	PropAuthMethod             byte = 21
	PropAuthData               byte = 22
	PropRequestProblemInfo     byte = 23
	PropWillDelayInterval      byte = 24
	PropRequestResponseInfo    byte = 25
	PropResponseInfo           byte = 26
	PropServerReference        byte = 28
	PropReasonString           byte = 31
	PropReceiveMaximum         byte = 33
	PropTopicAliasMaximum      byte = 34
	PropTopicAlias             byte = 35
	PropMaximumQOS             byte = 36
	PropRetainAvailable        byte = 37
	PropUser                   byte = 38
	PropMaximumPacketSize      byte = 39
	PropWildcardSubAvailable   byte = 40
	PropSubIDAvailable         byte = 41
	PropSharedSubAvailable     byte = 42
)

var version2protocolName = map[Version]string{
	Version31:  "MQIsdp", // 'M', 'Q', 'I', 's', 'd', 'p'
	Version311: "MQTT",   // 'M', 'Q', 'T', 'T'
	Version5:   "MQTT",   // 'M', 'Q', 'T', 'T'
}
var version2versionName = map[Version]string{
	Version31:  "MQTT3.1",   // 'M', 'Q', 'I', 's', 'd', 'p'
	Version311: "MQTT3.1.1", // 'M', 'Q', 'T', 'T'
	Version5:   "MQTT5",     // 'M', 'Q', 'T', 'T'
}

type PayloadFormat = byte

const (
	PayloadFormatBytes  PayloadFormat = 0
	PayloadFormatString PayloadFormat = 1
)

type (
	// Version MQTT版本
	Version byte
	// QoS 消息质量
	QoS = byte
	// Id  数据包ID
	Id = uint16

	Type = byte

	// Packet defines the interface for structs intended to hold
	// decoded MQTT packets, either from being read or before being
	// written
	Packet interface {
		// Encode encodes the packet struct into bytes and writes it into io.Writer.
		Encode(w io.Writer) (err error)
		// Decode read the packet bytes from io.Reader and decodes it into the packet struct
		Decode(r io.Reader) (err error)
		// String is mainly used in logging, debugging and testing.
		String() string
	}
)

func (v Version) String() string {
	if s, ok := version2versionName[v]; ok {
		return s
	}
	return "Unknown version"
}

func IsVersion3(version Version) bool {
	return version == Version31 || version == Version311
}

func IsVersion5(version Version) bool {
	return version == Version5
}
func writeUint16(w *bytes.Buffer, value uint16) {
	w.WriteByte(byte(value >> 8))
	w.WriteByte(byte(value))
}
func readUint16(r *bytes.Buffer) (uint16, error) {
	if r.Len() < 2 {
		return 0, xerror.ErrMalformed
	}
	return binary.BigEndian.Uint16(r.Next(2)), nil
}
func writeBinary(w *bytes.Buffer, b []byte) {
	writeUint16(w, uint16(len(b)))
	w.Write(b)
}

// UTF8EncodedStrings encodes the bytes into UTF-8 encoded strings, returns the encoded bytes, bytes size and error.
func UTF8EncodedStrings(data []byte) (b []byte, size int, err error) {
	dataSize := len(data)
	if dataSize > UTF8EncodedStringsMaxLen {
		return nil, 0, xerror.ErrMalformed
	}
	// http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.html#_Figure_1.1_Structure
	bufLen := dataSize + 2
	buf := make([]byte, bufLen)
	binary.BigEndian.PutUint16(buf, uint16(dataSize))
	copy(buf[2:], data)
	return buf, bufLen, nil
}
func UTF8DecodedStrings(mustUTF8 bool, r *bytes.Buffer) (b []byte, err error) {
	if r.Len() < 2 {
		return nil, xerror.ErrMalformed
	}
	// strings for length
	length := int(binary.BigEndian.Uint16(r.Next(2)))
	if r.Len() < length {
		return nil, xerror.ErrMalformed
	}
	// strings
	b = r.Next(length)
	if mustUTF8 {
		if !ValidUTF8(b) {
			return nil, xerror.ErrMalformed
		}
	}
	return b, nil
}

// ValidUTF8 returns whether the given bytes is in UTF-8 form.
func ValidUTF8(b []byte) bool {
	for {
		if len(b) == 0 {
			return true
		}
		ru, size := utf8.DecodeRune(b)

		// The data SHOULD NOT include encodings of the Unicode [Unicode] code points listed below.
		// If a receiver (Server or Client) receives a Control Packet containing any of them
		// it MAY close the Network Connection:
		//
		//    	U+0001..U+001F control characters
		//		U+007F..U+009F control characters
		if ru >= '\u0000' && ru <= '\u001f' { //[MQTT-1.5.3-2]
			return false
		}
		if ru >= '\u007f' && ru <= '\u009f' {
			return false
		}
		if ru == utf8.RuneError {
			return false
		}
		if !utf8.ValidRune(ru) {
			return false
		}
		if size == 0 {
			return true
		}
		b = b[size:]
	}
}

//EncodeRemainLength puts the length int into bytes
func EncodeRemainLength(length int) (result []byte, err error) {
	if length <= RemainLength1ByteMax {
		result = make([]byte, 1)
	} else if length <= RemainLength2ByteMax {
		result = make([]byte, 2)
	} else if length <= RemainLength3ByteMax {
		result = make([]byte, 3)
	} else if length <= RemainLength4ByteMax {
		result = make([]byte, 4)
	} else {
		return nil, xerror.ErrMalformed
	}
	// -------------------------------------------------
	//         	do
	//              encodedByte = X MOD 128
	//
	//              X = X DIV 128
	//
	//             // if there are more data to encode, set the top bit of this byte
	//
	//             if ( X > 0 )
	//
	//                 encodedByte = encodedByte OR 128
	//
	//             endif
	//
	//                 'output' encodedByte
	//
	//        while ( X > 0 )
	// -------------------------------------------------
	var index int
	for {
		encodedByte := length % 128
		length = length / 128
		// if there are more data to encode, set the top bit of this byte
		if length > 0 {
			encodedByte = encodedByte | 128
		}
		result[index] = byte(encodedByte)
		index++
		if !(length > 0) {
			break
		}
	}
	return result, nil
}

// DecodeRemainLength reads the remain length bytes from bufio.Reader and returns length int.
func DecodeRemainLength(r io.ByteReader) (int, error) {
	var multiplier uint32 = 1
	var value uint32
	for {
		encodedByte, err := r.ReadByte()
		if err != nil && err != io.EOF {
			return 0, err
		}
		value += uint32(encodedByte&127) * multiplier
		multiplier *= 128
		if multiplier > 128*128*128 {
			return 0, xerror.ErrMalformed
		}
		if (encodedByte & 128) == 0 {
			break
		}
	}
	return int(value), nil
}

func NewPacket(fixedHeader *FixedHeader, version Version, r io.Reader) (Packet, error) {
	switch fixedHeader.PacketType {
	case CONNECT:
		return NewConnect(fixedHeader, version, r)
	case CONNACK:
		return NewConnack(fixedHeader, version, r)
	case PUBLISH:
		return NewPublish(fixedHeader, version, r)
	case PUBACK:
		return NewPuback(fixedHeader, version, r)
	case PUBREC:
		return NewPubrec(fixedHeader, version, r)
	case PUBREL:
		return NewPubrel(fixedHeader, version, r)
	case PUBCOMP:
		return NewPubcomp(fixedHeader, version, r)
	case SUBSCRIBE:
		return NewSubscribe(fixedHeader, version, r)
	case SUBACK:
		return NewSuback(fixedHeader, version, r)
	case UNSUBSCRIBE:
		return NewUnsubscribe(fixedHeader, version, r)
	case PINGREQ:
		return NewPingreq(fixedHeader, r)
	case DISCONNECT:
		return NewDisconnect(fixedHeader, version, r)
	case UNSUBACK:
		return NewSuback(fixedHeader, version, r)
	case PINGRESP:
		return NewPingresp(fixedHeader, r)
	//case AUTH:
	default:
		return nil, xerror.ErrProtocol
	}
}

// encode 编码
func encode(fixedHeader *FixedHeader, readBuf *bytes.Buffer, w io.Writer) (err error) {
	fixedHeader.RemainLength = readBuf.Len()
	err = fixedHeader.Encode(w)
	if err != nil {
		return
	}
	_, err = readBuf.WriteTo(w)
	return err
}

// ValidV5Topic returns whether the given bytes is a valid MQTT V5 topic
func ValidV5Topic(p []byte) bool {
	if len(p) == 0 {
		return false
	}
	if bytes.HasPrefix(p, []byte("$share/")) {
		if len(p) < 9 {
			return false
		}
		if p[7] != '/' {
			subp := p[7:]
			for len(subp) > 0 {
				ru, size := utf8.DecodeRune(subp)
				if ru == utf8.RuneError {
					return false
				}
				if size == 1 {
					if subp[0] == '/' {
						return ValidTopicFilter(true, subp[1:])
					}
					if subp[0] == byte('+') || subp[0] == byte('#') {
						return false
					}
				}
				subp = subp[size:]
			}

		}
		return false
	}
	return ValidTopicFilter(true, p)
}
