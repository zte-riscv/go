// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import "testing"

// Benchmark for Zbs extension optimizations

// Bit Extract (BEXT) - variable shift
func BenchmarkBitExtract(b *testing.B) {
	const N = 64 * 8
	a := make([]uint64, N/64)
	for i := 0; i < b.N; i++ {
		var s uint64
		for j := uint64(0); j < N; j++ {
			s += (a[j/64] >> (j % 64)) & 1
		}
		globl = int64(s)
	}
}

// Bit Extract (BEXTI) - constant shift
func BenchmarkBitExtractConst(b *testing.B) {
	const N = 64
	a := make([]uint64, N)
	for i := 0; i < b.N; i++ {
		var s uint64
		for j := range a {
			s += (a[j] >> 37) & 1
		}
		globl = int64(s)
	}
}

// Bit Extract with mask (BEXTI) - (x>>c)&mask pattern
func BenchmarkBitExtractMask(b *testing.B) {
	const N = 64
	a := make([]uint64, N)
	for i := 0; i < b.N; i++ {
		var s uint64
		for j := range a {
			s += (a[j] >> 5) & 4 // Should optimize to BEXTI [7]
		}
		globl = int64(s)
	}
}

// Bit test with SNEZ - if x&mask != 0
//BEXT
func BenchmarkBitTest(b *testing.B) {
	const N = 64 * 8
	a := make([]uint64, N/64)
	for i := 0; i < b.N; i++ {
		var count int
		for j := uint64(0); j < N; j++ {
			if a[j/64]&(1<<(j%64)) != 0 {
				count++
			}
		}
		globl = int64(count)
	}
}

// Bit test with SEQZ - if x&mask == 0
//BEXT
func BenchmarkBitTestZero(b *testing.B) {
	const N = 64 * 8
	a := make([]uint64, N/64)
	for i := 0; i < b.N; i++ {
		var count int
		for j := uint64(0); j < N; j++ {
			if a[j/64]&(1<<(j%64)) == 0 {
				count++
			}
		}
		globl = int64(count)
	}
}

// Bit test with constant mask
//BEXTI
func BenchmarkBitTestConst(b *testing.B) {
	const N = 64
	a := make([]uint64, N)
	for i := 0; i < b.N; i++ {
		var count int
		for j := range a {
			if a[j]&(1<<37) != 0 {
				count++
			}
		}
		globl = int64(count)
	}
}

// 32-bit Bit Set (BSET) - variable shift
func BenchmarkBitSet32(b *testing.B) {
	const N = 32 * 8
	a := make([]uint32, N/32)
	for i := 0; i < b.N; i++ {
		for j := uint32(0); j < N; j++ {
			a[j/32] |= 1 << (j % 32)
		}
	}
}

// 32-bit Bit Clear (BCLR) - variable shift
func BenchmarkBitClear32(b *testing.B) {
	const N = 32 * 8
	a := make([]uint32, N/32)
	for i := 0; i < b.N; i++ {
		for j := uint32(0); j < N; j++ {
			a[j/32] &^= 1 << (j % 32)
		}
	}
}

// 32-bit Bit Toggle (BINV) - variable shift
func BenchmarkBitToggle32(b *testing.B) {
	const N = 32 * 8
	a := make([]uint32, N/32)
	for i := 0; i < b.N; i++ {
		for j := uint32(0); j < N; j++ {
			a[j/32] ^= 1 << (j % 32)
		}
	}
}

// 32-bit Bit Extract (BEXT) - variable shift
func BenchmarkBitExtract32(b *testing.B) {
	const N = 32 * 8
	a := make([]uint32, N/32)
	for i := 0; i < b.N; i++ {
		var s uint32
		for j := uint32(0); j < N; j++ {
			s += (a[j/32] >> (j % 32)) & 1
		}
		globl32 = int32(s)
	}
}

// 32-bit Bit Extract (BEXTI) - constant shift
func BenchmarkBitExtractConst32(b *testing.B) {
	const N = 64
	a := make([]uint32, N)
	for i := 0; i < b.N; i++ {
		var s uint32
		for j := range a {
			s += (a[j] >> 20) & 1
		}
		globl32 = int32(s)
	}
}

// Immediate bit operations - ORI/ANDI/XORI optimizations
//BSETI
func BenchmarkBitSetImmediate(b *testing.B) {
	const N = 64
	a := make([]uint64, N)
	for i := 0; i < b.N; i++ {
		for j := range a {
			a[j] |= 1024 // 1<<10, should use BSETI
		}
	}
}
//BCLRI
func BenchmarkBitClearImmediate(b *testing.B) {
	const N = 64
	a := make([]uint64, N)
	for i := 0; i < b.N; i++ {
		for j := range a {
			a[j] &^= 1024 // Should use BCLRI
		}
	}
}
//BINVI
func BenchmarkBitToggleImmediate(b *testing.B) {
	const N = 64
	a := make([]uint64, N)
	for i := 0; i < b.N; i++ {
		for j := range a {
			a[j] ^= 1024 // Should use BINVI
		}
	}
}

// Cascaded pattern - (BEXTI [c] (SRLI [d] x))
//BEXTI
func BenchmarkBitExtractCascaded(b *testing.B) {
	const N = 64
	a := make([]uint64, N)
	for i := 0; i < b.N; i++ {
		var s uint64
		for j := range a {
			// (a[j]>>5)&4 should merge to BEXTI [7]
			s += (a[j] >> 5) & 4
		}
		globl = int64(s)
	}
}

// Branch optimization - BEQZ/BNEZ with bit test
//BEXT
func BenchmarkBitTestBranch(b *testing.B) {
	const N = 64 * 8
	a := make([]uint64, N/64)
	for i := 0; i < b.N; i++ {
		var count int
		for j := uint64(0); j < N; j++ {
			if a[j/64]&(1<<(j%64)) != 0 {
				count++
			}
		}
		globl = int64(count)
	}
}

// Shift-right and test - (x>>c)&1 pattern
//BEXTI
func BenchmarkShiftRightAndTest(b *testing.B) {
	const N = 64
	a := make([]uint64, N)
	for i := 0; i < b.N; i++ {
		var count int
		for j := range a {
			if (a[j]>>37)&1 != 0 {
				count++
			}
		}
		globl = int64(count)
	}
}

//BEXTI
// 32-bit shift-right and test
func BenchmarkShiftRightAndTest32(b *testing.B) {
	const N = 64
	a := make([]uint32, N)
	for i := 0; i < b.N; i++ {
		var count int
		for j := range a {
			if (a[j]>>20)&1 != 0 {
				count++
			}
		}
		globl = int64(count)
	}
}
