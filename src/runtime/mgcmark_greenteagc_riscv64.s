// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// func extractHeapBitsSmall(hbits *byte, spanBase, addr, elemsize uintptr) uintptr
//
// On riscv64, goarch.PtrSize = 8 and ptrBits = 64, so we hardcode
// the following simplifications:
//   i   = (addr - spanBase) / 8 / 64 = diff >> 9
//   j   = (addr - spanBase) / 8 % 64 = (diff >> 3) & 63
//   bits = elemsize / 8 = elemsize >> 3
//   word0 address = hbits + 8*(i) = hbits + (diff >> 6)
//   word1 address = hbits + (diff >> 6) + 8
//
// Inputs:  A0=hbits, A1=spanBase, A2=addr, A3=elemsize
// Output:  A0=heap bits
TEXT runtime·extractHeapBitsSmall<ABIInternal>(SB), NOSPLIT|NOFRAME, $0-32

	// diff = addr - spanBase
	SUB	A1, A2			// A2 = diff

	// j = (diff >> 3) & 63
	SRLI	$3, A2, T0		// T0 = diff >> 3
	ANDI	$63, T0, T1		// T1 = j

	// bits = elemsize >> 3
	SRLI	$3, A3, T2		// T2 = bits

	// word0_ptr = hbits + 8*(diff>>9) = hbits + (diff>>9)<<3
	SRLI	$9, A2, T3		// T3 = diff >> 9 = word index
	SLLI	$3, T3, T3		// T3 = 8 * (diff >> 9)
	ADD	A0, T3			// T3 = hbits + 8*i = &word0

	// Check: j + bits > 64 ?
	ADD	T1, T2, T0		// T0 = j + bits
	ADDI	$-65, T0		// T0 = j + bits - 65
	BGEZ	T0, two_read		// if j + bits >= 65, two reads needed

one_read:
	// read = (*word0 >> j) & ((1 << bits) - 1)
	MOV	(T3), T4		// T4 = *word0
	SRL	T1, T4, T5		// T5 = *word0 >> j

	// If bits == 64, j must be 0 and mask is all ones; skip masking
	ADDI	$-64, T2, T4		// T4 = bits - 64
	BEQZ	T4, one_read_done

	// mask = (1 << bits) - 1
	MOV	$1, T4
	SLL	T2, T4, T4		// T4 = 1 << bits
	ADDI	$-1, T4		// T4 = (1 << bits) - 1
	AND	T4, T5			// T5 = (*word0 >> j) & mask

one_read_done:
	MOV	T5, A0			// A0 = result (return register)
	JMP	done

two_read:
	// bits0 = 64 - j
	MOV	$64, T4
	SUB	T1, T4			// T4 = 64 - j = bits0

	// bits1 = bits - bits0
	SUB	T4, T2, T5		// T5 = bits1

	// read = *word0 >> j
	MOV	(T3), A2		// A2 = *word0
	SRL	T1, A2, A0		// A0 = *word0 >> j

	// Advance to word1 pointer
	ADD	$8, T3			// T3 = &word1

	// (*word1 & ((1 << bits1) - 1)) << bits0
	MOV	(T3), A2		// A2 = *word1
	MOV	$1, A4
	SLL	T5, A4, A4		// A4 = 1 << bits1
	ADDI	$-1, A4		// A4 = (1 << bits1) - 1
	AND	A4, A2, A2		// A2 = *word1 & mask1
	SLL	T4, A2, A2		// A2 = (*word1 & mask1) << bits0

	// read |= (*word1 & mask1) << bits0
	OR	A2, A0			// A0 = A0 | A2

done:
	RET
