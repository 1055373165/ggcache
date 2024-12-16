// Package cache implements a distributed cache system with various features.
package cache

// ByteView holds an immutable view of bytes.
type ByteView struct {
	b []byte // Actual bytes stored
}

// Len returns the view's length.
func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice returns a copy of the data as a byte slice.
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// String returns the data as a string, making a copy if necessary.
func (v ByteView) String() string {
	return string(v.b)
}

// Bytes returns the underlying byte slice.
// Note: The returned slice should not be modified.
func (v ByteView) Bytes() []byte {
	return v.b
}

// cloneBytes returns a copy of the input byte slice.
// If the input is nil, it returns nil.
func cloneBytes(b []byte) []byte {
	if b == nil {
		return nil
	}
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
