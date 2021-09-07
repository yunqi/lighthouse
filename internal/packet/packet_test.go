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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUTF8EncodedStrings(t *testing.T) {
	b := []byte("123")
	data, size, err := UTF8EncodedStrings(b)
	assert.Nil(t, err)
	assert.Equal(t, len(b)+2, size)
	assert.Equal(t, append([]byte{0x00, 0x03}, b...), data)
}

func TestNewPacket(t *testing.T) {

}
