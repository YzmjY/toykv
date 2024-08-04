package bloomfilter

import "github.com/YzmjY/toykv/x"

type FilterPoliy interface {
	Name() string
	AppendFilter(keys [][]byte, dst []byte) []byte
	KeyMayMatch(key []byte, filter []byte) bool
}

type BloomFilterPoliy struct {
	bitsPerKey uint64
	k          uint64
}

func NewBloomFilterPoliy(bitsPerKey uint64) FilterPoliy {
	x.AssertTrue(bitsPerKey > 0)

	k := uint64(float64(bitsPerKey) * 0.69)
	if k > 30 {
		k = 30
	}
	if k < 1 {
		k = 1
	}

	return &BloomFilterPoliy{
		bitsPerKey: bitsPerKey,
		k:          k,
	}
}

func (*BloomFilterPoliy) Name() string {
	return "toykv.bloomfilter"
}

func (b *BloomFilterPoliy) AppendFilter(keys [][]byte, dst []byte) []byte {
	keysHash := make([]uint32, len(keys))
	for idx, key := range keys {
		keysHash[idx] = Hash(key)
	}

	return b.appendFilterHash(keysHash, dst)
}

func (b *BloomFilterPoliy) appendFilterHash(keysHash []uint32, dst []byte) []byte {
	nKeys := len(keysHash)
	nBits := nKeys * int(b.bitsPerKey)
	if nBits < 64 {
		nBits = 64
	}

	nBytes := (nBits + 7) / 8
	nBits = nBytes * 8

	c := len(dst)
	dst = extend(dst, nBytes+1)
	filter := dst[c:]

	for _, h := range keysHash {
		delta := h>>17 | h<<15
		for j := 0; j < int(b.k); j++ {
			bitPos := h % uint32(nBits)
			filter[bitPos/8] = filter[bitPos/8] | (1 << (bitPos % 8))
			h += delta
		}
	}

	filter[nBytes] = byte(b.k)
	return dst
}

func (b *BloomFilterPoliy) KeyMayMatch(key []byte, filter []byte) bool {
	keyHash := Hash(key)
	return b.KeyHashMayMatch(keyHash, filter)
}

func (b *BloomFilterPoliy) KeyHashMayMatch(hash uint32, filter []byte) bool {
	k := (uint8)(filter[len(filter)-1])

	nBits := uint32(8 * (len(filter) - 1))
	delta := hash>>17 | hash<<15
	for i := uint8(0); i < k; i++ {
		bitPos := hash % nBits
		if filter[bitPos/8]&(1<<(byte(bitPos%8))) == 0 {
			return false
		}

		hash += delta
	}

	return true
}

func extend(dst []byte, need int) []byte {
	var ans []byte
	if cap(dst) < len(dst)+need {
		ans = make([]byte, len(dst)+need)
		copy(ans, dst)
	} else {
		ans = dst[:len(dst)+need]
	}

	return ans
}

func Hash(b []byte) uint32 {
	const (
		seed = 0xbc9f1d34
		m    = 0xc6a4a793
	)
	h := uint32(seed) ^ uint32(len(b))*m
	for ; len(b) >= 4; b = b[4:] {
		h += uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
		h *= m
		h ^= h >> 16
	}
	switch len(b) {
	case 3:
		h += uint32(b[2]) << 16
		fallthrough
	case 2:
		h += uint32(b[1]) << 8
		fallthrough
	case 1:
		h += uint32(b[0])
		h *= m
		h ^= h >> 24
	}

	return h
}
