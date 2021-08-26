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
	"io"
	"sync"
)

var (
	bufReaderPool = &sync.Pool{}
	bufWriterPool = &sync.Pool{}
)

func NewBufReaderSize(r io.Reader, size int) *bufio.Reader {
	if v := bufReaderPool.Get(); v != nil {
		br := v.(*bufio.Reader)
		br.Reset(r)
		return br
	}
	return bufio.NewReaderSize(r, size)
}
func PutBufReader(r *bufio.Reader) {
	r.Reset(nil)
	bufReaderPool.Put(r)
}

func NewBufWriterSize(w io.Writer, size int) *bufio.Writer {
	if v := bufWriterPool.Get(); v != nil {
		bw := v.(*bufio.Writer)
		bw.Reset(w)
		return bw
	}
	return bufio.NewWriterSize(w, size)
}
func PutBufWriter(w *bufio.Writer) {
	w.Reset(nil)
	bufWriterPool.Put(w)
}
