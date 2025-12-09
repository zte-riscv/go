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

// Index vector to reorder final state from scratch layout {f,e,b,a,h,g,d,c}
// into {a,b,c,d,e,f,g,h}
// offsets: a=28, b=24, c=12, d=8, e=4, f=0, g=20, h=16
DATA	index_final<>+0(SB)/4, $28
DATA	index_final<>+4(SB)/4, $24
DATA	index_final<>+8(SB)/4, $12
DATA	index_final<>+12(SB)/4, $8
DATA	index_final<>+16(SB)/4, $4
DATA	index_final<>+20(SB)/4, $0
DATA	index_final<>+24(SB)/4, $20
DATA	index_final<>+28(SB)/4, $16
GLOBL	index_final<>(SB), RODATA, $64


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

    // ---------------- Rounds 4、5、6、7 (use W[4..7] + K[4..7]) ----------------
    // Load W[4..7]
    ADD $16, X5, X9                  // X9 = p_base + 16
    VLE32V     (X9), V1              // big-endian W4..7
    VREV8V     V1, V1                // little-endian
    // Add K[4..7]
    MOV  $·_K+16(SB), X7
    VLE32V     (X7), V31
    VADDVV     V1, V31, V1           // V1 = {w4+K4, w5+K5, w6+K6, w7+K7}
    // Save temp_kword for rounds 4、5、6、7 (offset +16 bytes)
    MOV temp_kword+40(FP), X7
    ADD $16, X7, X7
    VSE32V     V1, (X7)

    // VSHA2CLVV: rounds 4,5 (low words W4,W5)
    VSHA2CLVV  V1, V15, V16          // V16 = new {f,e,b,a}
    // Store temp_dig (pair index 2) at offset +128
    MOV temp_dig+32(FP), X7
    ADD $160, X7, X7
    VSE32V     V16, (X7)
    ADD $16, X7, X8
    VSE32V     V15, (X8)

    // VSHA2CHVV: rounds 6,7 (high words W6,W7)
    VSHA2CHVV  V1, V16, V15          // V15 = new {f,e,b,a}
    // Store temp_dig (pair index 3) at offset +192
    MOV temp_dig+32(FP), X7
    ADD $224, X7, X7
    VSE32V     V15, (X7)
    ADD $16, X7, X8
    VSE32V     V16, (X8)

    // ---------------- Rounds 8、9、10、11 (use W[8..11] + K[8..11]) ----------------
    // Load W[8..11]
    ADD $32, X5, X9                  // X9 = p_base + 32
    VLE32V     (X9), V1              // big-endian W8..11
    VREV8V     V1, V1                // little-endian
    // Add K[8..11]
    MOV  $·_K+32(SB), X7
    VLE32V     (X7), V31
    VADDVV     V1, V31, V1           // V1 = {w8+K8, w9+K9, w10+K10, w11+K11}
    // Save temp_kword for rounds 8、9、10、11 (offset +32 bytes)
    MOV temp_kword+40(FP), X7
    ADD $32, X7, X7
    VSE32V     V1, (X7)

    // VSHA2CLVV: rounds 8,9 (low words W8,W9)
    VSHA2CLVV  V1, V15, V16          // V16 = new {f,e,b,a}
    // Store temp_dig (pair index 4) at offset +288
    MOV temp_dig+32(FP), X7
    ADD $288, X7, X7
    VSE32V     V16, (X7)
    ADD $16, X7, X8
    VSE32V     V15, (X8)

    // VSHA2CHVV: rounds 10,11 (high words W10,W11)
    VSHA2CHVV  V1, V16, V15          // V15 = new {f,e,b,a}
    // Store temp_dig (pair index 5) at offset +352
    MOV temp_dig+32(FP), X7
    ADD $352, X7, X7
    VSE32V     V15, (X7)
    ADD $16, X7, X8
    VSE32V     V16, (X8)

    // ---------------- Rounds 12、13、14、15 (use W[12..15] + K[12..15]) ----------------
    // Load W[12..15]
    ADD $48, X5, X9                  // X9 = p_base + 48
    VLE32V     (X9), V1              // big-endian W12..15
    VREV8V     V1, V1                // little-endian
    // Add K[12..15]
    MOV  $·_K+48(SB), X7
    VLE32V     (X7), V31
    VADDVV     V1, V31, V1           // V1 = {w12+K12, w13+K13, w14+K14, w15+K15}
    // Save temp_kword for rounds 12、13、14、15 (offset +48 bytes)
    MOV temp_kword+40(FP), X7
    ADD $48, X7, X7
    VSE32V     V1, (X7)

    // VSHA2CLVV: rounds 12,13 (low words W12,W13)
    VSHA2CLVV  V1, V15, V16          // V16 = new {f,e,b,a}
    // Store temp_dig (pair index 6) at offset +416
    MOV temp_dig+32(FP), X7
    ADD $416, X7, X7
    VSE32V     V16, (X7)
    ADD $16, X7, X8
    VSE32V     V15, (X8)

    // VSHA2CHVV: rounds 14,15 (high words W14,W15)
    VSHA2CHVV  V1, V16, V15          // V15 = new {f,e,b,a}
    // Store temp_dig (pair index 7) at offset +480
    MOV temp_dig+32(FP), X7
    ADD $480, X7, X7
    VSE32V     V15, (X7)
    ADD $16, X7, X8
    VSE32V     V16, (X8)

    // ---------------- Rounds 16、17、18、19 (use W[16..19] + K[16..19]) ----------------
    // Recompute W16..19 via VSHA2MS (no buffering)
    // Load W[0..3] -> V1
    VLE32V     (X5), V1
    VREV8V     V1, V1
    // Load {W4, W9, W10, W11} -> V2 using index table
    MOV  $index_w4_w9_w10_w11<>(SB), X7
    VLE32V     (X7), V31
    ADD  $16, X5, X9                  // p_base + 16
    VLUXEI32V  (X9), V31, V2
    VREV8V     V2, V2
    // Load W[12..15] -> V3
    ADD  $48, X5, X9
    VLE32V     (X9), V3
    VREV8V     V3, V3
    // Generate W[16..19] into V1 (descending order {W19,W18,W17,W16}) — keep desc for later rounds
    VSHA2MSVV  V3, V2, V1
    // Add K[16..19] into V30 (preserve V1 desc)
    MOV  $·_K+64(SB), X7
    VLE32V     (X7), V31
    VADDVV     V1, V31, V30          // V30 = {w16+K16, w17+K17, w18+K18, w19+K19}
    // Save temp_kword for rounds 16..19 (offset +64 bytes)
    MOV temp_kword+40(FP), X7
    ADD $64, X7, X7
    VSE32V     V30, (X7)

    // VSHA2CLVV: rounds 16,17 (use low words W16,W17)
    VSHA2CLVV  V30, V15, V16          // V16 = new {f,e,b,a}
    // Store temp_dig (pair index 8) at offset +544
    MOV temp_dig+32(FP), X7
    ADD $544, X7, X7
    VSE32V     V16, (X7)
    ADD $16, X7, X8
    VSE32V     V15, (X8)

    // VSHA2CHVV: rounds 18,19 (use high words W18,W19)
    VSHA2CHVV  V30, V16, V15          // V15 = new {f,e,b,a}
    // Store temp_dig (pair index 9) at offset +608
    MOV temp_dig+32(FP), X7
    ADD $608, X7, X7
    VSE32V     V15, (X7)
    ADD $16, X7, X8
    VSE32V     V16, (X8)

    // ---------------- Rounds 20、21、22、23 (use W[20..23] + K[20..23]) ----------------
    // Generate W20..23 via VSHA2MS (use same regs as msg_expand)
    // Load W[4..7] -> V4 (asc)
    ADD $16, X5, X9
    VLE32V     (X9), V4
    VREV8V     V4, V4
    // Load {W8, W13, W14, W15} -> V5 (asc via index)
    MOV  $index_w4_w9_w10_w11<>(SB), X7
    VLE32V     (X7), V31
    ADD $32, X5, X9                  // p_base + 32
    VLUXEI32V  (X9), V31, V5
    VREV8V     V5, V5
    // VSHA2MS: vd=V4 gets {W23,W22,W21,W20} (descending), vs2=V1 (desc W19..16)
    VSHA2MSVV  V1, V5, V4
    // Add K[20..23] into V30 (preserve V1 desc, V4 desc)
    MOV  $·_K+80(SB), X7
    VLE32V     (X7), V31
    VADDVV     V4, V31, V30         // V30 = {w20+K20, w21+K21, w22+K22, w23+K23}
    // Save temp_kword for rounds 20..23 (offset +80 bytes)
    MOV temp_kword+40(FP), X7
    ADD $80, X7, X7
    VSE32V     V30, (X7)

    // VSHA2CLVV: rounds 20,21 (use low words W20,W21)
    VSHA2CLVV  V30, V15, V16          // V16 = new {f,e,b,a}
    // Store temp_dig (pair index 10) at offset +672
    MOV temp_dig+32(FP), X7
    ADD $672, X7, X7
    VSE32V     V16, (X7)
    ADD $16, X7, X8
    VSE32V     V15, (X8)

    // VSHA2CHVV: rounds 22,23 (use high words W22,W23)
    VSHA2CHVV  V30, V16, V15          // V15 = new {f,e,b,a}
    // Store temp_dig (pair index 11) at offset +736
    MOV temp_dig+32(FP), X7
    ADD $736, X7, X7
    VSE32V     V15, (X7)
    ADD $16, X7, X8
    VSE32V     V16, (X8)

    // ---------------- Rounds 24、25、26、27 (use W[24..27] + K[24..27]) ----------------
    // Prepare inputs for VSHA2MS (follow msg_expand regs):
    //  - vs2: V8 should hold W27..W24 (desc) after previous step; rebuild from V4 desc if needed
 
    //  - vs1: merge {W12,W13,W14,?} and {W19,W18,W17,W16} per mask {1,1,1,0}
    ADD	$32, X5, X7		// 常数相比上一轮加16
	VLE32V		(X7), V6	// 寄存器号加2	
	VREV8V		V6, V6			
	//VMV1RV		V6, V6			
	

	MOV	$vreg_mask<>(SB), X7	
	VLE32V		(X7), V0	// 掩码 {1,1,1,0}
	VMERGEVVM	V3, V1, V0, V7 	// 合并两个源操作数，1时选择第一个v中的元素
	// VMV1RV		V7, V10	
		
	VSHA2MSVV	V4, V7, V6		

    // Add K[24..27] into V30 (preserve V6 desc)
    MOV  $·_K+96(SB), X7
    VLE32V     (X7), V31
    VADDVV     V6, V31, V30          // V30 = {w24+K24, w25+K25, w26+K26, w27+K27}
    // Save temp_kword for rounds 24..27 (offset +96 bytes)
    MOV temp_kword+40(FP), X7
    ADD $96, X7, X7
    VSE32V     V30, (X7)

    // VSHA2CLVV: rounds 24,25 (use low words W24,W25)
    VSHA2CLVV  V30, V15, V16          // V16 = new {f,e,b,a}
    // Store temp_dig (pair index 12) at offset +800
    MOV temp_dig+32(FP), X7
    ADD $800, X7, X7
    VSE32V     V16, (X7)
    ADD $16, X7, X8
    VSE32V     V15, (X8)

    // VSHA2CHVV: rounds 26,27 (use high words W26,W27)
    VSHA2CHVV  V30, V16, V15          // V15 = new {f,e,b,a}
    // Store temp_dig (pair index 13) at offset +864
    MOV temp_dig+32(FP), X7
    ADD $864, X7, X7
    VSE32V     V15, (X7)
    ADD $16, X7, X8
    VSE32V     V16, (X8)

    // ---------------- Rounds 28、29、30、31 (use W[28..31] + K[28..31]) ----------------
    // Align registers to msg_expand style for W28..31:
    // vs2: V8 holds W27..W24 (desc) from previous step
    // vs1: merge V4(desc W23..20) and V6(desc W27..24) with mask -> V31 = {W23,W22,W21,W24}
    ADD	$48, X5, X7		// 常数相比上一轮加16
	VLE32V		(X7), V8	// 寄存器号加2	
	VREV8V		V8, V8					

	VMERGEVVM	V1, V4, V0, V9 	// 合并两个源操作数，1时选择第一个v中的元素	
		
	VSHA2MSVV	V6, V9, V8	
    // Add K[28..31] into V30 (preserve V9)
    MOV  $·_K+112(SB), X7
    VLE32V     (X7), V31
    VADDVV     V8, V31, V30           // V30 = {w28+K28, w29+K29, w30+K30, w31+K31}
    // Save temp_kword for rounds 28..31 (offset +112 bytes)
    MOV temp_kword+40(FP), X7
    ADD $112, X7, X7
    VSE32V     V30, (X7)

    // VSHA2CLVV: rounds 28,29 (use low words W28,W29)
    VSHA2CLVV  V30, V15, V16          // V16 = new {f,e,b,a}
    // Store temp_dig (pair index 14) at offset +928
    MOV temp_dig+32(FP), X7
    ADD $928, X7, X7
    VSE32V     V16, (X7)
    ADD $16, X7, X8
    VSE32V     V15, (X8)

    // VSHA2CHVV: rounds 30,31 (use high words W30,W31)
    VSHA2CHVV  V30, V16, V15          // V15 = new {f,e,b,a}
    // Store temp_dig (pair index 15) at offset +992
    MOV temp_dig+32(FP), X7
    ADD $992, X7, X7
    VSE32V     V15, (X7)
    ADD $16, X7, X8
    VSE32V     V16, (X8)



// 第五个
	VMERGEVVM	V4, V6, V0, V31 	// 合并两个源操作数，1时选择第一个v中的元素	
		
	VSHA2MSVV	V8, V31, V1		
	// v1 当前为 {W35, W34, W33, W32} (desc)
	VMV1RV		V1, V30				// 升序 {W32,W33,W34,W35}
	MOV	$·_K+128(SB), X7
	VLE32V		(X7), V31
	VADDVV		V30, V31, V30		// V30 = {w32+K32, w33+K33, w34+K34, w35+K35}
	// temp_kword offset +128
	MOV		temp_kword+40(FP), X7
	ADD	$128, X7, X7
	VSE32V		V30, (X7)
	// VSHA2CLVV rounds 32,33
	VSHA2CLVV	V30, V15, V16
	// temp_dig pair index 16 offset +1056
	MOV		temp_dig+32(FP), X7
	ADD	$1056, X7, X7
	VSE32V		V16, (X7)
	ADD	$16, X7, X8
	VSE32V		V15, (X8)
	// VSHA2CHVV rounds 34,35
	VSHA2CHVV	V30, V16, V15
	// temp_dig pair index 17 offset +1120
	MOV		temp_dig+32(FP), X7
	ADD	$1120, X7, X7
	VSE32V		V15, (X7)
	ADD	$16, X7, X8
	VSE32V		V16, (X8)
	

// 第六个
	VMERGEVVM	V6, V8, V0, V31 	// 合并两个源操作数，1时选择第一个v中的元素	
		
	VSHA2MSVV	V1, V31, V4		
    // v4 当前为 {W39, W38, W37, W36} (desc)
	VMV1RV		V4, V30				// 升序 {W36,W37,W38,W39}
	MOV	$·_K+144(SB), X7
	VLE32V		(X7), V31
	VADDVV		V30, V31, V30		// V30 = {w36+K36, w37+K37, w38+K38, w39+K39}
	// temp_kword offset +144
	MOV		temp_kword+40(FP), X7
	ADD	$144, X7, X7
	VSE32V		V30, (X7)
	// VSHA2CLVV rounds 36,37
	VSHA2CLVV	V30, V15, V16
	// temp_dig pair index 18 offset +1184
	MOV		temp_dig+32(FP), X7
	ADD	$1184, X7, X7
	VSE32V		V16, (X7)
	ADD	$16, X7, X8
	VSE32V		V15, (X8)
	// VSHA2CHVV rounds 38,39
	VSHA2CHVV	V30, V16, V15
	// temp_dig pair index 19 offset +1248
	MOV		temp_dig+32(FP), X7
	ADD	$1248, X7, X7
	VSE32V		V15, (X7)
	ADD	$16, X7, X8
	VSE32V		V16, (X8)


// 第七个
	VMERGEVVM	V8, V1, V0, V31 	// 合并两个源操作数，1时选择第一个v中的元素	
		
	VSHA2MSVV	V4, V31, V6		
	// v6 当前为 {W43, W42, W41, W40} (desc)
	VMV1RV		V6, V30				// 升序 {W40,W41,W42,W43}
	MOV	$·_K+160(SB), X7
	VLE32V		(X7), V31
	VADDVV		V30, V31, V30		// V30 = {w40+K40, w41+K41, w42+K42, w43+K43}
	// temp_kword offset +160
	MOV		temp_kword+40(FP), X7
	ADD	$160, X7, X7
	VSE32V		V30, (X7)
	// VSHA2CLVV rounds 40,41
	VSHA2CLVV	V30, V15, V16
	// temp_dig pair index 20 offset +1312
	MOV		temp_dig+32(FP), X7
	ADD	$1312, X7, X7
	VSE32V		V16, (X7)
	ADD	$16, X7, X8
	VSE32V		V15, (X8)
	// VSHA2CHVV rounds 42,43
	VSHA2CHVV	V30, V16, V15
	// temp_dig pair index 21 offset +1376
	MOV		temp_dig+32(FP), X7
	ADD	$1376, X7, X7
	VSE32V		V15, (X7)
	ADD	$16, X7, X8
	VSE32V		V16, (X8)

// 第八个
	VMERGEVVM	V1, V4, V0, V31 	// 合并两个源操作数，1时选择第一个v中的元素	
		
	VSHA2MSVV	V6, V31, V8		
	// v8 当前为 {W47, W46, W45, W44} (desc)
	VMV1RV		V8, V30				// 升序 {W44,W45,W46,W47}
	MOV	$·_K+176(SB), X7
	VLE32V		(X7), V31
	VADDVV		V30, V31, V30		// V30 = {w44+K44, w45+K45, w46+K46, w47+K47}
	// temp_kword offset +176
	MOV		temp_kword+40(FP), X7
	ADD	$176, X7, X7
	VSE32V		V30, (X7)
	// VSHA2CLVV rounds 44,45
	VSHA2CLVV	V30, V15, V16
	// temp_dig pair index 22 offset +1440
	MOV		temp_dig+32(FP), X7
	ADD	$1440, X7, X7
	VSE32V		V16, (X7)
	ADD	$16, X7, X8
	VSE32V		V15, (X8)
	// VSHA2CHVV rounds 46,47
	VSHA2CHVV	V30, V16, V15
	// temp_dig pair index 23 offset +1504
	MOV		temp_dig+32(FP), X7
	ADD	$1504, X7, X7
	VSE32V		V15, (X7)
	ADD	$16, X7, X8
	VSE32V		V16, (X8)

// 第九个
	VMERGEVVM	V4, V6, V0, V31 	// 合并两个源操作数，1时选择第一个v中的元素	
		
	VSHA2MSVV	V8, V31, V1		
	// v1 当前为 {W51, W50, W49, W48} (desc)
	VMV1RV		V1, V30				// 升序 {W48,W49,W50,W51}
	MOV	$·_K+192(SB), X7
	VLE32V		(X7), V31
	VADDVV		V30, V31, V30		// V30 = {w48+K48, w49+K49, w50+K50, w51+K51}
	// temp_kword offset +192
	MOV		temp_kword+40(FP), X7
	ADD	$192, X7, X7
	VSE32V		V30, (X7)
	// VSHA2CLVV rounds 48,49
	VSHA2CLVV	V30, V15, V16
	// temp_dig pair index 24 offset +1568
	MOV		temp_dig+32(FP), X7
	ADD	$1568, X7, X7
	VSE32V		V16, (X7)
	ADD	$16, X7, X8
	VSE32V		V15, (X8)
	// VSHA2CHVV rounds 50,51
	VSHA2CHVV	V30, V16, V15
	// temp_dig pair index 25 offset +1632
	MOV		temp_dig+32(FP), X7
	ADD	$1632, X7, X7
	VSE32V		V15, (X7)
	ADD	$16, X7, X8
	VSE32V		V16, (X8)

// 第十个
	VMERGEVVM	V6, V8, V0, V31 	// 合并两个源操作数，1时选择第一个v中的元素	
		
	VSHA2MSVV	V1, V31, V4		
	// v4 当前为 {W55, W54, W53, W52} (desc)
	VMV1RV		V4, V30				// 升序 {W52,W53,W54,W55}
	MOV	$·_K+208(SB), X7
	VLE32V		(X7), V31
	VADDVV		V30, V31, V30		// V30 = {w52+K52, w53+K53, w54+K54, w55+K55}
	// temp_kword offset +208
	MOV		temp_kword+40(FP), X7
	ADD	$208, X7, X7
	VSE32V		V30, (X7)
	// VSHA2CLVV rounds 52,53
	VSHA2CLVV	V30, V15, V16
	// temp_dig pair index 26 offset +1696
	MOV		temp_dig+32(FP), X7
	ADD	$1696, X7, X7
	VSE32V		V16, (X7)
	ADD	$16, X7, X8
	VSE32V		V15, (X8)
	// VSHA2CHVV rounds 54,55
	VSHA2CHVV	V30, V16, V15
	// temp_dig pair index 27 offset +1760
	MOV		temp_dig+32(FP), X7
	ADD	$1760, X7, X7
	VSE32V		V15, (X7)
	ADD	$16, X7, X8
	VSE32V		V16, (X8)

// 第十一个
	VMERGEVVM	V8, V1, V0, V31 	// 合并两个源操作数，1时选择第一个v中的元素	
		
	VSHA2MSVV	V4, V31, V6		
	// v6 当前为 {W59, W58, W57, W56} (desc)
	VMV1RV		V6, V30				// 升序 {W56,W57,W58,W59}
	MOV	$·_K+224(SB), X7
	VLE32V		(X7), V31
	VADDVV		V30, V31, V30		// V30 = {w56+K56, w57+K57, w58+K58, w59+K59}
	// temp_kword offset +224
	MOV		temp_kword+40(FP), X7
	ADD	$224, X7, X7
	VSE32V		V30, (X7)
	// VSHA2CLVV rounds 56,57
	VSHA2CLVV	V30, V15, V16
	// temp_dig pair index 28 offset +1824
	MOV		temp_dig+32(FP), X7
	ADD	$1824, X7, X7
	VSE32V		V16, (X7)
	ADD	$16, X7, X8
	VSE32V		V15, (X8)
	// VSHA2CHVV rounds 58,59
	VSHA2CHVV	V30, V16, V15
	// temp_dig pair index 29 offset +1888
	MOV		temp_dig+32(FP), X7
	ADD	$1888, X7, X7
	VSE32V		V15, (X7)
	ADD	$16, X7, X8
	VSE32V		V16, (X8)

// 第十二个
	VMERGEVVM	V1, V4, V0, V31 	// 合并两个源操作数，1时选择第一个v中的元素	
		
	VSHA2MSVV	V6, V31, V8		
	// v8 当前为 {W63, W62, W61, W60} (desc)
	VMV1RV		V8, V30				// 升序 {W60,W61,W62,W63}
	MOV	$·_K+240(SB), X7
	VLE32V		(X7), V31
	VADDVV		V30, V31, V30		// V30 = {w60+K60, w61+K61, w62+K62, w63+K63}
	// temp_kword offset +240
	MOV		temp_kword+40(FP), X7
	ADD	$240, X7, X7
	VSE32V		V30, (X7)
	// VSHA2CLVV rounds 60,61
	VSHA2CLVV	V30, V15, V16
	// temp_dig pair index 30 offset +1952
	MOV		temp_dig+32(FP), X7
	ADD	$1952, X7, X7
	VSE32V		V16, (X7)
	ADD	$16, X7, X8
	VSE32V		V15, (X8)
	// VSHA2CHVV rounds 62,63
	VSHA2CHVV	V30, V16, V15
	// temp_dig pair index 31 offset +2016
	MOV		temp_dig+32(FP), X7
	ADD	$2016, X7, X7
	VSE32V		V15, (X7)
	ADD	$16, X7, X8
	VSE32V		V16, (X8)
	
	// ---------- Final output of a..h (no accumulation) ----------
    VLE32V		(X6), V1
    ADD	$16, X6, X7
    VLE32V		(X7), V2


	// Scratch to reorder: use temp_dig tail
	// Store final state into scratch: [f,e,b,a] then [h,g,d,c]
	VSE32V		V15, (X6)
	VSE32V		V16, (X7)
	// Scalar load to avoid index ambiguity (corrected offsets)
	// Layout in scratch: 0:f,4:e,8:b,12:a,16:h,20:g,24:d,28:c
	MOVW	12(X6), X9          // a
	MOVW	8(X6), X10          // b
	MOVW	28(X6), X11         // c
	MOVW	24(X6), X12         // d
	MOVW	4(X6), X13          // e
	MOVW	0(X6), X14          // f
	MOVW	20(X6), X15         // g
	MOVW	16(X6), X5          // h

	// Store to digest
    MOV	dig+0(FP), X6
	MOVW	X9, 0(X6)
	MOVW	X10, 4(X6)
	MOVW	X11, 8(X6)
	MOVW	X12, 12(X6)
	ADD	$16, X6, X8
	MOVW	X13, 0(X8)
	MOVW	X14, 4(X8)
	MOVW	X15, 8(X8)
	MOVW	X5, 12(X8)

    VLE32V		(X6), V3
    VLE32V		(X7), V4

    VADDVV		V1, V3, V1
    VADDVV		V2, V4, V2

    VSE32V		V1, (X6)
    VSE32V		V2, (X7)

	RET

