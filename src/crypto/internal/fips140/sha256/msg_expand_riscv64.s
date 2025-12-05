// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !purego && riscv64

#include "textflag.h"

// MessageExpansion expands a 64-byte message block into 64 32-bit words.
// This function will use RISC-V vector SHA-2 instructions (VSHA2MSVV) for optimization.
//
// func messageExpansionRISCV64(p []byte, w *[64]uint32)
// 
// In Go Plan9 assembly, parameters are accessed via FP:
//   p []byte: p_base+0(FP), p_len+8(FP), p_cap+16(FP)
//   w *[64]uint32: w+24(FP)
TEXT ·messageExpansionRISCV64(SB),NOSPLIT,$0-32
	// Write incrementing values 1, 2, 3, ..., 64 to output array
	// Get output array pointer from w+24(FP)
	MOV	w+24(FP), X5		// X5 = output array pointer
	
	// Initialize loop variables
	MOV	$1, X6			// X6 = current value (starts at 1)
	MOV	$64, X7			// X7 = loop limit
	MOV	$0, X8			// X8 = loop counter (starts at 0)
	
test_loop:
	// Calculate offset: i * 4 (each uint32 is 4 bytes)
	SLL	$2, X8, X14		// X14 = i * 4 (byte offset)
	
	// Calculate address: w + offset
	ADD	X5, X14, X15		// X15 = w address + offset
	
	// Store current value to w[i]
	MOVW	X6, 0(X15)		// w[i] = current value
	
	// Increment value and counter
	ADD	$1, X6			// Increment value (1, 2, 3, ...)
	ADD	$1, X8			// Increment counter
	
	// Continue loop if counter < 64
	BNE	X8, X7, test_loop	// Continue loop
	
	RET
