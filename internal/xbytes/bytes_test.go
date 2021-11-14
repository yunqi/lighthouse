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
	"fmt"
	"github.com/stretchr/testify/assert"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
)

func TestGetNBytePool(t *testing.T) {
	m := sync.Map{}
	n := 100
	waitGroup := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		waitGroup.Add(1)
		go func(j int) {
			defer waitGroup.Done()
			k := j % 10
			t.Run(fmt.Sprintf("%d bytes", k), func(t *testing.T) {
				pool := GetNBytePool(k)
				value, _ := m.LoadOrStore(k, atomic.Value{})
				v := value.(atomic.Value)
				swap := v.CompareAndSwap(nil, pool)
				b := pool.Get().([]byte)
				assert.EqualValues(t, len(b), k)
				if !swap {
					assert.EqualValues(t, v.Load(), pool)
				}
			})

		}(i)
	}
	waitGroup.Wait()
}

func BenchmarkGetNBytePool(b *testing.B) {
	b.ReportAllocs()
	n := 20

	for i := 0; i < n; i++ {
		b.Run(strconv.Itoa(2<<i), func(b *testing.B) {
			b.ReportAllocs()
			pool := GetNBytePool(2 << i)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					waitGroup := sync.WaitGroup{}
					for j := 0; j < 100; j++ {
						waitGroup.Add(1)
						go func() {
							defer waitGroup.Done()
							a := pool.Get().([]byte)
							pool.Put(a)
						}()
						waitGroup.Wait()
					}

				}
			})
		})
	}
}
func BenchmarkNewBytes(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	n := 20
	for i := 0; i < n; i++ {
		b.Run(strconv.Itoa(2<<i), func(b *testing.B) {
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					waitGroup := sync.WaitGroup{}
					for j := 0; j < 100; j++ {
						waitGroup.Add(1)
						go func() {
							defer waitGroup.Done()
							a := make([]byte, 2<<i)
							_ = a
						}()
						waitGroup.Wait()
					}

				}
			})
		})
	}

}
