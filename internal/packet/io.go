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
	"bufio"
	"io"
)

type (
	// Reader is used to read data from bufio.Reader and create MQTT packet instance.
	Reader struct {
		buf     *bufio.Reader
		version Version
	}
	// Writer is used to encode MQTT packet into bytes and write it to bufio.Writer.
	Writer struct {
		buf *bufio.Writer
	}
	// ReadWriter warps Reader and Writer.
	ReadWriter struct {
		*Reader
		*Writer
	}
)

// NewReader returns a new Reader.
func NewReader(r io.Reader) *Reader {
	if buf, ok := r.(*bufio.Reader); ok {
		return &Reader{buf: buf, version: Version311}
	}
	return &Reader{buf: bufio.NewReaderSize(r, 2048), version: Version311}
}

// Read reads data from Reader and returns a  Packet instance.
// If any errors occurs, returns nil, error
func (r *Reader) Read() (Packet, error) {
	firstByte, err := r.buf.ReadByte()
	if err != nil {
		return nil, err
	}
	fh := &FixedHeader{PacketType: firstByte >> 4, Flags: firstByte & 15} //设置FixHeader
	length, err := DecodeRemainLength(r.buf)
	if err != nil {
		return nil, err
	}
	fh.RemainLength = length
	p, err := NewPacket(fh, r.version, r.buf)
	if err != nil {
		return nil, err
	}
	if p, ok := p.(*Connect); ok {
		r.version = p.Version
	}
	return p, err
}

// NewWriter returns a new Writer.
func NewWriter(w io.Writer) *Writer {
	if bufw, ok := w.(*bufio.Writer); ok {
		return &Writer{buf: bufw}
	}
	return &Writer{buf: bufio.NewWriterSize(w, 2048)}
}

// WritePacket writes the packet bytes to the Writer.
// Call Flush after WritePacket to flush buffered data to the underlying io.Writer.
func (w *Writer) WritePacket(p Packet) error {
	err := p.Encode(w.buf)
	if err != nil {
		return err
	}
	return nil
}

// Write raw bytes to the Writer.
// Call Flush after Write to flush buffered data to the underlying io.Writer.
func (w *Writer) Write(b []byte) error {
	_, err := w.buf.Write(b)
	if err != nil {
		return err
	}
	return nil
}

// WriteAndFlush writes and flush the packet bytes to the underlying io.Writer.
func (w *Writer) WriteAndFlush(p Packet) error {
	err := p.Encode(w.buf)
	if err != nil {
		return err
	}
	return w.Flush()
}

// Flush writes any buffered data to the underlying io.Writer.
func (w *Writer) Flush() error {
	return w.buf.Flush()
}
