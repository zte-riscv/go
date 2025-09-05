//go:build riscv64

#include "go_asm.h"
#include "textflag.h"

// 参数为三个 []byte，默认 ABI（ABI0）下从栈帧读取：
//   in:   in_base+0(FP),  in_len+8(FP),  in_cap+16(FP)
//   out0: out0_base+0(FP),out0_len+8(FP),out0_cap+16(FP)
//   out1: out1_base+0(FP),out1_len+8(FP),out1_cap+16(FP)

TEXT ·vlseg2Deinterleave(SB), NOSPLIT, $0-72
	// 从栈帧取参到工作寄存器
	MOV	in_base+0(FP), X10
	MOV	in_len+8(FP), X11
	// 每个 []byte 为 24 字节，out0 从 +24，out1 从 +48 开始
	MOV	out0_base+24(FP), X13
	MOV	out1_base+48(FP), X16
	// nPairs = in_len / 2
	SRLI	$1, X11, X11

	BEQZ	X11, done

loop:
	// 设置 vl = min(nPairs, VLEN/SEW)，SEW=E8, LMUL=M1，TA/MA
	// rd=X19 接收实际 vl，rs1=X11 提供 nPairs
	VSETVLI	X19, E8, M1, TA, MA, X11

	// 分段加载：v8=偶数元素，v9=奇数元素（以对为单位）
	VLSEG2E8V	(X10), V8

	// 分别存回两个输出缓冲
	VSE8V	V8, (X13)
	VSE8V	V9, (X16)

	// 指针前移：in += vl*2，out0/out1 各 += vl
	ADD	X19, X13
	ADD	X19, X16
	SLLI	$1, X19, X20
	ADD	X20, X10

	// nPairs -= vl
	SUB	X19, X11
	BNEZ	X11, loop

done:
	RET
