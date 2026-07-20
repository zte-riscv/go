// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !purego

#include "textflag.h"

// func encodeChunk(encode *[64]byte, dst, src []byte, n int)
//
// On entry (ABI0 calling convention), stack layout:
//   encode+0(FP)    - *[64]byte (8 bytes)
//   dst_base+8(FP)   - []byte ptr  (8 bytes)
//   dst_len+16(FP)   - []byte len  (8 bytes)
//   dst_cap+24(FP)   - []byte cap  (8 bytes)
//   src_base+32(FP)  - []byte ptr  (8 bytes)
//   src_len+40(FP)   - []byte len  (8 bytes)
//   src_cap+48(FP)   - []byte cap  (8 bytes)
//   n+56(FP)         - int         (8 bytes)
//
// Register usage:
//   X5  = encode table base (preserved across loop)
//   X6  = src pointer
//   X7  = dst pointer
//   X28 = remaining count (n)
//   X10, X11, X12, X13, X14, X15 = local temporaries
//   X31 = assembler temp (do not use)

TEXT ·encodeChunk(SB),NOSPLIT,$0-64
	MOV	n+56(FP), X28	// X28 = n
	BEQZ	X28, ret

	MOV	encode+0(FP), X5	// X5  = encode table base
	MOV	dst_base+8(FP), X7	// X7  = dst pointer
	MOV	src_base+32(FP), X6	// X6  = src pointer

	// If n < 12, use the single-group loop directly
	MOV	$12, X15
	BLT	X28, X15, tail

	// ================================================================
	// Main 4x unrolled loop: process 12 src bytes → 16 dst bytes per iteration
	// ================================================================
	PCALIGN	$16
loop4:
	// ──── Group 1: src[0..2] → dst[0..3] ────
	MOVBU	(X6), X10	// src[0]
	MOVBU	1(X6), X11	// src[1]
	MOVBU	2(X6), X12	// src[2]
	SLLI	$16, X10, X10
	SLLI	$8, X11, X11
	OR	X11, X10, X10
	OR	X12, X10, X10	// X10 = val1

	SRLI	$18, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, (X7)		// dst[0]

	SRLI	$12, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 1(X7)		// dst[1]

	SRLI	$6, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 2(X7)		// dst[2]

	ANDI	$0x3F, X10, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 3(X7)		// dst[3]

	// ──── Group 2: src[3..5] → dst[4..7] ────
	MOVBU	3(X6), X10	// src[3]
	MOVBU	4(X6), X11	// src[4]
	MOVBU	5(X6), X12	// src[5]
	SLLI	$16, X10, X10
	SLLI	$8, X11, X11
	OR	X11, X10, X10
	OR	X12, X10, X10	// X10 = val2

	SRLI	$18, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 4(X7)		// dst[4]

	SRLI	$12, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 5(X7)		// dst[5]

	SRLI	$6, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 6(X7)		// dst[6]

	ANDI	$0x3F, X10, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 7(X7)		// dst[7]

	// ──── Group 3: src[6..8] → dst[8..11] ────
	MOVBU	6(X6), X10	// src[6]
	MOVBU	7(X6), X11	// src[7]
	MOVBU	8(X6), X12	// src[8]
	SLLI	$16, X10, X10
	SLLI	$8, X11, X11
	OR	X11, X10, X10
	OR	X12, X10, X10	// X10 = val3

	SRLI	$18, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 8(X7)		// dst[8]

	SRLI	$12, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 9(X7)		// dst[9]

	SRLI	$6, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 10(X7)		// dst[10]

	ANDI	$0x3F, X10, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 11(X7)		// dst[11]

	// ──── Group 4: src[9..11] → dst[12..15] ────
	MOVBU	9(X6), X10	// src[9]
	MOVBU	10(X6), X11	// src[10]
	MOVBU	11(X6), X12	// src[11]
	SLLI	$16, X10, X10
	SLLI	$8, X11, X11
	OR	X11, X10, X10
	OR	X12, X10, X10	// X10 = val4

	SRLI	$18, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 12(X7)		// dst[12]

	SRLI	$12, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 13(X7)		// dst[13]

	SRLI	$6, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 14(X7)		// dst[14]

	ANDI	$0x3F, X10, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 15(X7)		// dst[15]

	// --- Advance pointers and loop ---
	ADD	$12, X6		// src += 12
	ADD	$16, X7		// dst += 16
	ADD	$-12, X28	// n -= 12
	MOV	$12, X15
	BGE	X28, X15, loop4

	BEQZ	X28, ret

	// ================================================================
	// Tail: single-group loop for remaining < 12 bytes
	// ================================================================
	PCALIGN	$16
tail:
	// --- Load 3 source bytes ---
	MOVBU	(X6), X10	// src[0]
	MOVBU	1(X6), X11	// src[1]
	MOVBU	2(X6), X12	// src[2]

	// --- Form 24-bit value: val = src[0]<<16 | src[1]<<8 | src[2] ---
	SLLI	$16, X10, X10
	SLLI	$8, X11, X11
	OR	X11, X10, X10
	OR	X12, X10, X10	// X10 = val

	// --- dst[0] = encode[val>>18 & 0x3F] ---
	SRLI	$18, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, (X7)

	// --- dst[1] = encode[val>>12 & 0x3F] ---
	SRLI	$12, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 1(X7)

	// --- dst[2] = encode[val>>6 & 0x3F] ---
	SRLI	$6, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 2(X7)

	// --- dst[3] = encode[val & 0x3F] ---
	ANDI	$0x3F, X10, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 3(X7)

	// --- Advance pointers ---
	ADD	$3, X6		// src += 3
	ADD	$4, X7		// dst += 4
	ADD	$-3, X28	// n -= 3
	BNEZ	X28, tail

ret:
	RET
