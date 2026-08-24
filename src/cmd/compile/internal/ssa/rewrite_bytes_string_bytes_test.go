// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import (
	"bytes"
	"fmt"
	"runtime"
	"testing"
)

var sinkBytes []byte

// TestRewriteBytesStringBytesOptimization tests that []byte(string([]byte))
// optimization produces correct results.
func TestRewriteBytesStringBytesOptimization(t *testing.T) {
	medium := make([]byte, 100)
	for i := range medium {
		medium[i] = byte(i % 256)
	}
	large := make([]byte, 10000)
	for i := range large {
		large[i] = byte(i % 256)
	}

	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"single", []byte{42}},
		{"small", []byte{1, 2, 3, 4, 5}},
		{"medium", medium},
		{"large", large},
		{"with_zero", []byte{0, 1, 2, 0, 3, 4}},
		{"all_zeros", make([]byte, 50)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := []byte(string(tt.data))
			if len(result) != len(tt.data) {
				t.Errorf("length mismatch: expected %d, got %d", len(tt.data), len(result))
			}

			if !bytes.Equal(result, tt.data) {
				t.Errorf("content mismatch: expected %v, got %v", tt.data, result)
			}

			if len(tt.data) > 0 && &result[0] == &tt.data[0] {
				t.Errorf("result and original data share the same underlying array, which is not expected copy behavior")
			}
		})
	}
}

// TestRewriteBytesStringBytesModification tests that modifying the original
// data does not affect the result (ensuring it's a copy).
func TestRewriteBytesStringBytesModification(t *testing.T) {
	original := []byte{1, 2, 3, 4, 5}
	result := []byte(string(original))

	original[0] = 99
	if result[0] == 99 {
		t.Errorf("modifying original data affected the result, indicating it's not a deep copy")
	}

	if result[0] != 1 {
		t.Errorf("result should maintain original value 1, but got %d", result[0])
	}
}

// TestRewriteBytesStringBytesNil tests nil and empty slice handling.
func TestRewriteBytesStringBytesNil(t *testing.T) {
	var nilSlice []byte
	result := []byte(string(nilSlice))
	if result != nil && len(result) != 0 {
		t.Errorf("nil slice should be converted to empty slice, but got %v", result)
	}

	empty := []byte{}
	result2 := []byte(string(empty))
	if len(result2) != 0 {
		t.Errorf("empty slice should remain empty, but got length %d", len(result2))
	}
}

// BenchmarkBytesStringBytesSubslice measures the cost of repeated
// []byte(string([]byte)) conversions on a subslice (e.g. s[4:l]).
func BenchmarkBytesStringBytesSubslice(b *testing.B) {
	// Large backing array; we'll always convert a subslice of it.
	buf := make([]byte, (1<<20)+64)
	for i := range buf {
		buf[i] = byte(i)
	}

	const off = 4
	const n = 4 << 10

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkBytes = []byte(string(buf[off : off+n]))
	}
	runtime.KeepAlive(buf)
}

// BenchmarkBytesStringBytesSubsliceSizes measures the same conversion
// across different subslice sizes.
func BenchmarkBytesStringBytesSubsliceSizes(b *testing.B) {
	sizes := []int{8, 32, 128, 512, 2 << 10, 8 << 10, 32 << 10, 128 << 10}
	max := 0
	for _, n := range sizes {
		if n > max {
			max = n
		}
	}

	buf := make([]byte, max+64)
	for i := range buf {
		buf[i] = byte(i)
	}

	const off = 4
	for _, n := range sizes {
		n := n
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				sinkBytes = []byte(string(buf[off : off+n]))
			}
			runtime.KeepAlive(buf)
		})
	}
}
