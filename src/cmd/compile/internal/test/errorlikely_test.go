// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

package test

import (
	"internal/testenv"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func errorLikelyAssemblyDump(t *testing.T, code string) []byte {
	testenv.MustHaveGoBuild(t)

	// Create source
	dir := t.TempDir()
	src := filepath.Join(dir, "test.go")
	f, err := os.OpenFile(src, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		t.Fatalf("could not create source file: %v", err)
	}
	_, err = f.Write([]byte(code))
	if err != nil {
		t.Fatalf("could not write source file: %v", err)
	}
	err = f.Close()
	if err != nil {
		t.Fatalf("could not close source file: %v", err)
	}

	// Compile source
	cmd := testenv.Command(t, testenv.GoToolPath(t), "tool", "compile", "-d=blockpredict=2", "-S", src)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("could not build target: %v\n%s", err, out)
	}
	return out
}

func errorLikelyCheckOrder(t *testing.T, s string, order []string) {
	index := 0
	for i, v := range order {
		if p := strings.Index(s[index:], v); p == -1 {
			t.Fatalf("expect order[%d](%s), read from index %d\n%s", i, v, index, s[index:])
		} else {
			index += p + len(v)
		}
	}
}

func TestErrorLikelyIfCondErrorNil(t *testing.T) {
	t.Parallel()
	expectOrder := []string{
		"main.main STEXT",
		`go:string."likely1`,
		`go:string."likely2`,
		"RET",
		`go:string."unlikely2`,
		`go:string."unlikely1`,
		"CALL\truntime.morestack_noctxt(SB)",
	}

	src := `
 	  package main
 	 
 	  func Foo1() error
 	  func Foo2() error
 	 
 	  func main() {
 	  	err := Foo1()
 	  	if err != nil {
 	  		println("unlikely1")
 	  	} else {
 	  		println("likely1")
 	  	}
 	 
 	  	err = Foo2()
 	  	if err == nil {
 	  		println("likely2")
 	  	} else {
 	  		println("unlikely2")
 	  	}
 	  }
 	  `
	output := errorLikelyAssemblyDump(t, src)
	errorLikelyCheckOrder(t, string(output), expectOrder)
}

func TestErrorLikelyIfBodyReturnError(t *testing.T) {
	t.Parallel()
	expectOrder := []string{
		"main.Foo STEXT",
		`go:string."likely1`,
		`go:string."likely2`,
		"RET",
		`go:string."unlikely2`,
		`go:string."unlikely1`,
		"CALL\truntime.morestack_noctxt(SB)",
	}

	src := `
 	  package main
 	 
 	  type err struct{}
 	 
 	  func (e err) Error() string {
 	  	return "Error"
 	  }
 	 
 	  func Bar1() bool
 	  func Bar2() bool
 	 
 	  func Foo() (int, error) {
 	  	if Bar1() {
 	  		println("unlikely1")
 	  		return 1, err{}
 	  	} else {
 	  		println("likely1")
 	  	}
 	 
 	  	if Bar2() {
 	  		println("unlikely2")
 	  	} else {
 	  		println("likely2")
 	  		return 0, nil
 	  	}
 	 
 	  	return 1, err{}
 	  }
 	  `
	output := errorLikelyAssemblyDump(t, src)
	errorLikelyCheckOrder(t, string(output), expectOrder)
}
