// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reflect_test

import (
	. "reflect"
	"testing"
)

// BenchmarkFuncOfConcurrent benchmarks FuncOf called concurrently.
func BenchmarkFuncOfConcurrent(b *testing.B) {
	// Warm up the cache so all timed calls hit the fast path.
	FuncOf([]Type{TypeOf(0)}, []Type{TypeOf(0)}, false)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = FuncOf([]Type{TypeOf(0)}, []Type{TypeOf(0)}, false)
		}
	})
}

// BenchmarkFuncOfSerial benchmarks FuncOf called sequentially.
func BenchmarkFuncOfSerial(b *testing.B) {
	FuncOf([]Type{TypeOf(0)}, []Type{TypeOf(0)}, false) // warm up cache
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FuncOf([]Type{TypeOf(0)}, []Type{TypeOf(0)}, false)
	}
}
