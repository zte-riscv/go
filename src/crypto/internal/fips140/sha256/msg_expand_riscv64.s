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


DATA	vreg_mask<>+0(SB)/4, $1
DATA	vreg_mask<>+4(SB)/4, $1
DATA	vreg_mask<>+8(SB)/4, $1
DATA	vreg_mask<>+12(SB)/4, $0
GLOBL	vreg_mask<>(SB), RODATA, $16


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

// 第一个
	VLE32V		(X5), V1		// V1 = {W[0], W[1], W[2], W[3]} (big-endian)
	VREV8V		V1, V1			// V1 = {W[0], W[1], W[2], W[3]} (little-endian)
	//VMV1RV		V1, V1			// V1 = {W[3], W[2], W[1], W[0]}

	MOV	$index_w4_w9_w10_w11<>(SB), X7	// X7 = address of index data
	VLE32V		(X7), V31		// V31 = {0, 20, 24, 28} (byte offsets)
	ADD	$16, X5, X7		// X7 = p_base + 16 (address of w[4])
	VLUXEI32V	(X7), V31, V2		// V2 = {W[4], W[9], W[10], W[11]} (big-endian)
	VREV8V		V2, V2			// V2 = {W[4], W[9], W[10], W[11]} (little-endian)
	//VMV1RV		V2, V2			// V2 = {W[11], W[10], W[9], W[4]}
	
	ADD	$48, X5, X7		// X7 = p_base + 48 (offset for W[12])
	VLE32V		(X7), V3		// V3 = {W[12], W[13], W[14], W[15]} (big-endian)
	VREV8V		V3, V3			// V3 = {W[12], W[13], W[14], W[15]} (little-endian)
	//VMV1RV		V3, V3			// V3 = {W[15], W[14], W[13], W[12]}
	VSHA2MSVV	V3, V2, V1		// V1 = {W[19], W[18], W[17], W[16]}

	// Store W[16..19] to w[16..19]
	ADD	$64, X6, X7		// X7 = w + 64 (offset for W[16])
	VSE32V		V1, (X7)		// Store W[16..19] from V1
	

// 第二个
	ADD	$16, X5, X7		// X7 = p_base + 16 (address of W[4])
	VLE32V		(X7), V4		// V4 = {W[4], W[5], W[6], W[7]} (big-endian)
	VREV8V		V4, V4			// Convert to little-endian
	//VMV1RV		V4, V4			// V4 = {W[7], W[6], W[5], W[4]}
	
	MOV	$index_w4_w9_w10_w11<>(SB), X7	// X7 = address of index data
	VLE32V		(X7), V31		// V31 = {0, 20, 24, 28} (byte offsets)
	ADD	$32, X5, X7		// X7 = p_base + 32 (address of W[8])
	VLUXEI32V	(X7), V31, V5		// V5 = {W[8], W[13], W[14], W[15]} (big-endian)
	VREV8V		V5, V5			// Convert to little-endian
	//VMV1RV		V5, V5			// V5 = {W[15], W[14], W[13], W[8]}
	
	VSHA2MSVV	V1, V5, V4		// V4 = {W[23], W[22], W[21], W[20]}
	
	// Store W[20..23] to w[20..23]
	ADD	$80, X6, X7		// X7 = w + 80 (offset for W[20], 20 * 4 = 80)
	VSE32V		V4, (X7)		// Store W[20..23] from V4


// 第三个，arg2需要利用之前的结果并使用掩码
	ADD	$32, X5, X7		// 常数相比上一轮加16
	VLE32V		(X7), V6	// 寄存器号加2	
	VREV8V		V6, V6			
	//VMV1RV		V6, V6			
	

	MOV	$vreg_mask<>(SB), X7	
	VLE32V		(X7), V0	// 掩码 {1,1,1,0}
	VMERGEVVM	V3, V1, V0, V7 	// 合并两个源操作数，1时选择第一个v中的元素
	// VMV1RV		V7, V10	
		
	VSHA2MSVV	V4, V7, V6		
	
	ADD	$96, X6, X7		// 常数相比上一轮加16
	VSE32V		V6, (X7)


// 第四个
	ADD	$48, X5, X7		// 常数相比上一轮加16
	VLE32V		(X7), V8	// 寄存器号加2	
	VREV8V		V8, V8					

	VMERGEVVM	V1, V4, V0, V9 	// 合并两个源操作数，1时选择第一个v中的元素	
		
	VSHA2MSVV	V6, V9, V8		
	
	ADD	$112, X6, X7		// 常数相比上一轮加16
	VSE32V		V8, (X7)


// 第五个
	VMERGEVVM	V4, V6, V0, V31 	// 合并两个源操作数，1时选择第一个v中的元素	
		
	VSHA2MSVV	V8, V31, V1		
	
	ADD	$128, X6, X7		// 常数相比上一轮加16
	VSE32V		V1, (X7)	// 覆盖

// 第六个
	VMERGEVVM	V6, V8, V0, V31 	// 合并两个源操作数，1时选择第一个v中的元素	
		
	VSHA2MSVV	V1, V31, V4		
	
	ADD	$144, X6, X7		// 常数相比上一轮加16
	VSE32V		V4, (X7)	// 覆盖

// 第七个
	VMERGEVVM	V8, V1, V0, V31 	// 合并两个源操作数，1时选择第一个v中的元素	
		
	VSHA2MSVV	V4, V31, V6		
	
	ADD	$160, X6, X7		// 常数相比上一轮加16
	VSE32V		V6, (X7)	// 覆盖

// 第八个
	VMERGEVVM	V1, V4, V0, V31 	// 合并两个源操作数，1时选择第一个v中的元素	
		
	VSHA2MSVV	V6, V31, V8		
	
	ADD	$176, X6, X7		// 常数相比上一轮加16
	VSE32V		V8, (X7)	// 覆盖

// 第九个
	VMERGEVVM	V4, V6, V0, V31 	// 合并两个源操作数，1时选择第一个v中的元素	
		
	VSHA2MSVV	V8, V31, V1		
	
	ADD	$192, X6, X7		// 常数相比上一轮加16
	VSE32V		V1, (X7)	// 覆盖

// 第十个
	VMERGEVVM	V6, V8, V0, V31 	// 合并两个源操作数，1时选择第一个v中的元素	
		
	VSHA2MSVV	V1, V31, V4		
	
	ADD	$208, X6, X7		// 常数相比上一轮加16
	VSE32V		V4, (X7)	// 覆盖

// 第十一个
	VMERGEVVM	V8, V1, V0, V31 	// 合并两个源操作数，1时选择第一个v中的元素	
		
	VSHA2MSVV	V4, V31, V6		
	
	ADD	$224, X6, X7		// 常数相比上一轮加16
	VSE32V		V6, (X7)	// 覆盖

// 第十二个
	VMERGEVVM	V1, V4, V0, V31 	// 合并两个源操作数，1时选择第一个v中的元素	
		
	VSHA2MSVV	V6, V31, V8		
	
	ADD	$240, X6, X7		// 常数相比上一轮加16
	VSE32V		V8, (X7)	// 覆盖

	RET

