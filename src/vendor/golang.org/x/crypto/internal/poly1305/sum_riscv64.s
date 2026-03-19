// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build gc && !purego

#define LOAD32U(base, offset, tmp, dest) \
	MOVBU	(offset+0*1)(base), dest; \
	MOVBU	(offset+1*1)(base), tmp; \
	SLL	$8, tmp; \
	OR	tmp, dest; \
	MOVBU	(offset+2*1)(base), tmp; \
	SLL	$16, tmp; \
	OR	tmp, dest; \
	MOVBU	(offset+3*1)(base), tmp; \
	SLL	$24, tmp; \
	OR	tmp, dest

#define LOAD64U(base, offset, tmp1, tmp2, dst) \
	LOAD32U(base, offset, tmp1, dst); \
	LOAD32U(base, offset+4, tmp1, tmp2); \
	SLL	$32, tmp2; \
	OR	tmp2, dst

// func update(state *macState, msg []byte)
TEXT ·update(SB), $0-32
	MOV	state+0(FP), X5
	MOV	msg_base+8(FP), X6
	MOV	msg_len+16(FP), X7

	MOV	$0x10, X8

	AND	$7, X6, X28

	MOV	(0*8)(X5), X9		// h0
	MOV	(1*8)(X5), X10		// h1
	MOV	(2*8)(X5), X11		// h2
	MOV	(3*8)(X5), X12		// r0
	MOV	(4*8)(X5), X13		// r1

	BLT	X7, X8, bytes_between_0_and_15

loop:
	BEQZ	X28, aligned_load

	LOAD64U(X6,0*8, X16, X18, X15)
	LOAD64U(X6,1*8, X19, X20, X17)
	JMP	block

aligned_load:
	MOV	(0*8)(X6), X15		// msg[0:8]
	MOV	(1*8)(X6), X17		// msg[8:16]

block:
	ADD	X15, X9, X9			// h0 (x1 + y1 = z1', if z1' < x1 then z1' overflow)
	ADD	X17, X10, X22
	SLTU	X15, X9, X19	// h0.carry
	SLTU	X10, X22, X23
	ADD	X22, X19, X10		// h1
	SLTU	X22, X10, X19
	OR	X19, X23, X19		// h1.carry
	ADD	$0x01, X19, X19
	ADD	X11, X19, X11		// h2

	ADD	$16, X6, X6			// msg = msg[16:]

multiply:
	MUL	X9, X12, X15		// h0r0.lo
	MULHU	X9, X12, X16	// h0r0.hi
	MUL	X10, X12, X14		// h1r0.lo
	MULHU	X10, X12, X17	// h1r0.hi
	ADD	X14, X16, X16
	SLTU	X14, X16, X19
	ADD	X19, X17, X17
	MUL	X11, X12, X20
	ADD	X17, X20, X20
	MUL	X9, X13, X14		// h0r1.lo
	MULHU	X9, X13, X17	// h0r1.hi
	ADD	X14, X16, X16
	SLTU	X14, X16, X19
	ADD	X19, X17, X17
	MOV	X17, X9
	MUL	X11, X13, X21		// h2r1
	MUL	X10, X13, X14		// h1r1.lo
	MULHU	X10, X13, X17	// h1r1.hi
	ADD	X14, X20, X20
	ADD	X17, X21, X22
	SLTU	X14, X20, X19
	ADD	X22, X19, X21
	ADD	X9, X20, X20
	SLTU	X9, X20, X19
	ADD	X19, X21, X21
	AND	$3, X20, X11
	AND	$-4, X20, X18
	ADD	X18, X15, X9
	ADD	X21, X16, X22
	SLTU	X18, X9, X19
	SLTU	X21, X22, X23
	ADD	X22, X19, X10
	SLTU	X22, X10, X19
	OR	X19, X23, X19
	ADD	X19, X11, X11
	SLL	$62, X21, X22
	SRL	$2, X20, X23
	SRL	$2, X21, X21
	OR	X22, X23, X20
	ADD	X20, X9, X9
	ADD	X21, X10, X22
	SLTU	X20, X9, X19
	SLTU	X21, X22, X23
	ADD	X22, X19, X10
	SLTU	X22, X10, X19
	OR	X19, X23, X19
	ADD	X19, X11, X11

	SUB	$16, X7, X7
	BGE	X7, X8, loop

bytes_between_0_and_15:
	BEQ	X7, X0, done
	MOV	$1, X15
	XOR	X16, X16
	ADD	X7, X6, X6

flush_buffer:
	MOVBU	-1(X6), X20
	SRL	$56, X15, X19
	SLL	$8, X16, X23
	SLL	$8, X15, X15
	OR	X19, X23, X16
	XOR	X20, X15, X15
	SUB	$1, X7, X7
	SUB	$1, X6, X6
	BNE	X7, X0, flush_buffer

	ADD	X15, X9, X9
	SLTU	X15, X9, X19
	ADD	X16, X10, X22
	SLTU	X16, X22, X23
	ADD	X22, X19, X10
	SLTU	X22, X10, X19
	OR	X19, X23, X19
	ADD	X11, X19, X11

	MOV	$16, X7
	JMP	multiply

done:
	MOV	X9, (0*8)(X5)
	MOV	X10, (1*8)(X5)
	MOV	X11, (2*8)(X5)
	RET
