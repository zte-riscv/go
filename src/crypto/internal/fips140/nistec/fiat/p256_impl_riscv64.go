// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build riscv64 && !purego

package fiat

//go:noescape
func p256Mul(out, a, b *p256MontgomeryDomainFieldElement)

//go:noescape
func p256Square(out1 *p256MontgomeryDomainFieldElement, arg1 *p256MontgomeryDomainFieldElement)

//go:noescape
func p256SquareOld(out1 *p256MontgomeryDomainFieldElement, arg1 *p256MontgomeryDomainFieldElement)

//go:noescape
func p256Add(out1 *p256MontgomeryDomainFieldElement, arg1 *p256MontgomeryDomainFieldElement, arg2 *p256MontgomeryDomainFieldElement)

//go:noescape
func p256Sub(out1 *p256MontgomeryDomainFieldElement, arg1 *p256MontgomeryDomainFieldElement, arg2 *p256MontgomeryDomainFieldElement)
