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

package binary

import (
	"encoding/binary"
	"errors"
	"io"
)

var InvalidLengthErr = errors.New("invalid length")

func WriteUint16(w io.Writer, i uint16) error {
	_, err := w.Write([]byte{byte(i >> 8), byte(i)})
	return err
}

func ReadUint16(r io.Reader) (uint16, error) {
	data := make([]byte, 2)
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
	var data = make([]byte, 1)
	if b {
		data[0] = 1
	}
	_, err := w.Write(data)
	return err
}

func ReadBool(r io.Reader) (bool, error) {
	b := make([]byte, 1)
	_, err := r.Read(b)
	if err != nil {
		return false, err
	}
	return b[0] == 1, nil
}

//------------

func ReadUint32(r io.Reader) (uint32, error) {
	data := make([]byte, 4)
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
	_, err := w.Write([]byte{
		byte(i >> 24),
		byte(i >> 16),
		byte(i >> 8),
		byte(i),
	})
	return err

}

//------------

func WriteString(w io.Writer, s []byte) (err error) {
	// length
	err = WriteUint16(w, uint16(len(s)))
	if err == nil {
		_, err = w.Write(s)
	}
	return
}
func ReadString(r io.Reader) (b []byte, err error) {
	nBytes := make([]byte, 2)
	_, err = io.ReadFull(r, nBytes)
	if err != nil {
		return nil, err
	}

	length := int(binary.BigEndian.Uint16(nBytes))
	payload := make([]byte, length)

	_, err = io.ReadFull(r, payload)
	if err != nil {
		return nil, err
	}

	return payload, nil
}
