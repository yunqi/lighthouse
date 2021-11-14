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

package xbinary

import (
	"encoding/binary"
	"errors"
	"github.com/yunqi/lighthouse/internal/xbytes"
	"io"
)

var (
	InvalidLengthErr = errors.New("invalid length")

	read1BytesPool = xbytes.GetNBytePool(1)
	read2BytesPool = xbytes.GetNBytePool(2)
	read4BytesPool = xbytes.GetNBytePool(4)
)

func WriteUint16(w io.Writer, i uint16) error {
	_, err := w.Write([]byte{byte(i >> 8), byte(i)})
	return err
}

func ReadUint16(r io.Reader) (uint16, error) {
	data := read2BytesPool.Get().([]byte)
	defer read2BytesPool.Put(data)
	n, err := r.Read(data)
	if err != nil {
		return 0, err
	}
	if n < 2 {
		return 0, InvalidLengthErr
	}
	return binary.BigEndian.Uint16(data), nil
}

//-----------------

func WriteBool(w io.Writer, b bool) error {
	data := read1BytesPool.Get().([]byte)
	defer read1BytesPool.Put(data)
	if b {
		data[0] = 1
	} else {
		data[0] = 0
	}
	_, err := w.Write(data)
	return err
}

func ReadBool(r io.Reader) (bool, error) {
	b := read1BytesPool.Get().([]byte)
	defer read1BytesPool.Put(b)
	_, err := r.Read(b)
	if err != nil {
		return false, err
	}
	return b[0] == 1, nil
}

//------------

func ReadUint32(r io.Reader) (uint32, error) {
	data := read4BytesPool.Get().([]byte)
	defer read4BytesPool.Put(data)
	n, err := r.Read(data)
	if err != nil {
		return 0, err
	}
	if n < 4 {
		return 0, InvalidLengthErr
	}
	return binary.BigEndian.Uint32(data), nil
}

func WriteUint32(w io.Writer, i uint32) error {
	data := read4BytesPool.Get().([]byte)
	defer read4BytesPool.Put(data)
	data[0] = byte(i >> 24)
	data[1] = byte(i >> 16)
	data[2] = byte(i >> 8)
	data[3] = byte(i)
	_, err := w.Write(data)
	return err

}

//------------

func WriteBytes(w io.Writer, s []byte) (err error) {
	// length
	err = WriteUint16(w, uint16(len(s)))
	if err == nil {
		_, err = w.Write(s)
	}
	return
}

func ReadBytes(r io.Reader) (b []byte, err error) {
	nBytes := read2BytesPool.Get().([]byte)
	defer read2BytesPool.Put(nBytes)
	_, err = io.ReadFull(r, nBytes)
	if err != nil {
		return nil, err
	}

	length := int(binary.BigEndian.Uint16(nBytes))
	if length == 0 {
		return nil, nil
	}
	pool := xbytes.GetNBytePool(length)
	payload := pool.Get().([]byte)
	defer pool.Put(payload)
	_, err = io.ReadFull(r, payload)
	if err != nil {
		return nil, err
	}

	return payload, nil
}
