package skiplist

import (
	"sync/atomic"
	"unsafe"
)

const (
	offsetSize = int(unsafe.Sizeof(uint32(0)))

	nodeAlign = int(unsafe.Sizeof(uint64(0))) - 1
)

type Arena struct {
	n atomic.Uint32
	buf  []byte
}

func newArena(n int64) *Arena {

}

func (s *Arena) size() int64 {

}

func (s *Arena) allocNode(height int) uint32 {

}

func(s *Arena) allocVal()
