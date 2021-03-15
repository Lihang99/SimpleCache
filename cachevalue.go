package SimpleCache

type CacheValue struct {
	b []byte
}

func (v CacheValue) Len() int {
	return len(v.b)
}

func (v CacheValue) ByteSlice() []byte {
	//如果不复制一份 将会直接导致原始切片的修改
	return cloneBytes(v.b)
}

func (v CacheValue) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
