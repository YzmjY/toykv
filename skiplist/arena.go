package skiplist

import (
	"sync/atomic"
	"unsafe"

	"github.com/YzmjY/toykv/val"
)

const (
	offsetSize = int(unsafe.Sizeof(uint32(0)))

	nodeAlign = int(unsafe.Sizeof(uint64(0))) - 1
)

// Arena Skiplist中的内存管理，
type Arena struct {
	n   atomic.Uint32
	buf []byte
}

func newArena(n int64) *Arena {
	ans := Arena{}
	ans.buf = make([]byte, 0, n)
	ans.n.Store(1)

	return &ans
}

func (s *Arena) size() int64 {
	return int64(s.n.Load())
}

func (s *Arena) allocNode(height int) uint32 {
	// 不是每个node都占满了maxHeight
	unused := (maxHeight - height) * offsetSize

	size := uint32(MaxNodeSize - unused + nodeAlign)

	// 预占size的大小
	end := s.n.Add(size)

	// 地址对齐
	return (end - size + uint32(nodeAlign)) & ^uint32(nodeAlign)
}

func (s *Arena) allocVal(val val.ValueStruct) uint32 {
	size := val.EncodeSize()

	end := s.n.Add(size)
	st := end - size
	val.Encode(s.buf[st:end])

	return st
}

func (s *Arena) allocKey(key []byte) uint32 {
	size := uint32(len(key))
	end := s.n.Add(size)

	st := end - size

	copy(s.buf[st:end], key)

	return st
}

func (s *Arena) getKey(offset uint32, size uint16) []byte {
	return s.buf[offset : offset+uint32(size)]
}

func (s *Arena) getVal(offset uint32, size uint32) (ret val.ValueStruct) {
	ret.Decode(s.buf[offset : offset+size])
	return
}

func (s *Arena) getNode(offset uint32) *node {
	if offset == 0 {
		return nil
	}
	// 强转为node指针
	return (*node)(unsafe.Pointer(&s.buf[offset]))
}

func (s *Arena) getNodeOffset(nd *node) uint32 {
	if nd == nil {
		return 0
	}

	// |-----|-----
	// ^     ^
	// 0	*node
	end := uintptr(unsafe.Pointer(nd))
	st := uintptr(unsafe.Pointer(&s.buf[0]))
	return uint32(end - st)
}
