TEXT ·myf(SB), $0-0
#ifdef GORISCV64OPT_TEST
	SUB $1, X10, X10
#else
	ADD $1, X10, X10
    ADD $1, X10, X10
    ADD $1, X10, X10
#endif
    RET
    