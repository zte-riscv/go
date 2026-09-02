// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
package runtime_test

import (
	"internal/goexperiment"
	"runtime"
	"testing"
)

func TestClearSpan(t *testing.T) {
	needzero, cleared := runtime.TryClearSpanFunc()
	if goexperiment.ClearSpan {
		if needzero != 0 || !cleared {
			t.Fatalf("ClearSpan enabled: expected needzero=0 and cleared=true, got needzero=%d cleared=%v", needzero, cleared)
		}
	} else if needzero == 0 || cleared {
		t.Fatalf("ClearSpan disabled: expected needzero!=0 and cleared=false, got needzero=%d cleared=%v", needzero, cleared)
	}
}
