// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "asm_riscv64.h"
#include "go_asm.h"
#include "textflag.h"

// See memmove Go doc for important implementation constraints.

// func memmove(to, from unsafe.Pointer, n uintptr)
TEXT runtime·memmove<ABIInternal>(SB),NOSPLIT|NOFRAME,$0-24
	// X10 = to
	// X11 = from
	// X12 = n
	BEQ	X10, X11, done
	BEQZ	X12, done

	// If the destination is ahead of the source, start at the end of the
	// buffer and go backward.
	BGTU	X10, X11, backward

#ifndef EnableSmallSizeMemVector
	// If less than 8 bytes, do single byte copies.
	MOV	$8, X9
	BLT	X12, X9, f_loop4_check
#endif

#ifndef hasV
	MOVB	internal∕cpu·RISCV64+const_offsetRISCV64HasV(SB), X5
	BEQZ	X5, f_memmove_scalar
#endif

#ifndef EnableSmallSizeMemVector
	// Use vector if destination and source are not 8 byte aligned.
	OR	X10, X11, X5
	AND	$7, X5
	BNEZ	X5, f_vector_loop

	// Use scalar if destination and source are 8 byte aligned and n <= 64 bytes.
	SUB	$64, X12, X6
	BLEZ	X6, f_loop_check

	// Copy whole pointers with 64-bit elements so a concurrent GC never
	// observes a partially updated pointer.
	AND	$7, X12, X5
	BEQZ	X5, f_vector64_start
#endif

#ifdef EnableSmallSizeMemVector
	// Same for the small-size variant: aligned whole-pointer copies use
	// 64-bit elements regardless of size.
	OR	X10, X11, X5
	OR	X12, X5
	AND	$7, X5
	BEQZ	X5, f_vector64_start

	PCALIGN	$16
f_vector_dispatch:
	MOV		$16, X6
#ifdef VLen_256
	SLLI	$1, X6
#endif
#ifdef VLen_512
	SLLI	$2, X6
#endif
	BGEU	X6, X12, f_vector_single
	SLLI	$2, X6
	BGTU	X12, X6, f_vector_loop
	SRLI	$1, X6
	BGTU	X12, X6, f_vector_quarter

// Copy (vlen+1)..(2*vlen) bytes
	PCALIGN	$16
f_vector_double:
	VSETVLI	X12, E8, M2, TA, MA, X5
	VLE8V	(X11), V8
	VSE8V	V8, (X10)
	RET

// Copy 1..(vlen) bytes
	PCALIGN	$16
f_vector_single:
	VSETVLI	X12, E8, M1, TA, MA, X5
	VLE8V	(X11), V8
	VSE8V	V8, (X10)
	RET

// Copy (2*vlen+1)..(4*vlen) bytes
	PCALIGN	$16
f_vector_quarter:
	VSETVLI	X12, E8, M4, TA, MA, X5
	VLE8V	(X11), V8
	VSE8V	V8, (X10)
	RET
#endif

// Copy (4*vlen+1).. bytes
	PCALIGN	$16
f_vector_loop:
	VSETVLI	X12, E8, M8, TA, MA, X5
	VLE8V	(X11), V8
	VSE8V	V8, (X10)
	ADD	X5, X10
	ADD	X5, X11
	SUB	X5, X12
	BNEZ	X12, f_vector_loop
	RET

f_vector64_start:
	// Copy whole pointers with 64-bit elements. X12 counts words here.
	SRLI	$3, X12, X12

	VSETVLI	X12, E64, M8, TA, MA, X5
	SLLI	$3, X5, X6

	// Load a whole group before storing it, so an overlapping destination
	// inside this group cannot clobber the source before it is read.
	PCALIGN	$16
f_vector64_loop:
	VLE64V	(X11), V8
	VSE64V	V8, (X10)
	ADD	X6, X10
	ADD	X6, X11
	SUB	X5, X12
	BGEU	X12, X5, f_vector64_loop
	BEQZ	X12, done
	VSETVLI	X12, E64, M8, TA, MA, X5
	VLE64V	(X11), V8
	VSE64V	V8, (X10)
	RET

#ifndef hasV
f_memmove_scalar:
	// Check alignment - if alignment differs we have to do one byte at a time.
	AND	$7, X10, X5
	AND	$7, X11, X6
	BNE	X5, X6, f_loop8_unaligned_check
	BEQZ	X5, f_loop_check

	// Move one byte at a time until we reach 8 byte alignment.
	SUB	X5, X9, X5
	SUB	X5, X12, X12
f_align:
	SUB	$1, X5
	MOVB	0(X11), X14
	MOVB	X14, 0(X10)
	ADD	$1, X10
	ADD	$1, X11
	BNEZ	X5, f_align
#endif

#ifndef EnableSmallSizeMemVector
f_loop_check:
	MOV	$16, X9
	BLT	X12, X9, f_loop8_check
	MOV	$32, X9
	BLT	X12, X9, f_loop16_check
	MOV	$64, X9
	BLT	X12, X9, f_loop32_check

	PCALIGN	$16
f_loop64:
	MOV	0(X11), X14
	MOV	8(X11), X15
	MOV	16(X11), X16
	MOV	24(X11), X17
	MOV	32(X11), X18
	MOV	40(X11), X19
	MOV	48(X11), X20
	MOV	56(X11), X21
	MOV	X14, 0(X10)
	MOV	X15, 8(X10)
	MOV	X16, 16(X10)
	MOV	X17, 24(X10)
	MOV	X18, 32(X10)
	MOV	X19, 40(X10)
	MOV	X20, 48(X10)
	MOV	X21, 56(X10)
	ADD	$64, X10
	ADD	$64, X11
	SUB	$64, X12
	BGE	X12, X9, f_loop64
	BEQZ	X12, done

f_loop32_check:
	MOV	$32, X9
	BLT	X12, X9, f_loop16_check
f_loop32:
	MOV	0(X11), X14
	MOV	8(X11), X15
	MOV	16(X11), X16
	MOV	24(X11), X17
	MOV	X14, 0(X10)
	MOV	X15, 8(X10)
	MOV	X16, 16(X10)
	MOV	X17, 24(X10)
	ADD	$32, X10
	ADD	$32, X11
	SUB	$32, X12
	BGE	X12, X9, f_loop32
	BEQZ	X12, done

f_loop16_check:
	MOV	$16, X9
	BLT	X12, X9, f_loop8_check
f_loop16:
	MOV	0(X11), X14
	MOV	8(X11), X15
	MOV	X14, 0(X10)
	MOV	X15, 8(X10)
	ADD	$16, X10
	ADD	$16, X11
	SUB	$16, X12
	BGE	X12, X9, f_loop16
	BEQZ	X12, done

f_loop8_check:
	MOV	$8, X9
	BLT	X12, X9, f_loop4_check
f_loop8:
	MOV	0(X11), X14
	MOV	X14, 0(X10)
	ADD	$8, X10
	ADD	$8, X11
	SUB	$8, X12
	BGE	X12, X9, f_loop8
	BEQZ	X12, done
	JMP	f_loop4_check

f_loop8_unaligned_check:
	MOV	$8, X9
	BLT	X12, X9, f_loop4_check
f_loop8_unaligned:
	MOVB	0(X11), X14
	MOVB	1(X11), X15
	MOVB	2(X11), X16
	MOVB	3(X11), X17
	MOVB	4(X11), X18
	MOVB	5(X11), X19
	MOVB	6(X11), X20
	MOVB	7(X11), X21
	MOVB	X14, 0(X10)
	MOVB	X15, 1(X10)
	MOVB	X16, 2(X10)
	MOVB	X17, 3(X10)
	MOVB	X18, 4(X10)
	MOVB	X19, 5(X10)
	MOVB	X20, 6(X10)
	MOVB	X21, 7(X10)
	ADD	$8, X10
	ADD	$8, X11
	SUB	$8, X12
	BGE	X12, X9, f_loop8_unaligned

f_loop4_check:
	MOV	$4, X9
	BLT	X12, X9, f_loop1
f_loop4:
	MOVB	0(X11), X14
	MOVB	1(X11), X15
	MOVB	2(X11), X16
	MOVB	3(X11), X17
	MOVB	X14, 0(X10)
	MOVB	X15, 1(X10)
	MOVB	X16, 2(X10)
	MOVB	X17, 3(X10)
	ADD	$4, X10
	ADD	$4, X11
	SUB	$4, X12
	BGE	X12, X9, f_loop4

f_loop1:
	BEQZ	X12, done
	MOVB	0(X11), X14
	MOVB	X14, 0(X10)
	ADD	$1, X10
	ADD	$1, X11
	SUB	$1, X12
	JMP	f_loop1
#endif

backward:
	ADD	X10, X12, X10
	ADD	X11, X12, X11

#ifndef EnableSmallSizeMemVector
	// If less than 8 bytes, do single byte copies.
	MOV	$8, X9
	BLT	X12, X9, b_loop4_check
#endif

#ifndef hasV
	MOVB	internal∕cpu·RISCV64+const_offsetRISCV64HasV(SB), X5
	BEQZ	X5, b_memmove_scalar
#endif

#ifndef EnableSmallSizeMemVector
	// Use vector if destination and source are not 8 byte aligned.
	OR	X10, X11, X5
	AND	$7, X5
	BNEZ	X5, b_vector_loop

	// Use scalar if destination and source are 8 byte aligned and n <= 64 bytes.
	SUB	$64, X12, X6
	BLEZ	X6, b_loop_check

	// Copy whole pointers with 64-bit elements. X10/X11 are end pointers.
	AND	$7, X12, X5
	BEQZ	X5, b_vector64_start
#endif

#ifdef EnableSmallSizeMemVector
	// Aligned whole-pointer copies use 64-bit elements regardless of size.
	OR	X10, X11, X5
	OR	X12, X5
	AND	$7, X5
	BEQZ	X5, b_vector64_start
#endif

#ifdef EnableSmallSizeMemVector
	PCALIGN	$16
b_vector_dispatch:
	MOV		$16, X6
#ifdef VLen_256
	SLLI	$1, X6
#endif
#ifdef VLen_512
	SLLI	$2, X6
#endif
	BGEU	X6, X12, b_vector_single
	SLLI	$2, X6
	BGTU	X12, X6, b_vector_loop
	SRLI	$1, X6
	BGTU	X12, X6, b_vector_quarter

// Copy (vlen+1)..(2*vlen) bytes
	PCALIGN	$16
b_vector_double:
	VSETVLI	X12, E8, M2, TA, MA, X5
	SUB	X5, X10
	SUB	X5, X11
	VLE8V	(X11), V8
	VSE8V	V8, (X10)
	RET

// Copy 1..(vlen) bytes
	PCALIGN	$16
b_vector_single:
	VSETVLI	X12, E8, M1, TA, MA, X5
	SUB	X5, X10
	SUB	X5, X11
	VLE8V	(X11), V8
	VSE8V	V8, (X10)
	RET

// Copy (2*vlen+1)..(4*vlen) bytes
	PCALIGN	$16
b_vector_quarter:
	VSETVLI	X12, E8, M4, TA, MA, X5
	SUB	X5, X10
	SUB	X5, X11
	VLE8V	(X11), V8
	VSE8V	V8, (X10)
	RET
#endif

// Copy (4*vlen+1).. bytes
	PCALIGN	$16
b_vector_loop:
	VSETVLI	X12, E8, M8, TA, MA, X5
	SUB	X5, X10
	SUB	X5, X11
	VLE8V	(X11), V8
	VSE8V	V8, (X10)
	SUB	X5, X12
	BNEZ	X12, b_vector_loop
	RET

b_vector64_start:
	// X10/X11 point one past the end; X12 counts bytes here.
	SRLI	$3, X12, X12

	VSETVLI	X12, E64, M8, TA, MA, X5
	SLLI	$3, X5, X6

	// Copy full groups from the end, then the shorter head group. Loading
	// a whole group before storing it keeps an overlapping source intact.
	PCALIGN	$16
b_vector64_loop:
	SUB	X6, X10
	SUB	X6, X11
	VLE64V	(X11), V8
	VSE64V	V8, (X10)
	SUB	X5, X12
	BGEU	X12, X5, b_vector64_loop
	BEQZ	X12, done
	VSETVLI	X12, E64, M8, TA, MA, X5
	SLLI	$3, X5, X6
	SUB	X6, X10
	SUB	X6, X11
	VLE64V	(X11), V8
	VSE64V	V8, (X10)
	RET

#ifndef hasV
b_memmove_scalar:
	// Check alignment - if alignment differs we have to do one byte at a time.
	AND	$7, X10, X5
	AND	$7, X11, X6
	BNE	X5, X6, b_loop8_unaligned_check
	BEQZ	X5, b_loop_check

	// Move one byte at a time until we reach 8 byte alignment.
	SUB	X5, X12, X12
b_align:
	SUB	$1, X5
	SUB	$1, X10
	SUB	$1, X11
	MOVB	0(X11), X14
	MOVB	X14, 0(X10)
	BNEZ	X5, b_align
#endif

#ifndef EnableSmallSizeMemVector
b_loop_check:
	MOV	$16, X9
	BLT	X12, X9, b_loop8_check
	MOV	$32, X9
	BLT	X12, X9, b_loop16_check
	MOV	$64, X9
	BLT	X12, X9, b_loop32_check
b_loop64:
	SUB	$64, X10
	SUB	$64, X11
	MOV	0(X11), X14
	MOV	8(X11), X15
	MOV	16(X11), X16
	MOV	24(X11), X17
	MOV	32(X11), X18
	MOV	40(X11), X19
	MOV	48(X11), X20
	MOV	56(X11), X21
	MOV	X14, 0(X10)
	MOV	X15, 8(X10)
	MOV	X16, 16(X10)
	MOV	X17, 24(X10)
	MOV	X18, 32(X10)
	MOV	X19, 40(X10)
	MOV	X20, 48(X10)
	MOV	X21, 56(X10)
	SUB	$64, X12
	BGE	X12, X9, b_loop64
	BEQZ	X12, done

b_loop32_check:
	MOV	$32, X9
	BLT	X12, X9, b_loop16_check
b_loop32:
	SUB	$32, X10
	SUB	$32, X11
	MOV	0(X11), X14
	MOV	8(X11), X15
	MOV	16(X11), X16
	MOV	24(X11), X17
	MOV	X14, 0(X10)
	MOV	X15, 8(X10)
	MOV	X16, 16(X10)
	MOV	X17, 24(X10)
	SUB	$32, X12
	BGE	X12, X9, b_loop32
	BEQZ	X12, done

b_loop16_check:
	MOV	$16, X9
	BLT	X12, X9, b_loop8_check
b_loop16:
	SUB	$16, X10
	SUB	$16, X11
	MOV	0(X11), X14
	MOV	8(X11), X15
	MOV	X14, 0(X10)
	MOV	X15, 8(X10)
	SUB	$16, X12
	BGE	X12, X9, b_loop16
	BEQZ	X12, done

b_loop8_check:
	MOV	$8, X9
	BLT	X12, X9, b_loop4_check
b_loop8:
	SUB	$8, X10
	SUB	$8, X11
	MOV	0(X11), X14
	MOV	X14, 0(X10)
	SUB	$8, X12
	BGE	X12, X9, b_loop8
	BEQZ	X12, done
	JMP	b_loop4_check

b_loop8_unaligned_check:
	MOV	$8, X9
	BLT	X12, X9, b_loop4_check
b_loop8_unaligned:
	SUB	$8, X10
	SUB	$8, X11
	MOVB	0(X11), X14
	MOVB	1(X11), X15
	MOVB	2(X11), X16
	MOVB	3(X11), X17
	MOVB	4(X11), X18
	MOVB	5(X11), X19
	MOVB	6(X11), X20
	MOVB	7(X11), X21
	MOVB	X14, 0(X10)
	MOVB	X15, 1(X10)
	MOVB	X16, 2(X10)
	MOVB	X17, 3(X10)
	MOVB	X18, 4(X10)
	MOVB	X19, 5(X10)
	MOVB	X20, 6(X10)
	MOVB	X21, 7(X10)
	SUB	$8, X12
	BGE	X12, X9, b_loop8_unaligned

b_loop4_check:
	MOV	$4, X9
	BLT	X12, X9, b_loop1
b_loop4:
	SUB	$4, X10
	SUB	$4, X11
	MOVB	0(X11), X14
	MOVB	1(X11), X15
	MOVB	2(X11), X16
	MOVB	3(X11), X17
	MOVB	X14, 0(X10)
	MOVB	X15, 1(X10)
	MOVB	X16, 2(X10)
	MOVB	X17, 3(X10)
	SUB	$4, X12
	BGE	X12, X9, b_loop4

b_loop1:
	BEQZ	X12, done
	SUB	$1, X10
	SUB	$1, X11
	MOVB	0(X11), X14
	MOVB	X14, 0(X10)
	SUB	$1, X12
	JMP	b_loop1
#endif

done:
	RET
