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
	"golang.org/x/sync/singleflight"
	"io"
	"strconv"
	"sync"
)

var (
	bufReaderPool = &sync.Pool{}
	bufWriterPool = &sync.Pool{}
	bytesPoolMap  = &sync.Map{}
	singleFlight  = &singleflight.Group{}
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

// -----------------

// GetBufferReaderSize returns a bufio.Reader.
func GetBufferReaderSize(r io.Reader, size int) *bufio.Reader {
	if v := bufReaderPool.Get(); v != nil {
		br := v.(*bufio.Reader)
		br.Reset(r)
		return br
	}
	return bufio.NewReaderSize(r, size)
}

// GetBufferReader returns a bufio.Reader.
func GetBufferReader(r io.Reader) *bufio.Reader {
	if v := bufReaderPool.Get(); v != nil {
		br := v.(*bufio.Reader)
		br.Reset(r)
		return br
	}
	return bufio.NewReader(r)
}

// PutBufferReader recycles a bufio.Reader.
func PutBufferReader(r *bufio.Reader) {
	r.Reset(nil)
	bufReaderPool.Put(r)
}

// -----------------

// GetBufferWriterSize returns a bufio.Writer.
func GetBufferWriterSize(w io.Writer, size int) *bufio.Writer {
	if v := bufWriterPool.Get(); v != nil {
		bw := v.(*bufio.Writer)
		bw.Reset(w)
		return bw
	}
	return bufio.NewWriterSize(w, size)
}

// GetBufferWriter returns a bufio.Writer.
func GetBufferWriter(w io.Writer) *bufio.Writer {
	if v := bufWriterPool.Get(); v != nil {
		bw := v.(*bufio.Writer)
		bw.Reset(w)
		return bw
	}
	return bufio.NewWriter(w)
}

// PutBufWriter recycles a bufio.Writer.
func PutBufWriter(w *bufio.Writer) {
	w.Reset(nil)
	bufWriterPool.Put(w)
}
