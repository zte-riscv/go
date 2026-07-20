// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !purego

#include "textflag.h"

// func encodeChunk(encode *[64]byte, dst, src []byte, n int)
//
// On entry (ABI0 calling convention):
//   X10 = encode *[64]byte  (base64 lookup table: 64 bytes)
//   X11 = dst_base
//   X12 = dst_len
//   X13 = dst_cap
//   X14 = src_base
//   X15 = src_len
//   X16 = src_cap
//   X17 = n  (number of source bytes to encode, must be a multiple of 3)
//
// This function encodes n bytes from src into dst using the base64
// encode table. Each 3-byte source block produces 4 output bytes.
// Unlike the pure Go version, this hand-written assembly eliminates
// bounds checks, slice indexing overhead, and nil-check the encode
// pointer on every iteration.

TEXT ·encodeChunk(SB),NOSPLIT,$0-73
	// Load parameters from stack frame (ABI0 convention)
	// Stack layout:
	//   encode+0(FP)    - *[64]byte (8 bytes)
	//   dst_base+8(FP)  - []byte base (8 bytes)
	//   dst_len+16(FP)  - []byte len (8 bytes)
	//   dst_cap+24(FP)  - []byte cap (8 bytes)
	//   src_base+32(FP) - []byte base (8 bytes)
	//   src_len+40(FP)  - []byte len (8 bytes)
	//   src_cap+48(FP)  - []byte cap (8 bytes)
	//   n+56(FP)        - int (8 bytes)
	// Total: 64 bytes + 8 (return address) = 72, but Go uses 73 for alignment

	MOV	n+56(FP), X28	// X28 = n
	BEQZ	X28, ret

	MOV	encode+0(FP), X5	// X5 = encode table base
	MOV	dst_base+8(FP), X7	// X7 = dst pointer
	MOV	src_base+32(FP), X6	// X6 = src pointer

	PCALIGN	$16
loop:
	// --- Load 3 source bytes into X11, X12, X13 ---
	MOVBU	(X6), X11	// X11 = src[0]
	MOVBU	1(X6), X12	// X12 = src[1]
	MOVBU	2(X6), X13	// X13 = src[2]

	// --- Form 24-bit value: val = src[0]<<16 | src[1]<<8 | src[2] ---
	SLLI	$16, X11, X11	// X11 = src[0] << 16
	SLLI	$8, X12, X12	// X12 = src[1] << 8
	OR	X12, X11, X11	// X11 = src[0]<<16 | src[1]<<8
	OR	X13, X11, X11	// X11 = val

	// --- dst[0] = encode[val>>18 & 0x3F] ---
	SRLI	$18, X11, X12	// X12 = val >> 18
	ANDI	$0x3F, X12, X12	// X12 = idx0
	ADD	X5, X12, X12	// X12 = &encode[idx0]
	MOVBU	(X12), X12	// X12 = encode[idx0]
	MOVB	X12, (X7)	// dst[0] = encode[idx0]

	// --- dst[1] = encode[val>>12 & 0x3F] ---
	SRLI	$12, X11, X12	// X12 = val >> 12
	ANDI	$0x3F, X12, X12	// X12 = idx1
	ADD	X5, X12, X12	// X12 = &encode[idx1]
	MOVBU	(X12), X12	// X12 = encode[idx1]
	MOVB	X12, 1(X7)	// dst[1] = encode[idx1]

	// --- dst[2] = encode[val>>6 & 0x3F] ---
	SRLI	$6, X11, X12	// X12 = val >> 6
	ANDI	$0x3F, X12, X12	// X12 = idx2
	ADD	X5, X12, X12	// X12 = &encode[idx2]
	MOVBU	(X12), X12	// X12 = encode[idx2]
	MOVB	X12, 2(X7)	// dst[2] = encode[idx2]

	// --- dst[3] = encode[val & 0x3F] ---
	ANDI	$0x3F, X11, X12	// X12 = idx3
	ADD	X5, X12, X12	// X12 = &encode[idx3]
	MOVBU	(X12), X12	// X12 = encode[idx3]
	MOVB	X12, 3(X7)	// dst[3] = encode[idx3]

	// --- Advance pointers ---
	ADD	$3, X6		// src += 3
	ADD	$4, X7		// dst += 4
	ADD	$-3, X28	// n   -= 3
	BNEZ	X28, loop

ret:
	RET
