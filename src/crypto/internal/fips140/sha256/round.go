// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sha256

import "math/bits"

// BlockGenericWithTrace performs SHA-256 block compression similar to blockGeneric,
// but also saves the intermediate hash state (a, b, c, d, e, f, g, h) for each of the 64 rounds.
//
// Input:
//
//	dig *Digest - the digest to update
//	p []byte - the message block (64 bytes)
//
// Output:
//
//	temp_dig *[64][8]uint32 - array to store intermediate states for each round
//	         temp_dig[i] contains [a, b, c, d, e, f, g, h] after round i
//	temp_kword *[64]uint32 - array to store w[i]+_K[i] for each round before computation
//	         temp_kword[i] contains w[i]+_K[i] for round i
func blockGenericWithTrace(dig *Digest, p []byte, temp_dig *[64][8]uint32, temp_kword *[64]uint32) {
	var w [64]uint32
	h0, h1, h2, h3, h4, h5, h6, h7 := dig.h[0], dig.h[1], dig.h[2], dig.h[3], dig.h[4], dig.h[5], dig.h[6], dig.h[7]

	// Process one 64-byte chunk
	if len(p) >= chunk {
		// Expand message schedule W[0..15]
		for i := 0; i < 16; i++ {
			j := i * 4
			w[i] = uint32(p[j])<<24 | uint32(p[j+1])<<16 | uint32(p[j+2])<<8 | uint32(p[j+3])
		}

		// Expand message schedule W[16..63]
		for i := 16; i < 64; i++ {
			v1 := w[i-2]
			t1 := (bits.RotateLeft32(v1, -17)) ^ (bits.RotateLeft32(v1, -19)) ^ (v1 >> 10)
			v2 := w[i-15]
			t2 := (bits.RotateLeft32(v2, -7)) ^ (bits.RotateLeft32(v2, -18)) ^ (v2 >> 3)
			w[i] = t1 + w[i-7] + t2 + w[i-16]
		}

		// Initialize working variables
		a, b, c, d, e, f, g, h := h0, h1, h2, h3, h4, h5, h6, h7

		// Main compression loop - 64 rounds
		for i := 0; i < 64; i++ {
			// Save w[i]+_K[i] before round computation
			temp_kword[i] = w[i] + _K[i]

			// Save state before round computation
			// temp_dig[i][0] = a
			// temp_dig[i][1] = b
			// temp_dig[i][2] = c
			// temp_dig[i][3] = d
			// temp_dig[i][4] = e
			// temp_dig[i][5] = f
			// temp_dig[i][6] = g
			// temp_dig[i][7] = h

			// Round function
			t1 := h + ((bits.RotateLeft32(e, -6)) ^ (bits.RotateLeft32(e, -11)) ^ (bits.RotateLeft32(e, -25))) + ((e & f) ^ (^e & g)) + _K[i] + w[i]
			t2 := ((bits.RotateLeft32(a, -2)) ^ (bits.RotateLeft32(a, -13)) ^ (bits.RotateLeft32(a, -22))) + ((a & b) ^ (a & c) ^ (b & c))

			h = g
			g = f
			f = e
			e = d + t1
			d = c
			c = b
			b = a
			a = t1 + t2

			// Save state after round
			temp_dig[i][0] = a
			temp_dig[i][1] = b
			temp_dig[i][2] = c
			temp_dig[i][3] = d
			temp_dig[i][4] = e
			temp_dig[i][5] = f
			temp_dig[i][6] = g
			temp_dig[i][7] = h
		}

		// Add compressed chunk to hash
		h0 += a
		h1 += b
		h2 += c
		h3 += d
		h4 += e
		h5 += f
		h6 += g
		h7 += h
	}

	// Update digest
	dig.h[0], dig.h[1], dig.h[2], dig.h[3], dig.h[4], dig.h[5], dig.h[6], dig.h[7] = h0, h1, h2, h3, h4, h5, h6, h7
}
