// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build riscv64 && debug

package fiat

import (
	"fmt"
	"math/bits"
)

// p256SquareArm64StyleDebug is a debug version that prints intermediate results
func p256SquareArm64StyleDebug(out1 *p256MontgomeryDomainFieldElement, arg1 *p256MontgomeryDomainFieldElement) {
	x0 := arg1[0]
	x1 := arg1[1]
	x2 := arg1[2]
	x3 := arg1[3]

	fmt.Printf("Input: x0=%016x x1=%016x x2=%016x x3=%016x\n", x0, x1, x2, x3)

	// =====================================
	// Compute cross products: x[1:] * x[0]
	// =====================================
	var acc1, acc2, acc3, acc4, acc5, acc6, acc7 uint64
	var carry uint64

	// x0 * x1
	lo, hi := bits.Mul64(x0, x1)
	acc1 = lo
	acc2 = hi
	fmt.Printf("After x0*x1: acc1=%016x acc2=%016x\n", acc1, acc2)

	// x0 * x2
	lo, hi = bits.Mul64(x0, x2)
	acc2, carry = bits.Add64(acc2, lo, 0)
	acc3 = hi
	acc3, carry = bits.Add64(acc3, carry, 0)
	fmt.Printf("After x0*x2: acc2=%016x acc3=%016x carry=%016x\n", acc2, acc3, carry)

	// x0 * x3
	lo, hi = bits.Mul64(x0, x3)
	acc3, carry = bits.Add64(acc3, lo, carry)
	acc4 = hi
	acc4, carry = bits.Add64(acc4, carry, 0)
	fmt.Printf("After x0*x3: acc3=%016x acc4=%016x carry=%016x\n", acc3, acc4, carry)

	// =====================================
	// x[2:] * x[1]
	// =====================================
	// x1 * x2
	lo, hi = bits.Mul64(x1, x2)
	acc3, carry = bits.Add64(acc3, lo, 0)
	acc4, carry = bits.Add64(acc4, hi, carry)
	acc5 = carry
	fmt.Printf("After x1*x2: acc3=%016x acc4=%016x acc5=%016x\n", acc3, acc4, acc5)

	// x1 * x3
	lo, hi = bits.Mul64(x1, x3)
	acc4, carry = bits.Add64(acc4, lo, 0)
	acc5, carry = bits.Add64(acc5, hi, carry)
	fmt.Printf("After x1*x3: acc4=%016x acc5=%016x carry=%016x\n", acc4, acc5, carry)

	// =====================================
	// x[3] * x[2]
	// =====================================
	// x2 * x3
	lo, hi = bits.Mul64(x2, x3)
	acc5, carry = bits.Add64(acc5, lo, 0)
	acc6 = hi
	acc6, _ = bits.Add64(acc6, carry, 0)
	fmt.Printf("After x2*x3: acc5=%016x acc6=%016x\n", acc5, acc6)

	acc7 = 0
	fmt.Printf("Before *2: acc1=%016x acc2=%016x acc3=%016x acc4=%016x acc5=%016x acc6=%016x acc7=%016x\n",
		acc1, acc2, acc3, acc4, acc5, acc6, acc7)

	// =====================================
	// Multiply cross products by 2
	// =====================================
	acc1, carry = bits.Add64(acc1, acc1, 0)
	acc2, carry = bits.Add64(acc2, acc2, carry)
	acc3, carry = bits.Add64(acc3, acc3, carry)
	acc4, carry = bits.Add64(acc4, acc4, carry)
	acc5, carry = bits.Add64(acc5, acc5, carry)
	acc6, carry = bits.Add64(acc6, acc6, carry)
	acc7, _ = bits.Add64(acc7, acc7, carry)
	fmt.Printf("After *2: acc1=%016x acc2=%016x acc3=%016x acc4=%016x acc5=%016x acc6=%016x acc7=%016x\n",
		acc1, acc2, acc3, acc4, acc5, acc6, acc7)

	// =====================================
	// Add missing products (squares)
	// =====================================
	var acc0 uint64
	// x0 * x0
	lo, hi = bits.Mul64(x0, x0)
	acc0 = lo
	acc1, carry = bits.Add64(acc1, hi, 0)
	fmt.Printf("After x0*x0: acc0=%016x acc1=%016x carry=%016x\n", acc0, acc1, carry)

	// x1 * x1
	lo, hi = bits.Mul64(x1, x1)
	acc2, carry = bits.Add64(acc2, lo, carry)
	acc3, carry = bits.Add64(acc3, hi, carry)
	fmt.Printf("After x1*x1: acc2=%016x acc3=%016x carry=%016x\n", acc2, acc3, carry)

	// x2 * x2
	lo, hi = bits.Mul64(x2, x2)
	acc4, carry = bits.Add64(acc4, lo, carry)
	acc5, carry = bits.Add64(acc5, hi, carry)
	fmt.Printf("After x2*x2: acc4=%016x acc5=%016x carry=%016x\n", acc4, acc5, carry)

	// x3 * x3
	lo, hi = bits.Mul64(x3, x3)
	acc6, carry = bits.Add64(acc6, lo, carry)
	acc7, _ = bits.Add64(acc7, hi, carry)
	fmt.Printf("After x3*x3: acc6=%016x acc7=%016x\n", acc6, acc7)
	fmt.Printf("After adding squares: acc0=%016x acc1=%016x acc2=%016x acc3=%016x acc4=%016x acc5=%016x acc6=%016x acc7=%016x\n",
		acc0, acc1, acc2, acc3, acc4, acc5, acc6, acc7)

	// =====================================
	// First reduction step
	// =====================================
	const const1 = 0xffffffff00000001
	t0 := acc0 << 32
	t1 := acc0 >> 32
	lo, hi = bits.Mul64(acc0, const1)
	t2 := lo
	acc0 = hi

	acc1, carry = bits.Add64(acc1, t0, 0)
	acc2, carry = bits.Add64(acc2, t1, carry)
	acc3, carry = bits.Add64(acc3, t2, carry)
	acc0, _ = bits.Add64(acc0, carry, 0)
	fmt.Printf("After first reduction: acc0=%016x acc1=%016x acc2=%016x acc3=%016x\n", acc0, acc1, acc2, acc3)

	// =====================================
	// Second reduction step
	// =====================================
	t0 = acc1 << 32
	t1 = acc1 >> 32
	lo, hi = bits.Mul64(acc1, const1)
	t2 = lo
	acc1 = hi

	acc2, carry = bits.Add64(acc2, t0, 0)
	acc3, carry = bits.Add64(acc3, t1, carry)
	acc0, carry = bits.Add64(acc0, t2, carry)
	acc1, _ = bits.Add64(acc1, carry, 0)
	fmt.Printf("After second reduction: acc0=%016x acc1=%016x acc2=%016x acc3=%016x\n", acc0, acc1, acc2, acc3)

	// =====================================
	// Third reduction step
	// =====================================
	t0 = acc2 << 32
	t1 = acc2 >> 32
	lo, hi = bits.Mul64(acc2, const1)
	t2 = lo
	acc2 = hi

	acc3, carry = bits.Add64(acc3, t0, 0)
	acc0, carry = bits.Add64(acc0, t1, carry)
	acc1, carry = bits.Add64(acc1, t2, carry)
	acc2, _ = bits.Add64(acc2, carry, 0)
	fmt.Printf("After third reduction: acc0=%016x acc1=%016x acc2=%016x acc3=%016x\n", acc0, acc1, acc2, acc3)

	// =====================================
	// Last reduction step
	// =====================================
	t0 = acc3 << 32
	t1 = acc3 >> 32
	lo, hi = bits.Mul64(acc3, const1)
	t2 = lo
	acc3 = hi

	acc0, carry = bits.Add64(acc0, t0, 0)
	acc1, carry = bits.Add64(acc1, t1, carry)
	acc2, carry = bits.Add64(acc2, t2, carry)
	acc3, _ = bits.Add64(acc3, carry, 0)
	fmt.Printf("After last reduction: acc0=%016x acc1=%016x acc2=%016x acc3=%016x\n", acc0, acc1, acc2, acc3)

	// =====================================
	// Add bits [511:256] of the sqr result
	// =====================================
	acc0, carry = bits.Add64(acc0, acc4, 0)
	acc1, carry = bits.Add64(acc1, acc5, carry)
	acc2, carry = bits.Add64(acc2, acc6, carry)
	acc3, carry = bits.Add64(acc3, acc7, carry)
	finalCarry := carry
	fmt.Printf("After adding high bits: acc0=%016x acc1=%016x acc2=%016x acc3=%016x finalCarry=%016x\n",
		acc0, acc1, acc2, acc3, finalCarry)

	// =====================================
	// Conditional subtraction
	// =====================================
	const p0 = 0xffffffffffffffff
	const p1 = 0x00000000ffffffff
	const p3 = 0xffffffff00000001

	var t0_sub, t1_sub, t2_sub, t3_sub uint64
	var borrow uint64

	t0_sub, borrow = bits.Sub64(acc0, p0, 0)
	t1_sub, borrow = bits.Sub64(acc1, p1, borrow)
	t2_sub, borrow = bits.Sub64(acc2, 0, borrow)
	t3_sub, borrow = bits.Sub64(acc3, p3, borrow)
	_, finalBorrow := bits.Sub64(finalCarry, 0, borrow)

	fmt.Printf("Conditional subtraction: t0_sub=%016x t1_sub=%016x t2_sub=%016x t3_sub=%016x finalBorrow=%016x\n",
		t0_sub, t1_sub, t2_sub, t3_sub, finalBorrow)

	if finalBorrow == 0 {
		out1[0] = t0_sub
		out1[1] = t1_sub
		out1[2] = t2_sub
		out1[3] = t3_sub
	} else {
		out1[0] = acc0
		out1[1] = acc1
		out1[2] = acc2
		out1[3] = acc3
	}
	fmt.Printf("Final output: out1[0]=%016x out1[1]=%016x out1[2]=%016x out1[3]=%016x\n",
		out1[0], out1[1], out1[2], out1[3])
}
