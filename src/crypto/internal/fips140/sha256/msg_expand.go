// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sha256

import "math/bits"

// MessageExpansion expands a 64-byte message block into 64 32-bit words.
// This is the message schedule (W[0] to W[63]) used in SHA-256.
//
// Input: p - 64 bytes of message data (big-endian)
// Output: w - 64 uint32 words (written to the provided array)
func messageExpansionGeneric(p []byte, w *[64]uint32) {
	// W[0-15]: Load directly from message (big-endian)
	for i := 0; i < 16; i++ {
		j := i * 4
		w[i] = uint32(p[j])<<24 | uint32(p[j+1])<<16 | uint32(p[j+2])<<8 | uint32(p[j+3])
	}

	// W[16-63]: Expand using SHA-256 message schedule formula
	// W[i] = σ1(W[i-2]) + W[i-7] + σ0(W[i-15]) + W[i-16]
	// σ0(x) = ROTR(7,x) XOR ROTR(18,x) XOR SHR(3,x)
	// σ1(x) = ROTR(17,x) XOR ROTR(19,x) XOR SHR(10,x)
	for i := 16; i < 64; i++ {
		v1 := w[i-2]
		t1 := (bits.RotateLeft32(v1, -17)) ^ (bits.RotateLeft32(v1, -19)) ^ (v1 >> 10) // σ1(W[i-2])

		v2 := w[i-15]
		t2 := (bits.RotateLeft32(v2, -7)) ^ (bits.RotateLeft32(v2, -18)) ^ (v2 >> 3) // σ0(W[i-15])

		w[i] = t1 + w[i-7] + t2 + w[i-16]
	}
}
