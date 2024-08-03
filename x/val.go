package x

import "encoding/binary"

// ValueStruct represents the value info that can be associated with a key, but also the internal
// Meta field.
type ValueStruct struct {
	Meta      byte
	UserMeta  byte
	ExpiresAt uint64
	Value     []byte

	Version uint64 // This field is not serialized. Only for internal usage.
}

func (v *ValueStruct) EncodeSize() uint32 {
	return uint32(len(v.Value) + sizeVarint(v.ExpiresAt) + 2)
}

func sizeVarint(x uint64) int {
	var n int
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}

func (v *ValueStruct) Encode(dst []byte) uint32 {
	need := v.EncodeSize()
	dst[0] = v.Meta
	dst[1] = v.UserMeta
	vSize := binary.PutUvarint(dst[2:], v.ExpiresAt)

	copy(dst[2+vSize:], v.Value)
	return uint32(need)
}

func (v *ValueStruct) Decode(src []byte) {
	if len(src) < 2 {
		panic("len(src) must be gt 2")
	}
	v.Meta = src[0]
	v.UserMeta = src[1]

	var size int
	v.ExpiresAt, size = binary.Uvarint(src[2:])
	v.Value = src[2+size:]
}
