package simple_search

import (
	"github.com/twmb/murmur3"
	"hash/adler32"
	"hash/fnv"
	"math/rand"
	"testing"
)

func TestBloomFilter(t *testing.T) {
	hf := []HashFunc{
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

	filter := NewBloomFilter(1<<15, hf)
	// 生成随机的字符串
	size := 1000
	input := make([][]byte, size)
	for i := 0; i < len(input); i++ {
		msg := make([]byte, size)
		for j := 0; j < len(msg); j++ {
			msg[j] = byte(rand.Intn(128))
		}
		input[i] = msg
		filter.Add(msg)
	}

	for i := 0; i < len(input); i++ {
		if !filter.Contains(input[i]) {
			t.Error("不包含", input[i])
		}
	}

}
