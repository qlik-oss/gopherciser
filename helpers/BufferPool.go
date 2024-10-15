package helpers

import (
	"bytes"
	"io"
	"sync"
)

type (
	BufferPool struct {
		pool sync.Pool
	}
)

var GlobalBufferPool = NewBufferPool()

func NewBufferPool() *BufferPool {
	return &BufferPool{sync.Pool{New: newBuffer}}
}

func newBuffer() interface{} {
	return bytes.NewBuffer(nil)
}

func (bp *BufferPool) Get() *bytes.Buffer {
	return bp.pool.Get().(*bytes.Buffer)
}

func (bp *BufferPool) Put(buf *bytes.Buffer) {
	buf.Reset()
	bp.pool.Put(buf)
}

func ReadAll(r io.Reader) ([]byte, error) {
	buf := GlobalBufferPool.Get()
	defer GlobalBufferPool.Put(buf)

	capacity := int64(bytes.MinRead)
	var err error
	// If the buffer overflows, we will get bytes.ErrTooLarge.
	// Return that as an error. Any other panic remains.
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = panicErr
		} else {
			panic(e)
		}
	}()
	if int64(int(capacity)) == capacity {
		buf.Grow(int(capacity))
	}
	_, err = buf.ReadFrom(r)
	return buf.Bytes(), err
}
