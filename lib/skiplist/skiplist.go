package skiplist

import (
	"sync/atomic"
	"unsafe"
)

const (
	maxHeight = 20
)

const MaxNodeSize = int(unsafe.Sizeof(node{}))

type node struct {
	value atomic.Uint64

	keyOffset uint32
	keySize   uint32
	height    uint16

	tower [maxHeight]atomic.Uint32
}

type Skiplist struct {

}

