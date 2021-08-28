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

package xerror

import (
	"fmt"
	"github.com/yunqi/lighthouse/internal/code"
)

var (
	ErrMalformed                     = NewError(code.MalformedPacket)
	ErrProtocol                      = NewError(code.ProtocolError)
	ErrV3UnacceptableProtocolVersion = NewError(code.V3UnacceptableProtocolVersion)
	ErrV3IdentifierRejected          = NewError(code.V3IdentifierRejected)
)

type (
	Error struct {
		// Code is the MQTT Reason Code
		Code code.Code
		ErrorDetails
	}
	// ErrorDetails wraps reason string and user property for diagnostics.
	ErrorDetails struct {
		// ReasonString is the reason string field in property.
		// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901029
		ReasonString []byte
		// UserProperties is the user property field in property.
		UserProperties []struct {
			Key   []byte
			Value []byte
		}
	}
)

func NewError(code code.Code) *Error {
	return &Error{Code: code}
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("operation error: Code = %x, reasonString: %s", e.Code, e.ReasonString)
}
