// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.greenteagc

package runtime_test

import (
	"internal/goarch"
	"math/rand"
	"runtime"
	"testing"
	"unsafe"
)

// extractHeapBitsSmallRef is a platform-independent reference for
// runtime.extractHeapBitsSmall, used to validate the riscv64 assembly
// implementation.
func extractHeapBitsSmallRef(hbits *byte, spanBase, addr, elemsize uintptr) uintptr {
	const ptrBits = 8 * goarch.PtrSize
	i := (addr - spanBase) / goarch.PtrSize / ptrBits
	j := (addr - spanBase) / goarch.PtrSize % ptrBits
	bits := elemsize / goarch.PtrSize
	word0 := (*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(hbits)) + goarch.PtrSize*i))
	word1 := (*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(hbits)) + goarch.PtrSize*(i+1)))

	var read uintptr
	if j+bits > ptrBits {
		bits0 := ptrBits - j
		bits1 := bits - bits0
		read = *word0 >> j
		read |= (*word1 & ((1 << bits1) - 1)) << bits0
	} else {
		read = (*word0 >> j) & ((1 << bits) - 1)
	}
	return read
}

// TestExtractHeapBitsSmall exercises both the one-read and two-read
// paths, the bits==64 special case, and multi-word indices, comparing
// against the reference implementation.
func TestExtractHeapBitsSmall(t *testing.T) {
	hbits := make([]byte, 128)
	for i := range hbits {
		hbits[i] = byte(i*13 + 7)
	}
	hb := &hbits[0]

	// spanBase only needs to be aligned to goarch.PtrSize.
	const spanBase = uintptr(1) << 20

	tests := []struct {
		name     string
		addr     uintptr
		elemsize uintptr
	}{
		// One-read path (condition: j+bits <= 64).
		{"j=0 bits=1", spanBase, 8},
		{"j=0 bits=8", spanBase, 64},
		{"j=1 bits=8", spanBase + 8, 64},
		{"j=8 bits=8", spanBase + 64, 64},
		{"j=55 bits=8", spanBase + 55*8, 64},   // 55+8=63 <= 64
		{"j=56 bits=8", spanBase + 56*8, 64},   // 56+8=64, exact boundary
		{"j=63 bits=1", spanBase + 63*8, 8},    // 63+1=64, exact boundary
		{"j=31 bits=32", spanBase + 31*8, 256}, // 31+32=63
		{"j=32 bits=32", spanBase + 32*8, 256}, // 32+32=64, exact boundary
		{"j=0 bits=64", spanBase, 512},         // bits==64, skip-mask special case

		// Two-read path (condition: j+bits > 64).
		{"j=33 bits=32", spanBase + 33*8, 256}, // 33+32=65, bits1=1
		{"j=57 bits=8", spanBase + 57*8, 64},   // 57+8=65, bits1=1
		{"j=60 bits=8", spanBase + 60*8, 64},
		{"j=63 bits=2", spanBase + 63*8, 16}, // 63+2=65, bits1=1
		{"j=63 bits=8", spanBase + 63*8, 64},
		{"j=32 bits=64", spanBase + 32*8, 512}, // j>0, bits=64 -> two reads
		{"j=1 bits=64", spanBase + 8, 512},
		{"j=63 bits=64", spanBase + 63*8, 512},
		{"j=40 bits=48", spanBase + 40*8, 384}, // 40+48=88 > 64

		// Multi-word indices (i > 0).
		{"i=1 j=0 bits=8", spanBase + 512, 64},
		{"i=1 j=1 bits=8", spanBase + 520, 64},
		{"i=1 j=60 bits=16", spanBase + 512 + 60*8, 128},
		{"i=1 j=63 bits=64", spanBase + 512 + 63*8, 512},
		{"i=2 j=0 bits=64", spanBase + 1024, 512},
		{"i=3 j=40 bits=24", spanBase + 3*512 + 40*8, 192},
		{"i=4 j=32 bits=64", spanBase + 4*512 + 32*8, 512},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.ExtractHeapBitsSmall(hb, spanBase, tt.addr, tt.elemsize)
			want := extractHeapBitsSmallRef(hb, spanBase, tt.addr, tt.elemsize)
			if got != want {
				t.Fatalf("extractHeapBitsSmall(hbits, 0x%x, 0x%x, %d) = 0x%x, want 0x%x",
					spanBase, tt.addr, tt.elemsize, got, want)
			}
		})
	}
}

// TestExtractHeapBitsSmallExplicit verifies known values against
// hand-computed expectations, independent of the reference
// implementation.
func TestExtractHeapBitsSmallExplicit(t *testing.T) {
	var buf [16]byte
	*(*uintptr)(unsafe.Pointer(&buf[0])) = 0x0123456789abcdef
	*(*uintptr)(unsafe.Pointer(&buf[8])) = 0x0f0f0f0f0f0f0f0f
	hb := (*byte)(unsafe.Pointer(&buf[0]))

	const spanBase = uintptr(1) << 20

	tests := []struct {
		name     string
		addr     uintptr
		elemsize uintptr
		want     uintptr
	}{
		// One read: (word0 >> 8) & 0xff = 0xcd.
		{"one_read j=8 bits=8", spanBase + 8*8, 64, 0xcd},
		// One read, whole low byte: word0 & 0xff = 0xef.
		{"one_read j=0 bits=8", spanBase, 64, 0xef},
		// One read, bits=64 special case: entire word0.
		{"one_read j=0 bits=64", spanBase, 512, 0x0123456789abcdef},
		// Two read: (word0 >> 60) | ((word1 & 0xf) << 4).
		// bits0 = 64-60 = 4, bits1 = 8-4 = 4.
		// word0>>60 = 0x0 (top nibble of 0x0123... is 0);
		// (0x0f & 0xf) << 4 = 0xf0.
		{"two_read j=60 bits=8", spanBase + 60*8, 64, 0xf0},
		// Two read: (word0 >> 63) | ((word1 & 0x7f) << 1).
		// bits0 = 64-63 = 1, bits1 = 8-1 = 7.
		// word0>>63 = 0x0 (top bit of 0x0123... is 0);
		// (0x0f & 0x7f) << 1 = 0x1e.
		{"two_read j=63 bits=8", spanBase + 63*8, 64, 0x1e},
		// Two read, bits1==1: (word0>>57) | ((word1&1)<<7).
		// word0>>57 = 0x0 (the 7 top bits of 0x0123... are all 0);
		// (0x0f & 1) << 7 = 0x80.
		{"two_read j=57 bits=8 bits1=1", spanBase + 57*8, 64, 0x80},
		// Two read, both parts 1 bit: (word0>>63) | ((word1&1)<<1).
		// word0>>63 = 0x0; (0x0f & 1) << 1 = 0x2.
		{"two_read j=63 bits=2 bits0=1 bits1=1", spanBase + 63*8, 16, 0x2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.ExtractHeapBitsSmall(hb, spanBase, tt.addr, tt.elemsize)
			if got != tt.want {
				t.Fatalf("extractHeapBitsSmall(hbits, 0x%x, 0x%x, %d) = 0x%x, want 0x%x",
					spanBase, tt.addr, tt.elemsize, got, tt.want)
			}
		})
	}
}

// TestExtractHeapBitsSmallAllOnes checks the bits==64, j==0 special
// case where the mask would be (1<<64)-1. The assembly skips masking
// entirely, and the result must be the full word.
func TestExtractHeapBitsSmallAllOnes(t *testing.T) {
	var buf [16]byte
	*(*uintptr)(unsafe.Pointer(&buf[0])) = ^uintptr(0)
	*(*uintptr)(unsafe.Pointer(&buf[8])) = ^uintptr(0)
	hb := (*byte)(unsafe.Pointer(&buf[0]))

	const spanBase = uintptr(1) << 20

	got := runtime.ExtractHeapBitsSmall(hb, spanBase, spanBase, 512)
	want := ^uintptr(0)
	if got != want {
		t.Fatalf("extractHeapBitsSmall(all-ones, j=0, bits=64) = 0x%x, want 0x%x", got, want)
	}
}

// TestExtractHeapBitsSmallRandom compares against the reference
// implementation over random inputs, covering both read paths.
func TestExtractHeapBitsSmallRandom(t *testing.T) {
	hbits := make([]byte, 128)
	hb := &hbits[0]

	const spanBase = uintptr(1) << 20

	// Fixed seed; the constant exceeds math.MaxInt64, so negate it.
	rng := rand.New(rand.NewSource(-0x61c8864680b583eb))
	for i := 0; i < 10000; i++ {
		for j := range hbits {
			hbits[j] = byte(rng.Uint64())
		}

		// bits in [1, 64]; elemsize is a multiple of 8.
		bits := uintptr(rng.Intn(64) + 1)
		elemsize := bits * 8

		// Cap diff so word0 and word1 stay within the 128-byte buffer:
		// word1 offset = 8*(diff/512+1) reaches 120 at diff = 512*14.
		maxDiff := uintptr(512 * 14)
		diff := uintptr(rng.Int63n(int64(maxDiff/8))) * 8
		addr := spanBase + diff

		got := runtime.ExtractHeapBitsSmall(hb, spanBase, addr, elemsize)
		want := extractHeapBitsSmallRef(hb, spanBase, addr, elemsize)
		if got != want {
			t.Fatalf("extractHeapBitsSmall(hbits, 0x%x, 0x%x, %d) = 0x%x, want 0x%x",
				spanBase, addr, elemsize, got, want)
		}
	}
}
