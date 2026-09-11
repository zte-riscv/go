// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"errors"
	"fmt"
	"testing"
)

const (
	bpChainDepth = 32
	bpMWDepth    = 16
)

var (
	bpErrBoom = errors.New("boom")

	bpSink     error
	bpSinkHits int
)

func bpFailAt(v, errRate int) bool {
	return (v*31+7)%100 < errRate
}

//go:noinline
func bpStep(i, iter, errRate int) error {
	if bpFailAt(i+iter, errRate) {
		return bpErrBoom
	}
	return nil
}

func bpWorkChain(depth, iter, errRate int) error {
	for i := 0; i < depth; i++ {
		if err := bpStep(i, iter, errRate); err != nil {
			return fmt.Errorf("bpStep %d: %w", i, err)
		}
	}
	return nil
}

func BenchmarkErrorCheckChain(b *testing.B) {
	for _, errRate := range []int{0, 10, 50, 100} {
		b.Run(fmt.Sprintf("errrate-%d", errRate), func(b *testing.B) {
			b.ReportAllocs()
			for iter := 0; iter < b.N; iter++ {
				bpSink = bpWorkChain(bpChainDepth, iter, errRate)
			}
		})
	}
}

//go:noinline
func bpMWTerminal(iter, errRate int) error {
	if bpFailAt(iter, errRate) {
		return bpErrBoom
	}
	return nil
}

func bpMWChain(level, iter, errRate int) error {
	if level == 0 {
		return bpMWTerminal(iter, errRate)
	}
	if err := bpMWChain(level-1, iter, errRate); err != nil {
		return fmt.Errorf("level %d: %w", level, err)
	}
	return nil
}

func BenchmarkMiddlewareChain(b *testing.B) {
	for _, errRate := range []int{0, 10, 50, 100} {
		b.Run(fmt.Sprintf("errrate-%d", errRate), func(b *testing.B) {
			b.ReportAllocs()
			for iter := 0; iter < b.N; iter++ {
				bpSink = bpMWChain(bpMWDepth, iter, errRate)
			}
		})
	}
}

//go:noinline
func bpStepBaseline(i, iter, rate int) int {
	if bpFailAt(i+iter, rate) {
		return 1
	}
	return 0
}

func bpWorkBaseline(depth, iter, rate int) int {
	hits := 0
	for i := 0; i < depth; i++ {
		if h := bpStepBaseline(i, iter, rate); h != 0 {
			hits++
		}
	}
	return hits
}

func BenchmarkNoErrorBaseline(b *testing.B) {
	for _, rate := range []int{0, 50} {
		b.Run(fmt.Sprintf("rate-%d", rate), func(b *testing.B) {
			b.ReportAllocs()
			for iter := 0; iter < b.N; iter++ {
				bpSinkHits = bpWorkBaseline(bpChainDepth, iter, rate)
			}
		})
	}
}
