// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package buildcfg

import (
	"os"
	"strings"
	"testing"
)

func TestConfigFlags(t *testing.T) {
	os.Setenv("GOAMD64", "v1")
	if goamd64() != 1 {
		t.Errorf("Wrong parsing of GOAMD64=v1")
	}
	os.Setenv("GOAMD64", "v4")
	if goamd64() != 4 {
		t.Errorf("Wrong parsing of GOAMD64=v4")
	}
	Error = nil
	os.Setenv("GOAMD64", "1")
	if goamd64(); Error == nil {
		t.Errorf("Wrong parsing of GOAMD64=1")
	}

	os.Setenv("GORISCV64", "rva20u64")
	if goriscv64() != 20 {
		t.Errorf("Wrong parsing of RISCV64=rva20u64")
	}
	os.Setenv("GORISCV64", "rva22u64")
	if goriscv64() != 22 {
		t.Errorf("Wrong parsing of RISCV64=rva22u64")
	}
	os.Setenv("GORISCV64", "rva23u64")
	if goriscv64() != 23 {
		t.Errorf("Wrong parsing of RISCV64=rva23u64")
	}
	Error = nil
	os.Setenv("GORISCV64", "rva22")
	if _ = goriscv64(); Error == nil {
		t.Errorf("Wrong parsing of RISCV64=rva22")
	}
	Error = nil
	os.Setenv("GOARM64", "v7.0")
	if _ = goarm64(); Error == nil {
		t.Errorf("Wrong parsing of GOARM64=7.0")
	}
	Error = nil
	os.Setenv("GOARM64", "8.0")
	if _ = goarm64(); Error == nil {
		t.Errorf("Wrong parsing of GOARM64=8.0")
	}
	Error = nil
	os.Setenv("GOARM64", "v8.0,lsb")
	if _ = goarm64(); Error == nil {
		t.Errorf("Wrong parsing of GOARM64=v8.0,lsb")
	}
	os.Setenv("GOARM64", "v8.0,lse")
	if goarm64().Version != "v8.0" || goarm64().LSE != true || goarm64().Crypto != false {
		t.Errorf("Wrong parsing of GOARM64=v8.0,lse")
	}
	os.Setenv("GOARM64", "v8.0,crypto")
	if goarm64().Version != "v8.0" || goarm64().LSE != false || goarm64().Crypto != true {
		t.Errorf("Wrong parsing of GOARM64=v8.0,crypto")
	}
	os.Setenv("GOARM64", "v8.0,crypto,lse")
	if goarm64().Version != "v8.0" || goarm64().LSE != true || goarm64().Crypto != true {
		t.Errorf("Wrong parsing of GOARM64=v8.0,crypto,lse")
	}
	os.Setenv("GOARM64", "v8.0,lse,crypto")
	if goarm64().Version != "v8.0" || goarm64().LSE != true || goarm64().Crypto != true {
		t.Errorf("Wrong parsing of GOARM64=v8.0,lse,crypto")
	}
	os.Setenv("GOARM64", "v9.0")
	if goarm64().Version != "v9.0" || goarm64().LSE != true || goarm64().Crypto != false {
		t.Errorf("Wrong parsing of GOARM64=v9.0")
	}
}

func TestParseGORISCV64(t *testing.T) {
	tests := []struct {
		in      string
		profile string
		level   int
		exts    Goriscv64Extensions
	}{
		{"rva20u64", "rva20u64", 20, Goriscv64Extensions{}},
		{"rva22u64", "rva22u64", 22, Goriscv64Extensions{}},
		{"rva23u64", "rva23u64", 23, Goriscv64Extensions{}},
		{"rva23u64,zacas", "rva23u64", 23, Goriscv64Extensions{Zacas: true}},
		{"rva23u64,zacas,zabha", "rva23u64", 23, Goriscv64Extensions{Zacas: true, Zabha: true}},
		{"rva23u64,misaligned_fast", "rva23u64", 23, Goriscv64Extensions{MisalignedFast: true}},
		{"rva23u64,ZACAS,ZABHA", "rva23u64", 23, Goriscv64Extensions{Zacas: true, Zabha: true}},
		{"rva23u64, ZaCaS , zAbHa ", "rva23u64", 23, Goriscv64Extensions{Zacas: true, Zabha: true}},
		{"rva23u64,", "rva23u64", 23, Goriscv64Extensions{}},
	}
	for _, tt := range tests {
		got, err := ParseGORISCV64(tt.in)
		if err != nil {
			t.Errorf("ParseGORISCV64(%q) returned error: %v", tt.in, err)
			continue
		}
		if got.Profile != tt.profile || got.Level != tt.level || got.Exts != tt.exts {
			t.Errorf("ParseGORISCV64(%q) = %#v, want profile %q, level %d, exts %#v",
				tt.in, got, tt.profile, tt.level, tt.exts)
		}
	}
}

func TestParseGORISCV64Invalid(t *testing.T) {
	// An empty string is rejected by the parser itself, though envOr treats
	// an empty GORISCV64 as unset and substitutes DefaultGORISCV64.
	if _, err := ParseGORISCV64(""); err == nil {
		t.Errorf("ParseGORISCV64(%q) = nil error, want error", "")
	}

	// Profiles outside the supported rva20/rva22/rva23 whitelist must be
	// rejected by the parser itself, not by a later goriscv64 switch.
	invalidProfiles := []string{
		"rva",
		"rva19u64",
		"rva21u64",
		"rva24u64",
		"rva99u64",
		"rva2xu64",
		"rvb23u64",
		// Both the profile and the extension are invalid; the profile error
		// is reported first.
		"rva19u64,foo",
	}
	for _, in := range invalidProfiles {
		if _, err := ParseGORISCV64(in); err == nil {
			t.Errorf("ParseGORISCV64(%q) = nil error, want error", in)
		}
	}

	// Invalid extensions, alone and combined with a valid or invalid profile.
	invalidExtensions := []string{
		"rva23u64,foo",
		"rva23u64,zacas,foo",
	}
	for _, in := range invalidExtensions {
		_, err := ParseGORISCV64(in)
		if err == nil {
			t.Errorf("ParseGORISCV64(%q) = nil error, want error", in)
			continue
		}
		if !strings.Contains(err.Error(), "invalid GORISCV64 extension") {
			t.Errorf("ParseGORISCV64(%q) error = %v, want extension error", in, err)
		}
	}

	// A bad profile must be reported as an error by both entry points,
	// rather than being silently accepted by goriscv64Extensions.
	for _, in := range append(invalidProfiles, invalidExtensions...) {
		Error = nil
		t.Setenv("GORISCV64", in)
		if _ = goriscv64(); Error == nil {
			t.Errorf("goriscv64() with GORISCV64=%q: Error = nil, want error", in)
		}
		Error = nil
		if ext := goriscv64Extensions(); ext != (Goriscv64Extensions{}) {
			t.Errorf("goriscv64Extensions() with GORISCV64=%q = %#v, want zero value", in, ext)
		}
		if Error == nil {
			t.Errorf("goriscv64Extensions() with GORISCV64=%q: Error = nil, want error", in)
		}
	}
}

func TestGoriscv64ExtensionsHas(t *testing.T) {
	ext := Goriscv64Extensions{Zacas: true, Zabha: false}
	if !ext.Has(Riscv64ExtZacas) {
		t.Errorf("Has(%q) = false, want true", Riscv64ExtZacas)
	}
	if ext.Has(Riscv64ExtZabha) {
		t.Errorf("Has(%q) = true, want false", Riscv64ExtZabha)
	}
	if ext.Has("nonexistent") {
		t.Errorf("Has(%q) = true, want false", "nonexistent")
	}
}

func TestGoriscv64Extensions(t *testing.T) {
	t.Setenv("GORISCV64", "rva23u64,zacas")
	ext := goriscv64Extensions()
	want := Goriscv64Extensions{Zacas: true, Zabha: false}
	if ext != want {
		t.Fatalf("extensions = %#v, want %#v", ext, want)
	}
}

func TestGoriscv64ExtensionsInvalid(t *testing.T) {
	t.Setenv("GORISCV64", "rva23u64,foo")
	ext := goriscv64Extensions()
	want := Goriscv64Extensions{Zacas: false, Zabha: false}
	if ext != want {
		t.Fatalf("extensions = %#v, want %#v on invalid extension", ext, want)
	}
}

func TestGoarm64FeaturesSupports(t *testing.T) {
	g, _ := ParseGoarm64("v9.3")

	if !g.Supports("v9.3") {
		t.Errorf("Wrong goarm64Features.Supports for v9.3, v9.3")
	}

	if g.Supports("v9.4") {
		t.Errorf("Wrong goarm64Features.Supports for v9.3, v9.4")
	}

	if !g.Supports("v8.8") {
		t.Errorf("Wrong goarm64Features.Supports for v9.3, v8.8")
	}

	if g.Supports("v8.9") {
		t.Errorf("Wrong goarm64Features.Supports for v9.3, v8.9")
	}

	if g.Supports(",lse") {
		t.Errorf("Wrong goarm64Features.Supports for v9.3, ,lse")
	}
}

func TestGogoarchTags(t *testing.T) {
	old_goarch := GOARCH
	old_goarm64 := GOARM64

	GOARCH = "arm64"

	os.Setenv("GOARM64", "v9.5")
	GOARM64 = goarm64()
	tags := gogoarchTags()
	want := []string{"arm64.v9.0", "arm64.v9.1", "arm64.v9.2", "arm64.v9.3", "arm64.v9.4", "arm64.v9.5",
		"arm64.v8.0", "arm64.v8.1", "arm64.v8.2", "arm64.v8.3", "arm64.v8.4", "arm64.v8.5", "arm64.v8.6", "arm64.v8.7", "arm64.v8.8", "arm64.v8.9"}
	if len(tags) != len(want) {
		t.Errorf("Wrong number of tags for GOARM64=v9.5")
	} else {
		for i, v := range tags {
			if v != want[i] {
				t.Error("Wrong tags for GOARM64=v9.5")
				break
			}
		}
	}

	GOARCH = old_goarch
	GOARM64 = old_goarm64
}

var goodFIPS = []string{
	"v1.0.0",
	"v1.0.1",
	"v1.2.0",
	"v1.2.3",
}

var badFIPS = []string{
	"v1.0.0-fips",
	"v1.0.0+fips",
	"1.0.0",
	"x1.0.0",
}

func TestIsFIPSVersion(t *testing.T) {
	// good
	for _, s := range goodFIPS {
		if !isFIPSVersion(s) {
			t.Errorf("isFIPSVersion(%q) = false, want true", s)
		}
	}
	// truncated
	const v = "v1.2.3"
	for i := 0; i < len(v); i++ {
		if isFIPSVersion(v[:i]) {
			t.Errorf("isFIPSVersion(%q) = true, want false", v[:i])
		}
	}
	// bad
	for _, s := range badFIPS {
		if isFIPSVersion(s) {
			t.Errorf("isFIPSVersion(%q) = true, want false", s)
		}
	}
}
