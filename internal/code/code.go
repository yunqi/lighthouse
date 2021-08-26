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

package code

// There are the possible Code in v311 connack packet.
// http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.html#_Table_3.1_-
const (
	V3Accepted                    Code = 0x00
	V3UnacceptableProtocolVersion Code = 0x01
	V3IdentifierRejected          Code = 0x02
	V3ServerUnavaliable           Code = 0x03
	V3BadUsernameorPassword       Code = 0x04
	V3NotAuthorized               Code = 0x05
)

// There are the possible reason Code in v5
const (
	Success                     Code = 0x00
	NormalDisconnection         Code = 0x00
	GrantedQoS0                 Code = 0x00
	GrantedQoS1                 Code = 0x01
	GrantedQoS2                 Code = 0x02
	DisconnectWithWillMessage   Code = 0x04
	NotMatchingSubscribers      Code = 0x10
	NoSubscriptionExisted       Code = 0x11
	ContinueAuthentication      Code = 0x18
	ReAuthenticate              Code = 0x19
	UnspecifiedError            Code = 0x80
	MalformedPacket             Code = 0x81
	ProtocolError               Code = 0x82
	ImplementationSpecificError Code = 0x83
	UnsupportedProtocolVersion  Code = 0x84
	ClientIdentifierNotValid    Code = 0x85
	BadUserNameOrPassword       Code = 0x86
	NotAuthorized               Code = 0x87
	ServerUnavailable           Code = 0x88
	ServerBusy                  Code = 0x89
	Banned                      Code = 0x8A
	BadAuthMethod               Code = 0x8C
	KeepAliveTimeout            Code = 0x8D
	SessionTakenOver            Code = 0x8E
	TopicFilterInvalid          Code = 0x8F
	TopicNameInvalid            Code = 0x90
	PacketIDInUse               Code = 0x91
	PacketIDNotFound            Code = 0x92
	RecvMaxExceeded             Code = 0x93
	TopicAliasInvalid           Code = 0x94
	PacketTooLarge              Code = 0x95
	MessageRateTooHigh          Code = 0x96
	QuotaExceeded               Code = 0x97
	AdminAction                 Code = 0x98
	PayloadFormatInvalid        Code = 0x99
	RetainNotSupported          Code = 0x9A
	QoSNotSupported             Code = 0x9B
	UseAnotherServer            Code = 0x9C
	ServerMoved                 Code = 0x9D
	SharedSubNotSupported       Code = 0x9E
	ConnectionRateExceeded      Code = 0x9F
	MaxConnectTime              Code = 0xA0
	SubIDNotSupported           Code = 0xA1
	WildcardSubNotSupported     Code = 0xA2
)

type (
	Code = byte
)
