// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sync

import (
	"internal/abi"
	"unsafe"
)

// NewBadHashTrieMap creates a new HashTrieMap for the provided key and value
// but with an intentionally bad hash function.
func NewBadHashTrieMap[K, V comparable]() *HashTrieMap[K, V] {
	// Stub out the good hash function with a terrible one.
	// Everything should still work as expected.
	var m HashTrieMap[K, V]
	m.init()
	m.keyHash = func(_ unsafe.Pointer, _ uintptr) uintptr {
		return 0
	}
	return &m
}

// NewTruncHashTrieMap creates a new HashTrieMap for the provided key and value
// but with an intentionally bad hash function.
func NewTruncHashTrieMap[K, V comparable]() *HashTrieMap[K, V] {
	// Stub out the good hash function with a terrible one.
	// Everything should still work as expected.
	var m HashTrieMap[K, V]
	var mx map[string]int
	mapType := abi.TypeOf(mx).MapType()
	hasher := mapType.Hasher
	m.keyHash = func(p unsafe.Pointer, n uintptr) uintptr {
		return hasher(p, n) & ((uintptr(1) << 4) - 1)
	}
	return &m
}

func NewXor1HashTrieMap[K comparable, V any](k1, k2 K) *HashTrieMap[K, V] {
	var m HashTrieMap[K, V]
	m.init()
	m.keyHash = func(p unsafe.Pointer, _ uintptr) uintptr {
		k := *(*K)(p)
		if k == k1 {
			return 0xFFFFFFFF_FFFFFFFE
		}
		if k == k2 {
			return 0xFFFFFFFF_FFFFFFFF
		}
		return 0
	}
	return &m
}

func NewFinalBitCollisionHashTrieMap[K comparable, V any](k1, k2, k3 K) *HashTrieMap[K, V] {
	const (
		sharedPrefix = uintptr(0x123456789ABC) << 15
		parentSlot   = uintptr(0x15)
		siblingSlot  = uintptr(0x2a)
		h1           = sharedPrefix | (siblingSlot << 8) | (parentSlot << 1)
		h2           = h1 | 1
		h3           = sharedPrefix | (siblingSlot << 8) | (siblingSlot << 1)
	)

	var m HashTrieMap[K, V]
	m.init()
	m.keyHash = func(p unsafe.Pointer, _ uintptr) uintptr {
		k := *(*K)(p)
		switch k {
		case k1:
			return h1
		case k2:
			return h2
		case k3:
			return h3
		default:
			return 0
		}
	}
	return &m
}
