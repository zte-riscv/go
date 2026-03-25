// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "asm_riscv64.h"
#include "go_asm.h"
#include "textflag.h"

TEXT ·CountString<ABIInternal>(SB),NOSPLIT|NOFRAME,$0
	// X10 = s_base
	// X11 = s_len
	// X12 = byte to count
	MOV	X12, X13 // byte to Count wanted in X13
	JMP	·Count<ABIInternal>(SB)

TEXT ·Count<ABIInternal>(SB),NOSPLIT|NOFRAME,$0
	// X10 = b_base
	// X11 = b_len
	// X12 = b_cap (unused)
	// X13 = byte to count
	MOV	X10, X14	// src pointer
	MOV	ZERO, X10	// reset counter
	AND	$0xff, X13	// make sure it's a byte to compare
#ifndef EnableSmallSizeMemVector
	SUB	$8, X11, X5
	BLEZ	X5, count_scalar
#endif
#ifndef hasV
	MOVB	internal∕cpu·RISCV64+const_offsetRISCV64HasV(SB), X5
	BEQZ	X5, count_scalar
#endif

#ifdef EnableSmallSizeMemVector
	PCALIGN	$16
count_vector_dispatch:
	// X6 = vlen bytes for E8,M1 (base 16 bytes for VLEN=128)
	MOV	$16, X6
#ifdef VLen_256
	SLLI	$1, X6
#endif
#ifdef VLen_512
	SLLI	$2, X6
#endif

	// 1..vlen bytes
	BGEU	X6, X11, count_vector_m1
	// (vlen+1)..(2*vlen) bytes
	SLLI	$1, X6
	BGEU	X6, X11, count_vector_m2
	// (2*vlen+1)..(4*vlen) bytes
	SLLI	$1, X6
	BGEU	X6, X11, count_vector_m4
#endif

	PCALIGN	$16
count_vector_loop:
	VSETVLI	X11, E8, M8, TA, MA, X5
	VLE8V	(X14), V8
	VMSEQVX	X13, V8, V1
	VCPOPM	V1, X15
	ADD	X15, X10	// add counter
	ADD	X5, X14
	SUB	X5, X11
	BEQZ	X11, done
	JMP	count_vector_loop

#ifdef EnableSmallSizeMemVector
	// Count 1..vlen bytes
	PCALIGN	$16
count_vector_m1:
	VSETVLI	X11, E8, M1, TA, MA, X5
	VLE8V	(X14), V8
	VMSEQVX	X13, V8, V1
	VCPOPM	V1, X15
	ADD	X15, X10
	RET

	// Count (vlen+1)..(2*vlen) bytes
	PCALIGN	$16
count_vector_m2:
	VSETVLI	X11, E8, M2, TA, MA, X5
	VLE8V	(X14), V8
	VMSEQVX	X13, V8, V1
	VCPOPM	V1, X15
	ADD	X15, X10
	RET

	// Count (2*vlen+1)..(4*vlen) bytes
	PCALIGN	$16
count_vector_m4:
	VSETVLI	X11, E8, M4, TA, MA, X5
	VLE8V	(X14), V8
	VMSEQVX	X13, V8, V1
	VCPOPM	V1, X15
	ADD	X15, X10
	RET
#endif

#ifndef EnableSmallSizeMemVector
	PCALIGN	$16
count_scalar:
	ADD	X14, X11	// end pointer
	PCALIGN	$16
count_scalar_loop:
	BEQ	X14, X11, done
	MOVBU	(X14), X15
	ADD	$1, X14
	BNE	X13, X15, count_scalar_loop
	ADD	$1, X10
	JMP	count_scalar_loop
#endif

done:
	RET
