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

package xio

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
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

//func BenchmarkGetNBytePool(b *testing.B) {
//	b.ReportAllocs()
//	n := 20
//
//	for i := 0; i < n; i++ {
//		b.Run(strconv.Itoa(2<<i), func(b *testing.B) {
//			b.ReportAllocs()
//			pool := GetNBytePool(2 << i)
//			b.RunParallel(func(pb *testing.PB) {
//				for pb.Next() {
//					a := pool.Get().([]byte)
//					pool.Put(a)
//				}
//			})
//		})
//	}
//}
//func BenchmarkNewBytes(b *testing.B) {
//	b.ReportAllocs()
//	b.ResetTimer()
//	n := 20
//	for i := 0; i < n; i++ {
//		b.Run(strconv.Itoa(2<<i), func(b *testing.B) {
//			b.ReportAllocs()
//			b.RunParallel(func(pb *testing.PB) {
//				for pb.Next() {
//					a := make([]byte, 2<<i)
//					//_ = bytes.NewBuffer(a)
//					_ = a
//				}
//			})
//		})
//	}
//
//}

func TestGetBufferReaderSize(t *testing.T) {
	buffer := bytes.NewBuffer(make([]byte, 1000))
	for i := 0; i < 1000; i++ {
		reader := GetBufferReaderSize(buffer, buffer.Len())
		PutBufferReader(reader)
	}

}

func BenchmarkGetBufferReaderSize(b *testing.B) {
	buffer := bytes.NewBuffer(make([]byte, 1000))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			reader := GetBufferReaderSize(buffer, buffer.Len())
			PutBufferReader(reader)
		}
	})
}
func BenchmarkBufferReaderSize(b *testing.B) {
	buffer := bytes.NewBuffer(make([]byte, 1000))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = bufio.NewReaderSize(buffer, buffer.Len())
		}
	})
}

func BenchmarkGetBufferReader(b *testing.B) {
	buffer := bytes.NewBuffer(make([]byte, 1000))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			reader := GetBufferReader(buffer)
			PutBufferReader(reader)
		}
	})
}
func BenchmarkBufferReader(b *testing.B) {
	buffer := bytes.NewBuffer(make([]byte, 1000))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = bufio.NewReader(buffer)
		}
	})
}

func TestGetBufferWriter(t *testing.T) {
	buffer := bytes.NewBuffer(make([]byte, 1000))
	for i := 0; i < 1000; i++ {
		writer := GetBufferWriter(buffer)
		assert.Equal(t, 0, writer.Buffered())
		assert.Equal(t, 4096, writer.Size())
		writer.WriteByte(1)
		assert.Equal(t, 1, writer.Buffered())
		assert.Equal(t, 4095, writer.Available())
		PutBufWriter(writer)
	}

}

func TestGetBufferWriterSize(t *testing.T) {
	buffer := bytes.NewBuffer(make([]byte, 1000))
	for i := 0; i < 1000; i++ {
		writer := GetBufferWriterSize(buffer, 1000)
		assert.Equal(t, 0, writer.Buffered())
		writer.WriteByte(1)
		assert.Equal(t, 1, writer.Buffered())
		PutBufWriter(writer)
	}
}

func BenchmarkGetBufferWriter(b *testing.B) {
	buffer := bytes.NewBuffer(make([]byte, 1000))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			writer := GetBufferWriter(buffer)
			writer.WriteByte(1)
			PutBufWriter(writer)
		}
	})
}

func BenchmarkGetBufferWriterSize(b *testing.B) {
	buffer := bytes.NewBuffer(make([]byte, 1000))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			writer := GetBufferWriterSize(buffer, 1000)
			writer.WriteByte(1)
			PutBufWriter(writer)
		}
	})
}

func BenchmarkBufferWriterSize(b *testing.B) {
	buffer := bytes.NewBuffer(make([]byte, 1000))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			writer := bufio.NewWriterSize(buffer, 1000)
			writer.WriteByte(1)
		}
	})
}

func BenchmarkBufferWriter(b *testing.B) {
	buffer := bytes.NewBuffer(make([]byte, 1000))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			writer := bufio.NewWriter(buffer)
			writer.WriteByte(1)
		}
	})
}

func TestGetBufferReader(t *testing.T) {
	b := make([]byte, 1000)
	buffer := bytes.NewBuffer(b)
	waitGroup := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		waitGroup.Add(1)
		go func(k int) {
			defer waitGroup.Done()

			reader := GetBufferReader(buffer)
			assert.NotNil(t, reader)
			PutBufferReader(reader)
		}(i)
	}

	waitGroup.Wait()
}
