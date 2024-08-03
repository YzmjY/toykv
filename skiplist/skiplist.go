package skiplist

import (
	"math/rand"
	"sync/atomic"
	"unsafe"

	"github.com/YzmjY/toykv/x"
)

const (
	maxHeight = 20
)

const MaxNodeSize = int(unsafe.Sizeof(node{}))

type node struct {
	value atomic.Uint64

	keyOffset uint32
	keySize   uint16
	height    uint16

	tower [maxHeight]atomic.Uint32
}

func (nd *node) key(arena *Arena) []byte {
	return arena.getKey(nd.keyOffset, nd.keySize)
}

func (nd *node) val(arena *Arena) x.ValueStruct {
	offset, size := decodeValue(nd.value.Load())
	return arena.getVal(offset, size)
}

func (nd *node) setValue(arena *Arena, v x.ValueStruct) {
	offset := arena.allocVal(v)
	size := v.EncodeSize()

	nd.value.Store(encodeValue(offset, size))
}

func (nd *node) setKey(arena *Arena, key []byte) {
	offset := arena.allocKey(key)
	nd.keyOffset = offset
	nd.keySize = uint16(len(key))
}

func (nd *node) getNextOffset(h int) uint32 {
	return nd.tower[h].Load()
}
func (nd *node) getNodeOffset(arena *Arena) uint32 {
	return arena.getNodeOffset(nd)
}
func (nd *node) casNextOffset(h int, old, new uint32) bool {
	return nd.tower[h].CompareAndSwap(old, new)
}

func encodeValue(offset uint32, size uint32) uint64 {
	return uint64(offset)<<32 | uint64(size)
}

func decodeValue(v uint64) (offset uint32, size uint32) {
	offset = uint32(v >> 32)
	size = uint32(v)
	return
}

type Skiplist struct {
	height atomic.Int32
	head   *node

	ref   atomic.Int32
	arena *Arena
}

func (s *Skiplist) IncrRef() {
	s.ref.Add(1)
}

func (s *Skiplist) DecrRef() {
	new := s.ref.Add(-1)
	if new > 0 {
		// still alive
		return
	}

	// make gc can reclaim memory
	s.arena = nil
	s.head = nil
}

func newNode(arena *Arena, key []byte, v x.ValueStruct, height int) *node {
	ndOffset := arena.allocNode(height)
	nd := arena.getNode(ndOffset)
	nd.setValue(arena, v)
	nd.setKey(arena, key)
	nd.height = uint16(height)

	return nd
}

func NewSkiplist(arenaSize int64) *Skiplist {
	arena := newArena(arenaSize)
	head := newNode(arena, nil, x.ValueStruct{}, maxHeight)

	s := &Skiplist{
		arena: arena,
		head:  head,
	}

	s.height.Store(1)
	s.ref.Store(1)
	return s
}

func (s *Skiplist) randomHeight() int {
	h := 1
	for h < maxHeight && (rand.Intn(4)+1) == 1 {
		// 1/4的概率继续增加高度
		h++
	}

	return h
}

func (s *Skiplist) getNext(nd *node, height int) *node {
	nextOffset := nd.getNextOffset(height)
	return s.arena.getNode(nextOffset)
}

func (s *Skiplist) getHeight() int32 {
	return s.height.Load()
}

// !!! 无锁实现
func (s *Skiplist) Put(key []byte, v x.ValueStruct) {
	// 实现无锁的插入
	height := s.getHeight()
	var (
		prev [maxHeight + 1]*node
		next [maxHeight + 1]*node
	)
	prev[height] = s.head
	next[height] = nil

	// 获取当前skiplist下，插入节点在各层上的前驱后继
	for i := int(height) - 1; i >= 0; i-- {
		prev[i], next[i] = s.findSpliceForLevel(key, prev[i+1], i)
		if prev[i] == next[i] {
			// exist
			prev[i].setValue(s.arena, v)
			return
		}
	}

	// 确定高度
	newHeight := s.randomHeight()
	newNode := newNode(s.arena, key, v, newHeight)

	// cas 设置skiplist高度
	height = s.getHeight() // 减少CAS失败的可能性
	for newHeight > int(height) {
		if s.height.CompareAndSwap(height, int32(newHeight)) {
			break
		}

		height = s.getHeight()
	}

	// 插入到skiplist中
	for i := 0; i < newHeight; i++ {
	loop:
		// insert from base level, so if anyone insert a same key
		// we will find it at level 0
		if prev[i] == nil {
			// 还没获取该层的前驱后继，可能是插入了其他node，导致skiplist 的层高变高了
			prev[i], next[i] = s.findSpliceForLevel(key, s.head, i)
		}

		nextOffset := s.arena.getNodeOffset(next[i])
		newNode.tower[i].Store(nextOffset) // 如果next已经变了，那么下面的cas会失败
		if prev[i].casNextOffset(i, nextOffset, newNode.getNodeOffset(s.arena)) {
			// prev[i] next[i]之间没有插入新元素，本层插入成功
			continue
		}

		// 有其他元素插入，重新获取
		prev[i], next[i] = s.findSpliceForLevel(key, s.head, i)
		if prev[i] == next[i] {
			prev[i].setValue(s.arena, v)
			return
		}
		goto loop
	}

}

func (s *Skiplist) Get(key []byte) x.ValueStruct {
	n, _ := s.findGreaterOrEqual(key)
	if n == nil {
		return x.ValueStruct{}
	}

	tarKey := n.key(s.arena)
	if !x.SameUserKey(tarKey, key) {
		return x.ValueStruct{}
	}

	val := n.val(s.arena)
	val.Version = x.ParseTs(tarKey)
	return val
}

func (s *Skiplist) Empty() bool {
	return s.findLast() == nil
}

func (s *Skiplist) findSpliceForLevel(key []byte, before *node, level int) (*node, *node) {
	for {
		next := s.getNext(before, level)
		if next == nil {
			return before, nil
		}

		// next not nil, compare keys
		nextKey := next.key(s.arena)
		if cmp := x.KeysCompare(key, nextKey); cmp == 0 {
			return next, next
		} else if cmp < 0 {
			return before, next
		} else {
			before = next
		}
	}
}

func (s *Skiplist) findLast() *node {
	cur := s.head
	height := s.getHeight() - 1

	for {
		next := s.getNext(cur, int(height))
		if next != nil {
			cur = next
			continue
		}

		// next == nil
		if height == 0 {
			// base level
			if cur == s.head {
				return nil // nothing in skiplist
			}

			return cur
		}

		// next == nil && not base level
		height--
	}
}

// findGreaterOrEqual 获取第一个大于等于key的node
func (s *Skiplist) findGreaterOrEqual(key []byte) (*node, bool) {
	// 最高层开始，判断next
	// next > key : 找最小的大于，下降一层
	// next == key : 返回
	// next <key: 找大的，前进一个
	cur := s.head
	height := s.getHeight() - 1

	for {
		next := s.getNext(cur, int(height))
		if next == nil {
			if height > 0 {
				height--
				continue
			}

			// base level
			return nil, false
		}

		nextKey := next.key(s.arena)
		cmp := x.KeysCompare(key, nextKey)
		if cmp == 0 {
			return next, true
		} else if cmp < 0 {
			if height > 0 {
				height--
				continue
			}

			return next, false
		} else {
			cur = next
		}
	}
}

func (s *Skiplist) findLessThan(key []byte, enableEqual bool) (*node, bool) {
	// 最高层开始，判断next
	// next> key :下降一层
	// next== key : 下降一层
	// next<key:前进
	cur := s.head
	height := s.getHeight() - 1

	for {
		next := s.getNext(cur, int(height))
		if next == nil {
			if height > 0 {
				height--
				continue
			}

			if cur == s.head {
				return nil, false
			}

			return cur, false
		}

		nextKey := next.key(s.arena)
		cmp := x.KeysCompare(key, nextKey)
		if cmp == 0 && enableEqual {
			return next, true
		}

		if cmp <= 0 {
			// key <= nextKey
			if height > 0 {
				height--
				continue
			}

			if cur == s.head {
				return nil, false
			}

			return cur, false
		} else {
			cur = next
		}
	}
}

func (s *Skiplist) NewIterator() *Iterator {
	s.IncrRef()
	return &Iterator{
		cur: s.head,
		skl: s,
	}
}

type Iterator struct {
	cur *node
	skl *Skiplist
}

func (iter *Iterator) Close() {
	iter.skl.DecrRef()
	iter.cur = nil
	iter.skl = nil
}

func (iter *Iterator) Next() {
	iter.cur = iter.skl.getNext(iter.cur, 0)
}

func (iter *Iterator) Prev() {
	iter.cur, _ = iter.skl.findLessThan(iter.cur.key(iter.skl.arena), false)
}

// Vaild 当前迭代器是否有效，即cur是一个有效元素
func (iter *Iterator) Vaild() bool {
	return !(iter.cur == nil || iter.cur == iter.skl.head)
}

// Seek 移动到第一个大于等于key的位置
func (iter *Iterator) Seek(key []byte) {
	iter.cur, _ = iter.skl.findGreaterOrEqual(key)
}

// SeekPrev 移动到最后一个小于等于key的位置
func (iter *Iterator) SeekPrev(key []byte) {
	iter.cur, _ = iter.skl.findLessThan(key, true)
}

// SeekToLast 移动到最后一个元素
func (iter *Iterator) SeekToLast() {
	iter.cur = iter.skl.findLast()
}

// SeekToFirst 移动到第一个元素
func (iter *Iterator) SeekToFirst() {
	iter.cur = iter.skl.getNext(iter.skl.head, 0)
}

func (iter *Iterator) Key() []byte {
	return iter.cur.key(iter.skl.arena)
}

func (iter *Iterator) Value() x.ValueStruct {
	return iter.cur.val(iter.skl.arena)
}

// UniIterator 单向迭代器，reversed参数表示迭代的方向
type UniIterator struct {
	reversed bool
	iter     *Iterator
}

var _ x.Iterator = &UniIterator{}

func (s *Skiplist) NewUinIterator(reversed bool) *UniIterator {
	return &UniIterator{
		reversed: reversed,
		iter:     s.NewIterator(),
	}
}

func (ui *UniIterator) Next() {
	if ui.reversed {
		ui.iter.Prev()
	} else {
		ui.iter.Next()
	}
}

func (ui *UniIterator) Vaild() bool {
	return ui.iter.Vaild()
}

func (ui *UniIterator) Rewind() {
	if !ui.reversed {
		ui.iter.SeekToFirst()
	} else {
		ui.iter.SeekToLast()
	}
}

func (ui *UniIterator) Seek(key []byte) {
	if ui.reversed {
		ui.iter.SeekPrev(key)
	} else {
		ui.iter.Seek(key)
	}
}

func (ui *UniIterator) Key() []byte {
	return ui.iter.Key()
}

func (ui *UniIterator) Value() x.ValueStruct {
	return ui.iter.Value()
}

func (ui *UniIterator) Close() {
	ui.iter.Close()
}
