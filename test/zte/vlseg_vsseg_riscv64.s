//go:build riscv64

#include "go_asm.h"
#include "textflag.h"

// Parameters: three []byte, read from the stack frame under default ABI (ABI0):
//   in:   in_base+0(FP),  in_len+8(FP),  in_cap+16(FP)
//   out0: out0_base+24(FP),out0_len+32(FP),out0_cap+40(FP)
//   out1: out1_base+48(FP),out1_len+56(FP),out1_cap+64(FP)
//   out3: out3_base+72(FP),out3_len+80(FP),out3_cap+88(FP)

TEXT ·vlseg2E8VAndVsseg2E8VForDeinterleaveAndInterleave(SB), NOSPLIT, $0-96
	// Load parameters from the stack frame into working registers
	MOV	in_base+0(FP), X10
	MOV	in_len+8(FP), X11
	// Each []byte descriptor is 24 bytes; out0 at +24, out1 at +48
	MOV	out0_base+24(FP), X13
	MOV	out1_base+48(FP), X16
	MOV	out3_base+72(FP), X17
	// nPairs = in_len / 2
	SRLI	$1, X11, X11

	BEQZ	X11, done

loop:
    // Set vl = min(nPairs, VLEN/SEW), SEW=E8, LMUL=M1, TA/MA
    // rd=X12 receives actual vl; rs1=X11 supplies nPairs (caller-saved)
    VSETVLI X11, E8, M1, TA, MA, X12

    // Segmented load: V8 holds segment 0, V9 holds segment 1
    VLSEG2E8V   (X10), V8

    // Store to two output buffers
    VSE8V   V8, (X13)
    VSE8V   V9, (X16)

    VSSEG2E8V   V8, (X17)

    // Advance pointers: in += vl*2; out0/out1 each += vl
    ADD X12, X13
    ADD X12, X16
    SLLI    $1, X12, X6
    ADD X6, X10

    ADD X12, X17
    ADD X12, X17

    // nPairs -= vl
    SUB X12, X11
    BNEZ    X11, loop

done:
	RET





// Parameters: four []byte, read from the stack frame under default ABI (ABI0):
//   in:   in_base+0(FP),  in_len+8(FP),  in_cap+16(FP)
//   out0: out0_base+24(FP),out0_len+32(FP),out0_cap+40(FP)
//   out1: out1_base+48(FP),out1_len+56(FP),out1_cap+64(FP)
//   out2: out2_base+72(FP),out2_len+80(FP),out2_cap+88(FP)
//   out3: out3_base+96(FP),out3_len+104(FP),out3_cap+112(FP)

TEXT ·vlseg3E8VAndVsseg3E8VForDeinterleaveAndInterleave(SB), NOSPLIT, $0-120
	// Load parameters from the stack frame into working registers
	MOV	in_base+0(FP), X10
	MOV	in_len+8(FP), X11
	// Each []byte descriptor is 24 bytes; out0 at +24, out1 at +48, out2 at +72, out3 at +96
	MOV	out0_base+24(FP), X13
	MOV	out1_base+48(FP), X16
	MOV	out2_base+72(FP), X17
	MOV	out3_base+96(FP), X18
	// nPairs = in_len / 3
	MOV $3, X6
  DIVU X6, X11, X11    // X11 = X11 / 3

	BEQZ	X11, done2

loop2:
    // Set vl = min(nPairs, VLEN/SEW), SEW=E8, LMUL=M1, TA/MA
    // rd=X12 receives actual vl; rs1=X11 supplies nPairs (caller-saved)
    VSETVLI X11, E8, M1, TA, MA, X12

    // Segmented load: V8,V9,V10 hold deinterleaved segments 0..2
    VLSEG3E8V   (X10), V8

    // Store to three output buffers
    VSE8V   V8, (X13)
    VSE8V   V9, (X16)
    VSE8V   V10, (X17)

    VSSEG3E8V   V8, (X18)

    // Advance pointers: in += vl*3; out0/out1/out2 each += vl
    ADD X12, X13
    ADD X12, X16
    ADD X12, X17

    ADD X12, X10
    ADD X12, X10
    ADD X12, X10


    ADD X12, X18
    ADD X12, X18
    ADD X12, X18

    // nPairs -= vl
    SUB X12, X11
    BNEZ    X11, loop2

done2:
	RET



// Parameters: three []int32, read from the stack frame under default ABI (ABI0):
//   in:   in_base+0(FP),  in_len+8(FP),  in_cap+16(FP)
//   out0: out0_base+24(FP),out0_len+32(FP),out0_cap+40(FP)
//   out1: out1_base+48(FP),out1_len+56(FP),out1_cap+64(FP)
//   out3: out3_base+72(FP),out3_len+80(FP),out3_cap+88(FP)

TEXT ·vlseg2E32VAndVsseg2E32VForDeinterleaveAndInterleave(SB), NOSPLIT, $0-96
	// Load parameters from the stack frame into working registers
	MOV	in_base+0(FP), X10
	MOV	in_len+8(FP), X11
	// Each slice descriptor is 24 bytes; out0 at +24, out1 at +48
	MOV	out0_base+24(FP), X19
	MOV	out1_base+48(FP), X16
	MOV	out3_base+72(FP), X17
	// nPairs = in_len / 2
	SRLI	$1, X11, X11

	BEQZ	X11, done3

loop3:
    // Set vl = min(nPairs, VLEN/SEW), SEW=E32, LMUL=M1, TA/MA
    // rd=X12 receives actual vl; rs1=X11 supplies nPairs (caller-saved)
    VSETVLI X11, E32, M1, TA, MA, X12

    // Segmented load: V8 holds segment 0, V9 holds segment 1
    VLSEG2E32V   (X10), V8

    // Store to two output buffers

    VSE32V   V8, (X19)
    VSE32V   V9, (X16)

    VSSEG2E32V   V8, (X17)

    // Advance pointers: in += vl*2; out0/out1 each += vl
    SLLI    $2, X12, X6
    ADD X6, X19
    ADD X6, X16

    SLLI    $3, X12, X6
    ADD X6, X10

    ADD X6, X17

    // nPairs -= vl
    SUB X12, X11
    BNEZ    X11, loop3

done3:
	RET


// Parameters: three []int64, read from the stack frame under default ABI (ABI0):
//   in:   in_base+0(FP),  in_len+8(FP),  in_cap+16(FP)
//   out0: out0_base+24(FP),out0_len+32(FP),out0_cap+40(FP)
//   out1: out1_base+48(FP),out1_len+56(FP),out1_cap+64(FP)
//   out3: out3_base+72(FP),out3_len+80(FP),out3_cap+88(FP)

TEXT ·vlseg2E64VAndVsseg2E64VForDeinterleaveAndInterleave(SB), NOSPLIT, $0-96
	// Load parameters from the stack frame into working registers
	MOV	in_base+0(FP), X10
	MOV	in_len+8(FP), X11
	// Each slice descriptor is 24 bytes; out0 at +24, out1 at +48
	MOV	out0_base+24(FP), X19
	MOV	out1_base+48(FP), X16
	MOV	out3_base+72(FP), X17
	// nPairs = in_len / 2
	SRLI	$1, X11, X11

	BEQZ	X11, done4

loop4:
    // Set vl = min(nPairs, VLEN/SEW), SEW=E32, LMUL=M1, TA/MA
    // rd=X12 receives actual vl; rs1=X11 supplies nPairs (caller-saved)
    VSETVLI X11, E64, M1, TA, MA, X12

    // Segmented load: V8 holds segment 0, V9 holds segment 1
    VLSEG2E64V   (X10), V8

    // Store to two output buffers

    VSE64V   V8, (X19)
    VSE64V   V9, (X16)

    VSSEG2E64V   V8, (X17)

    // Advance pointers: in += vl*2; out0/out1 each += vl
    SLLI    $3, X12, X6
    ADD X6, X19
    ADD X6, X16

    SLLI    $4, X12, X6
    ADD X6, X10

    ADD X6, X17

    // nPairs -= vl
    SUB X12, X11
    BNEZ    X11, loop4  

done4:
	RET
