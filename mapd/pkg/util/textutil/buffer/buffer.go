// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package buffer

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"

	"github.com/google/uuid"
)

// smallBufferSize is an initial allocation minimal capacity.
const smallBufferSize = 64

type Dispatcher struct {
	mux sync.RWMutex
	sub map[uuid.UUID]chan<- []byte
	buf *Buffer
	lim int
}

func (d *Dispatcher) String() string {
	return d.buf.String()
}

func (d *Dispatcher) Subscribe(ctx context.Context) <-chan []byte {
	d.mux.Lock()
	defer d.mux.Unlock()

	ch := make(chan []byte, d.lim)
	id := uuid.New()
	d.sub[id] = ch

	go func() {
		defer close(ch)

		<-ctx.Done()

		d.mux.Lock()
		defer d.mux.Unlock()

		delete(d.sub, id)
	}()

	ch <- d.buf.Bytes()
	return ch
}

func (d *Dispatcher) Write(p []byte) (n int, err error) {
	d.mux.RLock()
	defer d.mux.RUnlock()

	for _, sub := range d.sub {
		sub <- p
	}
	return d.buf.Write(p)
}

func (d *Dispatcher) WriteString(s string) (n int, err error) {
	return d.Write([]byte(s))
}

func NewDispatcher(limit int) *Dispatcher {
	return &Dispatcher{sub: make(map[uuid.UUID]chan<- []byte), buf: NewEmptyBuffer(limit), lim: limit}
}

// A Buffer is a variable-sized buffer of bytes with Read and Write methods.
// The zero value for Buffer is an empty buffer ready to use.
type Buffer struct {
	mux sync.RWMutex
	buf []byte // contents are the bytes buf[off : len(buf)]
	off int    // read at &buf[off], write at &buf[len(buf)]
	lim int    // limit on newlines
}

// ErrTooLarge is passed to panic if memory cannot be allocated to store data in a buffer.
var ErrTooLarge = errors.New("bytes.Buffer: too large")
var errNegativeRead = errors.New("bytes.Buffer: reader returned negative count from Read")

const maxInt = int(^uint(0) >> 1)

// Bytes returns a slice of length b.Len() holding the unread portion of the buffer.
// The slice is valid for use only until the next buffer modification (that is,
// only until the next call to a method like Read, Write, Reset, or Truncate).
// The slice aliases the buffer content at least until the next buffer modification,
// so immediate changes to the slice will affect the result of future reads.
func (b *Buffer) Bytes() []byte {
	b.mux.RLock()
	defer b.mux.RUnlock()

	return b.buf[b.off:]
}

// String returns the contents of the unread portion of the buffer
// as a string. If the Buffer is a nil pointer, it returns "<nil>".
//
// To build strings more efficiently, see the strings.Builder type.
func (b *Buffer) String() string {
	if b == nil {
		// Special case, useful in debugging.
		return "<nil>"
	}
	b.mux.RLock()
	defer b.mux.RUnlock()

	return string(b.buf[b.off:])
}

// empty reports whether the unread portion of the buffer is empty.
func (b *Buffer) empty() bool { return len(b.buf) <= b.off }

// Len returns the number of bytes of the unread portion of the buffer;
// b.Len() == len(b.Bytes()).
func (b *Buffer) Len() int {
	b.mux.RLock()
	defer b.mux.RUnlock()

	return len(b.buf) - b.off
}

// Cap returns the capacity of the buffer's underlying byte slice, that is, the
// total space allocated for the buffer's data.
func (b *Buffer) Cap() int {
	b.mux.RLock()
	defer b.mux.RUnlock()

	return cap(b.buf)
}

// Reset resets the buffer to be empty,
// but it retains the underlying storage for use by future writes.
// Reset is the same as Truncate(0).
func (b *Buffer) Reset() {
	b.mux.Lock()
	defer b.mux.Unlock()

	b.buf = b.buf[:0]
	b.off = 0
}

// tryGrowByReslice is a inlineable version of grow for the fast-case where the
// internal buffer only needs to be resliced.
// It returns the index where bytes should be written and whether it succeeded.
func (b *Buffer) tryGrowByReslice(n int) (int, bool) {
	if l := len(b.buf); n <= cap(b.buf)-l {
		b.buf = b.buf[:l+n]
		return l, true
	}
	return 0, false
}

// grow grows the buffer to guarantee space for n more bytes.
// It returns the index where bytes should be written.
// If the buffer can't grow it will panic with ErrTooLarge.
func (b *Buffer) grow(n int) int {
	m := len(b.buf) - b.off
	// If buffer is empty, reset to recover space.
	if m == 0 && b.off != 0 {
		b.buf = b.buf[:0]
		b.off = 0
	}
	// Try to grow by means of a reslice.
	if i, ok := b.tryGrowByReslice(n); ok {
		return i
	}
	if b.buf == nil && n <= smallBufferSize {
		b.buf = make([]byte, n, smallBufferSize)
		return 0
	}
	c := cap(b.buf)
	if n <= c/2-m {
		// We can slide things down instead of allocating a new
		// slice. We only need m+n <= c to slide, but
		// we instead let capacity get twice as large so we
		// don't spend all our time copying.
		copy(b.buf, b.buf[b.off:])
	} else if c > maxInt-c-n {
		panic(ErrTooLarge)
	} else {
		// Not enough space anywhere, we need to allocate.
		buf := makeSlice(2*c + n)
		copy(buf, b.buf[b.off:])
		b.buf = buf
	}
	// Restore b.off and len(b.buf).
	b.off = 0
	b.buf = b.buf[:m+n]
	return m
}

// resize trims the old lines from the buffer.
func (b *Buffer) resize() {
	for i := bytes.Count(b.buf, []byte{'\n'}); i > b.lim; i-- {
		r := bytes.IndexRune(b.buf, '\n')
		if b.off > r {
			b.off -= r + 1
		} else {
			b.off = 0
		}
		b.buf = b.buf[r+1:]
	}
}

// Write appends the contents of p to the buffer, growing the buffer as
// needed. The return value n is the length of p; err is always nil. If the
// buffer becomes too large, Write will panic with ErrTooLarge.
func (b *Buffer) Write(p []byte) (n int, err error) {
	b.mux.Lock()
	defer b.mux.Unlock()

	m, ok := b.tryGrowByReslice(len(p))
	if !ok {
		m = b.grow(len(p))
	}
	n = copy(b.buf[m:], p)
	b.resize()
	return n, nil
}

// WriteString appends the contents of s to the buffer, growing the buffer as
// needed. The return value n is the length of s; err is always nil. If the
// buffer becomes too large, WriteString will panic with ErrTooLarge.
func (b *Buffer) WriteString(s string) (n int, err error) {
	b.mux.Lock()
	defer b.mux.Unlock()

	m, ok := b.tryGrowByReslice(len(s))
	if !ok {
		m = b.grow(len(s))
	}
	n = copy(b.buf[m:], s)
	b.resize()
	return n, nil
}

// MinRead is the minimum slice size passed to a Read call by
// Buffer.ReadFrom. As long as the Buffer has at least MinRead bytes beyond
// what is required to hold the contents of r, ReadFrom will not grow the
// underlying buffer.
const MinRead = 512

// ReadFrom reads data from r until EOF and appends it to the buffer, growing
// the buffer as needed. The return value n is the number of bytes read. Any
// error except io.EOF encountered during the read is also returned. If the
// buffer becomes too large, ReadFrom will panic with ErrTooLarge.
func (b *Buffer) ReadFrom(r io.Reader) (n int64, err error) {
	b.mux.Lock()
	defer b.mux.Unlock()

	for {
		i := b.grow(MinRead)
		b.buf = b.buf[:i]
		m, e := r.Read(b.buf[i:cap(b.buf)])
		if m < 0 {
			panic(errNegativeRead)
		}

		b.buf = b.buf[:i+m]
		n += int64(m)
		b.resize()
		if e == io.EOF {
			return n, nil // e is EOF, so return nil explicitly
		}
		if e != nil {
			return n, e
		}
	}
}

// makeSlice allocates a slice of size n. If the allocation fails, it panics
// with ErrTooLarge.
func makeSlice(n int) []byte {
	// If the make fails, give a known error.
	defer func() {
		if recover() != nil {
			panic(ErrTooLarge)
		}
	}()
	return make([]byte, n)
}

// WriteTo writes data to w until the buffer is drained or an error occurs.
// The return value n is the number of bytes written; it always fits into an
// int, but it is int64 to match the io.WriterTo interface. Any error
// encountered during the write is also returned.
func (b *Buffer) WriteTo(w io.Writer) (n int64, err error) {
	b.mux.Lock()
	defer b.mux.Unlock()

	if nBytes := len(b.buf) - b.off; nBytes > 0 {
		m, e := w.Write(b.buf[b.off:])
		if m > nBytes {
			panic("bytes.Buffer.WriteTo: invalid Write count")
		}
		b.off += m
		n = int64(m)
		if e != nil {
			return n, e
		}
		// all bytes should have been written, by definition of
		// Write method in io.Writer
		if m != nBytes {
			return n, io.ErrShortWrite
		}
	}
	return n, nil
}

// Read reads the next len(p) bytes from the buffer or until the buffer
// is drained. The return value n is the number of bytes read. If the
// buffer has no data to return, err is io.EOF (unless len(p) is zero);
// otherwise it is nil.
func (b *Buffer) Read(p []byte) (n int, err error) {
	b.mux.Lock()
	defer b.mux.Unlock()

	if b.empty() {
		if len(p) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}
	n = copy(p, b.buf[b.off:])
	b.off += n
	return n, nil
}

// NewBuffer creates and initializes a new Buffer using buf as its
// initial contents. The new Buffer takes ownership of buf, and the
// caller should not use buf after this call. NewBuffer is intended to
// prepare a Buffer to read existing data. It can also be used to set
// the initial size of the internal buffer for writing. To do that,
// buf should have the desired capacity but a length of zero.
//
// In most cases, new(Buffer) (or just declaring a Buffer variable) is
// sufficient to initialize a Buffer.
func NewBuffer(buf []byte, limit int) *Buffer { return &Buffer{buf: buf, lim: limit} }

func NewEmptyBuffer(limit int) *Buffer { return &Buffer{lim: limit} }

// NewBufferString creates and initializes a new Buffer using string s as its
// initial contents. It is intended to prepare a buffer to read an existing
// string.
//
// In most cases, new(Buffer) (or just declaring a Buffer variable) is
// sufficient to initialize a Buffer.
func NewBufferString(s string, limit int) *Buffer {
	return &Buffer{buf: []byte(s), lim: limit}
}
