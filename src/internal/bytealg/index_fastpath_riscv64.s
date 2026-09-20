// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// func hasMisalignedFastBuild() bool
TEXT ·hasMisalignedFastBuild<ABIInternal>(SB), NOSPLIT, $0-0
#ifdef GORISCV64EXT_misaligned_fast
	MOV	$1, X10
#else
	MOV	$0, X10
#endif
	RET
