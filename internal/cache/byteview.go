// Package cache implements a distributed cache system with various features.
package cache

import "time"

// ByteView holds an immutable view of bytes.
type ByteView struct {
	b        []byte    // Actual bytes stored
	expireAt time.Time // 过期时间，零值表示永不过期
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

// IsExpired 检查值是否已过期
func (v ByteView) IsExpired() bool {
	// 零值时间表示永不过期
	return !v.expireAt.IsZero() && time.Now().After(v.expireAt)
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
