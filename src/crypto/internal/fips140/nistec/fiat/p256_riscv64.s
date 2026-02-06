// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !purego

#include "textflag.h"

// P256 constants
DATA p256const0<>+0x00(SB)/8, $0x00000000ffffffff
DATA p256const1<>+0x00(SB)/8, $0xffffffff00000001
GLOBL p256const0<>(SB), RODATA, $8
GLOBL p256const1<>(SB), RODATA, $8

// func p256Mul(out, a, b *p256MontgomeryDomainFieldElement)
// RISC-V assembly implementation of P256 Montgomery multiplication
// Register allocation:
//   X10 = out pointer
//   X11 = a pointer (temporary), later used as temp variable
//   X12 = b pointer (temporary)
//   X13-X16 = x0-x3 (input a)
//   X17-X20 = y0-y3 (input b)
//   X21-X24 = acc0-acc3 (accumulator low bits)
//   X25-X26, X28-X29 = acc4-acc7 (accumulator high bits)
//   X30 = const1 = 0xFFFFFFFF00000001
//   X5-X9, X12 = temporary variables
TEXT ·p256Mul(SB),NOSPLIT,$0-24
	MOV	out+0(FP), X10
	MOV	a+8(FP), X11
	MOV	b+16(FP), X12

	// Load const1
	MOV	$p256const1<>(SB), X30
	MOV	(X30), X30

	// Load input x (a): X13-X16 = x0-x3
	MOV	0(X11), X13
	MOV	8(X11), X14
	MOV	16(X11), X15
	MOV	24(X11), X16

	// Load input y (b): X17-X20 = y0-y3
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

	MOV	X6, X29			// Temporarily store carry_lo

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

	MOV	X6, X12			// Temporarily store carry_lo

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

TEXT ·p256Square(SB),NOSPLIT,$0-16
	MOV	out1+0(FP), X10
	MOV	arg1+8(FP), X11

	// Load input x: X5-X8 = x0-x3
	MOV	0(X11), X5
	MOV	8(X11), X6
	MOV	16(X11), X7
	MOV	24(X11), X8

	// =====================================
	// Compute cross products: x[1:] * x[0]
	// =====================================
	MULHU	X5, X6, X9
	MUL	X5, X6, X11
	MULHU	X5, X7, X12
	MUL	X5, X7, X13
	ADD	X13, X9, X9
	SLTU	X13, X9, X13
	MULHU	X5, X8, X14
	MUL	X5, X8, X15
	ADD	X15, X12, X12
	ADD	X12, X13, X13
	SLTU	X12, X13, X16
	SLTU	X15, X12, X12
	OR	X12, X16, X12
	ADD	X14, X12, X12
	MULHU	X6, X7, X14
	MUL	X6, X7, X15
	ADD	X15, X13, X13
	SLTU	X15, X13, X15
	ADD	X14, X12, X12
	SLTU	X14, X12, X14
	ADD	X12, X15, X15
	SLTU	X12, X15, X12
	OR	X14, X12, X12
	MULHU	X6, X8, X14
	MUL	X6, X8, X16
	ADD	X16, X15, X15
	SLTU	X16, X15, X16
	ADD	X14, X12, X12
	ADD	X16, X12, X12
	MULHU	X7, X8, X14
	MUL	X7, X8, X16
	ADD	X12, X16, X16
	SLTU	X12, X16, X12
	ADD	X14, X12, X12

	// =====================================
	// Multiply cross products by 2
	// =====================================
	ADD	X11, X11, X14
	SLTU	X11, X14, X11
	ADD	X9, X9, X17
	SLTU	X9, X17, X9
	ADD	X17, X11, X11
	SLTU	X17, X11, X17
	OR	X9, X17, X9
	ADD	X13, X13, X17
	ADD	X17, X9, X9
	SLTU	X13, X17, X13
	SLTU	X17, X9, X17
	OR	X17, X13, X13
	ADD	X15, X15, X17
	ADD	X17, X13, X13
	SLTU	X15, X17, X15
	SLTU	X17, X13, X17
	OR	X17, X15, X15
	ADD	X16, X16, X17
	ADD	X17, X15, X15
	SLTU	X16, X17, X16
	SLTU	X17, X15, X17
	OR	X17, X16, X16
	ADD	X12, X12, X17
	SLTU	X12, X17, X12
	ADD	X17, X16, X16
	SLTU	X17, X16, X17
	OR	X12, X17, X12

	// =====================================
	// Add missing products (squares)
	// =====================================
	MULHU	X5, X5, X17
	MUL	X5, X5, X18
	ADD	X14, X17, X5
	SLTU	X14, X5, X14
	MULHU	X6, X6, X17
	MUL	X6, X6, X19
	ADD	X11, X19, X6
	SLTU	X19, X6, X11
	ADD	X6, X14, X14
	SLTU	X6, X14, X6
	OR	X11, X6, X6
	ADD	X17, X9, X9
	ADD	X9, X6, X6
	SLTU	X9, X6, X11
	SLTU	X17, X9, X9
	OR	X11, X9, X9
	MULHU	X7, X7, X11
	MUL	X7, X7, X17
	ADD	X13, X17, X7
	SLTU	X17, X7, X13
	ADD	X7, X9, X9
	SLTU	X7, X9, X7
	OR	X13, X7, X7
	ADD	X15, X11, X13
	ADD	X13, X7, X7
	SLTU	X13, X7, X15
	SLTU	X11, X13, X11
	OR	X15, X11, X11
	MULHU	X8, X8, X13
	MUL	X8, X8, X15
	ADD	X16, X15, X8
	ADD	X8, X11, X11
	SLTU	X8, X11, X16
	SLTU	X15, X8, X8
	OR	X8, X16, X8
	ADD	X13, X12, X12
	ADD	X12, X8, X8

	// =====================================
	// First reduction step
	// =====================================
	// Load const1
	MOV	$p256const1<>(SB), X15
	MOV	(X15), X15

	SLLI	$32, X18, X12
	ADD	X12, X5, X5
	SLTU	X12, X5, X12
	SRLI	$32, X18, X13
	MULHU	X18, X15, X16
	MUL	X18, X15, X17
	ADD	X13, X14, X14
	SLTU	X13, X14, X13
	ADD	X14, X12, X12
	SLTU	X14, X12, X14
	OR	X14, X13, X13
	ADD	X17, X6, X6
	SLTU	X17, X6, X14
	ADD	X6, X13, X13
	SLTU	X6, X13, X6
	OR	X14, X6, X6
	ADD	X16, X6, X6

	// =====================================
	// Second reduction step
	// =====================================
	SLLI	$32, X5, X14
	ADD	X12, X14, X14
	SLTU	X12, X14, X12
	SRLI	$32, X5, X16
	MULHU	X5, X15, X17
	MUL	X5, X15, X18
	ADD	X13, X16, X5
	SLTU	X16, X5, X13
	ADD	X5, X12, X12
	SLTU	X5, X12, X5
	OR	X13, X5, X5
	ADD	X18, X6, X6
	SLTU	X18, X6, X13
	ADD	X6, X5, X5
	SLTU	X6, X5, X6
	OR	X13, X6, X6
	ADD	X17, X6, X6

	// =====================================
	// Third reduction step
	// =====================================
	SLLI	$32, X14, X13
	ADD	X12, X13, X13
	SLTU	X12, X13, X12
	SRLI	$32, X14, X16
	MULHU	X15, X14, X17
	MUL	X15, X14, X18
	ADD	X16, X5, X5
	SLTU	X16, X5, X14
	ADD	X5, X12, X12
	SLTU	X5, X12, X5
	OR	X14, X5, X5
	ADD	X6, X18, X14
	SLTU	X6, X14, X6
	ADD	X14, X5, X5
	SLTU	X14, X5, X14
	OR	X14, X6, X6
	ADD	X17, X6, X6

	// =====================================
	// Last reduction step
	// =====================================
	SLLI	$32, X13, X14
	ADD	X14, X12, X12
	SLTU	X14, X12, X14
	SRLI	$32, X13, X16
	MULHU	X15, X13, X17
	MUL	X15, X13, X18
	ADD	X16, X5, X5
	ADD	X5, X14, X13
	SLTU	X5, X13, X14
	SLTU	X16, X5, X5
	OR	X5, X14, X5
	ADD	X18, X6, X6
	ADD	X6, X5, X5
	SLTU	X6, X5, X14
	SLTU	X18, X6, X6
	OR	X6, X14, X6
	ADD	X17, X6, X6

	// =====================================
	// Add bits [511:256] of the sqr result
	// =====================================
	ADD	X9, X12, X12
	SLTU	X9, X12, X9
	ADD	X7, X13, X13
	SLTU	X7, X13, X7
	ADD	X13, X9, X9
	SLTU	X13, X9, X13
	OR	X7, X13, X7
	ADD	X5, X11, X11
	SLTU	X5, X11, X5
	ADD	X11, X7, X7
	SLTU	X11, X7, X11
	OR	X5, X11, X5
	ADD	X6, X8, X8
	SLTU	X6, X8, X6
	ADD	X8, X5, X5
	SLTU	X8, X5, X8

	// =====================================
	// Conditional subtraction
	// =====================================
	ADDI	$1, X12, X11
	SLTU	X11, X12, X13
	ADDI	$-1, X0, X14
	SRLI	$32, X14, X14
	SUB	X14, X9, X14
	SLTU	X14, X9, X16
	SUB	X13, X14, X13
	SLTU	X13, X14, X14
	OR	X16, X14, X14
	SUB	X14, X7, X14
	SLTU	X14, X7, X16
	SUB	X15, X5, X15
	SLTU	X15, X5, X17
	SUB	X16, X15, X16
	SLTU	X16, X15, X15
	OR	X17, X15, X15
	OR	X6, X8, X6
	SUB	X15, X6, X8
	BLTU	X6, X8, skip_sub
	MOV	X11, 0(X10)
	MOV	X13, 8(X10)
	MOV	X14, 16(X10)
	MOV	X16, 24(X10)
	JMP	done
skip_sub:
	MOV	X12, 0(X10)
	MOV	X9, 8(X10)
	MOV	X7, 16(X10)
	MOV	X5, 24(X10)
done:
	RET

TEXT ·p256SquareCDisassemble(SB),NOSPLIT,$0-16
	MOV	out1+0(FP), X10
	MOV	arg1+8(FP), X11

	// Load input parameters
	MOV	(X11), X13
	MOV	8(X11), X5
	MOV	16(X11), X6

	ADDI	$-1, X0, X14
	BSETI	$32, X0, X30
	MOV	24(X11), X29

	SLLI	$32, X14, X14
	ADDI	$-1, X30, X30

	MOV	X14, X31
	ADDI	$1, X14, X14

	
	MUL	X5, X13, X17
	MULHU	X13, X13, X12

	MULHU	X5, X13, X8
	MUL	X13, X6, X28

	MUL	X5, X5, X11
	MUL	X13, X13, X18
	SH1ADD	X12, X17, X15
	MULHU	X13, X6, X19
	MUL	X13, X29, X16
	SRLI	$63, X17, X17
	ADD	X8, X28, X28
	SLTU	X12, X15, X12
	MUL	X5, X6, X21
	MULHU	X13, X29, X13
	ADD	X12, X11, X11
	SH1ADD	X17, X28, X17
	SLLI	$32, X18, X7
	SLTU	X8, X28, X8
	ADD	X17, X11, X9
	ADD	X15, X7, X7
	ADD	X19, X16, X16
	SLTU	X12, X11, X12
	SLTU	X11, X9, X11
	SRLI	$32, X18, X20
	SLLI	$1, X28, X22
	SLTU	X15, X7, X15
	ADD	X12, X11, X11
	ADD	X16, X8, X8
	SRLI	$63, X28, X12
	MULHU	X5, X5, X28
	ADD	X20, X15, X15
	SLTU	X19, X16, X19
	SLTU	X22, X17, X17
	SLTU	X16, X8, X16
	SLTU	X20, X15, X22
	ADD	X21, X8, X8
	ADD	X19, X16, X20
	ADD	X15, X9, X9
	MUL	X14, X18, X19
	ADD	X12, X17, X17
	ADD	X11, X28, X28
	SLTU	X15, X9, X15
	SH1ADD	X17, X8, X17
	ADD	X22, X15, X15
	MULHU	X5, X6, X22
	SLTU	X11, X28, X11
	ADD	X17, X28, X12
	SLLI	$32, X7, X16
	SLTU	X21, X8, X21
	MULHU	X14, X18, X18
	ADD	X15, X19, X19
	SLTU	X28, X12, X28
	ADD	X16, X9, X9
	MUL	X6, X6, X23
	ADD	X11, X28, X28
	ADD	X19, X12, X12
	SLLI	$1, X8, X11
	SLTU	X15, X19, X15
	SLTU	X11, X17, X17
	ADD	X22, X13, X13
	SLTU	X19, X12, X19
	SRLI	$63, X8, X8
	SLTU	X16, X9, X16
	ADD	X15, X19, X19
	ADD	X17, X8, X8
	ADD	X20, X13, X15
	MUL	X14, X7, X17
	SRLI	$32, X7, X20
	SLTU	X22, X13, X22
	ADD	X15, X21, X21
	ADD	X16, X20, X11
	SLTU	X13, X15, X13
	SLTU	X15, X21, X15
	MULHU	X14, X7, X7
	ADD	X11, X12, X12
	SLTU	X20, X11, X20
	ADD	X22, X13, X13
	ADD	X18, X17, X17
	SLTU	X11, X12, X11
	ADD	X13, X15, X15
	ADD	X19, X17, X16
	ADD	X20, X11, X11
	SLLI	$32, X9, X19
	MUL	X5, X29, X20
	ADD	X19, X12, X12
	MULHU	X5, X29, X5
	SLTU	X18, X17, X18
	ADD	X11, X16, X13
	SLTU	X17, X16, X11
	SLTU	X19, X12, X19
	SLTU	X16, X13, X16
	MUL	X14, X9, X17
	ADD	X18, X11, X11
	ADD	X20, X21, X21
	ADD	X19, X13, X13
	ADD	X11, X16, X22
	SRLI	$32, X9, X16
	SH1ADD	X8, X21, X11
	ADD	X15, X5, X18
	SLTU	X20, X21, X20
	ADD	X23, X28, X5
	ADD	X13, X16, X16
	SLTU	X19, X13, X19
	ADD	X20, X18, X18
	SLTU	X13, X16, X15
	ADD	X11, X5, X20
	MUL	X6, X29, X8
	SLTU	X28, X5, X28
	ADD	X19, X15, X15
	SLTU	X5, X20, X5
	ADD	X17, X7, X19
	ADD	X28, X5, X5
	SLLI	$1, X21, X17
	ADD	X22, X19, X28
	MULHU	X6, X6, X22
	SLTU	X17, X11, X11
	SRLI	$63, X21, X17
	ADD	X8, X18, X13
	MULHU	X6, X29, X6
	ADD	X17, X11, X11
	SLLI	$32, X12, X21
	SLTU	X19, X28, X17
	MULHU	X14, X9, X9
	SH1ADD	X11, X13, X8
	SLTU	X7, X19, X11
	ADD	X22, X5, X7
	ADD	X21, X16, X16
	ADD	X11, X17, X19
	SLTU	X5, X7, X11
	MUL	X14, X12, X5
	SLTU	X21, X16, X21
	SLTU	X18, X13, X18
	ADD	X28, X15, X15
	SLLI	$1, X13, X17
	ADD	X20, X16, X16
	ADD	X6, X18, X18
	SLTU	X17, X8, X17
	SLTU	X28, X15, X28
	ADD	X7, X8, X8
	ADD	X21, X15, X15
	SRLI	$32, X12, X6
	SRLI	$63, X13, X13
	ADD	X19, X28, X28
	SLTU	X7, X8, X7
	ADD	X17, X13, X13
	ADD	X15, X6, X6
	ADD	X9, X5, X5
	MUL	X29, X29, X17
	SLTU	X20, X16, X20
	SH1ADD	X13, X18, X22
	ADD	X11, X7, X19
	ADD	X20, X8, X8
	SLTU	X15, X6, X13
	MULHU	X29, X29, X11
	SLTU	X21, X15, X21
	ADD	X28, X5, X15
	SLLI	$1, X18, X28
	ADD	X8, X6, X6
	ADD	X22, X19, X7
	SRLI	$63, X18, X18
	SLTU	X28, X22, X28
	ADDI	$1, X16, X29
	SLTU	X20, X8, X20
	ADD	X18, X28, X28
	ADD	X7, X17, X17
	ADD	X21, X13, X13
	SLTU	X8, X6, X8
	ORN	X16, X29, X18
	MULHU	X14, X12, X12
	ADD	X15, X13, X13
	ADD	X20, X8, X8
	SLTU	X19, X7, X19
	ADD	X28, X11, X11
	SLTU	X7, X17, X7
	ADD	X14, X6, X28
	SRLI	$63, X18, X18
	SLTU	X9, X5, X9
	ADD	X19, X7, X7
	SLTU	X5, X15, X5
	XOR	X31, X6, X31
	SLTU	X15, X13, X15
	SUB	X18, X28, X28
	ADD	X8, X13, X13
	ADD	X13, X17, X17
	ADD	X7, X11, X11
	ADD	X9, X5, X5
	AND	X28, X31, X31
	ADD	X11, X12, X12
	ADD	X5, X15, X15
	SLTU	X8, X13, X8
	SRLI	$63, X31, X31
	SLTU	X13, X17, X13
	ADD	X12, X15, X15
	SUB	X31, X17, X31
	BSETI	$32, X0, X5
	ADD	X8, X13, X13
	ANDN	X17, X31, X7
	ADDI	$-2, X5, X5
	SLTU	X11, X12, X11
	ADD	X15, X13, X13
	SRLI	$63, X7, X7
	SLTU	X12, X15, X12
	ADD	X13, X30, X30
	XOR	X5, X13, X5
	ADD	X11, X12, X12
	ANDN	X13, X14, X14
	SUB	X7, X30, X30
	SLTU	X15, X13, X15
	AND	X30, X5, X11
	ADD	X12, X15, X15
	OR	X11, X14, X14
	SRLI	$63, X14, X14
	SUB	X14, X15, X14
	ANDN	X15, X14, X15
	BGEZ	X15, skip_sel
	MOV	X16, X29
	MOV	X6, X28
	MOV	X17, X31
	MOV	X13, X30
skip_sel:
	MOV	X29, (X10)
	MOV	X28, 8(X10)
	MOV	X31, 16(X10)
	MOV	X30, 24(X10)
	RET

