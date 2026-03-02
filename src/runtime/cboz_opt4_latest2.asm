TEXT runtime_test.BenchmarkClearFat1024(SB) /home/10356270/dev/zte-riscv/go/src/runtime/memmove_test.go
  memmove_test.go:699	0x2f39f8		010db303		MOV 16(X27), X6					
  memmove_test.go:699	0x2f39fc		00236863		BLTU X6, X2, 4(PC)				
  memmove_test.go:699	0x2f3a00		2ae4			MOV X10, 8(X2)					
  memmove_test.go:699	0x2f3a02		c668f2ef		JAL X5, runtime.morestack_noctxt-tramp1(SB)	
  memmove_test.go:699	0x2f3a06		2265			MOV 8(X2), X10					
  memmove_test.go:699	0x2f3a08		ff1ff06f		JMP runtime_test.BenchmarkClearFat1024(SB)	
  memmove_test.go:699	0x2f3a0c		fc113c23		MOV X1, -40(X2)					
  memmove_test.go:699	0x2f3a10		fd810113		ADDI $-40, X2, X2				
  memmove_test.go:699	0x2f3a14		06e0			MOV X1, (X2)					
  export_test.go:1451	0x2f3a16		2af8			MOV X10, 48(X2)					
  memmove_test.go:700	0x2f3a18		000e1517		AUIPC $225, X10					
  memmove_test.go:700	0x2f3a1c		a4850513		ADDI $-1464, X10, X10				
  memmove_test.go:700	0x2f3a20		a61380ef		CALL runtime.newobject-tramp1(SB)		
  memmove_test.go:700	0x2f3a24		2af0			MOV X10, 32(X2)					
  export_test.go:1451	0x2f3a26		0054f297		AUIPC $1359, X5					
  export_test.go:1451	0x2f3a2a		13e2c283		MOVBU 318(X5), X5				
  memmove_test.go:701	0x2f3a2e		0100			MOV X0, X0					
  export_test.go:1451	0x2f3a30		02028e63		BEQZ X5, 15(PC)					
  export_test.go:1452	0x2f3a34		000cd297		AUIPC $205, X5					
  export_test.go:1452	0x2f3a38		f6c28293		ADDI $-148, X5, X5				
  export_test.go:1452	0x2f3a3c		00526f97		AUIPC $1318, X31				
  export_test.go:1452	0x2f3a40		3a5fb223		MOV X5, 932(X31)				
  export_test.go:1452	0x2f3a44		0054f297		AUIPC $1359, X5					
  export_test.go:1452	0x2f3a48		77c2e283		MOVWU 1916(X5), X5				
  export_test.go:1452	0x2f3a4c		00028c63		BEQZ X5, 6(PC)					
  export_test.go:1452	0x2f3a50		00526297		AUIPC $1318, X5					
  export_test.go:1452	0x2f3a54		3982b283		MOV 920(X5), X5					
  export_test.go:1452	0x2f3a58		b81900ef		CALL runtime.gcWriteBarrier2-tramp1(SB)		
  export_test.go:1452	0x2f3a5c		00ac3023		MOV X10, (X24)					
  export_test.go:1452	0x2f3a60		005c3423		MOV X5, 8(X24)					
  export_test.go:1452	0x2f3a64		00526f97		AUIPC $1318, X31				
  export_test.go:1452	0x2f3a68		38afb223		MOV X10, 900(X31)				
  memmove_test.go:702	0x2f3a6c		4275			MOV 48(X2), X10					
  memmove_test.go:702	0x2f3a6e		8c3b00ef		CALL testing.(*B).ResetTimer-tramp0(SB)		
  memmove_test.go:703	0x2f3a72		4275			MOV 48(X2), X10					
  memmove_test.go:703	0x2f3a74		8272			MOV 32(X2), X5					
  memmove_test.go:703	0x2f3a76		0143			MOV X0, X6					
  memmove_test.go:703	0x2f3a78		0840006f		JMP 33(PC)					
  memmove_test.go:704	0x2f3a7c		40028493		ADDI $1024, X5, X9				
  memmove_test.go:704	0x2f3a80		fc04f593		ANDI $-64, X9, X11				
  memmove_test.go:704	0x2f3a84		03f28393		ADDI $63, X5, X7				
  memmove_test.go:704	0x2f3a88		fc03f393		ANDI $-64, X7, X7				
  memmove_test.go:704	0x2f3a8c		407283b3		SUB X7, X5, X7					
  memmove_test.go:704	0x2f3a90		00705863		BGE X0, X7, 4(PC)				
  memmove_test.go:704	0x2f3a94		0002a023		MOVW X0, (X5)					
  memmove_test.go:704	0x2f3a98		9102			ADDI $4, X5, X5					
  memmove_test.go:704	0x2f3a9a		f113			ADDI $-4, X7, X7				
  memmove_test.go:704	0x2f3a9c		fe704ce3		BLT X0, X7, -2(PC)				
  memmove_test.go:704	0x2f3aa0		40b283b3		SUB X11, X5, X7					
  memmove_test.go:704	0x2f3aa4		0042a00f		CBOZERO X5					
  memmove_test.go:704	0x2f3aa8		04028293		ADDI $64, X5, X5				
  memmove_test.go:704	0x2f3aac		fc038393		ADDI $-64, X7, X7				
  memmove_test.go:704	0x2f3ab0		fe704ae3		BLT X0, X7, -3(PC)				
  memmove_test.go:704	0x2f3ab4		409283b3		SUB X9, X5, X7					
  memmove_test.go:704	0x2f3ab8		00705863		BGE X0, X7, 4(PC)				
  memmove_test.go:704	0x2f3abc		0002a023		MOVW X0, (X5)					
  memmove_test.go:704	0x2f3ac0		9102			ADDI $4, X5, X5					
  memmove_test.go:704	0x2f3ac2		f113			ADDI $-4, X7, X7				
  memmove_test.go:704	0x2f3ac4		fe704ce3		BLT X0, X7, -2(PC)				
  memmove_test.go:704	0x2f3ac8		0300006f		JMP 12(PC)					
  memmove_test.go:704	0x2f3acc		40028393		ADDI $1024, X5, X7				
  memmove_test.go:704	0x2f3ad0		0002a023		MOVW X0, (X5)					
  memmove_test.go:704	0x2f3ad4		0002a223		MOVW X0, 4(X5)					
  memmove_test.go:704	0x2f3ad8		0002a423		MOVW X0, 8(X5)					
  memmove_test.go:704	0x2f3adc		0002a623		MOVW X0, 12(X5)					
  memmove_test.go:704	0x2f3ae0		0002a823		MOVW X0, 16(X5)					
  memmove_test.go:704	0x2f3ae4		0002aa23		MOVW X0, 20(X5)					
  memmove_test.go:704	0x2f3ae8		0002ac23		MOVW X0, 24(X5)					
  memmove_test.go:704	0x2f3aec		0002ae23		MOVW X0, 28(X5)					
  memmove_test.go:704	0x2f3af0		02028293		ADDI $32, X5, X5				
  memmove_test.go:704	0x2f3af4		fc539ee3		BNE X7, X5, -9(PC)				
  memmove_test.go:703	0x2f3af8		0503			ADDI $1, X6, X6					
  memmove_test.go:704	0x2f3afa		8272			MOV 32(X2), X5					
  memmove_test.go:703	0x2f3afc		1c853383		MOV 456(X10), X7				
  memmove_test.go:703	0x2f3b00		f6734ee3		BLT X6, X7, -33(PC)				
  memmove_test.go:706	0x2f3b04		8260			MOV (X2), X1					
  memmove_test.go:706	0x2f3b06		02810113		ADDI $40, X2, X2				
  memmove_test.go:706	0x2f3b0a		00008067		RET						
