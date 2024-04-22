package service

/*
1. ensure data immutability
2. effectively avoid unnecessary data copying
3. encapsulates the processing logic of byte slices and provides some practical methods,
such as obtaining the byte length, returning a copy of the byte slice, etc.
4. the caller does not need to care about the specific storage and management method of data,
and only needs to operate the data through the interface provided by ByteView.
*/
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

// Implement the interfaces.Value interface so that it can be used as a cached value
func (bv ByteView) Len() int {
	return len(bv.b)
}
