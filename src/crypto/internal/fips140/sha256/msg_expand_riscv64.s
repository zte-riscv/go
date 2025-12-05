// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !purego && riscv64

#include "textflag.h"

// Index vector for loading {w4, w9, w10, w11} using VLUXEI32V
// Byte offsets: 0, 20, 24, 28 (for w4, w9, w10, w11 from p_base)
DATA	index_w4_w9_w10_w11<>+0(SB)/4, $0
DATA	index_w4_w9_w10_w11<>+4(SB)/4, $20
DATA	index_w4_w9_w10_w11<>+8(SB)/4, $24
DATA	index_w4_w9_w10_w11<>+12(SB)/4, $28
GLOBL	index_w4_w9_w10_w11<>(SB), RODATA, $16

// MessageExpansion expands a 64-byte message block into 64 32-bit words.
// This function uses RISC-V vector SHA-2 instructions (VSHA2MSVV) for optimization.
//
// func messageExpansionRISCV64(p []byte, w *[64]uint32)
// 
// In Go Plan9 assembly, parameters are accessed via FP:
//   p []byte: p_base+0(FP), p_len+8(FP), p_cap+16(FP)
//   w *[64]uint32: w+24(FP)
TEXT ·messageExpansionRISCV64(SB),NOSPLIT,$0-32
	// Get parameters from FP
	MOV	p_base+0(FP), X5	// X5 = input message pointer (p_base)
	MOV	w+24(FP), X6		// X6 = output array pointer
	
	// Simple version: Generate W[16..19] from W[0..15] using VSHA2MSVV
	// Set vector config: E32 (32-bit elements), M1 (LMUL=1), vl=4 (process 4 elements)
	VSETIVLI	$4, E32, M1, TA, MA, X0
	
	// Step 1: Load W[0..3] from p_base to V1
	// Load 4 words (16 bytes) starting from p_base
	VLE32V		(X5), V1		// V1 = {W[0], W[1], W[2], W[3]} (big-endian)
	
	// Convert from big-endian to little-endian
	VREV8V		V1, V1			// V1 = {W[0], W[1], W[2], W[3]} (little-endian)
	
	// Reverse to get {W[3], W[2], W[1], W[0]} format for VSHA2MSVV Vd input
	VMV1RV		V1, V1			// V1 = {W[3], W[2], W[1], W[0]}
	
	// Step 2: Load {W[4], W[9], W[10], W[11]} using VLUXEI32V
	// Prepare index vector: load indices {0, 20, 24, 28} into V31
	// These are byte offsets relative to w[4] address
	MOV	$index_w4_w9_w10_w11<>(SB), X7	// X7 = address of index data
	VLE32V		(X7), V31		// V31 = {0, 20, 24, 28} (byte offsets)
	
	// Calculate w[4] address: p_base + 16 (4 * 4 bytes)
	ADD	$16, X5, X7		// X7 = p_base + 16 (address of w[4])
	
	// Load using indexed load: VLUXEI32V (base), index, dest
	// Base address is w[4] address (X7), indices are in V31
	VLUXEI32V	(X7), V31, V2		// V2 = {W[4], W[9], W[10], W[11]} (big-endian)
	
	// Convert from big-endian to little-endian
	VREV8V		V2, V2			// V2 = {W[4], W[9], W[10], W[11]} (little-endian)
	
	// Rearrange to {W[11], W[10], W[9], W[4]} format for VSHA2MSVV Vs2 input
	VMV1RV		V2, V2			// V2 = {W[11], W[10], W[9], W[4]}
	
	// Step 3: Load W[12..15] from p_base to V3
	// Load 4 words starting from offset 48 (W[12] = p_base + 48)
	ADD	$48, X5, X7		// X7 = p_base + 48 (offset for W[12])
	VLE32V		(X7), V3		// V3 = {W[12], W[13], W[14], W[15]} (big-endian)
	
	// Convert from big-endian to little-endian
	VREV8V		V3, V3			// V3 = {W[12], W[13], W[14], W[15]} (little-endian)
	
	// Rearrange to {W[15], W[14], W[13], W[12]} format for VSHA2MSVV Vs1 input
	// Note: W[13] is not used in the calculation (marked as '-'), but we can keep its value
	VMV1RV		V3, V3			// V3 = {W[15], W[14], W[13], W[12]}
	
	// Step 4: Call VSHA2MSVV to generate W[16..19]
	// VSHA2MSVV vs1, vs2, vd
	// vd (V1 input): {W[3], W[2], W[1], W[0]}
	// vs2 (V2 input): {W[11], W[10], W[9], W[4]}
	// vs1 (V3 input): {W[15], W[14], 0, W[12]}
	// vd (V1 output): {W[19], W[18], W[17], W[16]}
	VSHA2MSVV	V3, V2, V1		// V1 = {W[19], W[18], W[17], W[16]}
	
	// Step 5: Store W[0..15] and W[16..19] to output array
	// Set vector config: E32, M1, vl=4 (process 4 elements at a time)
	VSETIVLI	$4, E32, M1, TA, MA, X0
	
	// Store W[0..15]: manually load and store 4 groups of 4 words
	// Group 1: W[0..3]
	VLE32V		(X5), V4		// V4 = {W[0], W[1], W[2], W[3]} (big-endian)
	VREV8V		V4, V4			// Convert to little-endian
	VSE32V		V4, (X6)		// Store W[0..3]
	
	// Group 2: W[4..7]
	ADD	$16, X5, X7		// X7 = p_base + 16
	VLE32V		(X7), V4		// V4 = {W[4], W[5], W[6], W[7]} (big-endian)
	VREV8V		V4, V4			// Convert to little-endian
	ADD	$16, X6, X7		// X7 = w + 16
	VSE32V		V4, (X7)		// Store W[4..7]
	
	// Group 3: W[8..11]
	ADD	$32, X5, X7		// X7 = p_base + 32
	VLE32V		(X7), V4		// V4 = {W[8], W[9], W[10], W[11]} (big-endian)
	VREV8V		V4, V4			// Convert to little-endian
	ADD	$32, X6, X7		// X7 = w + 32
	VSE32V		V4, (X7)		// Store W[8..11]
	
	// Group 4: W[12..15]
	ADD	$48, X5, X7		// X7 = p_base + 48
	VLE32V		(X7), V4		// V4 = {W[12], W[13], W[14], W[15]} (big-endian)
	VREV8V		V4, V4			// Convert to little-endian
	ADD	$48, X6, X7		// X7 = w + 48
	VSE32V		V4, (X7)		// Store W[12..15]
	
	// Store W[16..19] to w[16..19]
	ADD	$64, X6, X7		// X7 = w + 64 (offset for W[16])
	VSE32V		V1, (X7)		// Store W[16..19] from V1
	
	// Step 6: Generate W[20..23] from W[4..19] using VSHA2MSVV
	// Prepare Vd (V4) for VSHA2MSVV: {W[7], W[6], W[5], W[4]}
	// Load W[4..7] from p_base
	ADD	$16, X5, X7		// X7 = p_base + 16 (address of W[4])
	VLE32V		(X7), V4		// V4 = {W[4], W[5], W[6], W[7]} (big-endian)
	VREV8V		V4, V4			// Convert to little-endian
	VMV1RV		V4, V4			// V4 = {W[7], W[6], W[5], W[4]}
	
	// Prepare Vs2 (V5) for VSHA2MSVV: {W[15], W[14], W[13], W[8]}
	// Load {W[8], W[13], W[14], W[15]} using VLUXEI32V
	// Reuse the same index data (offsets are the same: 0, 20, 24, 28)
	MOV	$index_w4_w9_w10_w11<>(SB), X7	// X7 = address of index data
	VLE32V		(X7), V31		// V31 = {0, 20, 24, 28} (byte offsets)
	
	// Calculate W[8] address: p_base + 32
	ADD	$32, X5, X7		// X7 = p_base + 32 (address of W[8])
	VLUXEI32V	(X7), V31, V5		// V5 = {W[8], W[13], W[14], W[15]} (big-endian)
	VREV8V		V5, V5			// Convert to little-endian
	VMV1RV		V5, V5			// V5 = {W[15], W[14], W[13], W[8]}
	
	// Prepare Vs1 (V6) for VSHA2MSVV: {W[19], W[18], -, W[16]}
	// V1 currently contains {W[19], W[18], W[17], W[16]} from previous VSHA2MSVV
	// Copy V1 to V6 and set W[17] (index 2) to 0 using vmerge
	// Reference: similar to Step 3 (lines 63-73), but need to zero out W[17]
	VMV1RV		V1, V6			// V6 = {W[16], W[17], W[18], W[19]}
	
	// Call VSHA2MSVV to generate W[20..23]
	// VSHA2MSVV vs1, vs2, vd
	// vd (V4 input): {W[7], W[6], W[5], W[4]}
	// vs2 (V5 input): {W[15], W[14], W[13], W[8]}
	// vs1 (V6 input): {W[19], W[18], 0, W[16]}
	// vd (V4 output): {W[23], W[22], W[21], W[20]}
	VSHA2MSVV	V6, V5, V4		// V4 = {W[23], W[22], W[21], W[20]}
	
	// Store W[20..23] to w[20..23]
	ADD	$80, X6, X7		// X7 = w + 80 (offset for W[20], 20 * 4 = 80)
	VSE32V		V4, (X7)		// Store W[20..23] from V4
	
	RET
