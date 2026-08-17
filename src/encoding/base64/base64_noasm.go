// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !riscv64 || purego

package base64

func encodeChunk(encode *[64]byte, dst, src []byte, n int) {
	si := 0
	di := 0
	for si < n {
		// Convert 3x 8bit source bytes into 4 bytes
		val := uint(src[si+0])<<16 | uint(src[si+1])<<8 | uint(src[si+2])

		dst[di+0] = encode[val>>18&0x3F]
		dst[di+1] = encode[val>>12&0x3F]
		dst[di+2] = encode[val>>6&0x3F]
		dst[di+3] = encode[val&0x3F]

		si += 3
		di += 4
	}
}
