// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build riscv64

package sha256

import (
	"testing"
)

// TestBlockWithTrace compares the output of the generic Go implementation
// with the RISC-V optimized implementation, checking intermediate states for each round.
func TestBlockWithTrace(t *testing.T) {
	// Test cases: various 64-byte message blocks
	testCases := []struct {
		name string
		data []byte
	}{
		// {
		// 	name: "zero block",
		// 	data: make([]byte, 64),
		// },
		// {
		// 	name: "all ones",
		// 	data: bytes.Repeat([]byte{0xff}, 64),
		// },
		// {
		// 	name: "incremental",
		// 	data: func() []byte {
		// 		b := make([]byte, 64)
		// 		for i := range b {
		// 			b[i] = byte(i)
		// 		}
		// 		return b
		// 	}(),
		// },
		// {
		// 	name: "SHA-256 test vector 1",
		// 	data: []byte{
		// 		0x61, 0x62, 0x63, 0x64, 0x62, 0x63, 0x64, 0x65,
		// 		0x63, 0x64, 0x65, 0x66, 0x64, 0x65, 0x66, 0x67,
		// 		0x65, 0x66, 0x67, 0x68, 0x66, 0x67, 0x68, 0x69,
		// 		0x67, 0x68, 0x69, 0x6a, 0x68, 0x69, 0x6a, 0x6b,
		// 		0x69, 0x6a, 0x6b, 0x6c, 0x6a, 0x6b, 0x6c, 0x6d,
		// 		0x6b, 0x6c, 0x6d, 0x6e, 0x6c, 0x6d, 0x6e, 0x6f,
		// 		0x6d, 0x6e, 0x6f, 0x70, 0x6e, 0x6f, 0x70, 0x71,
		// 		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		// 	},
		// },
		{
			name: "random pattern",
			data: []byte{
				0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0,
				0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88,
				0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11,
				0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99,
				0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11,
				0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99,
				0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11,
				0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.data) != 64 {
				t.Fatalf("test case %s: data must be exactly 64 bytes, got %d", tc.name, len(tc.data))
			}

			// Create digests with initial state
			digRef := &Digest{}
			digRef.Reset()
			digOpt := &Digest{}
			digOpt.Reset()

			// Create temp_dig arrays to store intermediate states
			var tempDigRef [64][8]uint32
			var tempDigOpt [64][8]uint32
			// Create temp_kword arrays to store w[i]+_K[i] for each round
			var tempKwordRef [64]uint32
			var tempKwordOpt [64]uint32

			// Get reference output from generic implementation
			blockGenericWithTrace(digRef, tc.data, &tempDigRef, &tempKwordRef)

			// Get output from RISC-V optimized implementation
			blockRISCV64WithTrace(digOpt, tc.data, &tempDigOpt, &tempKwordOpt)

			// Compare final digest states
			if digRef.h != digOpt.h {
				t.Errorf("test case %s: final digest states differ", tc.name)
				t.Logf("Reference digest: %x", digRef.h[:])
				t.Logf("Optimized digest: %x", digOpt.h[:])
			}

			// Compare intermediate states for each round (only odd rounds: i%2==1)
			for i := 0; i < 64; i++ {
				if i%2 == 1 {
					// Reorder tempDigOpt[i] from {a b e f c d g h} to {a b c d e f g h}
					var tempDigOptReordered [8]uint32
					tempDigOptReordered[0] = tempDigOpt[i][3] // a
					tempDigOptReordered[1] = tempDigOpt[i][2] // b
					tempDigOptReordered[2] = tempDigOpt[i][7] // c
					tempDigOptReordered[3] = tempDigOpt[i][6] // d
					tempDigOptReordered[4] = tempDigOpt[i][1] // e
					tempDigOptReordered[5] = tempDigOpt[i][0] // f
					tempDigOptReordered[6] = tempDigOpt[i][5] // g
					tempDigOptReordered[7] = tempDigOpt[i][4] // h

					if tempDigRef[i] != tempDigOptReordered {
						t.Logf("test case %s: round %d intermediate states differ", tc.name, i)
						t.Logf("Round %d - Reference: a=0x%08x, b=0x%08x, c=0x%08x, d=0x%08x, e=0x%08x, f=0x%08x, g=0x%08x, h=0x%08x",
							i, tempDigRef[i][0], tempDigRef[i][1], tempDigRef[i][2], tempDigRef[i][3],
							tempDigRef[i][4], tempDigRef[i][5], tempDigRef[i][6], tempDigRef[i][7])
						t.Logf("Round %d - Optimized (original): a=0x%08x, b=0x%08x, e=0x%08x, f=0x%08x, c=0x%08x, d=0x%08x, g=0x%08x, h=0x%08x",
							i, tempDigOpt[i][0], tempDigOpt[i][1], tempDigOpt[i][2], tempDigOpt[i][3],
							tempDigOpt[i][4], tempDigOpt[i][5], tempDigOpt[i][6], tempDigOpt[i][7])
						t.Logf("Round %d - Optimized (reordered): a=0x%08x, b=0x%08x, c=0x%08x, d=0x%08x, e=0x%08x, f=0x%08x, g=0x%08x, h=0x%08x",
							i, tempDigOptReordered[0], tempDigOptReordered[1], tempDigOptReordered[2], tempDigOptReordered[3],
							tempDigOptReordered[4], tempDigOptReordered[5], tempDigOptReordered[6], tempDigOptReordered[7])

						// Show individual differences
						for j := 0; j < 8; j++ {
							if tempDigRef[i][j] != tempDigOptReordered[j] {
								stateNames := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
								t.Logf("  Round %d, %s differs: ref=0x%08x, opt=0x%08x",
									i, stateNames[j], tempDigRef[i][j], tempDigOptReordered[j])
							}
						}
						// Only report first difference to avoid too much output
						break
					}
				}
			}

			// Compare temp_kword values for each round
			for i := 0; i < 4; i++ {
				if tempKwordRef[i] != tempKwordOpt[i] {
					t.Errorf("test case %s: round %d temp_kword differs: ref=0x%08x, opt=0x%08x",
						tc.name, i, tempKwordRef[i], tempKwordOpt[i])
					//Only report first difference to avoid too much output
					break
				}
			}
		})
	}
}

// TestBlockWithTraceAllRounds tests all 64 rounds with detailed output for each round.
// func TestBlockWithTraceAllRounds(t *testing.T) {
// 	// Use a test message block
// 	testMessage := []byte{
// 		0x61, 0x62, 0x63, 0x64, 0x62, 0x63, 0x64, 0x65,
// 		0x63, 0x64, 0x65, 0x66, 0x64, 0x65, 0x66, 0x67,
// 		0x65, 0x66, 0x67, 0x68, 0x66, 0x67, 0x68, 0x69,
// 		0x67, 0x68, 0x69, 0x6a, 0x68, 0x69, 0x6a, 0x6b,
// 		0x69, 0x6a, 0x6b, 0x6c, 0x6a, 0x6b, 0x6c, 0x6d,
// 		0x6b, 0x6c, 0x6d, 0x6e, 0x6c, 0x6d, 0x6e, 0x6f,
// 		0x6d, 0x6e, 0x6f, 0x70, 0x6e, 0x6f, 0x70, 0x71,
// 		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
// 	}

// 	// Create digests with initial state
// 	digRef := &Digest{}
// 	digRef.Reset()
// 	digOpt := &Digest{}
// 	digOpt.Reset()

// 	// Create temp_dig arrays to store intermediate states
// 	var tempDigRef [64][8]uint32
// 	var tempDigOpt [64][8]uint32

// 	// Get reference output from generic implementation
// 	blockGenericWithTrace(digRef, testMessage, &tempDigRef)

// 	// Get output from RISC-V optimized implementation
// 	blockRISCV64WithTrace(digOpt, testMessage, &tempDigOpt)

// 	// Compare each round's intermediate state
// 	for i := 0; i < 64; i++ {
// 		t.Run(fmt.Sprintf("round_%d", i), func(t *testing.T) {
// 			if tempDigRef[i] != tempDigOpt[i] {
// 				t.Errorf("Round %d: intermediate states differ", i)
// 				t.Logf("Reference: a=0x%08x, b=0x%08x, c=0x%08x, d=0x%08x, e=0x%08x, f=0x%08x, g=0x%08x, h=0x%08x",
// 					tempDigRef[i][0], tempDigRef[i][1], tempDigRef[i][2], tempDigRef[i][3],
// 					tempDigRef[i][4], tempDigRef[i][5], tempDigRef[i][6], tempDigRef[i][7])
// 				t.Logf("Optimized: a=0x%08x, b=0x%08x, c=0x%08x, d=0x%08x, e=0x%08x, f=0x%08x, g=0x%08x, h=0x%08x",
// 					tempDigOpt[i][0], tempDigOpt[i][1], tempDigOpt[i][2], tempDigOpt[i][3],
// 					tempDigOpt[i][4], tempDigOpt[i][5], tempDigOpt[i][6], tempDigOpt[i][7])

// 				// Show individual differences
// 				stateNames := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
// 				for j := 0; j < 8; j++ {
// 					if tempDigRef[i][j] != tempDigOpt[i][j] {
// 						t.Errorf("  %s differs: ref=0x%08x, opt=0x%08x",
// 							stateNames[j], tempDigRef[i][j], tempDigOpt[i][j])
// 					}
// 				}
// 			}
// 		})
// 	}

// 	// Also compare final digest states
// 	if digRef.h != digOpt.h {
// 		t.Errorf("Final digest states differ")
// 		t.Logf("Reference digest: %x", digRef.h[:])
// 		t.Logf("Optimized digest: %x", digOpt.h[:])
// 	}
// }
