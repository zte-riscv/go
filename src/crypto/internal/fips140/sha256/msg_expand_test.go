// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build riscv64

package sha256

import (
	"bytes"
	"testing"
)

// TestMessageExpansion compares the output of the generic Go implementation
// with the RISC-V optimized implementation (when available).
func TestMessageExpansion(t *testing.T) {
	// Test cases: various 64-byte message blocks
	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "zero block",
			data: make([]byte, 64),
		},
		{
			name: "all ones",
			data: bytes.Repeat([]byte{0xff}, 64),
		},
		{
			name: "incremental",
			data: func() []byte {
				b := make([]byte, 64)
				for i := range b {
					b[i] = byte(i)
				}
				return b
			}(),
		},
		{
			name: "SHA-256 test vector 1",
			data: []byte{
				0x61, 0x62, 0x63, 0x64, 0x62, 0x63, 0x64, 0x65,
				0x63, 0x64, 0x65, 0x66, 0x64, 0x65, 0x66, 0x67,
				0x65, 0x66, 0x67, 0x68, 0x66, 0x67, 0x68, 0x69,
				0x67, 0x68, 0x69, 0x6a, 0x68, 0x69, 0x6a, 0x6b,
				0x69, 0x6a, 0x6b, 0x6c, 0x6a, 0x6b, 0x6c, 0x6d,
				0x6b, 0x6c, 0x6d, 0x6e, 0x6c, 0x6d, 0x6e, 0x6f,
				0x6d, 0x6e, 0x6f, 0x70, 0x6e, 0x6f, 0x70, 0x71,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
		},
		{
			name: "random pattern",
			data: []byte{
				0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0,
				0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88,
				0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11,
				0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99,
				0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11,
				0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99,
				0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11,
				0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.data) != 64 {
				t.Fatalf("test case %s: data must be exactly 64 bytes, got %d", tc.name, len(tc.data))
			}

			// Get reference output from generic implementation
			// We'll call the function directly from msg_expand.go
			var reference [64]uint32
			messageExpansionGeneric(tc.data, &reference)

			// Get output from RISC-V optimized implementation
			// This will call messageExpansionRISCV64 if available, or fall back to generic
			var optimized [64]uint32
			messageExpansionRISCV64(tc.data, &optimized)

			// Compare outputs
			if !bytes.Equal(uint32ArrayToBytes(reference), uint32ArrayToBytes(optimized)) {
				t.Errorf("test case %s: outputs differ", tc.name)
				t.Logf("Reference W[0-15]: %x", reference[0:16])
				t.Logf("Optimized W[0-15]: %x", optimized[0:16])
				t.Logf("Reference W[16-31]: %x", reference[16:32])
				t.Logf("Optimized W[16-31]: %x", optimized[16:32])
				t.Logf("Reference W[32-47]: %x", reference[32:48])
				t.Logf("Optimized W[32-47]: %x", optimized[32:48])
				t.Logf("Reference W[48-63]: %x", reference[48:64])
				t.Logf("Optimized W[48-63]: %x", optimized[48:64])

				// Find first difference
				for i := 0; i < 64; i++ {
					if reference[i] != optimized[i] {
						t.Errorf("First difference at W[%d]: reference=0x%08x, optimized=0x%08x", i, reference[i], optimized[i])
						break
					}
				}
			}
		})
	}
}

// Helper function to convert uint32 array to byte slice for comparison
func uint32ArrayToBytes(arr [64]uint32) []byte {
	result := make([]byte, 64*4)
	for i, v := range arr {
		result[i*4+0] = byte(v >> 24)
		result[i*4+1] = byte(v >> 16)
		result[i*4+2] = byte(v >> 8)
		result[i*4+3] = byte(v)
	}
	return result
}
