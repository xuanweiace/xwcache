package xwcache

type ImmutableByte struct {
	b []byte
}

// 其实不需要带指针的，因为他的所有方法都不涉及修改
func (i *ImmutableByte) Len() int64 {
	return int64(len(i.b))
}

func (i *ImmutableByte) GetByteSlice() []byte {
	return cloneByteSlice(i.b)
}
func (v ImmutableByte) String() string {
	return string(v.b)
}

func cloneByteSlice(b []byte) (c []byte) {
	c = make([]byte, len(b)) //注意这里必须make
	copy(c, b)
	return
}
