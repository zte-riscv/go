// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !purego && riscv64

package sha256

// MessageExpansion expands a 64-byte message block into 64 32-bit words.
// This is the message schedule (W[0] to W[63]) used in SHA-256.
//
// Input: p - 64 bytes of message data (big-endian)
// Output: w - 64 uint32 words (written to the provided array)
//
// This function will be implemented in assembly using RISC-V vector instructions.
//
//go:noescape
func messageExpansionRISCV64(p []byte, w *[64]uint32)
