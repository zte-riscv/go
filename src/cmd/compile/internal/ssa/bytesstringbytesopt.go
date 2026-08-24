// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/reflectdata"
	"cmd/compile/internal/types"
	"cmd/internal/obj"
)

// bytesStringBytesOpt optimizes []byte(string([]byte)) pattern to use makeslicecopy directly.
func bytesStringBytesOpt(f *Func) {
	if f.Config.arch != "riscv64" {
		return
	}
	if !base.Flag.BytesStringBytesOpt {
		return
	}

	type selectNInfo struct {
		users [4]*Value // SelectN users indexed by AuxInt (0-3)
		count int       // total number of SelectN users
	}
	selectNUsers := make(map[*Value]*selectNInfo)

	for _, b := range f.Blocks {
		for _, v := range b.Values {
			if v.Op != OpSelectN || len(v.Args) != 1 {
				continue
			}
			parent := v.Args[0]
			info := selectNUsers[parent]
			if info == nil {
				info = &selectNInfo{}
				selectNUsers[parent] = info
			}
			idx := auxIntToInt64(v.AuxInt)
			if idx >= 0 && idx < 4 {
				info.users[idx] = v
			}
			info.count++
		}
	}

	type matchInfo struct {
		v       *Value // stringtoslicebyte call
		call    *Value // slicebytetostring call
		sp      *Value // SelectN[0] from call (string ptr)
		sl      *Value // SelectN[1] from call (string len)
		memSel  *Value // SelectN[2] from call (memory)
		ptrSel  *Value // SelectN[0] from v (slice ptr)
		lenSel  *Value // SelectN[1] from v (slice len)
		capSel  *Value // SelectN[2] from v (slice cap)
		memSel2 *Value // SelectN[3] from v (memory)
		ptr     *Value // original byte slice ptr
		length  *Value // original byte slice length
		mem0    *Value // original memory
		block   *Block // block containing v
	}

	var matches []matchInfo

	for _, b := range f.Blocks {
		for _, v := range b.Values {
			if v.Op != OpRISCV64CALLstatic {
				continue
			}
			aux := auxToCall(v.Aux)
			if !isSameCall(aux, "runtime.stringtoslicebyte") {
				continue
			}
			if len(v.Args) != 4 {
				continue
			}

			tmpBuf := v.Args[0]
			sp := v.Args[1]
			sl := v.Args[2]
			memSel := v.Args[3]

			isNil := tmpBuf.Op == OpConstNil ||
				(tmpBuf.Op == OpRISCV64MOVDconst && auxIntToInt64(tmpBuf.AuxInt) == 0)
			if !isNil {
				continue
			}

			if sp.Op != OpSelectN || sl.Op != OpSelectN || memSel.Op != OpSelectN {
				continue
			}
			if auxIntToInt64(sp.AuxInt) != 0 || auxIntToInt64(sl.AuxInt) != 1 || auxIntToInt64(memSel.AuxInt) != 2 {
				continue
			}
			call := sp.Args[0]
			if call != sl.Args[0] || call != memSel.Args[0] {
				continue
			}
			if call.Op != OpRISCV64CALLstatic {
				continue
			}
			callAux := auxToCall(call.Aux)
			if !isSameCall(callAux, "runtime.slicebytetostring") {
				continue
			}
			if len(call.Args) != 4 {
				continue
			}
			ptr := call.Args[1]
			length := call.Args[2]
			mem0 := call.Args[3]

			if sp.Uses != 1 || sl.Uses != 1 || memSel.Uses != 1 {
				continue
			}

			callInfo := selectNUsers[call]
			if callInfo == nil || callInfo.count != 3 {
				continue
			}

			if call.Block != b {
				continue
			}

			vInfo := selectNUsers[v]
			if vInfo == nil || vInfo.count != 4 {
				continue
			}
			ptrSel := vInfo.users[0]
			lenSel := vInfo.users[1]
			capSel := vInfo.users[2]
			memSel2 := vInfo.users[3]

			if ptrSel == nil || lenSel == nil || capSel == nil || memSel2 == nil {
				continue
			}
			if memSel2.Uses == 0 {
				continue
			}

			matches = append(matches, matchInfo{
				v: v, call: call,
				sp: sp, sl: sl, memSel: memSel,
				ptrSel: ptrSel, lenSel: lenSel, capSel: capSel, memSel2: memSel2,
				ptr: ptr, length: length, mem0: mem0,
				block: b,
			})
		}
	}

	for _, m := range matches {
		b := m.block
		v := m.v
		call := m.call
		ptr := m.ptr
		length := m.length
		mem0 := m.mem0
		ptrSel := m.ptrSel
		lenSel := m.lenSel
		capSel := m.capSel
		memSel2 := m.memSel2
		sp := m.sp
		sl := m.sl
		memSel := m.memSel

		cfgTypes := &f.Config.Types
		uptrType := types.Types[types.TUNSAFEPTR]

		sb := b.NewValue0(v.Pos, OpSB, cfgTypes.Uintptr)
		taddr := b.NewValue1A(v.Pos, OpRISCV64MOVaddr, cfgTypes.BytePtr, reflectdata.TypeLinksym(types.Types[types.TUINT8]), sb)

		from := ptr
		if ptr.Type != uptrType {
			from = b.NewValue2(v.Pos, OpConvert, uptrType, ptr, mem0)
		}

		argTypes := []*types.Type{cfgTypes.BytePtr, cfgTypes.Int, cfgTypes.Int, uptrType}
		resTypes := []*types.Type{uptrType}
		sym := f.Config.ctxt.LookupABI("runtime.makeslicecopy", obj.ABIInternal)
		aux2 := StaticAuxCall(sym, f.ABI1.ABIAnalyzeTypes(argTypes, resTypes))
		mc := b.NewValue0A(v.Pos, OpRISCV64CALLstatic, aux2.LateExpansionResultType(), aux2)
		mc.AddArg5(taddr, length, length, from, mem0)
		mc.AuxInt = aux2.ArgWidth()

		rawPtr := b.NewValue0(v.Pos, OpSelectN, uptrType)
		rawPtr.AuxInt = 0
		rawPtr.AddArg(mc)

		mem1 := b.NewValue0(v.Pos, OpSelectN, types.TypeMem)
		mem1.AuxInt = int64ToAuxInt(1)
		mem1.AddArg(mc)

		slicePtrType := v.Type.FieldType(0)
		newPtr := b.NewValue2(v.Pos, OpConvert, slicePtrType, rawPtr, mem1)

		ptrSel.copyOf(newPtr)
		lenSel.copyOf(length)
		capSel.copyOf(length)
		memSel2.copyOf(mem1)

		// Dead intermediate values: use type-compatible replacements to
		// properly disconnect reference counts while preserving type
		// consistency until the next deadcode pass cleans them up.
		sp.copyOf(ptr)
		sl.copyOf(length)
		memSel.copyOf(mem0)

		v.copyOf(mem1)
		call.copyOf(mem0)
	}
}
