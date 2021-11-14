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

package xbytes

import (
	"golang.org/x/sync/singleflight"
	"strconv"
	"sync"
)

var (
	bytesPoolMap = &sync.Map{}
	singleFlight = &singleflight.Group{}
)

// GetNBytePool returns a buffer sync.Pool.
// It is recommended to use n Byte greater than or equal to 64.
func GetNBytePool(nByte int) *sync.Pool {
	byteSizeStr := strconv.Itoa(nByte)
	pool, _, _ := singleFlight.Do(byteSizeStr, func() (interface{}, error) {
		if val, ok := bytesPoolMap.Load(byteSizeStr); ok {
			return val, nil
		}
		pool := &sync.Pool{New: func() interface{} {
			return make([]byte, nByte)
		}}
		bytesPoolMap.Store(byteSizeStr, pool)
		return pool, nil
	})
	return pool.(*sync.Pool)
}
