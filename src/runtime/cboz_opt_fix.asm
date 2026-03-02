TEXT runtime_test.BenchmarkClearFat1024(SB) /home/10356270/dev/zte-riscv/go/src/runtime/memmove_test.go
  memmove_test.go:699	0x2f3538		010db303		MOV 16(X27), X6					
  memmove_test.go:699	0x2f353c		00236863		BLTU X6, X2, 4(PC)				
  memmove_test.go:699	0x2f3540		2ae4			MOV X10, 8(X2)					
  memmove_test.go:699	0x2f3542		d66902ef		JAL X5, runtime.morestack_noctxt-tramp1(SB)	
  memmove_test.go:699	0x2f3546		2265			MOV 8(X2), X10					
  memmove_test.go:699	0x2f3548		ff1ff06f		JMP runtime_test.BenchmarkClearFat1024(SB)	
  memmove_test.go:699	0x2f354c		fc113c23		MOV X1, -40(X2)					
  memmove_test.go:699	0x2f3550		fd810113		ADDI $-40, X2, X2				
  memmove_test.go:699	0x2f3554		06e0			MOV X1, (X2)					
  export_test.go:1451	0x2f3556		2af8			MOV X10, 48(X2)					
  memmove_test.go:700	0x2f3558		000e1517		AUIPC $225, X10					
  memmove_test.go:700	0x2f355c		f0850513		ADDI $-248, X10, X10				
  memmove_test.go:700	0x2f3560		e49380ef		CALL runtime.newobject-tramp1(SB)		
  memmove_test.go:700	0x2f3564		2af0			MOV X10, 32(X2)					
  export_test.go:1451	0x2f3566		0054f297		AUIPC $1359, X5					
  export_test.go:1451	0x2f356a		5fe2c283		MOVBU 1534(X5), X5				
  memmove_test.go:701	0x2f356e		0100			MOV X0, X0					
  export_test.go:1451	0x2f3570		02028e63		BEQZ X5, 15(PC)					
  export_test.go:1452	0x2f3574		000cd297		AUIPC $205, X5					
  export_test.go:1452	0x2f3578		42c28293		ADDI $1068, X5, X5				
  export_test.go:1452	0x2f357c		00527f97		AUIPC $1319, X31				
  export_test.go:1452	0x2f3580		865fb223		MOV X5, -1948(X31)				
  export_test.go:1452	0x2f3584		00550297		AUIPC $1360, X5					
  export_test.go:1452	0x2f3588		c3c2e283		MOVWU -964(X5), X5				
  export_test.go:1452	0x2f358c		00028c63		BEQZ X5, 6(PC)					
  export_test.go:1452	0x2f3590		00527297		AUIPC $1319, X5					
  export_test.go:1452	0x2f3594		8582b283		MOV -1960(X5), X5				
  export_test.go:1452	0x2f3598		918910ef		CALL runtime.gcWriteBarrier2-tramp1(SB)		
  export_test.go:1452	0x2f359c		00ac3023		MOV X10, (X24)					
  export_test.go:1452	0x2f35a0		005c3423		MOV X5, 8(X24)					
  export_test.go:1452	0x2f35a4		00527f97		AUIPC $1319, X31				
  export_test.go:1452	0x2f35a8		84afb223		MOV X10, -1980(X31)				
  memmove_test.go:702	0x2f35ac		4275			MOV 48(X2), X10					
  memmove_test.go:702	0x2f35ae		8e3b00ef		CALL testing.(*B).ResetTimer-tramp0(SB)		
  memmove_test.go:703	0x2f35b2		4275			MOV 48(X2), X10					
  memmove_test.go:703	0x2f35b4		8272			MOV 32(X2), X5					
  memmove_test.go:703	0x2f35b6		0143			MOV X0, X6					
  memmove_test.go:703	0x2f35b8		0840006f		JMP 33(PC)					
  memmove_test.go:704	0x2f35bc		40028493		ADDI $1024, X5, X9				
  memmove_test.go:704	0x2f35c0		fc04f593		ANDI $-64, X9, X11				
  memmove_test.go:704	0x2f35c4		03f28393		ADDI $63, X5, X7				
  memmove_test.go:704	0x2f35c8		fc03f393		ANDI $-64, X7, X7				
  memmove_test.go:704	0x2f35cc		407283b3		SUB X7, X5, X7					
  memmove_test.go:704	0x2f35d0		00705863		BGE X0, X7, 4(PC)				
  memmove_test.go:704	0x2f35d4		00028023		MOVB X0, (X5)					
  memmove_test.go:704	0x2f35d8		8502			ADDI $1, X5, X5					
  memmove_test.go:704	0x2f35da		fd13			ADDI $-1, X7, X7				
  memmove_test.go:704	0x2f35dc		fe704ce3		BLT X0, X7, -2(PC)				
  memmove_test.go:704	0x2f35e0		40b283b3		SUB X11, X5, X7					
  memmove_test.go:704	0x2f35e4		0042a00f		CBOZERO X5					
  memmove_test.go:704	0x2f35e8		04028293		ADDI $64, X5, X5				
  memmove_test.go:704	0x2f35ec		fc038393		ADDI $-64, X7, X7				
  memmove_test.go:704	0x2f35f0		fe704ae3		BLT X0, X7, -3(PC)				
  memmove_test.go:704	0x2f35f4		409283b3		SUB X9, X5, X7					
  memmove_test.go:704	0x2f35f8		00705863		BGE X0, X7, 4(PC)				
  memmove_test.go:704	0x2f35fc		00028023		MOVB X0, (X5)					
  memmove_test.go:704	0x2f3600		8502			ADDI $1, X5, X5					
  memmove_test.go:704	0x2f3602		fd13			ADDI $-1, X7, X7				
  memmove_test.go:704	0x2f3604		fe704ce3		BLT X0, X7, -2(PC)				
  memmove_test.go:704	0x2f3608		0300006f		JMP 12(PC)					
  memmove_test.go:704	0x2f360c		40028393		ADDI $1024, X5, X7				
  memmove_test.go:704	0x2f3610		0002a023		MOVW X0, (X5)					
  memmove_test.go:704	0x2f3614		0002a223		MOVW X0, 4(X5)					
  memmove_test.go:704	0x2f3618		0002a423		MOVW X0, 8(X5)					
  memmove_test.go:704	0x2f361c		0002a623		MOVW X0, 12(X5)					
  memmove_test.go:704	0x2f3620		0002a823		MOVW X0, 16(X5)					
  memmove_test.go:704	0x2f3624		0002aa23		MOVW X0, 20(X5)					
  memmove_test.go:704	0x2f3628		0002ac23		MOVW X0, 24(X5)					
  memmove_test.go:704	0x2f362c		0002ae23		MOVW X0, 28(X5)					
  memmove_test.go:704	0x2f3630		02028293		ADDI $32, X5, X5				
  memmove_test.go:704	0x2f3634		fc539ee3		BNE X7, X5, -9(PC)				
  memmove_test.go:703	0x2f3638		0503			ADDI $1, X6, X6					
  memmove_test.go:704	0x2f363a		8272			MOV 32(X2), X5					
  memmove_test.go:703	0x2f363c		1c853383		MOV 456(X10), X7				
  memmove_test.go:703	0x2f3640		f6734ee3		BLT X6, X7, -33(PC)				
  memmove_test.go:706	0x2f3644		8260			MOV (X2), X1					
  memmove_test.go:706	0x2f3646		02810113		ADDI $40, X2, X2				
  memmove_test.go:706	0x2f364a		00008067		RET						
