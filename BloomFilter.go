package simple_search

import (
	"github.com/twmb/murmur3"
	"hash/adler32"
	"hash/fnv"
	"sync"
	"sync/atomic"
)

type HashFunc func(msg []byte) int64

type BloomFilter struct {
	// 用于取余的掩码。取余之后，得到的余数就是bitmap的pos值。
	bmMask int64
	bitmap *BitMap
	// 线程安全的哈希函数集合
	hashFunc []HashFunc
}

func DefaultHashFunc() []HashFunc {
	return []HashFunc{
		func(msg []byte) int64 {
			h := murmur3.New128()
			h.Write(msg)
			sum := h.Sum(nil)
			result := int64(0)
			return result | (int64(sum[0]) << 56) | (int64(sum[1]) << 48) | (int64(sum[2]) << 40) | (int64(sum[3]) << 32) |
				(int64(sum[4] << 24)) | (int64(sum[5]) << 16) | (int64(sum[6]) << 8) | (int64(sum[7]))
		},
		func(msg []byte) int64 {
			h := murmur3.New128()
			h.Write(msg)
			sum := h.Sum(nil)
			result := int64(0)
			return result | (int64(sum[8]) << 56) | (int64(sum[9]) << 48) | (int64(sum[10]) << 40) | (int64(sum[11]) << 32) |
				(int64(sum[12] << 24)) | (int64(sum[13]) << 16) | (int64(sum[14]) << 8) | (int64(sum[15]))
		},
		func(msg []byte) int64 {
			a32 := adler32.New()
			a32.Write(msg)
			sum := a32.Sum(nil)
			result := int64(0)
			return result | (int64(sum[0]) << 24) | (int64(sum[1]) << 16) | (int64(sum[2]) << 8) | (int64(sum[3]))
		},
		func(msg []byte) int64 {
			f := fnv.New128()
			f.Write(msg)
			sum := f.Sum(nil)
			result := int64(0)
			return result | (int64(sum[0]) << 56) | (int64(sum[1]) << 48) | (int64(sum[2]) << 40) | (int64(sum[3]) << 32) |
				(int64(sum[4] << 24)) | (int64(sum[5]) << 16) | (int64(sum[6]) << 8) | (int64(sum[7]))
		},
		func(msg []byte) int64 {
			f := fnv.New128()
			f.Write(msg)
			sum := f.Sum(nil)
			result := int64(0)
			return result | (int64(sum[8]) << 56) | (int64(sum[9]) << 48) | (int64(sum[10]) << 40) | (int64(sum[11]) << 32) |
				(int64(sum[12] << 24)) | (int64(sum[13]) << 16) | (int64(sum[14]) << 8) | (int64(sum[15]))
		},
	}
}

func NewBloomFilter(size int, hashFunc []HashFunc) *BloomFilter {
	if hashFunc == nil || len(hashFunc) == 0 || size <= 0 {
		panic("参数不正确")
	}

	bmSize := int64(1)
	for size > 0 {
		size >>= 1
		bmSize <<= 1
	}
	mask := bmSize - 1
	bm := NewBitMap(int(bmSize))
	return &BloomFilter{
		bmMask:   mask,
		bitmap:   bm,
		hashFunc: hashFunc,
	}
}

func (bf *BloomFilter) Add(msg []byte) {
	wg := sync.WaitGroup{}
	for i := 0; i < len(bf.hashFunc); i++ {
		wg.Add(1)
		go func(index int) {
			pos := bf.hashFunc[index](msg) & bf.bmMask
			bf.bitmap.Set(int(pos), true)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func (bf *BloomFilter) Contains(msg []byte) bool {
	count := int32(0)
	wg := sync.WaitGroup{}
	for i := 0; i < len(bf.hashFunc); i++ {
		wg.Add(1)
		go func(index int) {
			pos := bf.hashFunc[index](msg) & bf.bmMask
			if bf.bitmap.Get(int(pos)) {
				atomic.AddInt32(&count, 1)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	return count == int32(len(bf.hashFunc))
}
