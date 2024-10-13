package internal

type ByteView struct {
	b []byte
}

func (bv ByteView) Bytes() []byte {
	return cloneBytes(bv.b)
}

func cloneBytes(b []byte) []byte {
	newB := make([]byte, len(b))
	copy(newB, b)
	return newB
}

func (bv ByteView) String() string {
	return string(bv.b)
}

// 实现 Value 接口
func (bv ByteView) Len() int {
	return len(bv.b)
}
