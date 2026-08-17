// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build riscv64 && !purego

package base64

import (
	"internal/cpu"
	"unsafe"
)

// Offsets into internal/cpu records for use in assembly.
const offsetRISCV64HasZbb = unsafe.Offsetof(cpu.RISCV64.HasZbb)

//go:noescape
func encodeChunk(encode *[64]byte, dst, src []byte, n int)
