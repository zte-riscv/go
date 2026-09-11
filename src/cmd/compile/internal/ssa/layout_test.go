// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import (
	"cmd/compile/internal/types"
	"testing"
)

func TestLayoutOrderWithPredictedBranch(t *testing.T) {
	c := testConfig(t)
	fun := c.Fun("b1",
		Bloc("b1",
			Valu("mem", OpInitMem, types.TypeMem, 0, nil),
			Valu("b", OpConstBool, types.Types[types.TBOOL], 0, nil),
			If("b", "b2", "b3")),
		Bloc("b2",
			Goto("b5")),
		Bloc("b3",
			Goto("b4")),
		Bloc("b4",
			Goto("b5")),
		Bloc("b5",
			Exit("mem")),
	)
	fun.blocks["b1"].Likely = BranchUnlikely

	// Func:
	//    b1
	//   /  \
	//  b2  b3(likely)
	//  \    \
	//   \   b4
	//    \  /
	//     b5
	// we expect the blocks to be in the order of b1->b3->b4->b5.
	expectedOrder := [5]ID{1, 3, 4, 5, 2}

	CheckFunc(fun.f)
	layout(fun.f)
	CheckFunc(fun.f)

	for i, b := range fun.f.Blocks {
		if b.ID != expectedOrder[i] {
			t.Errorf("block layout order want %d, got %d", expectedOrder[i], b.ID)
		}
	}
}
