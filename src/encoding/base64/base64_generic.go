// Copyright 2026 Sun Yimin. All rights reserved.
// Use of this source code is governed by a BSD 3-Clause-style
// license that can be found in the LICENSE file.

//go:build !riscv64 || purego

package base64

func encode(enc *Encoding, dst, src []byte) {
	encodeGeneric(enc, dst, src)
}

func decode(enc *Encoding, dst, src []byte) (int, error) {
	return decodeGeneric(enc, dst, src)
}