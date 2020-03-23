package helpers

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type (
	Buffer struct {
		B     *strings.Builder
		Error error
	}
)

// NewBuffer creates and initializes a new string Buffer.
func NewBuffer() *Buffer {
	return &Buffer{B: &strings.Builder{}}
}

// WriteString to buffer, errors written to stderr
func (buffer *Buffer) WriteString(s string) {
	if buffer == nil || buffer.B == nil {
		return
	}

	if _, err := buffer.B.WriteString(s); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "buffer write error: %v\n", err)
		buffer.Error = err
	}
}

// Write bytes to buffer, errors written to stderr
func (buffer *Buffer) WriteBytes(p []byte) {
	if buffer == nil || buffer.B == nil {
		return
	}

	if _, err := buffer.B.Write(p); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "buffer write error: %v\n", err)
		buffer.Error = err
	}
}

// WriteByte appends the byte c to the buffer, growing the buffer as needed.
func (buffer *Buffer) WriteByte(c byte) {
	if buffer == nil || buffer.B == nil {
		return
	}

	if err := buffer.B.WriteByte(c); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "buffer write error: %v\n", err)
		buffer.Error = err
	}
}

// WriteRune appends the UTF-8 encoding of Unicode code point r to the buffer, errors written to stderr
func (buffer *Buffer) WriteRune(r rune) {
	if buffer == nil || buffer.B == nil {
		return
	}

	if _, err := buffer.B.WriteRune(r); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "buffer write error: %v\n", err)
		buffer.Error = err
	}
}

// WriteTo writes data to w until the buffer is drained or an error occurs., errors written to stderr
func (buffer *Buffer) WriteTo(w io.Writer) {
	if buffer == nil || buffer.B == nil {
		return
	}

	if _, err := w.Write([]byte(buffer.B.String())); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "buffer write error: %v\n", err)
		buffer.Error = err
	}
}

// String returns the contents of the unread portion of the buffer
// as a string. If the Buffer is a nil pointer, it returns "<nil>".
func (buffer *Buffer) String() string {
	if buffer == nil {
		return "<nil>"
	}
	return buffer.B.String()
}

// Reset resets the buffer to be empty,
func (buffer *Buffer) Reset() {
	buffer.B.Reset()
}
