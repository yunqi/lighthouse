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
	"testing"
)

func TestNewConnect(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		fixedHeader := &FixedHeader{
			PacketType:   CONNECT,
			Flags:        FixedHeaderFlagReserved,
			RemainLength: 10,
		}
		connectBytes := bytes.NewBuffer([]byte{1, 2})
		// TODO
		NewConnect(fixedHeader, Version311, connectBytes)
	})
}
