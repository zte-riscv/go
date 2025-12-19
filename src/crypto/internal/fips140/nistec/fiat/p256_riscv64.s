// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// P256 常量
DATA p256const0<>+0x00(SB)/8, $0x00000000ffffffff
DATA p256const1<>+0x00(SB)/8, $0xffffffff00000001
GLOBL p256const0<>(SB), RODATA, $8
GLOBL p256const1<>(SB), RODATA, $8

// func p256Mul(out, a, b *p256MontgomeryDomainFieldElement)
// RISC-V 汇编实现的 P256 蒙哥马利乘法
// 寄存器分配：
//   X10 = out 指针
//   X11 = a 指针 (临时), 后用作临时变量
//   X12 = b 指针 (临时)
//   X13-X16 = x0-x3 (输入 a)
//   X17-X20 = y0-y3 (输入 b)
//   X21-X24 = acc0-acc3 (累加器低位)
//   X25-X26, X28-X29 = acc4-acc7 (累加器高位)
//   X30 = const1 = 0xFFFFFFFF00000001
//   X5-X9, X12 = 临时变量
TEXT ·p256Mul(SB),NOSPLIT,$0-24
	MOV	out+0(FP), X10
	MOV	a+8(FP), X11
	MOV	b+16(FP), X12

	// 加载 const1
	MOV	$p256const1<>(SB), X30
	MOV	(X30), X30

	// 加载输入 x (a): X13-X16 = x0-x3
	MOV	0(X11), X13
	MOV	8(X11), X14
	MOV	16(X11), X15
	MOV	24(X11), X16

	// 加载输入 y (b): X17-X20 = y0-y3
	MOV	0(X12), X17
	MOV	8(X12), X18
	MOV	16(X12), X19
	MOV	24(X12), X20

	// =====================================
	// Round 0: y[0] * x
	// =====================================
	MUL	X17, X13, X21		// acc0
	MULHU	X17, X13, X22		// acc1

	MUL	X17, X14, X5
	MULHU	X17, X14, X23
	ADD	X22, X5, X22
	SLTU	X5, X22, X6
	ADD	X23, X6, X23

	MUL	X17, X15, X5
	MULHU	X17, X15, X24
	ADD	X23, X5, X23
	SLTU	X5, X23, X6
	ADD	X24, X6, X24

	MUL	X17, X16, X5
	MULHU	X17, X16, X25
	ADD	X24, X5, X24
	SLTU	X5, X24, X6
	ADD	X25, X6, X25

	// {x25 x24 x23 x22 x21}

	// =====================================
	// First reduction step
	// =====================================
	SLL	$32, X21, X5		// t0 = acc0 << 32
	SRL	$32, X21, X7		// t1 = acc0 >> 32
	MUL	X21, X30, X8		// t2 = lo(acc0 * const1)
	MULHU	X21, X30, X21		// acc0 = hi(acc0 * const1)

	ADD	X22, X5, X22
	SLTU	X5, X22, X6

	ADD	X23, X7, X23
	SLTU	X7, X23, X9
	ADD	X23, X6, X23
	SLTU	X6, X23, X5
	ADD	X9, X5, X6

	ADD	X24, X8, X24
	SLTU	X8, X24, X9
	ADD	X24, X6, X24
	SLTU	X6, X24, X5
	ADD	X9, X5, X6

	ADD	X21, X6, X21

	// =====================================
	// Round 1: y[1] * x
	// =====================================
	MULHU	X18, X13, X7
	MULHU	X18, X14, X8
	MULHU	X18, X15, X9
	MULHU	X18, X16, X26

	MUL	X18, X13, X5
	ADD	X22, X5, X22
	SLTU	X5, X22, X6

	MUL	X18, X14, X5
	ADD	X23, X5, X23
	SLTU	X5, X23, X12
	ADD	X23, X6, X23
	SLTU	X6, X23, X5
	ADD	X12, X5, X6

	MUL	X18, X15, X5
	ADD	X24, X5, X24
	SLTU	X5, X24, X12
	ADD	X24, X6, X24
	SLTU	X6, X24, X5
	ADD	X12, X5, X6

	MUL	X18, X16, X5
	ADD	X25, X5, X25
	SLTU	X5, X25, X12
	ADD	X25, X6, X25
	SLTU	X6, X25, X5
	ADD	X12, X5, X6

	MOV	X6, X29			// 暂存 carry_lo

	ADD	X23, X7, X23
	SLTU	X7, X23, X6

	ADD	X24, X8, X24
	SLTU	X8, X24, X5
	ADD	X24, X6, X24
	SLTU	X6, X24, X12
	ADD	X5, X12, X6

	ADD	X25, X9, X25
	SLTU	X9, X25, X5
	ADD	X25, X6, X25
	SLTU	X6, X25, X12
	ADD	X5, X12, X6

	ADD	X26, X29, X26		// acc5 = hi(y1*x3) + carry_lo
	ADD	X26, X6, X26		// acc5 += carry_hi

	// =====================================
	// Second reduction step
	// =====================================
	SLL	$32, X22, X5
	SRL	$32, X22, X7
	MUL	X22, X30, X8
	MULHU	X22, X30, X22

	ADD	X23, X5, X23
	SLTU	X5, X23, X6

	ADD	X24, X7, X24
	SLTU	X7, X24, X9
	ADD	X24, X6, X24
	SLTU	X6, X24, X5
	ADD	X9, X5, X6

	ADD	X21, X8, X21
	SLTU	X8, X21, X9
	ADD	X21, X6, X21
	SLTU	X6, X21, X5
	ADD	X9, X5, X6

	ADD	X22, X6, X22

	// =====================================
	// Round 2: y[2] * x
	// =====================================
	MULHU	X19, X13, X7
	MULHU	X19, X14, X8
	MULHU	X19, X15, X9
	MULHU	X19, X16, X28

	MUL	X19, X13, X5
	ADD	X23, X5, X23
	SLTU	X5, X23, X6

	MUL	X19, X14, X5
	ADD	X24, X5, X24
	SLTU	X5, X24, X12
	ADD	X24, X6, X24
	SLTU	X6, X24, X5
	ADD	X12, X5, X6

	MUL	X19, X15, X5
	ADD	X25, X5, X25
	SLTU	X5, X25, X12
	ADD	X25, X6, X25
	SLTU	X6, X25, X5
	ADD	X12, X5, X6

	MUL	X19, X16, X5
	ADD	X26, X5, X26
	SLTU	X5, X26, X12
	ADD	X26, X6, X26
	SLTU	X6, X26, X5
	ADD	X12, X5, X6

	MOV	X6, X29

	ADD	X24, X7, X24
	SLTU	X7, X24, X6

	ADD	X25, X8, X25
	SLTU	X8, X25, X5
	ADD	X25, X6, X25
	SLTU	X6, X25, X12
	ADD	X5, X12, X6

	ADD	X26, X9, X26
	SLTU	X9, X26, X5
	ADD	X26, X6, X26
	SLTU	X6, X26, X12
	ADD	X5, X12, X6

	ADD	X28, X29, X28
	ADD	X28, X6, X28

	MOV	X0, X29

	// =====================================
	// Third reduction step
	// =====================================
	SLL	$32, X23, X5
	SRL	$32, X23, X7
	MUL	X23, X30, X8
	MULHU	X23, X30, X23

	ADD	X24, X5, X24
	SLTU	X5, X24, X6

	ADD	X21, X7, X21
	SLTU	X7, X21, X9
	ADD	X21, X6, X21
	SLTU	X6, X21, X5
	ADD	X9, X5, X6

	ADD	X22, X8, X22
	SLTU	X8, X22, X9
	ADD	X22, X6, X22
	SLTU	X6, X22, X5
	ADD	X9, X5, X6

	ADD	X23, X6, X23

	// =====================================
	// Round 3: y[3] * x
	// =====================================
	MULHU	X20, X13, X7
	MULHU	X20, X14, X8
	MULHU	X20, X15, X9
	MULHU	X20, X16, X29

	MUL	X20, X13, X5
	ADD	X24, X5, X24
	SLTU	X5, X24, X6

	MUL	X20, X14, X5
	ADD	X25, X5, X25
	SLTU	X5, X25, X12
	ADD	X25, X6, X25
	SLTU	X6, X25, X5
	ADD	X12, X5, X6

	MUL	X20, X15, X5
	ADD	X26, X5, X26
	SLTU	X5, X26, X12
	ADD	X26, X6, X26
	SLTU	X6, X26, X5
	ADD	X12, X5, X6

	MUL	X20, X16, X5
	ADD	X28, X5, X28
	SLTU	X5, X28, X12
	ADD	X28, X6, X28
	SLTU	X6, X28, X5
	ADD	X12, X5, X6

	MOV	X6, X12			// 暂存 carry_lo

	ADD	X25, X7, X25
	SLTU	X7, X25, X6

	ADD	X26, X8, X26
	SLTU	X8, X26, X5
	ADD	X26, X6, X26
	SLTU	X6, X26, X7
	ADD	X5, X7, X6

	ADD	X28, X9, X28
	SLTU	X9, X28, X5
	ADD	X28, X6, X28
	SLTU	X6, X28, X7
	ADD	X5, X7, X6

	ADD	X29, X12, X29
	ADD	X29, X6, X29

	// =====================================
	// Last reduction step
	// =====================================
	SLL	$32, X24, X5
	SRL	$32, X24, X7
	MUL	X24, X30, X8
	MULHU	X24, X30, X24

	ADD	X21, X5, X21
	SLTU	X5, X21, X6

	ADD	X22, X7, X22
	SLTU	X7, X22, X9
	ADD	X22, X6, X22
	SLTU	X6, X22, X5
	ADD	X9, X5, X6

	ADD	X23, X8, X23
	SLTU	X8, X23, X9
	ADD	X23, X6, X23
	SLTU	X6, X23, X5
	ADD	X9, X5, X6

	ADD	X24, X6, X24

	// =====================================
	// Add bits [511:256]
	// =====================================
	ADD	X21, X25, X21
	SLTU	X25, X21, X6

	ADD	X22, X26, X22
	SLTU	X26, X22, X9
	ADD	X22, X6, X22
	SLTU	X6, X22, X5
	ADD	X9, X5, X6

	ADD	X23, X28, X23
	SLTU	X28, X23, X9
	ADD	X23, X6, X23
	SLTU	X6, X23, X5
	ADD	X9, X5, X6

	ADD	X24, X29, X24
	SLTU	X29, X24, X9
	ADD	X24, X6, X24
	SLTU	X6, X24, X5
	ADD	X9, X5, X25

	// =====================================
	// Conditional subtraction
	// =====================================
	MOV	$p256const0<>(SB), X5
	MOV	(X5), X5		// X5 = const0 = 0x00000000FFFFFFFF
	MOV	$-1, X6			// X6 = p0 = 0xFFFFFFFFFFFFFFFF

	// Step 1: t0, b0 = Sub64(acc0, p0, 0)
	ADD	$1, X21, X13		// X13 = t0 = acc0 + 1
	SLTU	X6, X21, X7		// X7 = b0 = (acc0 < p0)

	// Step 2: t1, b1 = Sub64(acc1, p1, b0)
	SUB	X5, X22, X8		// X8 = tmp = acc1 - const0
	SLTU	X5, X22, X9		// X9 = (acc1 < const0)
	SUB	X7, X8, X14		// X14 = t1 = tmp - b0
	SLTU	X7, X8, X12		// X12 = (tmp < b0)
	OR	X9, X12, X7		// X7 = b1

	// Step 3: t2, b2 = Sub64(acc2, 0, b1)
	SUB	X7, X23, X15		// X15 = t2 = acc2 - b1
	SLTU	X7, X23, X7		// X7 = b2 = (acc2 < b1)

	// Step 4: t3, b3 = Sub64(acc3, p3, b2)
	SUB	X30, X24, X8		// X8 = tmp = acc3 - const1
	SLTU	X30, X24, X9		// X9 = (acc3 < const1)
	SUB	X7, X8, X16		// X16 = t3 = tmp - b2
	SLTU	X7, X8, X12		// X12 = (tmp < b2)
	OR	X9, X12, X7		// X7 = b3

	// Step 5: final_borrow = (carry < b3)
	SLTU	X7, X25, X8		// X8 = (carry < b3) = final_borrow

	// Conditional select: mask = final_borrow ? -1 : 0
	NEG	X8, X8			// X8 = mask

	// Select result
	XOR	X21, X13, X9
	AND	X8, X9, X9
	XOR	X9, X13, X13

	XOR	X22, X14, X9
	AND	X8, X9, X9
	XOR	X9, X14, X14

	XOR	X23, X15, X9
	AND	X8, X9, X9
	XOR	X9, X15, X15

	XOR	X24, X16, X9
	AND	X8, X9, X9
	XOR	X9, X16, X16

	// Write result
	MOV	X13, 0(X10)
	MOV	X14, 8(X10)
	MOV	X15, 16(X10)
	MOV	X16, 24(X10)

	RET

// func p256MulWithDebug(out, a, b *p256MontgomeryDomainFieldElement, states *[8]DebugState)
// 带调试输出的实现，每个 DebugState 包含 8 个 uint64 (64 bytes)
// states 指向 8 个 DebugState = 512 bytes
TEXT ·p256MulWithDebug(SB),NOSPLIT,$0-32
	MOV	out+0(FP), X10
	MOV	a+8(FP), X11
	MOV	b+16(FP), X12
	MOV	states+24(FP), X31	// states 指针 (注意：在函数内部使用后要恢复)

	// 加载 const1
	MOV	$p256const1<>(SB), X30
	MOV	(X30), X30

	// 加载输入 x (a): X13-X16 = x0-x3
	MOV	0(X11), X13
	MOV	8(X11), X14
	MOV	16(X11), X15
	MOV	24(X11), X16

	// 加载输入 y (b): X17-X20 = y0-y3
	MOV	0(X12), X17
	MOV	8(X12), X18
	MOV	16(X12), X19
	MOV	24(X12), X20

	// 保存 states 指针到栈
	MOV	states+24(FP), X11	// 重新加载 states 指针 (X11 不再需要)

	// =====================================
	// Round 0: y[0] * x
	// =====================================
	MUL	X17, X13, X21		// acc0
	MULHU	X17, X13, X22		// acc1

	MUL	X17, X14, X5
	MULHU	X17, X14, X23
	ADD	X22, X5, X22
	SLTU	X5, X22, X6
	ADD	X23, X6, X23

	MUL	X17, X15, X5
	MULHU	X17, X15, X24
	ADD	X23, X5, X23
	SLTU	X5, X23, X6
	ADD	X24, X6, X24

	MUL	X17, X16, X5
	MULHU	X17, X16, X25
	ADD	X24, X5, X24
	SLTU	X5, X24, X6
	ADD	X25, X6, X25

	// 初始化 acc5, acc6, acc7 为 0
	MOV	X0, X26
	MOV	X0, X28
	MOV	X0, X29

	// 保存 After Round 0 状态 (state[0], offset 0)
	MOV	X21, 0(X11)	// acc0
	MOV	X22, 8(X11)	// acc1
	MOV	X23, 16(X11)	// acc2
	MOV	X24, 24(X11)	// acc3
	MOV	X25, 32(X11)	// acc4
	MOV	X26, 40(X11)	// acc5
	MOV	X28, 48(X11)	// acc6
	MOV	X29, 56(X11)	// acc7

	// =====================================
	// First reduction step
	// =====================================
	SLL	$32, X21, X5		// t0 = acc0 << 32
	SRL	$32, X21, X7		// t1 = acc0 >> 32
	MUL	X21, X30, X8		// t2 = lo(acc0 * const1)
	MULHU	X21, X30, X21		// acc0 = hi(acc0 * const1)

	ADD	X22, X5, X22
	SLTU	X5, X22, X6

	ADD	X23, X7, X23
	SLTU	X7, X23, X9
	ADD	X23, X6, X23
	SLTU	X6, X23, X5
	ADD	X9, X5, X6

	ADD	X24, X8, X24
	SLTU	X8, X24, X9
	ADD	X24, X6, X24
	SLTU	X6, X24, X5
	ADD	X9, X5, X6

	ADD	X21, X6, X21

	// 保存 After Reduce 1 状态 (state[1], offset 64)
	MOV	X21, 64(X11)
	MOV	X22, 72(X11)
	MOV	X23, 80(X11)
	MOV	X24, 88(X11)
	MOV	X25, 96(X11)
	MOV	X26, 104(X11)
	MOV	X28, 112(X11)
	MOV	X29, 120(X11)

	// =====================================
	// Round 1: y[1] * x
	// =====================================
	// 保存高位
	MULHU	X18, X13, X7
	MULHU	X18, X14, X8
	MULHU	X18, X15, X9
	MULHU	X18, X16, X26

	// 加低位
	MUL	X18, X13, X5
	ADD	X22, X5, X22
	SLTU	X5, X22, X6

	MUL	X18, X14, X5
	ADD	X23, X5, X23
	SLTU	X5, X23, X12
	ADD	X23, X6, X23
	SLTU	X6, X23, X5
	ADD	X12, X5, X6

	MUL	X18, X15, X5
	ADD	X24, X5, X24
	SLTU	X5, X24, X12
	ADD	X24, X6, X24
	SLTU	X6, X24, X5
	ADD	X12, X5, X6

	MUL	X18, X16, X5
	ADD	X25, X5, X25
	SLTU	X5, X25, X12
	ADD	X25, X6, X25
	SLTU	X6, X25, X5
	ADD	X12, X5, X6

	MOV	X6, X29		// 暂存 carry_lo

	// 加高位
	ADD	X23, X7, X23
	SLTU	X7, X23, X6

	ADD	X24, X8, X24
	SLTU	X8, X24, X5
	ADD	X24, X6, X24
	SLTU	X6, X24, X12
	ADD	X5, X12, X6

	ADD	X25, X9, X25
	SLTU	X9, X25, X5
	ADD	X25, X6, X25
	SLTU	X6, X25, X12
	ADD	X5, X12, X6

	ADD	X26, X29, X26		// acc5 = hi(y1*x3) + carry_lo
	ADD	X26, X6, X26		// acc5 += carry_hi

	MOV	X0, X28			// acc6 = 0
	MOV	X0, X29			// acc7 = 0

	// 保存 After Round 1 状态 (state[2], offset 128)
	MOV	X21, 128(X11)
	MOV	X22, 136(X11)
	MOV	X23, 144(X11)
	MOV	X24, 152(X11)
	MOV	X25, 160(X11)
	MOV	X26, 168(X11)
	MOV	X28, 176(X11)
	MOV	X29, 184(X11)

	// =====================================
	// Second reduction step
	// =====================================
	SLL	$32, X22, X5
	SRL	$32, X22, X7
	MUL	X22, X30, X8
	MULHU	X22, X30, X22

	ADD	X23, X5, X23
	SLTU	X5, X23, X6

	ADD	X24, X7, X24
	SLTU	X7, X24, X9
	ADD	X24, X6, X24
	SLTU	X6, X24, X5
	ADD	X9, X5, X6

	ADD	X21, X8, X21
	SLTU	X8, X21, X9
	ADD	X21, X6, X21
	SLTU	X6, X21, X5
	ADD	X9, X5, X6

	ADD	X22, X6, X22

	// 保存 After Reduce 2 状态 (state[3], offset 192)
	MOV	X21, 192(X11)
	MOV	X22, 200(X11)
	MOV	X23, 208(X11)
	MOV	X24, 216(X11)
	MOV	X25, 224(X11)
	MOV	X26, 232(X11)
	MOV	X28, 240(X11)
	MOV	X29, 248(X11)

	// =====================================
	// Round 2: y[2] * x
	// =====================================
	MULHU	X19, X13, X7
	MULHU	X19, X14, X8
	MULHU	X19, X15, X9
	MULHU	X19, X16, X28

	MUL	X19, X13, X5
	ADD	X23, X5, X23
	SLTU	X5, X23, X6

	MUL	X19, X14, X5
	ADD	X24, X5, X24
	SLTU	X5, X24, X12
	ADD	X24, X6, X24
	SLTU	X6, X24, X5
	ADD	X12, X5, X6

	MUL	X19, X15, X5
	ADD	X25, X5, X25
	SLTU	X5, X25, X12
	ADD	X25, X6, X25
	SLTU	X6, X25, X5
	ADD	X12, X5, X6

	MUL	X19, X16, X5
	ADD	X26, X5, X26
	SLTU	X5, X26, X12
	ADD	X26, X6, X26
	SLTU	X6, X26, X5
	ADD	X12, X5, X6

	MOV	X6, X29

	ADD	X24, X7, X24
	SLTU	X7, X24, X6

	ADD	X25, X8, X25
	SLTU	X8, X25, X5
	ADD	X25, X6, X25
	SLTU	X6, X25, X12
	ADD	X5, X12, X6

	ADD	X26, X9, X26
	SLTU	X9, X26, X5
	ADD	X26, X6, X26
	SLTU	X6, X26, X12
	ADD	X5, X12, X6

	ADD	X28, X29, X28
	ADD	X28, X6, X28

	MOV	X0, X29

	// 保存 After Round 2 状态 (state[4], offset 256)
	MOV	X21, 256(X11)
	MOV	X22, 264(X11)
	MOV	X23, 272(X11)
	MOV	X24, 280(X11)
	MOV	X25, 288(X11)
	MOV	X26, 296(X11)
	MOV	X28, 304(X11)
	MOV	X29, 312(X11)

	// =====================================
	// Third reduction step
	// =====================================
	SLL	$32, X23, X5
	SRL	$32, X23, X7
	MUL	X23, X30, X8
	MULHU	X23, X30, X23

	ADD	X24, X5, X24
	SLTU	X5, X24, X6

	ADD	X21, X7, X21
	SLTU	X7, X21, X9
	ADD	X21, X6, X21
	SLTU	X6, X21, X5
	ADD	X9, X5, X6

	ADD	X22, X8, X22
	SLTU	X8, X22, X9
	ADD	X22, X6, X22
	SLTU	X6, X22, X5
	ADD	X9, X5, X6

	ADD	X23, X6, X23

	// 保存 After Reduce 3 状态 (state[5], offset 320)
	MOV	X21, 320(X11)
	MOV	X22, 328(X11)
	MOV	X23, 336(X11)
	MOV	X24, 344(X11)
	MOV	X25, 352(X11)
	MOV	X26, 360(X11)
	MOV	X28, 368(X11)
	MOV	X29, 376(X11)

	// =====================================
	// Round 3: y[3] * x
	// =====================================
	MULHU	X20, X13, X7
	MULHU	X20, X14, X8
	MULHU	X20, X15, X9
	MULHU	X20, X16, X29

	MUL	X20, X13, X5
	ADD	X24, X5, X24
	SLTU	X5, X24, X6

	MUL	X20, X14, X5
	ADD	X25, X5, X25
	SLTU	X5, X25, X12
	ADD	X25, X6, X25
	SLTU	X6, X25, X5
	ADD	X12, X5, X6

	MUL	X20, X15, X5
	ADD	X26, X5, X26
	SLTU	X5, X26, X12
	ADD	X26, X6, X26
	SLTU	X6, X26, X5
	ADD	X12, X5, X6

	MUL	X20, X16, X5
	ADD	X28, X5, X28
	SLTU	X5, X28, X12
	ADD	X28, X6, X28
	SLTU	X6, X28, X5
	ADD	X12, X5, X6

	MOV	X6, X12		// 暂存 carry_lo

	ADD	X25, X7, X25
	SLTU	X7, X25, X6

	ADD	X26, X8, X26
	SLTU	X8, X26, X5
	ADD	X26, X6, X26
	SLTU	X6, X26, X7
	ADD	X5, X7, X6

	ADD	X28, X9, X28
	SLTU	X9, X28, X5
	ADD	X28, X6, X28
	SLTU	X6, X28, X7
	ADD	X5, X7, X6

	ADD	X29, X12, X29
	ADD	X29, X6, X29

	// 保存 After Round 3 状态 (state[6], offset 384)
	MOV	X21, 384(X11)
	MOV	X22, 392(X11)
	MOV	X23, 400(X11)
	MOV	X24, 408(X11)
	MOV	X25, 416(X11)
	MOV	X26, 424(X11)
	MOV	X28, 432(X11)
	MOV	X29, 440(X11)

	// =====================================
	// Last reduction step
	// =====================================
	SLL	$32, X24, X5
	SRL	$32, X24, X7
	MUL	X24, X30, X8
	MULHU	X24, X30, X24

	ADD	X21, X5, X21
	SLTU	X5, X21, X6

	ADD	X22, X7, X22
	SLTU	X7, X22, X9
	ADD	X22, X6, X22
	SLTU	X6, X22, X5
	ADD	X9, X5, X6

	ADD	X23, X8, X23
	SLTU	X8, X23, X9
	ADD	X23, X6, X23
	SLTU	X6, X23, X5
	ADD	X9, X5, X6

	ADD	X24, X6, X24

	// 保存 After Reduce 4 状态 (state[7], offset 448)
	MOV	X21, 448(X11)
	MOV	X22, 456(X11)
	MOV	X23, 464(X11)
	MOV	X24, 472(X11)
	MOV	X25, 480(X11)
	MOV	X26, 488(X11)
	MOV	X28, 496(X11)
	MOV	X29, 504(X11)

	// =====================================
	// Add bits [511:256]
	// =====================================
	ADD	X21, X25, X21
	SLTU	X25, X21, X6

	ADD	X22, X26, X22
	SLTU	X26, X22, X9
	ADD	X22, X6, X22
	SLTU	X6, X22, X5
	ADD	X9, X5, X6

	ADD	X23, X28, X23
	SLTU	X28, X23, X9
	ADD	X23, X6, X23
	SLTU	X6, X23, X5
	ADD	X9, X5, X6

	ADD	X24, X29, X24
	SLTU	X29, X24, X9
	ADD	X24, X6, X24
	SLTU	X6, X24, X5
	ADD	X9, X5, X25

	// =====================================
	// Conditional subtraction
	// P256 素数 p = [p0, p1, p2, p3] = [0xFFFFFFFFFFFFFFFF, 0x00000000FFFFFFFF, 0, 0xFFFFFFFF00000001]
	// 计算 (t0, t1, t2, t3) = (acc0, acc1, acc2, acc3) - (p0, p1, 0, p3)
	// 如果发生借位（结果为负），使用原值；否则使用减法结果
	// =====================================
	// 当前寄存器状态：
	// X21=acc0, X22=acc1, X23=acc2, X24=acc3, X25=carry (0 或 1)
	// X30 = const1 = 0xFFFFFFFF00000001
	// X10 = out 指针
	
	// 加载常量
	MOV	$p256const0<>(SB), X5
	MOV	(X5), X5		// X5 = const0 = 0x00000000FFFFFFFF
	MOV	$-1, X6			// X6 = p0 = 0xFFFFFFFFFFFFFFFF
	
	// ========== Step 1: t0, b0 = Sub64(acc0, p0, 0) ==========
	// t0 = acc0 - p0 = acc0 - (-1) = acc0 + 1 (mod 2^64)
	// b0 = (acc0 < p0) ? 1 : 0
	// 只有当 acc0 == 0xFFFFFFFFFFFFFFFF 时，b0 = 0
	ADD	$1, X21, X13		// X13 = t0 = acc0 + 1
	SLTU	X6, X21, X7		// X7 = b0 = (acc0 < p0)
	
	// ========== Step 2: t1, b1 = Sub64(acc1, p1, b0) ==========
	// t1 = acc1 - p1 - b0
	// b1 = (acc1 < p1 + b0) 考虑溢出
	// 使用两步：tmp = acc1 - p1, t1 = tmp - b0
	// b1 = (acc1 < p1) || (tmp < b0)
	SUB	X5, X22, X8		// X8 = tmp = acc1 - const0
	SLTU	X5, X22, X9		// X9 = (acc1 < const0) ? 1 : 0，即第一次借位
	// 注意：SLTU a, b, c 的语义是 c = (b < a)
	// 所以 SLTU X5, X22, X9 = X9 = (X22 < X5) = (acc1 < const0)
	SUB	X7, X8, X14		// X14 = t1 = tmp - b0
	SLTU	X7, X8, X12		// X12 = (tmp < b0) ? 1 : 0，即第二次借位
	OR	X9, X12, X7		// X7 = b1 = 第一次借位 || 第二次借位
	
	// ========== Step 3: t2, b2 = Sub64(acc2, 0, b1) ==========
	// t2 = acc2 - 0 - b1 = acc2 - b1
	// b2 = (acc2 < b1)
	SUB	X7, X23, X15		// X15 = t2 = acc2 - b1
	SLTU	X7, X23, X7		// X7 = b2 = (acc2 < b1)
	
	// ========== Step 4: t3, b3 = Sub64(acc3, p3, b2) ==========
	// t3 = acc3 - p3 - b2
	// b3 = (acc3 < p3 + b2) 考虑溢出
	SUB	X30, X24, X8		// X8 = tmp = acc3 - const1
	SLTU	X30, X24, X9		// X9 = (acc3 < const1)
	SUB	X7, X8, X16		// X16 = t3 = tmp - b2
	SLTU	X7, X8, X12		// X12 = (tmp < b2)
	OR	X9, X12, X7		// X7 = b3
	
	// ========== Step 5: final_borrow = (carry < b3) ==========
	// 如果 final_borrow = 1，使用原值；否则使用减法结果
	SLTU	X7, X25, X8		// X8 = (carry < b3) = final_borrow
	
	// ========== 条件选择 ==========
	// mask = final_borrow ? 0xFFFFFFFFFFFFFFFF : 0
	// result = (orig & mask) | (sub_result & ~mask)
	// 等价于 result = sub_result ^ ((orig ^ sub_result) & mask)
	NEG	X8, X8			// X8 = mask = 0 - final_borrow = (final_borrow ? -1 : 0)
	
	// 选择 result0
	XOR	X21, X13, X9		// X9 = acc0 ^ t0
	AND	X8, X9, X9		// X9 = (acc0 ^ t0) & mask
	XOR	X9, X13, X13		// X13 = result0
	
	// 选择 result1
	XOR	X22, X14, X9
	AND	X8, X9, X9
	XOR	X9, X14, X14		// X14 = result1
	
	// 选择 result2
	XOR	X23, X15, X9
	AND	X8, X9, X9
	XOR	X9, X15, X15		// X15 = result2
	
	// 选择 result3
	XOR	X24, X16, X9
	AND	X8, X9, X9
	XOR	X9, X16, X16		// X16 = result3
	
	// 写入结果
	MOV	X13, 0(X10)
	MOV	X14, 8(X10)
	MOV	X15, 16(X10)
	MOV	X16, 24(X10)

	RET
