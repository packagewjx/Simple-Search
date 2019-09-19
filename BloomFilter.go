package simple_search

import (
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
