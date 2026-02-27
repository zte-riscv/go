// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build riscv64 && !purego

package fiat

import "math/bits"

//go:noescape
func p256Mul(out, a, b *p256MontgomeryDomainFieldElement)

// p256Square squares a field element following ARM64 assembly algorithm
// This is a reference implementation for debugging purposes
func p256Square(out1 *p256MontgomeryDomainFieldElement, arg1 *p256MontgomeryDomainFieldElement) {
	x0 := arg1[0]
	x1 := arg1[1]
	x2 := arg1[2]
	x3 := arg1[3]

	// =====================================
	// Compute cross products: x[1:] * x[0]
	// =====================================
	var acc1, acc2, acc3, acc4, acc5, acc6, acc7 uint64
	var carry uint64

	// x0 * x1
	// MUL x0, x1, acc1
	// UMULH x0, x1, acc2
	hi, lo := bits.Mul64(x0, x1) // bits.Mul64 returns (hi, lo)
	acc1 = lo
	acc2 = hi

	// x0 * x2
	// MUL x0, x2, t0
	// ADDS t0, acc2, acc2  (acc2 = acc2 + lo, sets C flag)
	// UMULH x0, x2, acc3   (acc3 = hi, overwrites previous value)
	hi, lo = bits.Mul64(x0, x2)
	acc2, carry = bits.Add64(acc2, lo, 0) // carry1 from ADDS
	acc3 = hi                             // UMULH overwrites acc3

	// x0 * x3
	// MUL x0, x3, t0
	// ADCS t0, acc3, acc3  (acc3 = acc3 + lo + C from ADDS above)
	// UMULH x0, x3, acc4   (acc4 = hi)
	// ADC $0, acc4, acc4   (acc4 = acc4 + C from ADCS above)
	hi, lo = bits.Mul64(x0, x3)
	acc3, carry = bits.Add64(acc3, lo, carry) // Use carry1 from ADDS
	acc4 = hi
	acc4, carry = bits.Add64(acc4, 0, carry) // Use carry from ADCS

	// =====================================
	// x[2:] * x[1]
	// =====================================
	// x1 * x2
	// MUL x1, x2, t0
	// ADDS t0, acc3  (acc3 = acc3 + lo, sets C flag)
	// UMULH x1, x2, t1  (t1 = hi)
	// ADCS t1, acc4  (acc4 = acc4 + hi + C from ADDS)
	// ADC $0, ZR, acc5  (acc5 = C from ADCS)
	hi, lo = bits.Mul64(x1, x2)
	acc3, carry = bits.Add64(acc3, lo, 0)     // carry2 from ADDS
	acc4, carry = bits.Add64(acc4, hi, carry) // Use carry2
	acc5 = carry                              // ADC $0, ZR, acc5

	// x1 * x3
	// MUL x1, x3, t0
	// ADDS t0, acc4  (acc4 = acc4 + lo, sets C flag)
	// UMULH x1, x3, t1  (t1 = hi)
	// ADC t1, acc5  (acc5 = acc5 + hi + C from ADDS)
	hi, lo = bits.Mul64(x1, x3)
	acc4, carry = bits.Add64(acc4, lo, 0) // carry3 from ADDS
	acc5, _ = bits.Add64(acc5, hi, carry) // Use carry3, no further carry needed

	// =====================================
	// x[3] * x[2]
	// =====================================
	// x2 * x3
	// MUL x2, x3, t0
	// ADDS t0, acc5  (acc5 = acc5 + lo, sets C flag)
	// UMULH x2, x3, acc6  (acc6 = hi)
	// ADC $0, acc6  (acc6 = acc6 + C from ADDS)
	hi, lo = bits.Mul64(x2, x3)
	acc5, carry = bits.Add64(acc5, lo, 0) // carry4 from ADDS
	acc6 = hi
	acc6, _ = bits.Add64(acc6, 0, carry) // Use carry4

	acc7 = 0

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

	// =====================================
	// Add missing products (squares)
	// =====================================
	var acc0 uint64
	// x0 * x0
	// MUL x0, x0, acc0
	// UMULH x0, x0, t0
	// ADDS t0, acc1, acc1
	hi, lo = bits.Mul64(x0, x0) // bits.Mul64 returns (hi, lo)
	acc0 = lo
	acc1, carry = bits.Add64(acc1, hi, 0)

	// x1 * x1
	// MUL x1, x1, t0
	// ADCS t0, acc2, acc2
	// UMULH x1, x1, t1
	// ADCS t1, acc3, acc3
	hi, lo = bits.Mul64(x1, x1)
	acc2, carry = bits.Add64(acc2, lo, carry)
	acc3, carry = bits.Add64(acc3, hi, carry)

	// x2 * x2
	// MUL x2, x2, t0
	// ADCS t0, acc4, acc4
	// UMULH x2, x2, t1
	// ADCS t1, acc5, acc5
	hi, lo = bits.Mul64(x2, x2)
	acc4, carry = bits.Add64(acc4, lo, carry)
	acc5, carry = bits.Add64(acc5, hi, carry)

	// x3 * x3
	// MUL x3, x3, t0
	// ADCS t0, acc6, acc6
	// UMULH x3, x3, t1
	// ADCS t1, acc7, acc7
	hi, lo = bits.Mul64(x3, x3)
	acc6, carry = bits.Add64(acc6, lo, carry)
	acc7, _ = bits.Add64(acc7, hi, carry)

	// =====================================
	// First reduction step
	// =====================================
	// Following ARM64: ADDS acc0<<32, acc1, acc1 (sets C)
	//                  LSR $32, acc0, t0
	//                  MUL acc0, const1, t1
	//                  UMULH acc0, const1, acc0 (overwrites acc0)
	//                  ADCS t0, acc2, acc2 (uses C from ADDS)
	//                  ADCS t1, acc3, acc3 (uses C from previous ADCS)
	//                  ADC $0, acc0, acc0 (uses C from previous ADCS)
	const const1 = 0xffffffff00000001
	t0 := acc0 << 32
	acc1, carry = bits.Add64(acc1, t0, 0) // ADDS, sets carry1
	t1 := acc0 >> 32
	hi, lo = bits.Mul64(acc0, const1) // bits.Mul64 returns (hi, lo)
	t2 := lo
	acc0 = hi                                 // UMULH overwrites acc0
	acc2, carry = bits.Add64(acc2, t1, carry) // ADCS, uses carry1
	acc3, carry = bits.Add64(acc3, t2, carry) // ADCS, uses carry from previous
	acc0, _ = bits.Add64(acc0, 0, carry)      // ADC, uses carry from previous

	// =====================================
	// Second reduction step
	// =====================================
	// Following ARM64: ADDS acc1<<32, acc2, acc2 (sets C)
	//                  LSR $32, acc1, t0
	//                  MUL acc1, const1, t1
	//                  UMULH acc1, const1, acc1 (overwrites acc1)
	//                  ADCS t0, acc3, acc3 (uses C from ADDS)
	//                  ADCS t1, acc0, acc0 (uses C from previous ADCS)
	//                  ADC $0, acc1, acc1 (uses C from previous ADCS)
	t0 = acc1 << 32
	acc2, carry = bits.Add64(acc2, t0, 0) // ADDS, sets carry1
	t1 = acc1 >> 32
	hi, lo = bits.Mul64(acc1, const1) // bits.Mul64 returns (hi, lo)
	t2 = lo
	acc1 = hi                                 // UMULH overwrites acc1
	acc3, carry = bits.Add64(acc3, t1, carry) // ADCS, uses carry1
	acc0, carry = bits.Add64(acc0, t2, carry) // ADCS, uses carry from previous
	acc1, _ = bits.Add64(acc1, 0, carry)      // ADC, uses carry from previous

	// =====================================
	// Third reduction step
	// =====================================
	// Following ARM64: ADDS acc2<<32, acc3, acc3 (sets C)
	//                  LSR $32, acc2, t0
	//                  MUL acc2, const1, t1
	//                  UMULH acc2, const1, acc2 (overwrites acc2)
	//                  ADCS t0, acc0, acc0 (uses C from ADDS)
	//                  ADCS t1, acc1, acc1 (uses C from previous ADCS)
	//                  ADC $0, acc2, acc2 (uses C from previous ADCS)
	t0 = acc2 << 32
	acc3, carry = bits.Add64(acc3, t0, 0) // ADDS, sets carry1
	t1 = acc2 >> 32
	hi, lo = bits.Mul64(acc2, const1) // bits.Mul64 returns (hi, lo)
	t2 = lo
	acc2 = hi                                 // UMULH overwrites acc2
	acc0, carry = bits.Add64(acc0, t1, carry) // ADCS, uses carry1
	acc1, carry = bits.Add64(acc1, t2, carry) // ADCS, uses carry from previous
	acc2, _ = bits.Add64(acc2, 0, carry)      // ADC, uses carry from previous

	// =====================================
	// Last reduction step
	// =====================================
	// Following ARM64: ADDS acc3<<32, acc0, acc0 (sets C)
	//                  LSR $32, acc3, t0
	//                  MUL acc3, const1, t1
	//                  UMULH acc3, const1, acc3 (overwrites acc3)
	//                  ADCS t0, acc1, acc1 (uses C from ADDS)
	//                  ADCS t1, acc2, acc2 (uses C from previous ADCS)
	//                  ADC $0, acc3, acc3 (uses C from previous ADCS)
	t0 = acc3 << 32
	acc0, carry = bits.Add64(acc0, t0, 0) // ADDS, sets carry1
	t1 = acc3 >> 32
	hi, lo = bits.Mul64(acc3, const1) // bits.Mul64 returns (hi, lo)
	t2 = lo
	acc3 = hi                                 // UMULH overwrites acc3
	acc1, carry = bits.Add64(acc1, t1, carry) // ADCS, uses carry1
	acc2, carry = bits.Add64(acc2, t2, carry) // ADCS, uses carry from previous
	acc3, _ = bits.Add64(acc3, 0, carry)      // ADC, uses carry from previous

	// =====================================
	// Add bits [511:256] of the sqr result
	// =====================================
	acc0, carry = bits.Add64(acc0, acc4, 0)
	acc1, carry = bits.Add64(acc1, acc5, carry)
	acc2, carry = bits.Add64(acc2, acc6, carry)
	acc3, carry = bits.Add64(acc3, acc7, carry)
	finalCarry := carry

	// =====================================
	// Conditional subtraction
	// =====================================
	// Following ARM64 algorithm and p256SquareGeneric logic
	// SUBS $-1, acc0, t0 computes t0 = acc0 - (-1) = acc0 + 1
	// This checks if acc0 >= 0xffffffffffffffff (p0)
	const p0 = 0xffffffffffffffff
	const p1 = 0x00000000ffffffff
	const p3 = 0xffffffff00000001

	// Compute (acc0, acc1, acc2, acc3) - (p0, p1, 0, p3)
	// Following p256SquareGeneric pattern
	var t0_sub, t1_sub, t2_sub, t3_sub uint64
	var borrow uint64

	// Step 1: t0 = acc0 - p0 (SUBS $-1, acc0, t0 is acc0 + 1, checking acc0 >= p0)
	t0_sub, borrow = bits.Sub64(acc0, p0, 0)

	// Step 2: t1 = acc1 - p1 - borrow (SBCS const0, acc1, t1)
	t1_sub, borrow = bits.Sub64(acc1, p1, borrow)

	// Step 3: t2 = acc2 - 0 - borrow (SBCS $0, acc2, t2)
	t2_sub, borrow = bits.Sub64(acc2, 0, borrow)

	// Step 4: t3 = acc3 - p3 - borrow (SBCS const1, acc3, t3)
	t3_sub, borrow = bits.Sub64(acc3, p3, borrow)

	// Step 5: final borrow check (SBCS $0, acc4, acc4)
	// Check if finalCarry - 0 - borrow produces borrow
	_, finalBorrow := bits.Sub64(finalCarry, 0, borrow)

	// Conditional select: if finalBorrow == 0 (acc >= p), use subtracted; else use original
	// This matches p256CmovznzU64 logic: if arg1 == 0, select arg2 (subtracted), else arg3 (original)
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
}
