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

package xbitmap

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBitmap(t *testing.T) {

	size := MaxSize
	b := New(size)
	assert.Equal(t, size, b.Size())

	b.Set(1, 1)
	assert.EqualValues(t, 1, b.Get(1))

	b.Set(1, 0)
	assert.EqualValues(t, 0, b.Get(100))

	b.Set(size, 1)
	assert.EqualValues(t, 1, b.Get(size))

	b.Set(size, 0)
	assert.EqualValues(t, 0, b.Get(size))

	b.Set(MaxSize, 1)
	b.Get(MaxSize)
	assert.EqualValues(t, 1, b.Get(MaxSize))

}
