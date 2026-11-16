// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	. "runtime"
	"testing"
)

// TestGCRatioBounds verifies that GOGCRATIO is clamped to the safe range
// [1, 38] percent. The upper bound guarantees dedicated mark workers stay
// strictly below half of the Ps even with the pacer's 30% rounding error
// (maxUtilError), so the GC CPU limiter can always pull GC utilization back
// under its 50% threshold (see maxGCRatio).
func TestGCRatioBounds(t *testing.T) {
	// runtime.gogetenv reads the runtime's own envs snapshot (see
	// goenvs_unix), not the process environ, so drive it via SetEnvs
	// rather than t.Setenv/syscall.Setenv.
	orig := Envs()
	defer SetEnvs(orig)
	setEnv := func(kv string) {
		envs := make([]string, 0, len(orig)+1)
		for _, e := range orig {
			if len(e) < len(kv) || e[:len(kv)] != kv {
				envs = append(envs, e)
			}
		}
		SetEnvs(append(envs, kv))
	}

	testCases := []struct {
		name string
		env  string
		want float64
	}{
		{name: "Default", env: "", want: GCRatioDefault},
		{name: "BelowMin", env: "GOGCRATIO=0", want: 0.01},
		{name: "Min", env: "GOGCRATIO=1", want: 0.01},
		{name: "Mid", env: "GOGCRATIO=25", want: 0.25},
		{name: "Max", env: "GOGCRATIO=38", want: GCRatioMax},
		{name: "AboveMax", env: "GOGCRATIO=39", want: GCRatioMax},
		{name: "High", env: "GOGCRATIO=99", want: GCRatioMax},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.env == "" {
				SetEnvs(orig)
			} else {
				setEnv(tc.env)
			}
			if got := ReadGOGCRATIO(); got != tc.want {
				t.Errorf("ReadGOGCRATIO() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestGCRatioMaxBound anchors the safety invariant behind the GOGCRATIO
// upper clamp: with the pacer's rounding tolerance (maxUtilError = 30%),
// dedicated mark workers use at most 1.3*gcRatio of the Ps, and that must
// stay strictly below half of the Ps so the GC CPU limiter can always pull
// utilization back under its 50% threshold.
func TestGCRatioMaxBound(t *testing.T) {
	if got := 1.3 * GCRatioMax; !(got < 0.5) {
		t.Errorf("1.3*maxGCRatio = %v, want strictly < 0.5", got)
	}
}
