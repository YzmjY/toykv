package x

import (
	"bytes"
	"encoding/binary"
	"math"
)

func KeyWithTs(key []byte, ts uint64) []byte {
	out := make([]byte, len(key)+8)
	copy(out, key)
	binary.BigEndian.PutUint64(out[len(key):], math.MaxUint64-ts)
	return out
}

func ParseTs(key []byte) uint64 {
	AssertTrue(len(key) >= 8)
	return math.MaxUint64 - binary.BigEndian.Uint64(key[len(key)-8:])
}

func ParseUserKey(key []byte) []byte {
	AssertTrue(len(key) >= 8)
	return key[:len(key)-8]
}

// KeysCompare compare key ,return
// lhs == rhs : 0
// lhs < rhs : -1
// lhs > rhs : 1
func KeysCompare(lhs, rhs []byte) int {
	l := len(lhs)
	r := len(rhs)
	if cmp := bytes.Compare(lhs[:l-8], rhs[:r-8]); cmp != 0 {
		return cmp
	}
	return bytes.Compare(lhs[l-8:], rhs[r-8:])
}

func SameUserKey(lhs, rhs []byte) bool {
	return bytes.Equal(ParseUserKey(lhs), ParseUserKey(rhs))
}
