// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytealg

func hasMisalignedFastBuild() bool

// UseFastIndex reports whether callers should use bytealg.Index/IndexString on RISC-V.
func UseFastIndex() bool {
	return hasMisalignedFastBuild()
}
