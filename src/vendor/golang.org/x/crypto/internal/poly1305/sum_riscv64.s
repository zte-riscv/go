// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build gc && !purego && riscv64

#include "textflag.h"

// func update(state *macState, msg []byte)
TEXT ·update(SB),NOSPLIT,$0-32
	MOV	state+0(FP), X10	// state pointer
	MOV	msg_base+8(FP), X11	// msg base
	MOV	msg_len+16(FP), X12	// msg length

	MOV	$16, X13		// TagSize = 16

	// Load state: h0, h1, h2, r0, r1
	MOV	(X10), X5		// h0
	MOV	8(X10), X6		// h1
	MOV	16(X10), X7		// h2
	MOV	24(X10), X8		// r0
	MOV	32(X10), X9		// r1

	// Check if msg length >= 16
	BLTU	X12, X13, bytes_between_0_and_15

loop:
	// Load 16 bytes from msg
	MOV	(X11), X14		// msg[0:8]
	MOV	8(X11), X15		// msg[8:16]

	// h0 += msg[0:8]
	ADD	X14, X5, X5		// h0 = h0 + msg[0:8]
	SLTU	X14, X5, X16		// carry from h0

	// h1 += msg[8:16] + carry
	ADD	X15, X6, X17		// h1 + msg[8:16]
	SLTU	X15, X17, X18		// carry from h1 addition
	ADD	X17, X16, X6		// h1 = h1 + msg[8:16] + carry
	SLTU	X17, X6, X19		// carry from h1
	OR	X19, X18, X18		// combined carry

	// h2 += carry + 1 (for full 16-byte chunk)
	ADDI	$1, X18, X18		// carry + 1
	ADD	X7, X18, X7		// h2 += carry + 1

	// Advance msg pointer
	ADDI	$16, X11, X11		// msg = msg[16:]

multiply:
	// Multiply h by r: h * r mod 2^130 - 5
	// Compute h0r0 = h0 * r0 (128-bit result)
	MUL	X5, X8, X20		// h0r0.lo
	MULHU	X5, X8, X21		// h0r0.hi

	// Compute h1r0 = h1 * r0 (128-bit result)
	MUL	X6, X8, X22		// h1r0.lo
	MULHU	X6, X8, X23		// h1r0.hi

	// Compute h2r0 = h2 * r0 (h2 is small, so hi should be 0)
	MUL	X7, X8, X24		// h2r0.lo
	MULHU	X7, X8, X25		// h2r0.hi (should be 0, checked later)

	// Compute h0r1 = h0 * r1 (128-bit result)
	MUL	X5, X9, X26		// h0r1.lo
	MULHU	X5, X9, X18		// h0r1.hi (use X18 to avoid X27/G register)

	// Compute h1r1 = h1 * r1 (128-bit result)
	MUL	X6, X9, X28		// h1r1.lo
	MULHU	X6, X9, X29		// h1r1.hi

	// Compute h2r1 = h2 * r1 (h2 is small, so hi should be 0)
	MUL	X7, X9, X30		// h2r1.lo
	MULHU	X7, X9, X31		// h2r1.hi (should be 0, checked later)

	// Check for overflow in h2r0 and h2r1
	BNEZ	X25, overflow_h2r0
	BNEZ	X31, overflow_h2r1

	// Combine products:
	// m0 = h0r0
	// m1 = h1r0 + h0r1
	// m2 = h2r0 + h1r1
	// m3 = h2r1

	// m1 = h1r0 + h0r1 (128-bit addition)
	ADD	X22, X26, X17		// m1.lo = h1r0.lo + h0r1.lo
	SLTU	X22, X17, X16		// carry from lo
	ADD	X23, X18, X19		// m1.hi = h1r0.hi + h0r1.hi (X18 is h0r1.hi)
	ADD	X19, X16, X19		// m1.hi += carry

	// m2 = h2r0 + h1r1 (128-bit addition)
	ADD	X24, X28, X23		// m2.lo = h2r0.lo + h1r1.lo
	SLTU	X24, X23, X16		// carry from lo
	ADD	X25, X29, X25		// m2.hi = h2r0.hi + h1r1.hi
	ADD	X25, X16, X25		// m2.hi += carry

	// m3 = h2r1
	// X30 = m3.lo (h2r1.lo), X31 = m3.hi (h2r1.hi, should be 0)

	// Now combine into 64-bit limbs:
	// t0 = m0.lo
	// t1 = m1.lo + m0.hi
	// t2 = m2.lo + m1.hi
	// t3 = m3.lo + m2.hi

	MOV	X20, X5		// t0 = m0.lo (h0r0.lo)

	// t1 = m1.lo + m0.hi
	ADD	X17, X21, X6		// t1 = m1.lo + m0.hi
	SLTU	X17, X6, X16		// carry
	ADD	X19, X16, X19		// m1.hi += carry

	// t2 = m2.lo + m1.hi
	ADD	X23, X19, X7		// t2 = m2.lo + m1.hi
	SLTU	X23, X7, X16		// carry
	ADD	X25, X16, X25		// m2.hi += carry

	// t3 = m3.lo + m2.hi
	ADD	X30, X25, X14		// t3 = m3.lo + m2.hi

	// Reduce modulo 2^130 - 5:
	// Split t2 at bit 130: h2 = t2 & 3, cc = (t2 & ~3, t3)
	// Save t2 before masking
	MOV	X7, X15		// Save t2
	ANDI	$3, X7, X7		// h2 = t2 & 3
	ANDI	$-4, X15, X15		// cc.lo = t2 & ~3
	MOV	X14, X16		// cc.hi = t3

	// Add cc (c * 4) to h
	ADD	X15, X5, X5		// h0 += cc.lo
	SLTU	X15, X5, X17		// carry
	ADD	X16, X6, X19		// h1 + cc.hi
	SLTU	X16, X19, X18		// carry
	ADD	X19, X17, X6		// h1 += cc.hi + carry
	SLTU	X19, X6, X20		// carry
	OR	X20, X18, X18		// combined carry
	ADD	X7, X18, X7		// h2 += carry

	// Shift cc right by 2: cc >> 2
	SRLI	$2, X15, X15		// cc.lo >> 2
	ANDI	$3, X16, X21		// Get low 2 bits of cc.hi
	SLLI	$62, X21, X21		// Shift to high bits
	OR	X21, X15, X15		// Combine
	SRLI	$2, X16, X16		// cc.hi >> 2

	// Add (cc >> 2) to h
	ADD	X15, X5, X5		// h0 += (cc >> 2).lo
	SLTU	X15, X5, X17		// carry
	ADD	X16, X6, X19		// h1 + (cc >> 2).hi
	SLTU	X16, X19, X18		// carry
	ADD	X19, X17, X6		// h1 += (cc >> 2).hi + carry
	SLTU	X19, X6, X20		// carry
	OR	X20, X18, X18		// combined carry
	ADD	X7, X18, X7		// h2 += carry

	// Check if more 16-byte blocks to process
	ADDI	$-16, X12, X12		// msg_len -= 16
	BGEU	X12, X13, loop		// if msg_len >= 16, continue loop

bytes_between_0_and_15:
	BEQ	X12, X0, done		// if msg_len == 0, done

	// Handle remaining bytes (< 16)
	// Build buffer in registers: buf[0:8] in X14, buf[8:16] in X15
	// Initialize with 1 at position len(msg)
	MOV	$1, X14		// Start with 1 in low byte
	XOR	X15, X15, X15	// Clear high part
	ADD	X12, X11, X11	// msg = msg + len(msg) (point to end)

flush_buffer:
	// Load byte from end and shift into buffer
	MOVBU	-1(X11), X16		// Load byte from msg[-1]
	SRLI	$56, X14, X17		// Get high byte of X14
	SLLI	$8, X15, X18		// Shift X15 left by 8
	OR	X17, X18, X15		// Combine: X15 = (X15 << 8) | (X14 >> 56)
	SLLI	$8, X14, X14		// Shift X14 left by 8
	XOR	X16, X14, X14		// XOR byte into X14
	ADDI	$-1, X12, X12		// Decrement length
	ADDI	$-1, X11, X11		// Decrement pointer
	BNEZ	X12, flush_buffer	// Continue if more bytes

	// Add buffer to h
	ADD	X14, X5, X5		// h0 += buf[0:8]
	SLTU	X14, X5, X16		// carry
	ADD	X15, X6, X17		// h1 + buf[8:16]
	SLTU	X15, X17, X18		// carry
	ADD	X17, X16, X6		// h1 += buf[8:16] + carry
	SLTU	X17, X6, X19		// carry
	OR	X19, X18, X18		// combined carry
	ADD	X7, X18, X7		// h2 += carry

	// Process as if it were a full 16-byte block
	MOV	$16, X12		// Set length to 16 for multiply
	JMP	multiply

overflow_h2r0:
	// Panic: h2r0.hi != 0
	// This should not happen in normal operation
	// For now, we'll just continue (could call panic)
	JMP	multiply

overflow_h2r1:
	// Panic: h2r1.hi != 0
	// This should not happen in normal operation
	// For now, we'll just continue (could call panic)
	JMP	multiply

done:
	// Store updated state
	MOV	X5, (X10)		// h0
	MOV	X6, 8(X10)		// h1
	MOV	X7, 16(X10)		// h2
	RET

