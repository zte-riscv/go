// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.greenteagc

// Export guts for testing.

package runtime

// ExtractHeapBitsSmall is exported for tests so that the riscv64
// assembly implementation can be validated from the runtime_test
// package.
var ExtractHeapBitsSmall = extractHeapBitsSmall
