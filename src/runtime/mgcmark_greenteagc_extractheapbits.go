// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.greenteagc && !riscv64 || purego

package runtime

import (
	"internal/goarch"
	"unsafe"
)

func extractHeapBitsSmall(hbits *byte, spanBase, addr, elemsize uintptr) uintptr {
	// These objects are always small enough that their bitmaps
	// fit in a single word, so just load the word or two we need.
	//
	// Mirrors mspan.writeHeapBitsSmall.
	//
	// We should be using heapBits(), but unfortunately it introduces
	// both bounds checks panics and throw which causes us to exceed
	// the nosplit limit in quite a few cases.
	i := (addr - spanBase) / goarch.PtrSize / ptrBits
	j := (addr - spanBase) / goarch.PtrSize % ptrBits
	bits := elemsize / goarch.PtrSize
	word0 := (*uintptr)(unsafe.Pointer(addb(hbits, goarch.PtrSize*(i+0))))
	word1 := (*uintptr)(unsafe.Pointer(addb(hbits, goarch.PtrSize*(i+1))))

	var read uintptr
	if j+bits > ptrBits {
		// Two reads.
		bits0 := ptrBits - j
		bits1 := bits - bits0
		read = *word0 >> j
		read |= (*word1 & ((1 << bits1) - 1)) << bits0
	} else {
		// One read.
		read = (*word0 >> j) & ((1 << bits) - 1)
	}
	return read
}
