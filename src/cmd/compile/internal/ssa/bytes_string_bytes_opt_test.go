// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa_test

import (
	"internal/testenv"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestBytesStringBytesOptCorrectness compiles and runs a test program
// with -bytesstringbytesopt enabled to verify the optimization produces
// correct results at runtime, not just correct assembly output.
func TestBytesStringBytesOptCorrectness(t *testing.T) {
	if runtime.GOARCH != "riscv64" {
		t.Skipf("optimization only applies to riscv64, got %s", runtime.GOARCH)
	}
	testenv.MustHaveGoBuild(t)

	tmpdir := t.TempDir()
	source := filepath.Join("testdata", "bytes_string_bytes.go")
	output := filepath.Join(tmpdir, "bytes_string_bytes.exe")

	build := testenv.Command(t, testenv.GoToolPath(t), "build", "-o", output, "-gcflags=-bytesstringbytesopt", source)
	buildOut, err := build.CombinedOutput()
	if err != nil {
		t.Fatalf("build with -bytesstringbytesopt failed: %v\n%s", err, buildOut)
	}

	run := testenv.Command(t, output)
	runOut, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("test program failed: %v\n%s", err, runOut)
	}
	if !strings.Contains(string(runOut), "PASS") {
		t.Fatalf("test program did not print PASS, output:\n%s", runOut)
	}
}
