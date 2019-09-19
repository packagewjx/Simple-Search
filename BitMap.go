package simple_search

type BitMap struct {
	bits []byte
}

func NewBitMap(initialCapacity int) *BitMap {
	if initialCapacity <= 0 {
		panic("容量有问题！")
	}
	return &BitMap{bits: make([]byte, initialCapacity>>3+1)}
}

var boolVal = []bool{false, true}

func (m *BitMap) Get(pos int) bool {
	if pos < 0 {
		panic("位置错误")
	}
	bytePos := pos >> 3
	// 扩容操作
	if bytePos >= len(m.bits) {
		n := make([]byte, bytePos*2)
		copy(n, m.bits)
		m.bits = n
		// 新的一定是空的，因此返回false即可
		return false
	}
	bitPos := pos & 7
	temp := m.bits[bytePos]
	temp = temp >> uint(bitPos) & 1
	return boolVal[temp]
}

func (m *BitMap) Set(pos int, val bool) {
	if pos < 0 {
		panic("位置错误")
	}
	bytePos := pos >> 3
	// 扩容操作
	if bytePos >= len(m.bits) {
		n := make([]byte, bytePos*2)
		copy(n, m.bits)
		m.bits = n
	}
	bitPos := pos & 7
	if val {
		m.bits[bytePos] = m.bits[bytePos] | (1 << uint(bitPos))
	} else {
		m.bits[bytePos] = m.bits[bytePos] & (1<<uint(bitPos) ^ 0xFF)
	}
}
