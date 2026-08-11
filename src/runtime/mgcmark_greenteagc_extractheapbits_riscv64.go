// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.greenteagc && riscv64 && !purego

package runtime

// Implemented in assembly in mgcmark_greenteagc_riscv64.s
func extractHeapBitsSmall(hbits *byte, spanBase, addr, elemsize uintptr) uintptr
