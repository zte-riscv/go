// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build riscv64

package fiat

import (
	"testing"
)

func TestP256Mul(t *testing.T) {
	t.Run("CompareAssemblyWithGo", func(t *testing.T) {
		testCases := []struct {
			name string
			a, b p256MontgomeryDomainFieldElement
		}{
			// Basic cases
			{
				name: "one_times_one",
				a:    p256MontgomeryDomainFieldElement{1, 0, 0, 0},
				b:    p256MontgomeryDomainFieldElement{1, 0, 0, 0},
			},
			{
				name: "max_low_times_two",
				a:    p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0, 0, 0},
				b:    p256MontgomeryDomainFieldElement{2, 0, 0, 0},
			},
			{
				name: "random_large_values",
				a:    p256MontgomeryDomainFieldElement{0x123456789abcdef0, 0x0fedcba987654321, 0x1111111111111111, 0x2222222222222222},
				b:    p256MontgomeryDomainFieldElement{0xaaaaaaaaaaaaaaaa, 0xbbbbbbbbbbbbbbbb, 0xcccccccccccccccc, 0xdddddddddddddddd},
			},
			// Zero cases
			{
				name: "zero_times_zero",
				a:    p256MontgomeryDomainFieldElement{0, 0, 0, 0},
				b:    p256MontgomeryDomainFieldElement{0, 0, 0, 0},
			},
			{
				name: "zero_times_one",
				a:    p256MontgomeryDomainFieldElement{0, 0, 0, 0},
				b:    p256MontgomeryDomainFieldElement{1, 0, 0, 0},
			},
			{
				name: "one_times_zero",
				a:    p256MontgomeryDomainFieldElement{1, 0, 0, 0},
				b:    p256MontgomeryDomainFieldElement{0, 0, 0, 0},
			},
			{
				name: "zero_times_max",
				a:    p256MontgomeryDomainFieldElement{0, 0, 0, 0},
				b:    p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
			},
			// Unit element cases
			{
				name: "one_times_max",
				a:    p256MontgomeryDomainFieldElement{1, 0, 0, 0},
				b:    p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
			},
			{
				name: "max_times_one",
				a:    p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
				b:    p256MontgomeryDomainFieldElement{1, 0, 0, 0},
			},
			// Maximum values
			{
				name: "max_times_max",
				a:    p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
				b:    p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
			},
			// Boundary values - P256 prime modulus components
			// P256 = 2^256 - 2^224 + 2^192 + 2^96 - 1
			// In Montgomery domain representation
			{
				name: "p256_prime_like_values",
				a:    p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0x00000000ffffffff, 0x0000000000000000, 0xffffffff00000001},
				b:    p256MontgomeryDomainFieldElement{0x0000000000000001, 0xffffffff00000000, 0xffffffffffffffff, 0x00000000fffffffe},
			},
			// Special bit patterns
			{
				name: "alternating_bits",
				a:    p256MontgomeryDomainFieldElement{0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x5555555555555555, 0xaaaaaaaaaaaaaaaa},
				b:    p256MontgomeryDomainFieldElement{0xaaaaaaaaaaaaaaaa, 0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x5555555555555555},
			},
			{
				name: "all_ones_pattern",
				a:    p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
				b:    p256MontgomeryDomainFieldElement{0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x3333333333333333, 0xcccccccccccccccc},
			},
			// High bit set cases
			{
				name: "high_bit_set",
				a:    p256MontgomeryDomainFieldElement{0x8000000000000000, 0x8000000000000000, 0x8000000000000000, 0x8000000000000000},
				b:    p256MontgomeryDomainFieldElement{0x8000000000000000, 0x8000000000000000, 0x8000000000000000, 0x8000000000000000},
			},
			// Low limb only
			{
				name: "low_limb_only",
				a:    p256MontgomeryDomainFieldElement{0x1234567890abcdef, 0, 0, 0},
				b:    p256MontgomeryDomainFieldElement{0xfedcba0987654321, 0, 0, 0},
			},
			// High limb only
			{
				name: "high_limb_only",
				a:    p256MontgomeryDomainFieldElement{0, 0, 0, 0x1234567890abcdef},
				b:    p256MontgomeryDomainFieldElement{0, 0, 0, 0xfedcba0987654321},
			},
			// Middle limbs
			{
				name: "middle_limbs",
				a:    p256MontgomeryDomainFieldElement{0, 0x1111111111111111, 0x2222222222222222, 0},
				b:    p256MontgomeryDomainFieldElement{0, 0x3333333333333333, 0x4444444444444444, 0},
			},
			// Small values
			{
				name: "small_values",
				a:    p256MontgomeryDomainFieldElement{2, 0, 0, 0},
				b:    p256MontgomeryDomainFieldElement{3, 0, 0, 0},
			},
			{
				name: "small_times_large",
				a:    p256MontgomeryDomainFieldElement{2, 0, 0, 0},
				b:    p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
			},
			// More random-like values
			{
				name: "random_pattern_1",
				a:    p256MontgomeryDomainFieldElement{0xdeadbeefcafebabe, 0x1234567890abcdef, 0xfedcba0987654321, 0xabcdef0123456789},
				b:    p256MontgomeryDomainFieldElement{0x1122334455667788, 0x99aabbccddeeff00, 0x0011223344556677, 0x8899aabbccddeeff},
			},
			{
				name: "random_pattern_2",
				a:    p256MontgomeryDomainFieldElement{0x0123456789abcdef, 0xfedcba9876543210, 0x13579bdf2468ace0, 0xfdb97531eca86420},
				b:    p256MontgomeryDomainFieldElement{0xf0e1d2c3b4a59687, 0x78695a4b3c2d1e0f, 0x0f1e2d3c4b5a6978, 0x8796a5b4c3d2e1f0},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var outGo, outAsm p256MontgomeryDomainFieldElement
				p256MulGeneric(&outGo, &tc.a, &tc.b)
				p256Mul(&outAsm, &tc.a, &tc.b)

				if outGo != outAsm {
					t.Errorf("Mismatch:\n"+
						"  Arg1: %016x %016x %016x %016x\n"+
						"  Arg2: %016x %016x %016x %016x\n"+
						"  Go:   %016x %016x %016x %016x\n"+
						"  Asm:  %016x %016x %016x %016x\n",
						tc.a[0], tc.a[1], tc.a[2], tc.a[3],
						tc.b[0], tc.b[1], tc.b[2], tc.b[3],
						outGo[0], outGo[1], outGo[2], outGo[3],
						outAsm[0], outAsm[1], outAsm[2], outAsm[3])
				}
			})
		}
	})
}

func TestP256Square(t *testing.T) {
	t.Run("CompareAssemblyWithGo", func(t *testing.T) {
		testCases := []struct {
			name string
			a    p256MontgomeryDomainFieldElement
		}{
			// Basic cases
			{
				name: "one_squared",
				a:    p256MontgomeryDomainFieldElement{1, 0, 0, 0},
			},
			{
				name: "two_squared",
				a:    p256MontgomeryDomainFieldElement{2, 0, 0, 0},
			},
			{
				name: "max_low_squared",
				a:    p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0, 0, 0},
			},
			{
				name: "random_large_value",
				a:    p256MontgomeryDomainFieldElement{0x123456789abcdef0, 0x0fedcba987654321, 0x1111111111111111, 0x2222222222222222},
			},
			// Zero cases
			{
				name: "zero_squared",
				a:    p256MontgomeryDomainFieldElement{0, 0, 0, 0},
			},
			// Unit element cases
			{
				name: "one_squared",
				a:    p256MontgomeryDomainFieldElement{1, 0, 0, 0},
			},
			{
				name: "max_squared",
				a:    p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
			},
			// Boundary values - P256 prime modulus components
			// P256 = 2^256 - 2^224 + 2^192 + 2^96 - 1
			// In Montgomery domain representation
			{
				name: "p256_prime_like_value",
				a:    p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0x00000000ffffffff, 0x0000000000000000, 0xffffffff00000001},
			},
			{
				name: "p256_one_in_montgomery",
				a:    p256MontgomeryDomainFieldElement{0x0000000000000001, 0xffffffff00000000, 0xffffffffffffffff, 0x00000000fffffffe},
			},
			// Special bit patterns
			{
				name: "alternating_bits",
				a:    p256MontgomeryDomainFieldElement{0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x5555555555555555, 0xaaaaaaaaaaaaaaaa},
			},
			{
				name: "all_ones_pattern",
				a:    p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
			},
			{
				name: "checkerboard_pattern",
				a:    p256MontgomeryDomainFieldElement{0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x3333333333333333, 0xcccccccccccccccc},
			},
			// High bit set cases
			{
				name: "high_bit_set",
				a:    p256MontgomeryDomainFieldElement{0x8000000000000000, 0x8000000000000000, 0x8000000000000000, 0x8000000000000000},
			},
			{
				name: "all_high_bits",
				a:    p256MontgomeryDomainFieldElement{0x8000000000000000, 0x8000000000000000, 0x8000000000000000, 0x8000000000000000},
			},
			// Low limb only
			{
				name: "low_limb_only",
				a:    p256MontgomeryDomainFieldElement{0x1234567890abcdef, 0, 0, 0},
			},
			{
				name: "low_limb_max",
				a:    p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0, 0, 0},
			},
			// High limb only
			{
				name: "high_limb_only",
				a:    p256MontgomeryDomainFieldElement{0, 0, 0, 0x1234567890abcdef},
			},
			{
				name: "high_limb_max",
				a:    p256MontgomeryDomainFieldElement{0, 0, 0, 0xffffffffffffffff},
			},
			// Middle limbs
			{
				name: "middle_limbs",
				a:    p256MontgomeryDomainFieldElement{0, 0x1111111111111111, 0x2222222222222222, 0},
			},
			{
				name: "second_limb_only",
				a:    p256MontgomeryDomainFieldElement{0, 0xffffffffffffffff, 0, 0},
			},
			{
				name: "third_limb_only",
				a:    p256MontgomeryDomainFieldElement{0, 0, 0xffffffffffffffff, 0},
			},
			// Small values
			{
				name: "small_value_2",
				a:    p256MontgomeryDomainFieldElement{2, 0, 0, 0},
			},
			{
				name: "small_value_3",
				a:    p256MontgomeryDomainFieldElement{3, 0, 0, 0},
			},
			{
				name: "small_value_256",
				a:    p256MontgomeryDomainFieldElement{256, 0, 0, 0},
			},
			// More random-like values
			{
				name: "random_pattern_1",
				a:    p256MontgomeryDomainFieldElement{0xdeadbeefcafebabe, 0x1234567890abcdef, 0xfedcba0987654321, 0xabcdef0123456789},
			},
			{
				name: "random_pattern_2",
				a:    p256MontgomeryDomainFieldElement{0x0123456789abcdef, 0xfedcba9876543210, 0x13579bdf2468ace0, 0xfdb97531eca86420},
			},
			{
				name: "random_pattern_3",
				a:    p256MontgomeryDomainFieldElement{0xf0e1d2c3b4a59687, 0x78695a4b3c2d1e0f, 0x0f1e2d3c4b5a6978, 0x8796a5b4c3d2e1f0},
			},
			// Edge cases for carry propagation
			{
				name: "carry_propagation_test_1",
				a:    p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0x0000000000000001, 0, 0},
			},
			{
				name: "carry_propagation_test_2",
				a:    p256MontgomeryDomainFieldElement{0, 0xffffffffffffffff, 0x0000000000000001, 0},
			},
			{
				name: "carry_propagation_test_3",
				a:    p256MontgomeryDomainFieldElement{0, 0, 0xffffffffffffffff, 0x0000000000000001},
			},
			// Power of two values (important for bit manipulation)
			{
				name: "power_of_two_1",
				a:    p256MontgomeryDomainFieldElement{0x0000000000000001, 0, 0, 0},
			},
			{
				name: "power_of_two_64",
				a:    p256MontgomeryDomainFieldElement{0x0000000000000000, 0x0000000000000001, 0, 0},
			},
			{
				name: "power_of_two_128",
				a:    p256MontgomeryDomainFieldElement{0x0000000000000000, 0x0000000000000000, 0x0000000000000001, 0},
			},
			{
				name: "power_of_two_192",
				a:    p256MontgomeryDomainFieldElement{0x0000000000000000, 0x0000000000000000, 0x0000000000000000, 0x0000000000000001},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var outGo, outAsm p256MontgomeryDomainFieldElement
				p256SquareGeneric(&outGo, &tc.a)
				p256Square(&outAsm, &tc.a)

				if outGo != outAsm {
					t.Errorf("Mismatch:\n"+
						"  Arg:  %016x %016x %016x %016x\n"+
						"  Go:   %016x %016x %016x %016x\n"+
						"  Asm:  %016x %016x %016x %016x\n",
						tc.a[0], tc.a[1], tc.a[2], tc.a[3],
						outGo[0], outGo[1], outGo[2], outGo[3],
						outAsm[0], outAsm[1], outAsm[2], outAsm[3])
					// Print detailed difference
					for i := 0; i < 4; i++ {
						if outGo[i] != outAsm[i] {
							diff := outGo[i] ^ outAsm[i]
							t.Logf("  Limb[%d] differs: Go=%016x Asm=%016x Diff=%016x",
								i, outGo[i], outAsm[i], diff)
							// Calculate numeric difference
							if outGo[i] > outAsm[i] {
								t.Logf("    Go is larger by %d (0x%x)",
									outGo[i]-outAsm[i], outGo[i]-outAsm[i])
							} else {
								t.Logf("    Asm is larger by %d (0x%x)",
									outAsm[i]-outGo[i], outAsm[i]-outGo[i])
							}
						}
					}
				}
			})
		}
	})
}

func TestP256SquareArm64Style(t *testing.T) {
	t.Run("CompareGenericWithArm64Style", func(t *testing.T) {
		testCases := []struct {
			name string
			arg  p256MontgomeryDomainFieldElement
		}{
			// Basic cases
			{
				name: "zero",
				arg:  p256MontgomeryDomainFieldElement{0, 0, 0, 0},
			},
			{
				name: "one",
				arg:  p256MontgomeryDomainFieldElement{1, 0, 0, 0},
			},
			{
				name: "two",
				arg:  p256MontgomeryDomainFieldElement{2, 0, 0, 0},
			},
			{
				name: "max_low",
				arg:  p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0, 0, 0},
			},
			{
				name: "random_large_value",
				arg:  p256MontgomeryDomainFieldElement{0x123456789abcdef0, 0x0fedcba987654321, 0x1111111111111111, 0x2222222222222222},
			},
			{
				name: "max_squared",
				arg:  p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
			},
			{
				name: "p256_prime_like_value",
				arg:  p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0x00000000ffffffff, 0x0000000000000000, 0xfffffff00000001},
			},
			{
				name: "p256_one_in_montgomery",
				arg:  p256MontgomeryDomainFieldElement{0x0000000000000001, 0xfffffff00000000, 0xffffffffffffffff, 0x00000000fffffffe},
			},
			{
				name: "alternating_bits",
				arg:  p256MontgomeryDomainFieldElement{0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x5555555555555555, 0xaaaaaaaaaaaaaaaa},
			},
			{
				name: "all_ones_pattern",
				arg:  p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
			},
			{
				name: "checkerboard_pattern",
				arg:  p256MontgomeryDomainFieldElement{0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x3333333333333333, 0xcccccccccccccccc},
			},
			{
				name: "high_bit_set",
				arg:  p256MontgomeryDomainFieldElement{0x8000000000000000, 0x8000000000000000, 0x8000000000000000, 0x8000000000000000},
			},
			{
				name: "all_high_bits",
				arg:  p256MontgomeryDomainFieldElement{0x8000000000000000, 0x8000000000000000, 0x8000000000000000, 0x8000000000000000},
			},
			{
				name: "random_pattern_1",
				arg:  p256MontgomeryDomainFieldElement{0xdeadbeefcafebabe, 0x1234567890abcdef, 0xfedcba0987654321, 0xabcdef0123456789},
			},
			{
				name: "random_pattern_2",
				arg:  p256MontgomeryDomainFieldElement{0x0123456789abcdef, 0xfedcba9876543210, 0x13579bdf2468ace0, 0xfdb97531eca86420},
			},
			{
				name: "random_pattern_3",
				arg:  p256MontgomeryDomainFieldElement{0xf0e1d2c3b4a59687, 0x78695a4b3c2d1e0f, 0x0f1e2d3c4b5a6978, 0x8796a5b4c3d2e1f0},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var outGeneric, outArm64Style p256MontgomeryDomainFieldElement

				p256SquareGeneric(&outGeneric, &tc.arg)
				p256SquareArm64Style(&outArm64Style, &tc.arg)

				if outGeneric != outArm64Style {
					t.Errorf("Mismatch for test case %s:", tc.name)
					t.Errorf("  Arg:  %016x %016x %016x %016x",
						tc.arg[0], tc.arg[1], tc.arg[2], tc.arg[3])
					t.Errorf("  Generic:    %016x %016x %016x %016x",
						outGeneric[0], outGeneric[1], outGeneric[2], outGeneric[3])
					t.Errorf("  Arm64Style: %016x %016x %016x %016x",
						outArm64Style[0], outArm64Style[1], outArm64Style[2], outArm64Style[3])

					for i := 0; i < 4; i++ {
						if outGeneric[i] != outArm64Style[i] {
							diff := outGeneric[i] ^ outArm64Style[i]
							if outGeneric[i] > outArm64Style[i] {
								t.Errorf("  Limb[%d] differs: Generic=%016x Arm64Style=%016x Diff=%016x",
									i, outGeneric[i], outArm64Style[i], diff)
								t.Errorf("    Generic is larger by %d (0x%x)",
									outGeneric[i]-outArm64Style[i], outGeneric[i]-outArm64Style[i])
							} else {
								t.Errorf("  Limb[%d] differs: Generic=%016x Arm64Style=%016x Diff=%016x",
									i, outGeneric[i], outArm64Style[i], diff)
								t.Errorf("    Arm64Style is larger by %d (0x%x)",
									outArm64Style[i]-outGeneric[i], outArm64Style[i]-outGeneric[i])
							}
						}
					}
				}
			})
		}
	})
}

func TestP256SquareCDisassemble(t *testing.T) {
	t.Run("CompareGenericWithCDisassemble", func(t *testing.T) {
		testCases := []struct {
			name string
			arg  p256MontgomeryDomainFieldElement
		}{
			// Basic cases
			{
				name: "zero",
				arg:  p256MontgomeryDomainFieldElement{0, 0, 0, 0},
			},
			{
				name: "one",
				arg:  p256MontgomeryDomainFieldElement{1, 0, 0, 0},
			},
			{
				name: "two",
				arg:  p256MontgomeryDomainFieldElement{2, 0, 0, 0},
			},
			{
				name: "max_low",
				arg:  p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0, 0, 0},
			},
			{
				name: "random_large_value",
				arg:  p256MontgomeryDomainFieldElement{0x123456789abcdef0, 0x0fedcba987654321, 0x1111111111111111, 0x2222222222222222},
			},
			{
				name: "max_squared",
				arg:  p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
			},
			{
				name: "p256_prime_like_value",
				arg:  p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0x00000000ffffffff, 0x0000000000000000, 0xfffffff00000001},
			},
			{
				name: "p256_one_in_montgomery",
				arg:  p256MontgomeryDomainFieldElement{0x0000000000000001, 0xfffffff00000000, 0xffffffffffffffff, 0x00000000fffffffe},
			},
			{
				name: "alternating_bits",
				arg:  p256MontgomeryDomainFieldElement{0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x5555555555555555, 0xaaaaaaaaaaaaaaaa},
			},
			{
				name: "all_ones_pattern",
				arg:  p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
			},
			{
				name: "checkerboard_pattern",
				arg:  p256MontgomeryDomainFieldElement{0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x3333333333333333, 0xcccccccccccccccc},
			},
			{
				name: "high_bit_set",
				arg:  p256MontgomeryDomainFieldElement{0x8000000000000000, 0x8000000000000000, 0x8000000000000000, 0x8000000000000000},
			},
			{
				name: "all_high_bits",
				arg:  p256MontgomeryDomainFieldElement{0x8000000000000000, 0x8000000000000000, 0x8000000000000000, 0x8000000000000000},
			},
			{
				name: "random_pattern_1",
				arg:  p256MontgomeryDomainFieldElement{0xdeadbeefcafebabe, 0x1234567890abcdef, 0xfedcba0987654321, 0xabcdef0123456789},
			},
			{
				name: "random_pattern_2",
				arg:  p256MontgomeryDomainFieldElement{0x0123456789abcdef, 0xfedcba9876543210, 0x13579bdf2468ace0, 0xfdb97531eca86420},
			},
			{
				name: "random_pattern_3",
				arg:  p256MontgomeryDomainFieldElement{0xf0e1d2c3b4a59687, 0x78695a4b3c2d1e0f, 0x0f1e2d3c4b5a6978, 0x8796a5b4c3d2e1f0},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var outGeneric, outArm64Style p256MontgomeryDomainFieldElement

				p256SquareGeneric(&outGeneric, &tc.arg)
				p256SquareCDisassemble(&outArm64Style, &tc.arg)

				if outGeneric != outArm64Style {
					t.Errorf("Mismatch for test case %s:", tc.name)
					t.Errorf("  Arg:  %016x %016x %016x %016x",
						tc.arg[0], tc.arg[1], tc.arg[2], tc.arg[3])
					t.Errorf("  Generic:    %016x %016x %016x %016x",
						outGeneric[0], outGeneric[1], outGeneric[2], outGeneric[3])
					t.Errorf("  Arm64Style: %016x %016x %016x %016x",
						outArm64Style[0], outArm64Style[1], outArm64Style[2], outArm64Style[3])

					for i := 0; i < 4; i++ {
						if outGeneric[i] != outArm64Style[i] {
							diff := outGeneric[i] ^ outArm64Style[i]
							if outGeneric[i] > outArm64Style[i] {
								t.Errorf("  Limb[%d] differs: Generic=%016x Arm64Style=%016x Diff=%016x",
									i, outGeneric[i], outArm64Style[i], diff)
								t.Errorf("    Generic is larger by %d (0x%x)",
									outGeneric[i]-outArm64Style[i], outGeneric[i]-outArm64Style[i])
							} else {
								t.Errorf("  Limb[%d] differs: Generic=%016x Arm64Style=%016x Diff=%016x",
									i, outGeneric[i], outArm64Style[i], diff)
								t.Errorf("    Arm64Style is larger by %d (0x%x)",
									outArm64Style[i]-outGeneric[i], outArm64Style[i]-outGeneric[i])
							}
						}
					}
				}
			})
		}
	})
}

func TestP256Add(t *testing.T) {
	t.Run("CompareGenericWithAssembly", func(t *testing.T) {
		testCases := []struct {
			name string
			arg1 p256MontgomeryDomainFieldElement
			arg2 p256MontgomeryDomainFieldElement
		}{
			// Basic cases
			{
				name: "zero_plus_zero",
				arg1: p256MontgomeryDomainFieldElement{0, 0, 0, 0},
				arg2: p256MontgomeryDomainFieldElement{0, 0, 0, 0},
			},
			{
				name: "zero_plus_one",
				arg1: p256MontgomeryDomainFieldElement{0, 0, 0, 0},
				arg2: p256MontgomeryDomainFieldElement{1, 0, 0, 0},
			},
			{
				name: "one_plus_zero",
				arg1: p256MontgomeryDomainFieldElement{1, 0, 0, 0},
				arg2: p256MontgomeryDomainFieldElement{0, 0, 0, 0},
			},
			{
				name: "one_plus_one",
				arg1: p256MontgomeryDomainFieldElement{1, 0, 0, 0},
				arg2: p256MontgomeryDomainFieldElement{1, 0, 0, 0},
			},
			{
				name: "two_plus_three",
				arg1: p256MontgomeryDomainFieldElement{2, 0, 0, 0},
				arg2: p256MontgomeryDomainFieldElement{3, 0, 0, 0},
			},
			// Maximum values
			{
				name: "max_plus_zero",
				arg1: p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
				arg2: p256MontgomeryDomainFieldElement{0, 0, 0, 0},
			},
			{
				name: "zero_plus_max",
				arg1: p256MontgomeryDomainFieldElement{0, 0, 0, 0},
				arg2: p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
			},
			{
				name: "max_low_plus_one",
				arg1: p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0, 0, 0},
				arg2: p256MontgomeryDomainFieldElement{1, 0, 0, 0},
			},
			// Random large values
			{
				name: "random_large_values_1",
				arg1: p256MontgomeryDomainFieldElement{0x123456789abcdef0, 0x0fedcba987654321, 0x1111111111111111, 0x2222222222222222},
				arg2: p256MontgomeryDomainFieldElement{0xaaaaaaaaaaaaaaaa, 0xbbbbbbbbbbbbbbbb, 0xcccccccccccccccc, 0xdddddddddddddddd},
			},
			{
				name: "random_large_values_2",
				arg1: p256MontgomeryDomainFieldElement{0xdeadbeefcafebabe, 0x1234567890abcdef, 0xfedcba0987654321, 0xabcdef0123456789},
				arg2: p256MontgomeryDomainFieldElement{0x1122334455667788, 0x99aabbccddeeff00, 0x0011223344556677, 0x8899aabbccddeeff},
			},
			// P256 prime-like values
			{
				name: "p256_prime_like_values",
				arg1: p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0x00000000ffffffff, 0x0000000000000000, 0xffffffff00000001},
				arg2: p256MontgomeryDomainFieldElement{0x0000000000000001, 0xffffffff00000000, 0xffffffffffffffff, 0x00000000fffffffe},
			},
			// Special bit patterns
			{
				name: "alternating_bits",
				arg1: p256MontgomeryDomainFieldElement{0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x5555555555555555, 0xaaaaaaaaaaaaaaaa},
				arg2: p256MontgomeryDomainFieldElement{0xaaaaaaaaaaaaaaaa, 0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x5555555555555555},
			},
			{
				name: "all_ones_pattern",
				arg1: p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
				arg2: p256MontgomeryDomainFieldElement{0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x3333333333333333, 0xcccccccccccccccc},
			},
			{
				name: "checkerboard_pattern",
				arg1: p256MontgomeryDomainFieldElement{0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x3333333333333333, 0xcccccccccccccccc},
				arg2: p256MontgomeryDomainFieldElement{0xaaaaaaaaaaaaaaaa, 0x5555555555555555, 0xcccccccccccccccc, 0x3333333333333333},
			},
			// High bit set cases
			{
				name: "high_bit_set",
				arg1: p256MontgomeryDomainFieldElement{0x8000000000000000, 0x8000000000000000, 0x8000000000000000, 0x8000000000000000},
				arg2: p256MontgomeryDomainFieldElement{0x8000000000000000, 0x8000000000000000, 0x8000000000000000, 0x8000000000000000},
			},
			// Low limb only
			{
				name: "low_limb_only",
				arg1: p256MontgomeryDomainFieldElement{0x1234567890abcdef, 0, 0, 0},
				arg2: p256MontgomeryDomainFieldElement{0xfedcba0987654321, 0, 0, 0},
			},
			// High limb only
			{
				name: "high_limb_only",
				arg1: p256MontgomeryDomainFieldElement{0, 0, 0, 0x1234567890abcdef},
				arg2: p256MontgomeryDomainFieldElement{0, 0, 0, 0xfedcba0987654321},
			},
			// Middle limbs
			{
				name: "middle_limbs",
				arg1: p256MontgomeryDomainFieldElement{0, 0x1111111111111111, 0x2222222222222222, 0},
				arg2: p256MontgomeryDomainFieldElement{0, 0x3333333333333333, 0x4444444444444444, 0},
			},
			// Small values
			{
				name: "small_values",
				arg1: p256MontgomeryDomainFieldElement{2, 0, 0, 0},
				arg2: p256MontgomeryDomainFieldElement{3, 0, 0, 0},
			},
			// More random-like values
			{
				name: "random_pattern_1",
				arg1: p256MontgomeryDomainFieldElement{0xdeadbeefcafebabe, 0x1234567890abcdef, 0xfedcba0987654321, 0xabcdef0123456789},
				arg2: p256MontgomeryDomainFieldElement{0x1122334455667788, 0x99aabbccddeeff00, 0x0011223344556677, 0x8899aabbccddeeff},
			},
			{
				name: "random_pattern_2",
				arg1: p256MontgomeryDomainFieldElement{0x0123456789abcdef, 0xfedcba9876543210, 0x13579bdf2468ace0, 0xfdb97531eca86420},
				arg2: p256MontgomeryDomainFieldElement{0xf0e1d2c3b4a59687, 0x78695a4b3c2d1e0f, 0x0f1e2d3c4b5a6978, 0x8796a5b4c3d2e1f0},
			},
			{
				name: "random_pattern_3",
				arg1: p256MontgomeryDomainFieldElement{0xf0e1d2c3b4a59687, 0x78695a4b3c2d1e0f, 0x0f1e2d3c4b5a6978, 0x8796a5b4c3d2e1f0},
				arg2: p256MontgomeryDomainFieldElement{0x0123456789abcdef, 0xfedcba9876543210, 0x13579bdf2468ace0, 0xfdb97531eca86420},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var outGeneric, outAssembly p256MontgomeryDomainFieldElement

				p256AddGeneric(&outGeneric, &tc.arg1, &tc.arg2)
				p256Add(&outAssembly, &tc.arg1, &tc.arg2)

				if outGeneric != outAssembly {
					t.Errorf("Mismatch for test case %s:", tc.name)
					t.Errorf("  Arg1: %016x %016x %016x %016x",
						tc.arg1[0], tc.arg1[1], tc.arg1[2], tc.arg1[3])
					t.Errorf("  Arg2: %016x %016x %016x %016x",
						tc.arg2[0], tc.arg2[1], tc.arg2[2], tc.arg2[3])
					t.Errorf("  Generic:    %016x %016x %016x %016x",
						outGeneric[0], outGeneric[1], outGeneric[2], outGeneric[3])
					t.Errorf("  Assembly:   %016x %016x %016x %016x",
						outAssembly[0], outAssembly[1], outAssembly[2], outAssembly[3])

					for i := 0; i < 4; i++ {
						if outGeneric[i] != outAssembly[i] {
							diff := outGeneric[i] ^ outAssembly[i]
							if outGeneric[i] > outAssembly[i] {
								t.Errorf("  Limb[%d] differs: Generic=%016x Assembly=%016x Diff=%016x",
									i, outGeneric[i], outAssembly[i], diff)
								t.Errorf("    Generic is larger by %d (0x%x)",
									outGeneric[i]-outAssembly[i], outGeneric[i]-outAssembly[i])
							} else {
								t.Errorf("  Limb[%d] differs: Generic=%016x Assembly=%016x Diff=%016x",
									i, outGeneric[i], outAssembly[i], diff)
								t.Errorf("    Assembly is larger by %d (0x%x)",
									outAssembly[i]-outGeneric[i], outAssembly[i]-outGeneric[i])
							}
						}
					}
				}
			})
		}
	})
}

func TestP256Sub(t *testing.T) {
	t.Run("CompareGenericWithAssembly", func(t *testing.T) {
		testCases := []struct {
			name string
			arg1 p256MontgomeryDomainFieldElement
			arg2 p256MontgomeryDomainFieldElement
		}{
			// Basic cases
			{
				name: "zero_minus_zero",
				arg1: p256MontgomeryDomainFieldElement{0, 0, 0, 0},
				arg2: p256MontgomeryDomainFieldElement{0, 0, 0, 0},
			},
			{
				name: "one_minus_zero",
				arg1: p256MontgomeryDomainFieldElement{1, 0, 0, 0},
				arg2: p256MontgomeryDomainFieldElement{0, 0, 0, 0},
			},
			{
				name: "zero_minus_one",
				arg1: p256MontgomeryDomainFieldElement{0, 0, 0, 0},
				arg2: p256MontgomeryDomainFieldElement{1, 0, 0, 0},
			},
			{
				name: "one_minus_one",
				arg1: p256MontgomeryDomainFieldElement{1, 0, 0, 0},
				arg2: p256MontgomeryDomainFieldElement{1, 0, 0, 0},
			},
			{
				name: "three_minus_two",
				arg1: p256MontgomeryDomainFieldElement{3, 0, 0, 0},
				arg2: p256MontgomeryDomainFieldElement{2, 0, 0, 0},
			},
			// Maximum values
			{
				name: "max_minus_zero",
				arg1: p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
				arg2: p256MontgomeryDomainFieldElement{0, 0, 0, 0},
			},
			{
				name: "zero_minus_max",
				arg1: p256MontgomeryDomainFieldElement{0, 0, 0, 0},
				arg2: p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
			},
			{
				name: "max_low_minus_one",
				arg1: p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0, 0, 0},
				arg2: p256MontgomeryDomainFieldElement{1, 0, 0, 0},
			},
			// Random large values
			{
				name: "random_large_values_1",
				arg1: p256MontgomeryDomainFieldElement{0x123456789abcdef0, 0x0fedcba987654321, 0x1111111111111111, 0x2222222222222222},
				arg2: p256MontgomeryDomainFieldElement{0xaaaaaaaaaaaaaaaa, 0xbbbbbbbbbbbbbbbb, 0xcccccccccccccccc, 0xdddddddddddddddd},
			},
			{
				name: "random_large_values_2",
				arg1: p256MontgomeryDomainFieldElement{0xdeadbeefcafebabe, 0x1234567890abcdef, 0xfedcba0987654321, 0xabcdef0123456789},
				arg2: p256MontgomeryDomainFieldElement{0x1122334455667788, 0x99aabbccddeeff00, 0x0011223344556677, 0x8899aabbccddeeff},
			},
			// P256 prime-like values
			{
				name: "p256_prime_like_values",
				arg1: p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0x00000000ffffffff, 0x0000000000000000, 0xffffffff00000001},
				arg2: p256MontgomeryDomainFieldElement{0x0000000000000001, 0xffffffff00000000, 0xffffffffffffffff, 0x00000000fffffffe},
			},
			// Special bit patterns
			{
				name: "alternating_bits",
				arg1: p256MontgomeryDomainFieldElement{0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x5555555555555555, 0xaaaaaaaaaaaaaaaa},
				arg2: p256MontgomeryDomainFieldElement{0xaaaaaaaaaaaaaaaa, 0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x5555555555555555},
			},
			{
				name: "all_ones_pattern",
				arg1: p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
				arg2: p256MontgomeryDomainFieldElement{0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x3333333333333333, 0xcccccccccccccccc},
			},
			{
				name: "checkerboard_pattern",
				arg1: p256MontgomeryDomainFieldElement{0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x3333333333333333, 0xcccccccccccccccc},
				arg2: p256MontgomeryDomainFieldElement{0xaaaaaaaaaaaaaaaa, 0x5555555555555555, 0xcccccccccccccccc, 0x3333333333333333},
			},
			// High bit set cases
			{
				name: "high_bit_set",
				arg1: p256MontgomeryDomainFieldElement{0x8000000000000000, 0x8000000000000000, 0x8000000000000000, 0x8000000000000000},
				arg2: p256MontgomeryDomainFieldElement{0x8000000000000000, 0x8000000000000000, 0x8000000000000000, 0x8000000000000000},
			},
			// Low limb only
			{
				name: "low_limb_only",
				arg1: p256MontgomeryDomainFieldElement{0x1234567890abcdef, 0, 0, 0},
				arg2: p256MontgomeryDomainFieldElement{0xfedcba0987654321, 0, 0, 0},
			},
			// High limb only
			{
				name: "high_limb_only",
				arg1: p256MontgomeryDomainFieldElement{0, 0, 0, 0x1234567890abcdef},
				arg2: p256MontgomeryDomainFieldElement{0, 0, 0, 0xfedcba0987654321},
			},
			// Middle limbs
			{
				name: "middle_limbs",
				arg1: p256MontgomeryDomainFieldElement{0, 0x1111111111111111, 0x2222222222222222, 0},
				arg2: p256MontgomeryDomainFieldElement{0, 0x3333333333333333, 0x4444444444444444, 0},
			},
			// Small values
			{
				name: "small_values",
				arg1: p256MontgomeryDomainFieldElement{3, 0, 0, 0},
				arg2: p256MontgomeryDomainFieldElement{2, 0, 0, 0},
			},
			// More random-like values
			{
				name: "random_pattern_1",
				arg1: p256MontgomeryDomainFieldElement{0xdeadbeefcafebabe, 0x1234567890abcdef, 0xfedcba0987654321, 0xabcdef0123456789},
				arg2: p256MontgomeryDomainFieldElement{0x1122334455667788, 0x99aabbccddeeff00, 0x0011223344556677, 0x8899aabbccddeeff},
			},
			{
				name: "random_pattern_2",
				arg1: p256MontgomeryDomainFieldElement{0x0123456789abcdef, 0xfedcba9876543210, 0x13579bdf2468ace0, 0xfdb97531eca86420},
				arg2: p256MontgomeryDomainFieldElement{0xf0e1d2c3b4a59687, 0x78695a4b3c2d1e0f, 0x0f1e2d3c4b5a6978, 0x8796a5b4c3d2e1f0},
			},
			{
				name: "random_pattern_3",
				arg1: p256MontgomeryDomainFieldElement{0xf0e1d2c3b4a59687, 0x78695a4b3c2d1e0f, 0x0f1e2d3c4b5a6978, 0x8796a5b4c3d2e1f0},
				arg2: p256MontgomeryDomainFieldElement{0x0123456789abcdef, 0xfedcba9876543210, 0x13579bdf2468ace0, 0xfdb97531eca86420},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var outGeneric, outAssembly p256MontgomeryDomainFieldElement

				p256SubGeneric(&outGeneric, &tc.arg1, &tc.arg2)
				p256Sub(&outAssembly, &tc.arg1, &tc.arg2)

				if outGeneric != outAssembly {
					t.Errorf("Mismatch for test case %s:", tc.name)
					t.Errorf("  Arg1: %016x %016x %016x %016x",
						tc.arg1[0], tc.arg1[1], tc.arg1[2], tc.arg1[3])
					t.Errorf("  Arg2: %016x %016x %016x %016x",
						tc.arg2[0], tc.arg2[1], tc.arg2[2], tc.arg2[3])
					t.Errorf("  Generic:    %016x %016x %016x %016x",
						outGeneric[0], outGeneric[1], outGeneric[2], outGeneric[3])
					t.Errorf("  Assembly:   %016x %016x %016x %016x",
						outAssembly[0], outAssembly[1], outAssembly[2], outAssembly[3])

					for i := 0; i < 4; i++ {
						if outGeneric[i] != outAssembly[i] {
							diff := outGeneric[i] ^ outAssembly[i]
							if outGeneric[i] > outAssembly[i] {
								t.Errorf("  Limb[%d] differs: Generic=%016x Assembly=%016x Diff=%016x",
									i, outGeneric[i], outAssembly[i], diff)
								t.Errorf("    Generic is larger by %d (0x%x)",
									outGeneric[i]-outAssembly[i], outGeneric[i]-outAssembly[i])
							} else {
								t.Errorf("  Limb[%d] differs: Generic=%016x Assembly=%016x Diff=%016x",
									i, outGeneric[i], outAssembly[i], diff)
								t.Errorf("    Assembly is larger by %d (0x%x)",
									outAssembly[i]-outGeneric[i], outAssembly[i]-outGeneric[i])
							}
						}
					}
				}
			})
		}
	})
}

func BenchmarkP256SquareArm64Style(b *testing.B) {
	testCases := []struct {
		name string
		arg  p256MontgomeryDomainFieldElement
	}{
		// Basic cases
		{
			name: "zero",
			arg:  p256MontgomeryDomainFieldElement{0, 0, 0, 0},
		},
		{
			name: "one",
			arg:  p256MontgomeryDomainFieldElement{1, 0, 0, 0},
		},
		{
			name: "two",
			arg:  p256MontgomeryDomainFieldElement{2, 0, 0, 0},
		},
		{
			name: "max_low",
			arg:  p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0, 0, 0},
		},
		{
			name: "random_large_value",
			arg:  p256MontgomeryDomainFieldElement{0x123456789abcdef0, 0x0fedcba987654321, 0x1111111111111111, 0x2222222222222222},
		},
		{
			name: "max_squared",
			arg:  p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
		},
		{
			name: "p256_prime_like_value",
			arg:  p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0x00000000ffffffff, 0x0000000000000000, 0xfffffff00000001},
		},
		{
			name: "p256_one_in_montgomery",
			arg:  p256MontgomeryDomainFieldElement{0x0000000000000001, 0xfffffff00000000, 0xffffffffffffffff, 0x00000000fffffffe},
		},
		{
			name: "alternating_bits",
			arg:  p256MontgomeryDomainFieldElement{0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x5555555555555555, 0xaaaaaaaaaaaaaaaa},
		},
		{
			name: "all_ones_pattern",
			arg:  p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
		},
		{
			name: "checkerboard_pattern",
			arg:  p256MontgomeryDomainFieldElement{0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x3333333333333333, 0xcccccccccccccccc},
		},
		{
			name: "high_bit_set",
			arg:  p256MontgomeryDomainFieldElement{0x8000000000000000, 0x8000000000000000, 0x8000000000000000, 0x8000000000000000},
		},
		{
			name: "all_high_bits",
			arg:  p256MontgomeryDomainFieldElement{0x8000000000000000, 0x8000000000000000, 0x8000000000000000, 0x8000000000000000},
		},
		{
			name: "random_pattern_1",
			arg:  p256MontgomeryDomainFieldElement{0xdeadbeefcafebabe, 0x1234567890abcdef, 0xfedcba0987654321, 0xabcdef0123456789},
		},
		{
			name: "random_pattern_2",
			arg:  p256MontgomeryDomainFieldElement{0x0123456789abcdef, 0xfedcba9876543210, 0x13579bdf2468ace0, 0xfdb97531eca86420},
		},
		{
			name: "random_pattern_3",
			arg:  p256MontgomeryDomainFieldElement{0xf0e1d2c3b4a59687, 0x78695a4b3c2d1e0f, 0x0f1e2d3c4b5a6978, 0x8796a5b4c3d2e1f0},
		},
	}
	b.ResetTimer()
	for _, tc := range testCases {
		var outArm64Style p256MontgomeryDomainFieldElement
		for i := 0; i < b.N; i++ {
			p256SquareArm64Style(&outArm64Style, &tc.arg)
		}
	}
}

func BenchmarkP256SquareGeneric(b *testing.B) {
	testCases := []struct {
		name string
		arg  p256MontgomeryDomainFieldElement
	}{
		// Basic cases
		{
			name: "zero",
			arg:  p256MontgomeryDomainFieldElement{0, 0, 0, 0},
		},
		{
			name: "one",
			arg:  p256MontgomeryDomainFieldElement{1, 0, 0, 0},
		},
		{
			name: "two",
			arg:  p256MontgomeryDomainFieldElement{2, 0, 0, 0},
		},
		{
			name: "max_low",
			arg:  p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0, 0, 0},
		},
		{
			name: "random_large_value",
			arg:  p256MontgomeryDomainFieldElement{0x123456789abcdef0, 0x0fedcba987654321, 0x1111111111111111, 0x2222222222222222},
		},
		{
			name: "max_squared",
			arg:  p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
		},
		{
			name: "p256_prime_like_value",
			arg:  p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0x00000000ffffffff, 0x0000000000000000, 0xfffffff00000001},
		},
		{
			name: "p256_one_in_montgomery",
			arg:  p256MontgomeryDomainFieldElement{0x0000000000000001, 0xfffffff00000000, 0xffffffffffffffff, 0x00000000fffffffe},
		},
		{
			name: "alternating_bits",
			arg:  p256MontgomeryDomainFieldElement{0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x5555555555555555, 0xaaaaaaaaaaaaaaaa},
		},
		{
			name: "all_ones_pattern",
			arg:  p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
		},
		{
			name: "checkerboard_pattern",
			arg:  p256MontgomeryDomainFieldElement{0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x3333333333333333, 0xcccccccccccccccc},
		},
		{
			name: "high_bit_set",
			arg:  p256MontgomeryDomainFieldElement{0x8000000000000000, 0x8000000000000000, 0x8000000000000000, 0x8000000000000000},
		},
		{
			name: "all_high_bits",
			arg:  p256MontgomeryDomainFieldElement{0x8000000000000000, 0x8000000000000000, 0x8000000000000000, 0x8000000000000000},
		},
		{
			name: "random_pattern_1",
			arg:  p256MontgomeryDomainFieldElement{0xdeadbeefcafebabe, 0x1234567890abcdef, 0xfedcba0987654321, 0xabcdef0123456789},
		},
		{
			name: "random_pattern_2",
			arg:  p256MontgomeryDomainFieldElement{0x0123456789abcdef, 0xfedcba9876543210, 0x13579bdf2468ace0, 0xfdb97531eca86420},
		},
		{
			name: "random_pattern_3",
			arg:  p256MontgomeryDomainFieldElement{0xf0e1d2c3b4a59687, 0x78695a4b3c2d1e0f, 0x0f1e2d3c4b5a6978, 0x8796a5b4c3d2e1f0},
		},
	}
	b.ResetTimer()
	for _, tc := range testCases {
		var outArm64Style p256MontgomeryDomainFieldElement
		for i := 0; i < b.N; i++ {
			p256SquareGeneric(&outArm64Style, &tc.arg)
		}
	}
}

func BenchmarkP256Square(b *testing.B) {
	testCases := []struct {
		name string
		arg  p256MontgomeryDomainFieldElement
	}{
		// Basic cases
		{
			name: "zero",
			arg:  p256MontgomeryDomainFieldElement{0, 0, 0, 0},
		},
		{
			name: "one",
			arg:  p256MontgomeryDomainFieldElement{1, 0, 0, 0},
		},
		{
			name: "two",
			arg:  p256MontgomeryDomainFieldElement{2, 0, 0, 0},
		},
		{
			name: "max_low",
			arg:  p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0, 0, 0},
		},
		{
			name: "random_large_value",
			arg:  p256MontgomeryDomainFieldElement{0x123456789abcdef0, 0x0fedcba987654321, 0x1111111111111111, 0x2222222222222222},
		},
		{
			name: "max_squared",
			arg:  p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
		},
		{
			name: "p256_prime_like_value",
			arg:  p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0x00000000ffffffff, 0x0000000000000000, 0xfffffff00000001},
		},
		{
			name: "p256_one_in_montgomery",
			arg:  p256MontgomeryDomainFieldElement{0x0000000000000001, 0xfffffff00000000, 0xffffffffffffffff, 0x00000000fffffffe},
		},
		{
			name: "alternating_bits",
			arg:  p256MontgomeryDomainFieldElement{0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x5555555555555555, 0xaaaaaaaaaaaaaaaa},
		},
		{
			name: "all_ones_pattern",
			arg:  p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
		},
		{
			name: "checkerboard_pattern",
			arg:  p256MontgomeryDomainFieldElement{0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x3333333333333333, 0xcccccccccccccccc},
		},
		{
			name: "high_bit_set",
			arg:  p256MontgomeryDomainFieldElement{0x8000000000000000, 0x8000000000000000, 0x8000000000000000, 0x8000000000000000},
		},
		{
			name: "all_high_bits",
			arg:  p256MontgomeryDomainFieldElement{0x8000000000000000, 0x8000000000000000, 0x8000000000000000, 0x8000000000000000},
		},
		{
			name: "random_pattern_1",
			arg:  p256MontgomeryDomainFieldElement{0xdeadbeefcafebabe, 0x1234567890abcdef, 0xfedcba0987654321, 0xabcdef0123456789},
		},
		{
			name: "random_pattern_2",
			arg:  p256MontgomeryDomainFieldElement{0x0123456789abcdef, 0xfedcba9876543210, 0x13579bdf2468ace0, 0xfdb97531eca86420},
		},
		{
			name: "random_pattern_3",
			arg:  p256MontgomeryDomainFieldElement{0xf0e1d2c3b4a59687, 0x78695a4b3c2d1e0f, 0x0f1e2d3c4b5a6978, 0x8796a5b4c3d2e1f0},
		},
	}
	b.ResetTimer()
	for _, tc := range testCases {
		var outArm64Style p256MontgomeryDomainFieldElement
		for i := 0; i < b.N; i++ {
			p256Square(&outArm64Style, &tc.arg)
		}
	}
}

func BenchmarkP256SquareCDisassemble(b *testing.B) {
	testCases := []struct {
		name string
		arg  p256MontgomeryDomainFieldElement
	}{
		// Basic cases
		{
			name: "zero",
			arg:  p256MontgomeryDomainFieldElement{0, 0, 0, 0},
		},
		{
			name: "one",
			arg:  p256MontgomeryDomainFieldElement{1, 0, 0, 0},
		},
		{
			name: "two",
			arg:  p256MontgomeryDomainFieldElement{2, 0, 0, 0},
		},
		{
			name: "max_low",
			arg:  p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0, 0, 0},
		},
		{
			name: "random_large_value",
			arg:  p256MontgomeryDomainFieldElement{0x123456789abcdef0, 0x0fedcba987654321, 0x1111111111111111, 0x2222222222222222},
		},
		{
			name: "max_squared",
			arg:  p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
		},
		{
			name: "p256_prime_like_value",
			arg:  p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0x00000000ffffffff, 0x0000000000000000, 0xfffffff00000001},
		},
		{
			name: "p256_one_in_montgomery",
			arg:  p256MontgomeryDomainFieldElement{0x0000000000000001, 0xfffffff00000000, 0xffffffffffffffff, 0x00000000fffffffe},
		},
		{
			name: "alternating_bits",
			arg:  p256MontgomeryDomainFieldElement{0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x5555555555555555, 0xaaaaaaaaaaaaaaaa},
		},
		{
			name: "all_ones_pattern",
			arg:  p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
		},
		{
			name: "checkerboard_pattern",
			arg:  p256MontgomeryDomainFieldElement{0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x3333333333333333, 0xcccccccccccccccc},
		},
		{
			name: "high_bit_set",
			arg:  p256MontgomeryDomainFieldElement{0x8000000000000000, 0x8000000000000000, 0x8000000000000000, 0x8000000000000000},
		},
		{
			name: "all_high_bits",
			arg:  p256MontgomeryDomainFieldElement{0x8000000000000000, 0x8000000000000000, 0x8000000000000000, 0x8000000000000000},
		},
		{
			name: "random_pattern_1",
			arg:  p256MontgomeryDomainFieldElement{0xdeadbeefcafebabe, 0x1234567890abcdef, 0xfedcba0987654321, 0xabcdef0123456789},
		},
		{
			name: "random_pattern_2",
			arg:  p256MontgomeryDomainFieldElement{0x0123456789abcdef, 0xfedcba9876543210, 0x13579bdf2468ace0, 0xfdb97531eca86420},
		},
		{
			name: "random_pattern_3",
			arg:  p256MontgomeryDomainFieldElement{0xf0e1d2c3b4a59687, 0x78695a4b3c2d1e0f, 0x0f1e2d3c4b5a6978, 0x8796a5b4c3d2e1f0},
		},
	}
	b.ResetTimer()
	for _, tc := range testCases {
		var outArm64Style p256MontgomeryDomainFieldElement
		for i := 0; i < b.N; i++ {
			p256SquareCDisassemble(&outArm64Style, &tc.arg)
		}
	}
}

func BenchmarkP256Add(b *testing.B) {
	testCases := []struct {
		name string
		arg1 p256MontgomeryDomainFieldElement
		arg2 p256MontgomeryDomainFieldElement
	}{
		// Basic cases
		{
			name: "zero_plus_zero",
			arg1: p256MontgomeryDomainFieldElement{0, 0, 0, 0},
			arg2: p256MontgomeryDomainFieldElement{0, 0, 0, 0},
		},
		{
			name: "one_plus_one",
			arg1: p256MontgomeryDomainFieldElement{1, 0, 0, 0},
			arg2: p256MontgomeryDomainFieldElement{1, 0, 0, 0},
		},
		{
			name: "max_plus_max",
			arg1: p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
			arg2: p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
		},
		{
			name: "random_large_values",
			arg1: p256MontgomeryDomainFieldElement{0x123456789abcdef0, 0x0fedcba987654321, 0x1111111111111111, 0x2222222222222222},
			arg2: p256MontgomeryDomainFieldElement{0xaaaaaaaaaaaaaaaa, 0xbbbbbbbbbbbbbbbb, 0xcccccccccccccccc, 0xdddddddddddddddd},
		},
		{
			name: "p256_prime_like_values",
			arg1: p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0x00000000ffffffff, 0x0000000000000000, 0xffffffff00000001},
			arg2: p256MontgomeryDomainFieldElement{0x0000000000000001, 0xffffffff00000000, 0xffffffffffffffff, 0x00000000fffffffe},
		},
		{
			name: "alternating_bits",
			arg1: p256MontgomeryDomainFieldElement{0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x5555555555555555, 0xaaaaaaaaaaaaaaaa},
			arg2: p256MontgomeryDomainFieldElement{0xaaaaaaaaaaaaaaaa, 0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x5555555555555555},
		},
		{
			name: "random_pattern_1",
			arg1: p256MontgomeryDomainFieldElement{0xdeadbeefcafebabe, 0x1234567890abcdef, 0xfedcba0987654321, 0xabcdef0123456789},
			arg2: p256MontgomeryDomainFieldElement{0x1122334455667788, 0x99aabbccddeeff00, 0x0011223344556677, 0x8899aabbccddeeff},
		},
	}
	b.ResetTimer()
	for _, tc := range testCases {
		var out p256MontgomeryDomainFieldElement
		for i := 0; i < b.N; i++ {
			p256Add(&out, &tc.arg1, &tc.arg2)
		}
	}
}

func BenchmarkP256AddGeneric(b *testing.B) {
	testCases := []struct {
		name string
		arg1 p256MontgomeryDomainFieldElement
		arg2 p256MontgomeryDomainFieldElement
	}{
		// Basic cases
		{
			name: "zero_plus_zero",
			arg1: p256MontgomeryDomainFieldElement{0, 0, 0, 0},
			arg2: p256MontgomeryDomainFieldElement{0, 0, 0, 0},
		},
		{
			name: "one_plus_one",
			arg1: p256MontgomeryDomainFieldElement{1, 0, 0, 0},
			arg2: p256MontgomeryDomainFieldElement{1, 0, 0, 0},
		},
		{
			name: "max_plus_max",
			arg1: p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
			arg2: p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
		},
		{
			name: "random_large_values",
			arg1: p256MontgomeryDomainFieldElement{0x123456789abcdef0, 0x0fedcba987654321, 0x1111111111111111, 0x2222222222222222},
			arg2: p256MontgomeryDomainFieldElement{0xaaaaaaaaaaaaaaaa, 0xbbbbbbbbbbbbbbbb, 0xcccccccccccccccc, 0xdddddddddddddddd},
		},
		{
			name: "p256_prime_like_values",
			arg1: p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0x00000000ffffffff, 0x0000000000000000, 0xffffffff00000001},
			arg2: p256MontgomeryDomainFieldElement{0x0000000000000001, 0xffffffff00000000, 0xffffffffffffffff, 0x00000000fffffffe},
		},
		{
			name: "alternating_bits",
			arg1: p256MontgomeryDomainFieldElement{0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x5555555555555555, 0xaaaaaaaaaaaaaaaa},
			arg2: p256MontgomeryDomainFieldElement{0xaaaaaaaaaaaaaaaa, 0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x5555555555555555},
		},
		{
			name: "random_pattern_1",
			arg1: p256MontgomeryDomainFieldElement{0xdeadbeefcafebabe, 0x1234567890abcdef, 0xfedcba0987654321, 0xabcdef0123456789},
			arg2: p256MontgomeryDomainFieldElement{0x1122334455667788, 0x99aabbccddeeff00, 0x0011223344556677, 0x8899aabbccddeeff},
		},
	}
	b.ResetTimer()
	for _, tc := range testCases {
		var out p256MontgomeryDomainFieldElement
		for i := 0; i < b.N; i++ {
			p256AddGeneric(&out, &tc.arg1, &tc.arg2)
		}
	}
}

func BenchmarkP256Sub(b *testing.B) {
	testCases := []struct {
		name string
		arg1 p256MontgomeryDomainFieldElement
		arg2 p256MontgomeryDomainFieldElement
	}{
		// Basic cases
		{
			name: "zero_minus_zero",
			arg1: p256MontgomeryDomainFieldElement{0, 0, 0, 0},
			arg2: p256MontgomeryDomainFieldElement{0, 0, 0, 0},
		},
		{
			name: "one_minus_one",
			arg1: p256MontgomeryDomainFieldElement{1, 0, 0, 0},
			arg2: p256MontgomeryDomainFieldElement{1, 0, 0, 0},
		},
		{
			name: "max_minus_max",
			arg1: p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
			arg2: p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
		},
		{
			name: "zero_minus_max",
			arg1: p256MontgomeryDomainFieldElement{0, 0, 0, 0},
			arg2: p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
		},
		{
			name: "random_large_values",
			arg1: p256MontgomeryDomainFieldElement{0x123456789abcdef0, 0x0fedcba987654321, 0x1111111111111111, 0x2222222222222222},
			arg2: p256MontgomeryDomainFieldElement{0xaaaaaaaaaaaaaaaa, 0xbbbbbbbbbbbbbbbb, 0xcccccccccccccccc, 0xdddddddddddddddd},
		},
		{
			name: "p256_prime_like_values",
			arg1: p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0x00000000ffffffff, 0x0000000000000000, 0xffffffff00000001},
			arg2: p256MontgomeryDomainFieldElement{0x0000000000000001, 0xffffffff00000000, 0xffffffffffffffff, 0x00000000fffffffe},
		},
		{
			name: "alternating_bits",
			arg1: p256MontgomeryDomainFieldElement{0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x5555555555555555, 0xaaaaaaaaaaaaaaaa},
			arg2: p256MontgomeryDomainFieldElement{0xaaaaaaaaaaaaaaaa, 0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x5555555555555555},
		},
		{
			name: "random_pattern_1",
			arg1: p256MontgomeryDomainFieldElement{0xdeadbeefcafebabe, 0x1234567890abcdef, 0xfedcba0987654321, 0xabcdef0123456789},
			arg2: p256MontgomeryDomainFieldElement{0x1122334455667788, 0x99aabbccddeeff00, 0x0011223344556677, 0x8899aabbccddeeff},
		},
	}
	b.ResetTimer()
	for _, tc := range testCases {
		var out p256MontgomeryDomainFieldElement
		for i := 0; i < b.N; i++ {
			p256Sub(&out, &tc.arg1, &tc.arg2)
		}
	}
}

func BenchmarkP256SubGeneric(b *testing.B) {
	testCases := []struct {
		name string
		arg1 p256MontgomeryDomainFieldElement
		arg2 p256MontgomeryDomainFieldElement
	}{
		// Basic cases
		{
			name: "zero_minus_zero",
			arg1: p256MontgomeryDomainFieldElement{0, 0, 0, 0},
			arg2: p256MontgomeryDomainFieldElement{0, 0, 0, 0},
		},
		{
			name: "one_minus_one",
			arg1: p256MontgomeryDomainFieldElement{1, 0, 0, 0},
			arg2: p256MontgomeryDomainFieldElement{1, 0, 0, 0},
		},
		{
			name: "max_minus_max",
			arg1: p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
			arg2: p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
		},
		{
			name: "zero_minus_max",
			arg1: p256MontgomeryDomainFieldElement{0, 0, 0, 0},
			arg2: p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff},
		},
		{
			name: "random_large_values",
			arg1: p256MontgomeryDomainFieldElement{0x123456789abcdef0, 0x0fedcba987654321, 0x1111111111111111, 0x2222222222222222},
			arg2: p256MontgomeryDomainFieldElement{0xaaaaaaaaaaaaaaaa, 0xbbbbbbbbbbbbbbbb, 0xcccccccccccccccc, 0xdddddddddddddddd},
		},
		{
			name: "p256_prime_like_values",
			arg1: p256MontgomeryDomainFieldElement{0xffffffffffffffff, 0x00000000ffffffff, 0x0000000000000000, 0xffffffff00000001},
			arg2: p256MontgomeryDomainFieldElement{0x0000000000000001, 0xffffffff00000000, 0xffffffffffffffff, 0x00000000fffffffe},
		},
		{
			name: "alternating_bits",
			arg1: p256MontgomeryDomainFieldElement{0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x5555555555555555, 0xaaaaaaaaaaaaaaaa},
			arg2: p256MontgomeryDomainFieldElement{0xaaaaaaaaaaaaaaaa, 0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x5555555555555555},
		},
		{
			name: "random_pattern_1",
			arg1: p256MontgomeryDomainFieldElement{0xdeadbeefcafebabe, 0x1234567890abcdef, 0xfedcba0987654321, 0xabcdef0123456789},
			arg2: p256MontgomeryDomainFieldElement{0x1122334455667788, 0x99aabbccddeeff00, 0x0011223344556677, 0x8899aabbccddeeff},
		},
	}
	b.ResetTimer()
	for _, tc := range testCases {
		var out p256MontgomeryDomainFieldElement
		for i := 0; i < b.N; i++ {
			p256SubGeneric(&out, &tc.arg1, &tc.arg2)
		}
	}
}
