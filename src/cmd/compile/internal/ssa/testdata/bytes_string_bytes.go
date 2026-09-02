// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"os"
)

//go:noinline
func bytesStringBytes(b []byte) []byte {
	return []byte(string(b))
}

func check(name string, input, result []byte) {
	if !bytes.Equal(result, input) {
		fmt.Fprintf(os.Stderr, "%s: content mismatch: input=%v result=%v\n", name, input, result)
		os.Exit(1)
	}
	if len(input) > 0 && &result[0] == &input[0] {
		fmt.Fprintf(os.Stderr, "%s: result shares backing array with input\n", name)
		os.Exit(1)
	}
}

func main() {
	check("nil", nil, bytesStringBytes(nil))
	check("empty", []byte{}, bytesStringBytes([]byte{}))
	check("single", []byte{42}, bytesStringBytes([]byte{42}))
	check("small", []byte{1, 2, 3, 4, 5}, bytesStringBytes([]byte{1, 2, 3, 4, 5}))

	medium := make([]byte, 100)
	for i := range medium {
		medium[i] = byte(i)
	}
	check("medium", medium, bytesStringBytes(medium))

	large := make([]byte, 10000)
	for i := range large {
		large[i] = byte(i % 256)
	}
	check("large", large, bytesStringBytes(large))

	// Verify independence: modifying input after conversion must not affect result.
	src := []byte{10, 20, 30}
	dst := bytesStringBytes(src)
	src[0] = 99
	if dst[0] == 99 {
		fmt.Fprintf(os.Stderr, "independence: modifying input affected result\n")
		os.Exit(1)
	}

	fmt.Println("PASS")
}
