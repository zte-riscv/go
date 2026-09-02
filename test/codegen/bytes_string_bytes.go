// asmcheck -gcflags=-bytesstringbytesopt

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

// This file contains code generation tests for the optimization
// that rewrites []byte(string([]byte)) to use runtime.makeslicecopy
// instead of runtime.slicebytetostring + runtime.stringtoslicebyte.

// BytesStringBytesOptimization tests that []byte(string([]byte))
// is optimized to use makeslicecopy instead of the two-step conversion.
func BytesStringBytesOptimization(data []byte) []byte {
	// Optimized: should use makeslicecopy directly
	// riscv64:`JAL.*runtime\.makeslicecopy`
	// riscv64:-`.*runtime\.slicebytetostring`
	// riscv64:-`.*runtime\.stringtoslicebyte`
	return []byte(string(data))
}

// BytesStringBytesSmall tests small slices.
func BytesStringBytesSmall() []byte {
	data := []byte{1, 2, 3, 4, 5}
	// riscv64:`JAL.*runtime\.makeslicecopy`
	// riscv64:-`.*runtime\.slicebytetostring`
	// riscv64:-`.*runtime\.stringtoslicebyte`
	return []byte(string(data))
}

// BytesStringBytesEmpty tests empty slices.
func BytesStringBytesEmpty() []byte {
	data := []byte{}
	// riscv64:`JAL.*runtime\.makeslicecopy`
	// riscv64:-`.*runtime\.slicebytetostring`
	// riscv64:-`.*runtime\.stringtoslicebyte`
	return []byte(string(data))
}

// BytesStringBytesLarge tests large slices.
func BytesStringBytesLarge() []byte {
	data := make([]byte, 1000)
	// riscv64:`JAL.*runtime\.makeslicecopy`
	// riscv64:-`.*runtime\.slicebytetostring`
	// riscv64:-`.*runtime\.stringtoslicebyte`
	return []byte(string(data))
}
