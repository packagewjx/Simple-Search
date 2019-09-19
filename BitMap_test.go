package simple_search

import "testing"

func TestBitMap(t *testing.T) {
	size := 1024
	bitmap := NewBitMap(size)
	for i := 0; i < size*2; i++ {
		bitmap.Set(i, true)
		if bitmap.Get(i) != true {
			t.Error("结果不对")
		}
		for j := 0; j < size; j++ {
			if i == j {
				continue
			}
			if bitmap.Get(j) != false {
				t.Error("位置不是false")
			}
		}
		bitmap.Set(i, false)
	}
}
