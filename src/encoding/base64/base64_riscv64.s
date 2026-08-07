// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !purego

#include "asm_riscv64.h"
#include "go_asm.h"
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
//   X30 = $12 (loop constant) / runtime check temp
//   X31 = assembler temp (do not use)

TEXT ·encodeChunk(SB),NOSPLIT,$0-64
	MOV	n+56(FP), X28	// X28 = n
	BEQZ	X28, ret

	MOV	encode+0(FP), X5	// X5  = encode table base
	MOV	dst_base+8(FP), X7	// X7  = dst pointer
	MOV	src_base+32(FP), X6	// X6  = src pointer

	// If n < 12, use the tail loop
	MOV	$12, X30
	BLT	X28, X30, tail

	// The 64-bit loads (MOV (X6), MOV 4(X6)) may be unaligned, so the
	// REV8 path is built only when misaligned_fast is enabled. REV8
	// itself requires Zbb.
#ifdef GORISCV64EXT_misaligned_fast
#ifndef hasZbb
	MOVB	internal∕cpu·RISCV64+const_offsetRISCV64HasZbb(SB), X30
	BEQZ	X30, scalar_4x
#endif

	MOV	$12, X30	// loop constant

	PCALIGN	$16
loop4:
	// Two 64-bit loads cover the 12 bytes
	MOV	(X6), X10	// src[0..7]
	MOV	4(X6), X11	// src[4..11]

	// REV8 gives big-endian byte order, so the 24-bit values line up
	// with simple shifts.
	REV8	X10, X10
	REV8	X11, X11

	// val1-4 = src[0..2], src[3..5], src[6..8], src[9..11]
	SRLI	$40, X10, X12		// val1
	SRLI	$16, X10, X13
	ANDI	$0xFFFFFF, X13, X13	// val2
	SRLI	$24, X11, X14
	ANDI	$0xFFFFFF, X14, X14	// val3
	ANDI	$0xFFFFFF, X11, X15	// val4

	// Group 1: val1 -> dst[0..3]
	SRLI	$18, X12, X10
	ANDI	$0x3F, X10, X10
	ADD	X5, X10, X10
	MOVBU	(X10), X11
	MOVB	X11, (X7)

	SRLI	$12, X12, X10
	ANDI	$0x3F, X10, X10
	ADD	X5, X10, X10
	MOVBU	(X10), X11
	MOVB	X11, 1(X7)

	SRLI	$6, X12, X10
	ANDI	$0x3F, X10, X10
	ADD	X5, X10, X10
	MOVBU	(X10), X11
	MOVB	X11, 2(X7)

	ANDI	$0x3F, X12, X10
	ADD	X5, X10, X10
	MOVBU	(X10), X11
	MOVB	X11, 3(X7)

	// Group 2: val2 -> dst[4..7]
	SRLI	$18, X13, X10
	ANDI	$0x3F, X10, X10
	ADD	X5, X10, X10
	MOVBU	(X10), X11
	MOVB	X11, 4(X7)

	SRLI	$12, X13, X10
	ANDI	$0x3F, X10, X10
	ADD	X5, X10, X10
	MOVBU	(X10), X11
	MOVB	X11, 5(X7)

	SRLI	$6, X13, X10
	ANDI	$0x3F, X10, X10
	ADD	X5, X10, X10
	MOVBU	(X10), X11
	MOVB	X11, 6(X7)

	ANDI	$0x3F, X13, X10
	ADD	X5, X10, X10
	MOVBU	(X10), X11
	MOVB	X11, 7(X7)

	// Group 3: val3 -> dst[8..11]
	SRLI	$18, X14, X10
	ANDI	$0x3F, X10, X10
	ADD	X5, X10, X10
	MOVBU	(X10), X11
	MOVB	X11, 8(X7)

	SRLI	$12, X14, X10
	ANDI	$0x3F, X10, X10
	ADD	X5, X10, X10
	MOVBU	(X10), X11
	MOVB	X11, 9(X7)

	SRLI	$6, X14, X10
	ANDI	$0x3F, X10, X10
	ADD	X5, X10, X10
	MOVBU	(X10), X11
	MOVB	X11, 10(X7)

	ANDI	$0x3F, X14, X10
	ADD	X5, X10, X10
	MOVBU	(X10), X11
	MOVB	X11, 11(X7)

	// Group 4: val4 -> dst[12..15]
	SRLI	$18, X15, X10
	ANDI	$0x3F, X10, X10
	ADD	X5, X10, X10
	MOVBU	(X10), X11
	MOVB	X11, 12(X7)

	SRLI	$12, X15, X10
	ANDI	$0x3F, X10, X10
	ADD	X5, X10, X10
	MOVBU	(X10), X11
	MOVB	X11, 13(X7)

	SRLI	$6, X15, X10
	ANDI	$0x3F, X10, X10
	ADD	X5, X10, X10
	MOVBU	(X10), X11
	MOVB	X11, 14(X7)

	ANDI	$0x3F, X15, X10
	ADD	X5, X10, X10
	MOVBU	(X10), X11
	MOVB	X11, 15(X7)

	ADD	$12, X6
	ADD	$16, X7
	ADD	$-12, X28
	BGE	X28, X30, loop4

	BEQZ	X28, ret
	JMP	tail
#endif // GORISCV64EXT_misaligned_fast

	// Scalar fallback: per-byte MOVBU loads, no REV8.
	// Used when misaligned_fast or Zbb is unavailable.
	PCALIGN	$16
scalar_4x:
	MOV	$12, X30
	BLT	X28, X30, tail

	PCALIGN	$16
scalar_loop4:
	// Group 1: src[0..2] -> dst[0..3]
	MOVBU	(X6), X10
	MOVBU	1(X6), X11
	MOVBU	2(X6), X12
	SLLI	$16, X10, X10
	SLLI	$8, X11, X11
	OR	X11, X10, X10
	OR	X12, X10, X10

	SRLI	$18, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, (X7)

	SRLI	$12, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 1(X7)

	SRLI	$6, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 2(X7)

	ANDI	$0x3F, X10, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 3(X7)

	// Group 2: src[3..5] -> dst[4..7]
	MOVBU	3(X6), X10
	MOVBU	4(X6), X11
	MOVBU	5(X6), X12
	SLLI	$16, X10, X10
	SLLI	$8, X11, X11
	OR	X11, X10, X10
	OR	X12, X10, X10

	SRLI	$18, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 4(X7)

	SRLI	$12, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 5(X7)

	SRLI	$6, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 6(X7)

	ANDI	$0x3F, X10, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 7(X7)

	// Group 3: src[6..8] -> dst[8..11]
	MOVBU	6(X6), X10
	MOVBU	7(X6), X11
	MOVBU	8(X6), X12
	SLLI	$16, X10, X10
	SLLI	$8, X11, X11
	OR	X11, X10, X10
	OR	X12, X10, X10

	SRLI	$18, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 8(X7)

	SRLI	$12, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 9(X7)

	SRLI	$6, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 10(X7)

	ANDI	$0x3F, X10, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 11(X7)

	// Group 4: src[9..11] -> dst[12..15]
	MOVBU	9(X6), X10
	MOVBU	10(X6), X11
	MOVBU	11(X6), X12
	SLLI	$16, X10, X10
	SLLI	$8, X11, X11
	OR	X11, X10, X10
	OR	X12, X10, X10

	SRLI	$18, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 12(X7)

	SRLI	$12, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 13(X7)

	SRLI	$6, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 14(X7)

	ANDI	$0x3F, X10, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 15(X7)

	ADD	$12, X6
	ADD	$16, X7
	ADD	$-12, X28
	BGE	X28, X30, scalar_loop4

	BEQZ	X28, ret

	// Tail: remaining < 12 bytes, one group per iteration
	PCALIGN	$16
tail:
	MOVBU	(X6), X10
	MOVBU	1(X6), X11
	MOVBU	2(X6), X12

	SLLI	$16, X10, X10
	SLLI	$8, X11, X11
	OR	X11, X10, X10
	OR	X12, X10, X10

	SRLI	$18, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, (X7)

	SRLI	$12, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 1(X7)

	SRLI	$6, X10, X11
	ANDI	$0x3F, X11, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 2(X7)

	ANDI	$0x3F, X10, X11
	ADD	X5, X11, X11
	MOVBU	(X11), X11
	MOVB	X11, 3(X7)

	ADD	$3, X6
	ADD	$4, X7
	ADD	$-3, X28
	BNEZ	X28, tail

ret:
	RET
