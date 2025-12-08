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


// VSHA2CL expects vs2={a,b,e,f} with element order [f,e,b,a] (index 0,1,2,3)
// Digest layout: h[0]=a(offset 0), h[1]=b(4), h[2]=c(8), h[3]=d(12),
//                h[4]=e(16), h[5]=f(20), h[6]=g(24), h[7]=h(28)
DATA	index_a_b_e_f<>+0(SB)/4, $20   // 元素[0] = f (偏移20)
DATA	index_a_b_e_f<>+4(SB)/4, $16   // 元素[1] = e (偏移16)
DATA	index_a_b_e_f<>+8(SB)/4, $4    // 元素[2] = b (偏移4)
DATA	index_a_b_e_f<>+12(SB)/4, $0   // 元素[3] = a (偏移0)
GLOBL	index_a_b_e_f<>(SB), RODATA, $16

// VSHA2CL expects vd={c,d,g,h} with element order [h,g,d,c] (index 0,1,2,3)
DATA	index_c_d_g_h<>+0(SB)/4, $28   // 元素[0] = h (偏移28)
DATA	index_c_d_g_h<>+4(SB)/4, $24   // 元素[1] = g (偏移24)
DATA	index_c_d_g_h<>+8(SB)/4, $12   // 元素[2] = d (偏移12)
DATA	index_c_d_g_h<>+12(SB)/4, $8   // 元素[3] = c (偏移8)
GLOBL	index_c_d_g_h<>(SB), RODATA, $16

// BlockRISCV64WithTrace performs SHA-256 block compression similar to blockGeneric,
// but also saves the intermediate hash state (a, b, c, d, e, f, g, h) for each of the 64 rounds.
//
// func blockRISCV64WithTrace(dig *Digest, p []byte, temp_dig *[64][8]uint32, temp_kword *[64]uint32)
//
// In Go Plan9 assembly, parameters are accessed via FP:
//   dig *Digest: dig+0(FP) (8 bytes on riscv64)
//   p []byte: p_base+8(FP), p_len+16(FP), p_cap+24(FP) (each 8 bytes on riscv64)
//   temp_dig *[64][8]uint32: temp_dig+32(FP) (8 bytes on riscv64)
//   temp_kword *[64]uint32: temp_kword+40(FP) (8 bytes on riscv64)
//
// Stack frame size: $0-48 (no local variables, 48 bytes for parameters: 8+8+8+8+8+8)
TEXT ·blockRISCV64WithTrace(SB),NOSPLIT,$0-48
	// TODO: Implement RISC-V optimized block function with trace
	// For now, this is a placeholder that will be implemented later

	MOV	p_base+8(FP), X5	// X5 = input message pointer (p_base)
    MOV	dig+0(FP), X6	// X6 = digest pointer

    VSETIVLI	$4, E32, M1, TA, MA, X0

    // 前四轮的值
    VLE32V		(X5), V1		// V1 = {W[0], W[1], W[2], W[3]} (big-endian)
    VREV8V		V1, V1			// Convert to little-endian

    // V1中为W[0..3]，加上k值
    MOV  $·_K+0(SB), X7         // 后续需要加对应的偏移量
    VLE32V		(X7), V31		// V32 = {K[0], K[1], K[2], K[3]}
    VADDVV		V1, V31, V1		// V1 = {W[0]+K[0], W[1]+K[1], W[2]+K[2], W[3]+K[3]}

    // 加载a, b, e, f
    MOV  $index_a_b_e_f<>(SB), X7
    VLE32V		(X7), V31
    VLUXEI32V	(X6), V31, V15  // V15 = {a, b, e, f}

    // 加载c, d, g, h
    MOV  $index_c_d_g_h<>(SB), X7
    VLE32V		(X7), V31
    VLUXEI32V	(X6), V31, V16  // V16 = {c, d, g, h}

    // v16、v15的值临时输出下
    MOV	temp_dig+32(FP), X7
    VSE32V		V15, (X7)
    ADD $16, X7, X8
    VSE32V		V16, (X8)

    // 把v1内元素的顺序反过来，先借助内存
    MOV	temp_dig+32(FP), X7
    VSE32V		V1, (X7)
    ADD $12, X7, X8
    MOV $-4, X9
    //VLSE32V		(X8), X9, V1

    // Go Plan9 汇编格式: VSHA2CLVV vs1, vs2, vd
    // vs1 = V1 = {W[0]+K[0], W[1]+K[1], ...}, VSHA2CL 使用低两个元素 [0] 和 [1]
    // vs2 = V15 = {a, b, e, f} 排列为 [f, e, b, a]
    // vd = V16 = {c, d, g, h} 排列为 [h, g, d, c], 输出为新的 {f,e,b,a}
    VSHA2CLVV   V1, V15, V16    // vs1=V1, vs2=V15, vd=V16 -> V16 = new {f,e,b,a}

    // store temp_dig
    MOV	temp_dig+32(FP), X7
    ADD $32, X7, X7
    VSE32V		V16, (X7)
    ADD $16, X7, X8
    VSE32V		V15, (X8)

    VSHA2CHVV   V1, V16, V15


    MOV	temp_dig+32(FP), X7
    ADD $96, X7, X7
    VSE32V		V15, (X7)
    ADD $16, X7, X8
    VSE32V		V16, (X8)

    // temp_kword is at temp_kword+40(FP)
    MOV	temp_kword+40(FP), X7
    VSE32V		V1, (X7)
	
	RET

