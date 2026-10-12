package classfile

import (
	"bufio"
	"fmt"
	"io"
)

// reader wraps a bufio.Reader with the big-endian fixed-width reads the
// class file format is built from, plus running error tracking so call
// sites can stay flat instead of checking err after every single read.
type reader struct {
	r   *bufio.Reader
	err error
}

func newReader(r io.Reader) *reader {
	return &reader{r: bufio.NewReaderSize(r, 32*1024)}
}

func (r *reader) u1() byte {
	if r.err != nil {
		return 0
	}
	b, err := r.r.ReadByte()
	if err != nil {
		r.err = fmt.Errorf("reading u1: %w", err)
		return 0
	}
	return b
}

func (r *reader) u2() uint16 {
	return uint16(r.u1())<<8 | uint16(r.u1())
}

func (r *reader) u4() uint32 {
	return uint32(r.u2())<<16 | uint32(r.u2())
}

// bytes reads n raw bytes.
func (r *reader) bytes(n uint32) []byte {
	if r.err != nil {
		return nil
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(r.r, buf); err != nil {
		r.err = fmt.Errorf("reading %d bytes: %w", n, err)
		return nil
	}
	return buf
}

// skip discards n bytes without allocating a buffer for them.
func (r *reader) skip(n uint32) {
	if r.err != nil {
		return
	}
	if _, err := io.CopyN(io.Discard, r.r, int64(n)); err != nil {
		r.err = fmt.Errorf("skipping %d bytes: %w", n, err)
	}
}
