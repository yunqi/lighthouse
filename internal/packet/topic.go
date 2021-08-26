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

import "unicode/utf8"

type (
	// Topic represents the MQTT Topic
	Topic struct {
		SubOptions
		Name string
	}
	// SubOptions is the subscription option of subscriptions.
	// For details: https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Subscription_Options
	SubOptions struct {
		// QoS is the QoS level of the subscription.
		// 0 = At most once delivery
		// 1 = At least once delivery
		// 2 = Exactly once delivery
		QoS QoS
		// RetainHandling specifies whether retained messages are sent when the subscription is established.
		// 0 = Send retained messages at the time of the subscribe
		// 1 = Send retained messages at subscribe only if the subscription does not currently exist
		// 2 = Do not send retained messages at the time of the subscribe
		RetainHandling byte
		// NoLocal is the No Local option.
		//  If the value is 1, Application Messages MUST NOT be forwarded to a connection with a ClientId equal to the ClientId of the publishing connection
		NoLocal bool
		// RetainAsPublished is the Retain As Published option.
		// If 1, Application Messages forwarded using this subscription keep the RETAIN flag they were published with.
		// If 0, Application Messages forwarded using this subscription have the RETAIN flag set to 0. Retained messages sent when the subscription is established have the RETAIN flag set to 1.
		RetainAsPublished bool
	}
)

// ValidTopicFilter  returns whether the bytes is a valid topic filter. [MQTT-4.7.1-2]  [MQTT-4.7.1-3]
func ValidTopicFilter(mustUTF8 bool, p []byte) bool {
	if len(p) == 0 {
		return false
	}
	var prevByte byte //前一个字节
	var isSetPrevByte bool

	for len(p) > 0 {
		ru, size := utf8.DecodeRune(p)
		if mustUTF8 && ru == utf8.RuneError {
			return false
		}
		plen := len(p)
		if p[0] == byte('#') && plen != 1 { // #一定是最后一个字符
			return false
		}
		if size == 1 && isSetPrevByte {
			// + 前（如果有前后字节）,一定是'/' [MQTT-4.7.1-2]  [MQTT-4.7.1-3]
			if (p[0] == byte('+') || p[0] == byte('#')) && prevByte != byte('/') {
				return false
			}

			if plen > 1 { // p[0] 不是最后一个字节
				if p[0] == byte('+') && p[1] != byte('/') { // + 后（如果有字节）,一定是 '/'
					return false
				}
			}
		}
		prevByte = p[0]
		isSetPrevByte = true
		p = p[size:]
	}
	return true
}
