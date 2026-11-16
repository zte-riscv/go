// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"runtime/debug"
	"testing"
)

// gcRatioNode is a heap-allocated object with a pointer-heavy layout: a
// slice of pointers plus a slice of words. A large live set of these
// objects, combined with a low GOGC, keeps the concurrent GC marking in
// almost continuous operation, so throughput is dominated by how many
// background mark workers the pacer dedicates (GOGCRATIO * GOMAXPROCS).
// See runtime/mgcpacer.go readGOGCRATIO.
type gcRatioNode struct {
	ptrs []*gcRatioNode
	data []uint64
	sum  uint64
}

func newGCRatioNode(ptrs, data int) *gcRatioNode {
	return &gcRatioNode{
		ptrs: make([]*gcRatioNode, ptrs),
		data: make([]uint64, data),
	}
}

// linkGCRatioNodes connects every node to a few of its successors so that
// GC marking has to follow real pointer chains rather than only scanning
// the pool roots.
func linkGCRatioNodes(pool []*gcRatioNode) {
	for i, n := range pool {
		for j := range n.ptrs {
			n.ptrs[j] = pool[(i+j+1)%len(pool)]
		}
	}
}

// runGCBackground measures an allocation-heavy workload over a large
// pointer-heavy live set in which objects are continuously replaced, so
// GC runs frequently. Whether marking is the bottleneck depends on GOGC:
// at the default 100 most of each cycle goes to the mutator and the
// number of background mark workers barely matters, while at low GOGC
// marking dominates and throughput is decided by how many workers the
// pacer dedicates (GOGCRATIO * GOMAXPROCS). Three wrappers below sample
// three GC-pressure regimes so the benchmark captures both the regular
// case (no regression expected) and the GC-bound case (large effect).
//
// Sub-benchmarks vary the live-set size and pointer density. Run with:
//
//	GOGCRATIO=25 GOMAXPROCS=N go test -run '^$' -bench BenchmarkGCBackground
func runGCBackground(b *testing.B, gogc int) {
	debug.SetGCPercent(gogc)

	for _, tc := range []struct {
		name string
		live int
		ptrs int
		data int
	}{
		{"gcbound", 32768, 32, 64},
		{"gcbound-large", 65536, 32, 64},
		{"gcbound-dense", 32768, 64, 32},
	} {
		b.Run(tc.name, func(b *testing.B) {
			pool := make([]*gcRatioNode, tc.live)
			for i := range pool {
				pool[i] = newGCRatioNode(tc.ptrs, tc.data)
			}
			linkGCRatioNodes(pool)
			idx := 0
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				pool[idx] = newGCRatioNode(tc.ptrs, tc.data)
				idx = (idx + 1) % tc.live
			}
		})
	}
}

// BenchmarkGCBackgroundDefault runs at GOGC=100, the regime of most real
// workloads: marking is far from the bottleneck, so GOGCRATIO should have
// a negligible effect. Guards against regressions in the common case.
func BenchmarkGCBackgroundDefault(b *testing.B) {
	runGCBackground(b, 100)
}

// BenchmarkGCBackgroundModerate runs at GOGC=20, an intermediate GC
// pressure where marking starts to consume a visible share of the CPU.
func BenchmarkGCBackgroundModerate(b *testing.B) {
	runGCBackground(b, 20)
}

// BenchmarkGCBackgroundDense runs at GOGC=5, where marking is the
// bottleneck and the number of background mark workers (governed by
// GOGCRATIO) directly decides how much of the available CPU goes to
// marking. Throughput is expected to vary strongly with the setting.
func BenchmarkGCBackgroundDense(b *testing.B) {
	runGCBackground(b, 5)
}
