// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package riscv64asm

import (
	"fmt"
	"strings"
)

// GNUSyntax returns the GNU assembler syntax for the instruction, as defined by GNU binutils.
// This form typically matches the syntax defined in the RISC-V Instruction Set Manual. See
// https://github.com/riscv/riscv-isa-manual/releases/download/Ratified-IMAFDQC/riscv-spec-20191213.pdf
func GNUSyntax(inst Inst) string {
	op := strings.ToLower(inst.Op.String())
	var args []string
	for _, a := range inst.Args {
		if a == nil {
			break
		}
		args = append(args, strings.ToLower(a.String()))
	}

	switch inst.Op {
	case ADDI, ADDIW, ANDI, ORI, SLLI, SLLIW, SRAI, SRAIW, SRLI, SRLIW, XORI:
		if inst.Op == ADDI {
			if inst.Args[1].(Reg) == X0 && inst.Args[0].(Reg) != X0 {
				op = "li"
				args[1] = args[2]
				args = args[:len(args)-1]
				break
			}

			if inst.Args[2].(Simm).Imm == 0 {
				if inst.Args[0].(Reg) == X0 && inst.Args[1].(Reg) == X0 {
					op = "nop"
					args = nil
				} else {
					op = "mv"
					args = args[:len(args)-1]
				}
			}
		}

		if inst.Op == ANDI && inst.Args[2].(Simm).Imm == 255 {
			op = "zext.b"
			args = args[:len(args)-1]
		}

		if inst.Op == ADDIW && inst.Args[2].(Simm).Imm == 0 {
			op = "sext.w"
			args = args[:len(args)-1]
		}

		if inst.Op == XORI && inst.Args[2].(Simm).String() == "-1" {
			op = "not"
			args = args[:len(args)-1]
		}

		if inst.Op == ORI && inst.Args[0].(Reg) == X0 {
			imm := inst.Args[2].(Simm).Imm
			switch imm & ((1 << 5) - 1) {
			case 0:
				op = "prefetch.i"
			case 1:
				op = "prefetch.r"
			case 3:
				op = "prefetch.w"
			}
			if imm == 0 {
				args[0] = fmt.Sprintf("(X%d)", inst.Args[1].(Reg))
			} else {
				args[0] = fmt.Sprintf("%s(X%d)", inst.Args[2].(Simm).String(), inst.Args[1].(Reg))
			}
			args = args[:len(args)-2]
		}

	case ADD:
		if inst.Args[1].(Reg) == X0 {
			if inst.Args[0].(Reg) == X0 {
				isZihintntl := true
				switch inst.Args[2].(Reg) {
				case X2:
					op = "ntl.p1"
				case X3:
					op = "ntl.pall"
				case X4:
					op = "ntl.s1"
				case X5:
					op = "ntl.all"
				default:
					isZihintntl = false
				}
				if isZihintntl {
					args = args[:0]
				}
			} else {
				op = "mv"
				args[1] = args[2]
				args = args[:len(args)-1]
			}
		}

	case BEQ:
		if inst.Args[1].(Reg) == X0 {
			op = "beqz"
			args[1] = args[2]
			args = args[:len(args)-1]
		}

	case BGE:
		if inst.Args[1].(Reg) == X0 {
			op = "bgez"
			args[1] = args[2]
			args = args[:len(args)-1]
		} else if inst.Args[0].(Reg) == X0 {
			op = "blez"
			args[0], args[1] = args[1], args[2]
			args = args[:len(args)-1]
		}

	case BLT:
		if inst.Args[1].(Reg) == X0 {
			op = "bltz"
			args[1] = args[2]
			args = args[:len(args)-1]
		} else if inst.Args[0].(Reg) == X0 {
			op = "bgtz"
			args[0], args[1] = args[1], args[2]
			args = args[:len(args)-1]
		}

	case BNE:
		if inst.Args[1].(Reg) == X0 {
			op = "bnez"
			args[1] = args[2]
			args = args[:len(args)-1]
		}

	case CSRRC:
		if inst.Args[0].(Reg) == X0 {
			op = "csrc"
			args[0], args[1] = args[1], args[2]
			args = args[:len(args)-1]
		}

	case CSRRCI:
		if inst.Args[0].(Reg) == X0 {
			op = "csrci"
			args[0], args[1] = args[1], args[2]
			args = args[:len(args)-1]
		}

	case CSRRS:
		if inst.Args[2].(Reg) == X0 {
			switch inst.Args[1].(CSR) {
			case FCSR:
				op = "frcsr"
				args = args[:len(args)-2]

			case FFLAGS:
				op = "frflags"
				args = args[:len(args)-2]

			case FRM:
				op = "frrm"
				args = args[:len(args)-2]

			// rdcycleh, rdinstreth and rdtimeh are RV-32 only instructions.
			// So not included there.
			case CYCLE:
				op = "rdcycle"
				args = args[:len(args)-2]

			case INSTRET:
				op = "rdinstret"
				args = args[:len(args)-2]

			case TIME:
				op = "rdtime"
				args = args[:len(args)-2]

			default:
				op = "csrr"
				args = args[:len(args)-1]
			}
		} else if inst.Args[0].(Reg) == X0 {
			op = "csrs"
			args[0], args[1] = args[1], args[2]
			args = args[:len(args)-1]
		}

	case CSRRSI:
		if inst.Args[0].(Reg) == X0 {
			op = "csrsi"
			args[0], args[1] = args[1], args[2]
			args = args[:len(args)-1]
		}

	case CSRRW:
		switch inst.Args[1].(CSR) {
		case FCSR:
			op = "fscsr"
			if inst.Args[0].(Reg) == X0 {
				args[0] = args[2]
				args = args[:len(args)-2]
			} else {
				args[1] = args[2]
				args = args[:len(args)-1]
			}

		case FFLAGS:
			op = "fsflags"
			if inst.Args[0].(Reg) == X0 {
				args[0] = args[2]
				args = args[:len(args)-2]
			} else {
				args[1] = args[2]
				args = args[:len(args)-1]
			}

		case FRM:
			op = "fsrm"
			if inst.Args[0].(Reg) == X0 {
				args[0] = args[2]
				args = args[:len(args)-2]
			} else {
				args[1] = args[2]
				args = args[:len(args)-1]
			}

		case CYCLE:
			if inst.Args[0].(Reg) == X0 && inst.Args[2].(Reg) == X0 {
				op = "unimp"
				args = nil
			}

		default:
			if inst.Args[0].(Reg) == X0 {
				op = "csrw"
				args[0], args[1] = args[1], args[2]
				args = args[:len(args)-1]
			}
		}

	case CSRRWI:
		if inst.Args[0].(Reg) == X0 {
			op = "csrwi"
			args[0], args[1] = args[1], args[2]
			args = args[:len(args)-1]
		}

	// When both pred and succ equals to iorw, the GNU objdump will omit them.
	case FENCE:
		if inst.Args[0].(MemOrder).String() == "iorw" &&
			inst.Args[1].(MemOrder).String() == "iorw" {
			args = nil
		}

	case FSGNJX_D:
		if inst.Args[1].(Reg) == inst.Args[2].(Reg) {
			op = "fabs.d"
			args = args[:len(args)-1]
		}

	case FSGNJX_S:
		if inst.Args[1].(Reg) == inst.Args[2].(Reg) {
			op = "fabs.s"
			args = args[:len(args)-1]
		}

	case FSGNJ_D:
		if inst.Args[1].(Reg) == inst.Args[2].(Reg) {
			op = "fmv.d"
			args = args[:len(args)-1]
		}

	case FSGNJ_S:
		if inst.Args[1].(Reg) == inst.Args[2].(Reg) {
			op = "fmv.s"
			args = args[:len(args)-1]
		}

	case FSGNJN_D:
		if inst.Args[1].(Reg) == inst.Args[2].(Reg) {
			op = "fneg.d"
			args = args[:len(args)-1]
		}

	case FSGNJN_S:
		if inst.Args[1].(Reg) == inst.Args[2].(Reg) {
			op = "fneg.s"
			args = args[:len(args)-1]
		}

	case MOP_RR_N:
		num := (inst.Enc >> 30) & 0x1 << 2
		num |= (inst.Enc >> 27) & 0x1 << 1
		num |= (inst.Enc >> 26) & 0x1 << 0
		op = fmt.Sprintf("mop.rr.%d", num)

		switch num {
		case moprr7:
			if inst.Args[0].(Reg) == X0 && inst.Args[1].(Reg) == X0 {
				switch inst.Args[2].(Reg) {
				case X1:
					op = "sspush.x1"
					args = nil
				case X5:
					op = "sspush.x5"
					args = nil
				}
			}
		}

	case MOP_R_N:
		num := (inst.Enc >> 30) & 0x1 << 4
		num |= (inst.Enc >> 27) & 0x1 << 3
		num |= (inst.Enc >> 26) & 0x1 << 2
		num |= (inst.Enc >> 21) & 0x1 << 1
		num |= (inst.Enc >> 20) & 0x1 << 0
		op = fmt.Sprintf("mop.r.%d", num)

		switch num {
		case mopr28:
			if inst.Args[0].(Reg) == X0 {
				switch inst.Args[1].(Reg) {
				case X1:
					op = "sspopchk.x1"
					args = nil
				case X5:
					op = "sspopchk.x5"
					args = nil
				}
			} else if inst.Args[1].(Reg) == X0 {
				op = "ssrdp"
				args = args[0:1]
			}
		}

	case JAL:
		if inst.Args[0].(Reg) == X0 {
			op = "j"
			args[0] = args[1]
			args = args[:len(args)-1]
		} else if inst.Args[0].(Reg) == X1 {
			op = "jal"
			args[0] = args[1]
			args = args[:len(args)-1]
		}

	case JALR:
		if inst.Args[0].(Reg) == X1 && inst.Args[1].(RegOffset).Ofs.Imm == 0 {
			args[0] = inst.Args[1].(RegOffset).OfsReg.String()
			args = args[:len(args)-1]
		}

		if inst.Args[0].(Reg) == X0 {
			if inst.Args[1].(RegOffset).OfsReg == X1 && inst.Args[1].(RegOffset).Ofs.Imm == 0 {
				op = "ret"
				args = nil
			} else if inst.Args[1].(RegOffset).Ofs.Imm == 0 {
				op = "jr"
				args[0] = inst.Args[1].(RegOffset).OfsReg.String()
				args = args[:len(args)-1]
			} else {
				op = "jr"
				args[0] = inst.Args[1].(RegOffset).String()
				args = args[:len(args)-1]
			}
		}

	case SLTIU:
		if inst.Args[2].(Simm).String() == "1" {
			op = "seqz"
			args = args[:len(args)-1]
		}

	case SLT:
		if inst.Args[1].(Reg) == X0 {
			op = "sgtz"
			args[1] = args[2]
			args = args[:len(args)-1]
		} else if inst.Args[2].(Reg) == X0 {
			op = "sltz"
			args = args[:len(args)-1]
		}

	case SLTU:
		if inst.Args[1].(Reg) == X0 {
			op = "snez"
			args[1] = args[2]
			args = args[:len(args)-1]
		}

	case SUB:
		if inst.Args[1].(Reg) == X0 {
			op = "neg"
			args[1] = args[2]
			args = args[:len(args)-1]
		}

	case SUBW:
		if inst.Args[1].(Reg) == X0 {
			op = "negw"
			args[1] = args[2]
			args = args[:len(args)-1]
		}
	}

	if args != nil {
		op += " " + strings.Join(args, ",")
	}
	return op
}
