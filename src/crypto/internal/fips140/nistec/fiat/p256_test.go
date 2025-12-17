// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build riscv64

package fiat

import (
	"fmt"
	"math/bits"
	"testing"
)

func TestP256Mul(t *testing.T) {
	t.Run("CompareAssemblyWithGo", func(t *testing.T) {
		testCases := []struct {
			a, b p256MontgomeryDomainFieldElement
		}{
			{
				a: p256MontgomeryDomainFieldElement{1, 0, 0, 0},
				b: p256MontgomeryDomainFieldElement{1, 0, 0, 0},
			},
			{
				a: p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0, 0, 0},
				b: p256MontgomeryDomainFieldElement{2, 0, 0, 0},
			},
			{
				a: p256MontgomeryDomainFieldElement{0x123456789abcdef0, 0x0fedcba987654321, 0x1111111111111111, 0x2222222222222222},
				b: p256MontgomeryDomainFieldElement{0xaaaaaaaaaaaaaaaa, 0xbbbbbbbbbbbbbbbb, 0xcccccccccccccccc, 0xdddddddddddddddd},
			},
		}

		for i, tc := range testCases {
			var outGo, outAsm p256MontgomeryDomainFieldElement
			p256MulGeneric(&outGo, &tc.a, &tc.b)
			p256Mul(&outAsm, &tc.a, &tc.b)

			if outGo != outAsm {
				t.Errorf("Test case %d mismatch:\n"+
					"  Arg1: %016x %016x %016x %016x\n"+
					"  Arg2: %016x %016x %016x %016x\n"+
					"  Go:   %016x %016x %016x %016x\n"+
					"  Asm:  %016x %016x %016x %016x\n",
					i,
					tc.a[0], tc.a[1], tc.a[2], tc.a[3],
					tc.b[0], tc.b[1], tc.b[2], tc.b[3],
					outGo[0], outGo[1], outGo[2], outGo[3],
					outAsm[0], outAsm[1], outAsm[2], outAsm[3])
			}
		}
	})
}

// DebugState 用于保存中间状态
type DebugState struct {
	Acc [8]uint64
}

// TestP256MulDebug 用于调试
func TestP256MulDebug(t *testing.T) {
	a := p256MontgomeryDomainFieldElement{0x123456789abcdef0, 0x0fedcba987654321, 0x1111111111111111, 0x2222222222222222}
	b := p256MontgomeryDomainFieldElement{0xaaaaaaaaaaaaaaaa, 0xbbbbbbbbbbbbbbbb, 0xcccccccccccccccc, 0xdddddddddddddddd}

	fmt.Println("=== P256 Mul Debug ===")
	fmt.Printf("Input A (x): [%016x, %016x, %016x, %016x]\n", a[0], a[1], a[2], a[3])
	fmt.Printf("Input B (y): [%016x, %016x, %016x, %016x]\n", b[0], b[1], b[2], b[3])

	// Go 实现
	fmt.Println("\n--- Go Implementation (ARM64 style) ---")
	outGo := p256MulDebugGo(&a, &b)
	fmt.Printf("Go Result: [%016x, %016x, %016x, %016x]\n", outGo[0], outGo[1], outGo[2], outGo[3])

	// 汇编实现
	fmt.Println("\n--- Assembly Implementation ---")
	var outAsm p256MontgomeryDomainFieldElement
	var asmStates [8]DebugState
	p256MulWithDebug(&outAsm, &a, &b, &asmStates)

	stages := []string{"After Round 0", "After Reduce 1", "After Round 1", "After Reduce 2",
		"After Round 2", "After Reduce 3", "After Round 3", "After Reduce 4"}
	fmt.Println("Asm Intermediate Values:")
	for i, name := range stages {
		s := asmStates[i]
		fmt.Printf("%s: [%016x, %016x, %016x, %016x, %016x, %016x, %016x, %016x]\n",
			name, s.Acc[0], s.Acc[1], s.Acc[2], s.Acc[3], s.Acc[4], s.Acc[5], s.Acc[6], s.Acc[7])
	}
	fmt.Printf("Asm Result: [%016x, %016x, %016x, %016x]\n", outAsm[0], outAsm[1], outAsm[2], outAsm[3])

	if outGo != outAsm {
		t.Errorf("Results don't match!")
	}
}

// p256MulDebugGo - ARM64 风格的实现
// 注意: bits.Mul64 返回 (hi, lo)，但 ARM64 约定 acc0=lo, acc1=hi
func p256MulDebugGo(arg1, arg2 *p256MontgomeryDomainFieldElement) p256MontgomeryDomainFieldElement {
	const const0 uint64 = 0x00000000ffffffff
	const const1 uint64 = 0xffffffff00000001

	x0, x1, x2, x3 := arg1[0], arg1[1], arg1[2], arg1[3]
	y0, y1, y2, y3 := arg2[0], arg2[1], arg2[2], arg2[3]

	var acc0, acc1, acc2, acc3, acc4, acc5, acc6, acc7 uint64
	var c uint64

	// Round 0: y[0] * x
	// 注意: bits.Mul64 返回 (hi, lo)
	hi, lo := bits.Mul64(y0, x0)
	acc0 = lo // ARM64: MUL -> low bits
	acc1 = hi // ARM64: UMULH -> high bits

	hi, lo = bits.Mul64(y0, x1)
	acc1, c = bits.Add64(acc1, lo, 0)
	acc2 = hi + c

	hi, lo = bits.Mul64(y0, x2)
	acc2, c = bits.Add64(acc2, lo, 0)
	acc3 = hi + c

	hi, lo = bits.Mul64(y0, x3)
	acc3, c = bits.Add64(acc3, lo, 0)
	acc4 = hi + c

	fmt.Printf("After Round 0:  acc=[%016x, %016x, %016x, %016x, %016x]\n", acc0, acc1, acc2, acc3, acc4)

	// First reduction step
	t0 := acc0 << 32
	t1 := acc0 >> 32
	t2hi, t2lo := bits.Mul64(acc0, const1)
	acc0 = t2hi

	acc1, c = bits.Add64(acc1, t0, 0)
	acc2, c = bits.Add64(acc2, t1, c)
	acc3, c = bits.Add64(acc3, t2lo, c)
	acc0 += c

	fmt.Printf("After Reduce 1: acc=[%016x, %016x, %016x, %016x, %016x]\n", acc0, acc1, acc2, acc3, acc4)

	// Round 1: y[1] * x
	// bits.Mul64 returns (hi, lo)
	h0, _ := bits.Mul64(y1, x0)
	h1, _ := bits.Mul64(y1, x1)
	h2, _ := bits.Mul64(y1, x2)
	h3, _ := bits.Mul64(y1, x3)

	lo = y1 * x0
	acc1, c = bits.Add64(acc1, lo, 0)
	lo = y1 * x1
	acc2, c = bits.Add64(acc2, lo, c)
	lo = y1 * x2
	acc3, c = bits.Add64(acc3, lo, c)
	lo = y1 * x3
	acc4, c = bits.Add64(acc4, lo, c)
	acc5 = c

	acc2, c = bits.Add64(acc2, h0, 0)
	acc3, c = bits.Add64(acc3, h1, c)
	acc4, c = bits.Add64(acc4, h2, c)
	acc5 += h3 + c

	fmt.Printf("After Round 1:  acc=[%016x, %016x, %016x, %016x, %016x, %016x]\n", acc0, acc1, acc2, acc3, acc4, acc5)

	// Second reduction step
	t0 = acc1 << 32
	t1 = acc1 >> 32
	t2hi, t2lo = bits.Mul64(acc1, const1)
	acc1 = t2hi

	acc2, c = bits.Add64(acc2, t0, 0)
	acc3, c = bits.Add64(acc3, t1, c)
	acc0, c = bits.Add64(acc0, t2lo, c)
	acc1 += c

	fmt.Printf("After Reduce 2: acc=[%016x, %016x, %016x, %016x, %016x, %016x]\n", acc0, acc1, acc2, acc3, acc4, acc5)

	// Round 2: y[2] * x
	// bits.Mul64 returns (hi, lo)
	h0, _ = bits.Mul64(y2, x0)
	h1, _ = bits.Mul64(y2, x1)
	h2, _ = bits.Mul64(y2, x2)
	h3, _ = bits.Mul64(y2, x3)

	lo = y2 * x0
	acc2, c = bits.Add64(acc2, lo, 0)
	lo = y2 * x1
	acc3, c = bits.Add64(acc3, lo, c)
	lo = y2 * x2
	acc4, c = bits.Add64(acc4, lo, c)
	lo = y2 * x3
	acc5, c = bits.Add64(acc5, lo, c)
	acc6 = c

	acc3, c = bits.Add64(acc3, h0, 0)
	acc4, c = bits.Add64(acc4, h1, c)
	acc5, c = bits.Add64(acc5, h2, c)
	acc6 += h3 + c

	fmt.Printf("After Round 2:  acc=[%016x, %016x, %016x, %016x, %016x, %016x, %016x]\n", acc0, acc1, acc2, acc3, acc4, acc5, acc6)

	// Third reduction step
	t0 = acc2 << 32
	t1 = acc2 >> 32
	t2hi, t2lo = bits.Mul64(acc2, const1)
	acc2 = t2hi

	acc3, c = bits.Add64(acc3, t0, 0)
	acc0, c = bits.Add64(acc0, t1, c)
	acc1, c = bits.Add64(acc1, t2lo, c)
	acc2 += c

	fmt.Printf("After Reduce 3: acc=[%016x, %016x, %016x, %016x, %016x, %016x, %016x]\n", acc0, acc1, acc2, acc3, acc4, acc5, acc6)

	// Round 3: y[3] * x
	// bits.Mul64 returns (hi, lo)
	h0, _ = bits.Mul64(y3, x0)
	h1, _ = bits.Mul64(y3, x1)
	h2, _ = bits.Mul64(y3, x2)
	h3, _ = bits.Mul64(y3, x3)

	lo = y3 * x0
	acc3, c = bits.Add64(acc3, lo, 0)
	lo = y3 * x1
	acc4, c = bits.Add64(acc4, lo, c)
	lo = y3 * x2
	acc5, c = bits.Add64(acc5, lo, c)
	lo = y3 * x3
	acc6, c = bits.Add64(acc6, lo, c)
	acc7 = c

	acc4, c = bits.Add64(acc4, h0, 0)
	acc5, c = bits.Add64(acc5, h1, c)
	acc6, c = bits.Add64(acc6, h2, c)
	acc7 += h3 + c

	fmt.Printf("After Round 3:  acc=[%016x, %016x, %016x, %016x, %016x, %016x, %016x, %016x]\n", acc0, acc1, acc2, acc3, acc4, acc5, acc6, acc7)

	// Last reduction step
	t0 = acc3 << 32
	t1 = acc3 >> 32
	t2hi, t2lo = bits.Mul64(acc3, const1)
	acc3 = t2hi

	acc0, c = bits.Add64(acc0, t0, 0)
	acc1, c = bits.Add64(acc1, t1, c)
	acc2, c = bits.Add64(acc2, t2lo, c)
	acc3 += c

	fmt.Printf("After Reduce 4: acc=[%016x, %016x, %016x, %016x, %016x, %016x, %016x, %016x]\n", acc0, acc1, acc2, acc3, acc4, acc5, acc6, acc7)

	// Add high bits
	acc0, c = bits.Add64(acc0, acc4, 0)
	acc1, c = bits.Add64(acc1, acc5, c)
	acc2, c = bits.Add64(acc2, acc6, c)
	acc3, c = bits.Add64(acc3, acc7, c)
	acc4 = c

	fmt.Printf("After Add High: acc=[%016x, %016x, %016x, %016x], carry=%d\n", acc0, acc1, acc2, acc3, acc4)

	// Conditional subtraction
	var b uint64
	r0, b := bits.Sub64(acc0, 0xffffffffffffffff, 0)
	r1, b := bits.Sub64(acc1, const0, b)
	r2, b := bits.Sub64(acc2, 0, b)
	r3, b := bits.Sub64(acc3, const1, b)
	_, b = bits.Sub64(acc4, 0, b)

	fmt.Printf("Sub result:     t=[%016x, %016x, %016x, %016x], borrow=%d\n", r0, r1, r2, r3, b)

	if b == 1 {
		return p256MontgomeryDomainFieldElement{acc0, acc1, acc2, acc3}
	}
	return p256MontgomeryDomainFieldElement{r0, r1, r2, r3}
}

//go:noescape
func p256MulWithDebug(out, a, b *p256MontgomeryDomainFieldElement, states *[8]DebugState)

func BenchmarkP256Mul(b *testing.B) {
	arg1 := p256MontgomeryDomainFieldElement{0x123456789abcdef0, 0x0fedcba987654321, 0x1111111111111111, 0x2222222222222222}
	arg2 := p256MontgomeryDomainFieldElement{0x123456789abcdef0, 0x0fedcba987654321, 0x1111111111111111, 0x2222222222222222}
	var outAsm p256MontgomeryDomainFieldElement
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p256Mul(&outAsm, &arg1, &arg2)
	}
}
