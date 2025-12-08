// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !purego && riscv64

package sha256

// BlockRISCV64WithTrace performs SHA-256 block compression similar to blockGeneric,
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
//
// This function will be implemented in assembly using RISC-V vector instructions.
// For now, it falls back to the generic implementation.
func blockRISCV64WithTrace(dig *Digest, p []byte, temp_dig *[64][8]uint32, temp_kword *[64]uint32)
