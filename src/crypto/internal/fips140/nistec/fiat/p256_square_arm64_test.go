// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build riscv64

package fiat

import (
	"testing"
)

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
