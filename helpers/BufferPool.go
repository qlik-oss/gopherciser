package helpers

import (
	"bytes"
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
