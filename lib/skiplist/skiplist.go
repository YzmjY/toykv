package skiplist

import (
	"sync/atomic"
	"unsafe"

	"github.com/YzmjY/toykv/lib/val"
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

func (nd *node) key(arena *Arena) []byte {

}

func (nd *node) val(arena *Arena) val.ValueStruct {

}

func (nd *node) setValue(arena *Arena, v val.ValueStruct) {

}

func (nd *node) getNextOffset(h int) uint32 {

}

func (nd *node) casNextOffset(h int, old, now uint32) bool {

}

type Skiplist struct {
}

func (s *Skiplist) IncrRef() {

}

func (s *Skiplist) DecrRef() {

}

func newNode(arena *Arena, key []byte, v val.ValueStruct, height int) *node {

}

func encodeValue(offset uint32, size uint16) uint64 {

}

func decodeValue(v uint64) (offset uint32, size uint32) {

}

func NewSkiplist(arenaSize int64) *Skiplist {

}

func (s *Skiplist) randomHeight() int {

}

func (s *Skiplist) getNext(nd *node, height int) *node {

}

func (s *Skiplist) getHeight() int32 {

}

func (s *Skiplist) Put(key []byte, v val.ValueStruct) {

}

func (s *Skiplist) Get(key []byte) val.ValueStruct {

}

func (s *Skiplist) Empty() bool {

}

func (s *Skiplist) findSpliceForLevel(key []byte, before *node, level int) (*node, *node) {

}

func (s *Skiplist) findLast() *node {

}

func (s *Skiplist) findGreaterOrEqual(key []byte) (*node, bool) {
	// 最高层开始，判断next
	// next > key : 找最小的大于，下降一层
	// next == key : 返回
	// next <key: 找大的，前进一个
}

func (s *Skiplist) findLessThan(key []byte) *node {
	// 最高层开始，判断next
	// next> key :下降一层
	// next== key : 下降一层
	// next<key:前进
}
