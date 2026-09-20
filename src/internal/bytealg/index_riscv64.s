// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

// func Index(a, b []byte) int
TEXT ·Index<ABIInternal>(SB),NOSPLIT,$0-56
	// X10 = a_base
	// X11 = a_len
	// X12 = a_cap (unused)
	// X13 = b_base
	// X14 = b_len
	// X15 = b_cap (unused)
	MOV	X13, X12
	MOV	X14, X13
	JMP	indexbody<>(SB)

// func IndexString(a, b string) int
TEXT ·IndexString<ABIInternal>(SB),NOSPLIT,$0-40
	// X10 = a_base
	// X11 = a_len
	// X12 = b_base
	// X13 = b_len
	JMP	indexbody<>(SB)

// On entry:
// X10 = a_base
// X11 = a_len
// X12 = b_base
// X13 = b_len (2 <= len <= 32)
// return in X10
TEXT indexbody<>(SB),NOSPLIT|NOFRAME,$0
	SUB	X13, X11, X14		// X14 = len(a) - len(b)
	ADD	X10, X14			// X14 = last valid start pointer
	ADD	$1, X10, X15		// X15 = a_base + 1, used by found:

	MOV	$8, X16
	BLT	X16, X13, greater_8	// if len(b) > 8
	BEQ	X16, X13, len_8

	MOV	$4, X16
	BLT	X16, X13, len_5_7	// if len(b) is 5..7
	BEQ	X16, X13, len_4

	MOV	$3, X16
	BEQ	X16, X13, len_3
	JMP	len_2

len_8:
	// Unaligned 8-byte load from needle (fast if misaligned access is good).
	MOV	(X12), X16
loop_8:
	BLTU	X14, X10, not_found
	// Unaligned 8-byte load from haystack candidate.
	MOV	(X10), X17
	ADD	$1, X10
	BNE	X16, X17, loop_8
	JMP	found

len_5_7:
	MOV	$7, X16
	BEQ	X16, X13, len_7
	MOV	$6, X16
	BEQ	X16, X13, len_6
	JMP	len_5

len_7:
	// 4-byte head uses potentially unaligned word load.
	MOVWU	(X12), X16
	MOVBU	4(X12), X17
	MOVBU	5(X12), X18
	MOVBU	6(X12), X19
loop_7:
	BLTU	X14, X10, not_found
	MOVWU	(X10), X20
	ADD	$1, X10
	BNE	X16, X20, loop_7
	MOVBU	3(X10), X20
	BNE	X17, X20, loop_7
	MOVBU	4(X10), X20
	BNE	X18, X20, loop_7
	MOVBU	5(X10), X20
	BNE	X19, X20, loop_7
	JMP	found

len_6:
	// 4-byte head uses potentially unaligned word load.
	MOVWU	(X12), X16
	MOVBU	4(X12), X17
	MOVBU	5(X12), X18
loop_6:
	BLTU	X14, X10, not_found
	MOVWU	(X10), X20
	ADD	$1, X10
	BNE	X16, X20, loop_6
	MOVBU	3(X10), X20
	BNE	X17, X20, loop_6
	MOVBU	4(X10), X20
	BNE	X18, X20, loop_6
	JMP	found

len_5:
	// 4-byte head uses potentially unaligned word load.
	MOVWU	(X12), X16
	MOVBU	4(X12), X17
loop_5:
	BLTU	X14, X10, not_found
	MOVWU	(X10), X20
	ADD	$1, X10
	BNE	X16, X20, loop_5
	MOVBU	3(X10), X20
	BNE	X17, X20, loop_5
	JMP	found

len_4:
	// Unaligned 4-byte load from needle.
	MOVWU	(X12), X16
loop_4:
	BLTU	X14, X10, not_found
	// Unaligned 4-byte load from haystack candidate.
	MOVWU	(X10), X17
	ADD	$1, X10
	BNE	X16, X17, loop_4
	JMP	found

len_3:
	MOVBU	0(X12), X16
	MOVBU	1(X12), X17
	MOVBU	2(X12), X18
loop_3:
	BLTU	X14, X10, not_found
	MOVBU	0(X10), X19
	ADD	$1, X10
	BNE	X16, X19, loop_3
	MOVBU	0(X10), X19
	BNE	X17, X19, loop_3
	MOVBU	1(X10), X19
	BNE	X18, X19, loop_3
	JMP	found

len_2:
	MOVBU	0(X12), X16
	MOVBU	1(X12), X17
loop_2:
	BLTU	X14, X10, not_found
	MOVBU	0(X10), X18
	ADD	$1, X10
	BNE	X16, X18, loop_2
	MOVBU	0(X10), X18
	BNE	X17, X18, loop_2

found:
	SUB	X15, X10
	RET

not_found:
	MOV	$-1, X10
	RET

greater_8:
	SUB	$9, X13, X21		// X21 = len(b) - 9; tail address uses X10+X21 after X10++
	MOV	$16, X16
	BLT	X16, X13, greater_16	// if len(b) > 16

len_9_16:
	// Unaligned 8-byte head and tail loads from needle.
	MOV	(X12), X16
	SUB	$8, X13, X17
	ADD	X12, X17
	MOV	(X17), X18
loop_9_16:
	BLTU	X14, X10, not_found
	// Unaligned 8-byte head load from haystack.
	MOV	(X10), X19
	ADD	$1, X10
	BNE	X16, X19, loop_9_16
	ADD	X10, X21, X20
	// Unaligned 8-byte tail load from haystack.
	MOV	(X20), X19
	BNE	X18, X19, loop_9_16
	JMP	found

greater_16:
	MOV	$24, X16
	BLT	X16, X13, len_25_32	// if len(b) > 24

len_17_24:
	// Unaligned 16-byte head and 8-byte tail loads from needle.
	MOV	0(X12), X16
	MOV	8(X12), X17
	SUB	$8, X13, X18
	ADD	X12, X18
	MOV	(X18), X19
loop_17_24:
	BLTU	X14, X10, not_found
	// Unaligned 16-byte head load from haystack.
	MOV	0(X10), X20
	MOV	8(X10), X22
	ADD	$1, X10
	BNE	X16, X20, loop_17_24
	BNE	X17, X22, loop_17_24
	ADD	X10, X21, X23
	// Unaligned 8-byte tail load from haystack.
	MOV	(X23), X20
	BNE	X19, X20, loop_17_24
	JMP	found

len_25_32:
	// Unaligned 24-byte head and 8-byte tail loads from needle.
	MOV	0(X12), X16
	MOV	8(X12), X17
	MOV	16(X12), X18
	SUB	$8, X13, X19
	ADD	X12, X19
	MOV	(X19), X20
loop_25_32:
	BLTU	X14, X10, not_found
	// Unaligned 24-byte head load from haystack.
	MOV	0(X10), X22
	MOV	8(X10), X23
	MOV	16(X10), X24
	ADD	$1, X10
	BNE	X16, X22, loop_25_32
	BNE	X17, X23, loop_25_32
	BNE	X18, X24, loop_25_32
	ADD	X10, X21, X25
	// Unaligned 8-byte tail load from haystack.
	MOV	(X25), X22
	BNE	X20, X22, loop_25_32
	JMP	found
