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
	// 标量版本：每次搬运一对字节 [a,b]，分别写入 out0/out1
sloop:
	MOVBU	(X10), X5
	MOVB	X5, (X13)
	ADDI	$1, X10
	MOVBU	(X10), X5
	MOVB	X5, (X16)
	ADDI	$1, X10
	ADDI	$1, X13
	ADDI	$1, X16
	ADDI	$-1, X11
	BNEZ	X11, sloop

done:
	RET
