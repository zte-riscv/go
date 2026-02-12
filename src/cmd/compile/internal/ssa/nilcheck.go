// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import (
	"cmd/compile/internal/ir"
	"cmd/internal/src"
	"fmt"
	"internal/buildcfg"
)

// nilcheckelim eliminates unnecessary nil checks.
// runs on machine-independent code.
func nilcheckelim(f *Func) {
	// A nil check is redundant if the same nil check was successful in a
	// dominating block. The efficacy of this pass depends heavily on the
	// efficacy of the cse pass.
	sdom := f.Sdom()

	// TODO: Eliminate more nil checks.
	// We can recursively remove any chain of fixed offset calculations,
	// i.e. struct fields and array elements, even with non-constant
	// indices: x is non-nil iff x.a.b[i].c is.

	type walkState int
	const (
		Work     walkState = iota // process nil checks and traverse to dominees
		ClearPtr                  // forget the fact that ptr is nil
	)

	type bp struct {
		block *Block // block, or nil in ClearPtr state
		ptr   *Value // if non-nil, ptr that is to be cleared in ClearPtr state
		op    walkState
	}

	work := make([]bp, 0, 256)
	work = append(work, bp{block: f.Entry})

	// map from value ID to known non-nil version of that value ID
	// (in the current dominator path being walked). This slice is updated by
	// walkStates to maintain the known non-nil values.
	// If there is extrinsic information about non-nil-ness, this map
	// points a value to itself. If a value is known non-nil because we
	// already did a nil check on it, it points to the nil check operation.
	nonNilValues := f.Cache.allocValueSlice(f.NumValues())
	defer f.Cache.freeValueSlice(nonNilValues)

	// make an initial pass identifying any non-nil values
	for _, b := range f.Blocks {
		for _, v := range b.Values {
			// a value resulting from taking the address of a
			// value, or a value constructed from an offset of a
			// non-nil ptr (OpAddPtr) implies it is non-nil
			// We also assume unsafe pointer arithmetic generates non-nil pointers. See #27180.
			// We assume that SlicePtr is non-nil because we do a bounds check
			// before the slice access (and all cap>0 slices have a non-nil ptr). See #30366.
			if v.Op == OpAddr || v.Op == OpLocalAddr || v.Op == OpAddPtr || v.Op == OpOffPtr || v.Op == OpAdd32 || v.Op == OpAdd64 || v.Op == OpSub32 || v.Op == OpSub64 || v.Op == OpSlicePtr {
				nonNilValues[v.ID] = v
			}
		}
	}

	for changed := true; changed; {
		changed = false
		for _, b := range f.Blocks {
			for _, v := range b.Values {
				// phis whose arguments are all non-nil
				// are non-nil
				if v.Op == OpPhi {
					argsNonNil := true
					for _, a := range v.Args {
						if nonNilValues[a.ID] == nil {
							argsNonNil = false
							break
						}
					}
					if argsNonNil {
						if nonNilValues[v.ID] == nil {
							changed = true
						}
						nonNilValues[v.ID] = v
					}
				}
			}
		}
	}

	// allocate auxiliary date structures for computing store order
	sset := f.newSparseSet(f.NumValues())
	defer f.retSparseSet(sset)
	storeNumber := f.Cache.allocInt32Slice(f.NumValues())
	defer f.Cache.freeInt32Slice(storeNumber)

	// perform a depth first walk of the dominee tree
	for len(work) > 0 {
		node := work[len(work)-1]
		work = work[:len(work)-1]

		switch node.op {
		case Work:
			b := node.block

			// First, see if we're dominated by an explicit nil check.
			if len(b.Preds) == 1 {
				p := b.Preds[0].b
				if p.Kind == BlockIf && p.Controls[0].Op == OpIsNonNil && p.Succs[0].b == b {
					if ptr := p.Controls[0].Args[0]; nonNilValues[ptr.ID] == nil {
						nonNilValues[ptr.ID] = ptr
						work = append(work, bp{op: ClearPtr, ptr: ptr})
					}
				}
			}

			// Next, order values in the current block w.r.t. stores.
			b.Values = storeOrder(b.Values, sset, storeNumber)

			pendingLines := f.cachedLineStarts // Holds statement boundaries that need to be moved to a new value/block
			pendingLines.clear()

			// Next, process values in the block.
			for _, v := range b.Values {
				switch v.Op {
				case OpIsNonNil:
					ptr := v.Args[0]
					if nonNilValues[ptr.ID] != nil {
						if v.Pos.IsStmt() == src.PosIsStmt { // Boolean true is a terrible statement boundary.
							pendingLines.add(v.Pos)
							v.Pos = v.Pos.WithNotStmt()
						}
						// This is a redundant explicit nil check.
						v.reset(OpConstBool)
						v.AuxInt = 1 // true
					}
				case OpNilCheck:
					ptr := v.Args[0]
					if nilCheck := nonNilValues[ptr.ID]; nilCheck != nil {
						// This is a redundant implicit nil check.
						// Logging in the style of the former compiler -- and omit line 1,
						// which is usually in generated code.
						if f.fe.Debug_checknil() && v.Pos.Line() > 1 {
							f.Warnl(v.Pos, "removed nil check")
						}
						if v.Pos.IsStmt() == src.PosIsStmt { // About to lose a statement boundary
							pendingLines.add(v.Pos)
						}
						v.Op = OpCopy
						v.SetArgs1(nilCheck)
						continue
					}
					// Record the fact that we know ptr is non nil, and remember to
					// undo that information when this dominator subtree is done.
					nonNilValues[ptr.ID] = v
					work = append(work, bp{op: ClearPtr, ptr: ptr})
					fallthrough // a non-eliminated nil check might be a good place for a statement boundary.
				default:
					if v.Pos.IsStmt() != src.PosNotStmt && !isPoorStatementOp(v.Op) && pendingLines.contains(v.Pos) {
						v.Pos = v.Pos.WithIsStmt()
						pendingLines.remove(v.Pos)
					}
				}
			}
			// This reduces the lost statement count in "go" by 5 (out of 500 total).
			for j := range b.Values { // is this an ordering problem?
				v := b.Values[j]
				if v.Pos.IsStmt() != src.PosNotStmt && !isPoorStatementOp(v.Op) && pendingLines.contains(v.Pos) {
					v.Pos = v.Pos.WithIsStmt()
					pendingLines.remove(v.Pos)
				}
			}
			if pendingLines.contains(b.Pos) {
				b.Pos = b.Pos.WithIsStmt()
				pendingLines.remove(b.Pos)
			}

			// Add all dominated blocks to the work list.
			for w := sdom[node.block.ID].child; w != nil; w = sdom[w.ID].sibling {
				work = append(work, bp{op: Work, block: w})
			}

		case ClearPtr:
			nonNilValues[node.ptr.ID] = nil
			continue
		}
	}

	// Optimize redundant Store-Move-Dereference chains for FieldByName functions
	// This should be done at the end of nilcheckelim, before expandCalls
	optimizeFieldByNameReturnValues(f)
}

// optimizeFieldByNameReturnValues eliminates redundant Store-Move-Dereference chains
// for FieldByName functions. Pattern:
//
//	SelectN -> Store(.autotmp) -> Move(.autotmp -> .autotmp2) -> Move(.autotmp2 -> returnLoc) -> Dereference(returnLoc) -> MakeResult
//
// Optimization: Directly use SelectN in MakeResult, skipping all intermediate operations.
func optimizeFieldByNameReturnValues(f *Func) {
	// Check if this is a FieldByName function
	if f.Name != "(*structType).FieldByName" && f.Name != "(*structType).FieldByNameFunc" {
		return
	}

	// Find all MakeResult operations
	for _, b := range f.Blocks {
		for _, v := range b.Values {
			if v.Op != OpMakeResult {
				continue
			}

			// Process each argument of MakeResult
			argsWithoutMem := v.Args[:len(v.Args)-1]
			optimized := false
			newArgs := make([]*Value, 0, len(v.Args))

			for _, arg := range argsWithoutMem {
				// Check if arg is Dereference(returnLoc, mem)
				if arg.Op != OpDereference || len(arg.Args) < 2 {
					newArgs = append(newArgs, arg)
					continue
				}

				returnLoc := arg.Args[0]

				// Check if returnLoc is LocalAddr (return value location)
				if returnLoc.Op != OpLocalAddr {
					newArgs = append(newArgs, arg)
					continue
				}

				// Search for Move operation that writes to returnLoc
				// We need to search in the same block and earlier blocks
				var moveToReturnLoc *Value
				var tempAddr *Value

				// Search backwards from the current block
				for searchBlock := b; searchBlock != nil; {
					for i := len(searchBlock.Values) - 1; i >= 0; i-- {
						v := searchBlock.Values[i]
						if v.Op == OpMove && len(v.Args) >= 2 {
							dst := v.Args[0]
							src := v.Args[1]

							// Check if this Move writes to returnLoc
							if dst == returnLoc {
								moveToReturnLoc = v
								tempAddr = src
								break
							}
						}
					}
					if moveToReturnLoc != nil {
						break
					}
					// Only search one predecessor to avoid complexity
					if len(searchBlock.Preds) == 0 {
						break
					}
					searchBlock = searchBlock.Preds[0].b
				}

				if moveToReturnLoc == nil || tempAddr == nil {
					newArgs = append(newArgs, arg)
					continue
				}

				// Check if tempAddr is LocalAddr (temporary variable)
				if tempAddr.Op != OpLocalAddr {
					newArgs = append(newArgs, arg)
					continue
				}

				// Find the Store operation that writes SelectN to tempAddr
				// We may need to trace through multiple Move operations
				// Pattern: Store(.autotmp_21) -> Move(.autotmp_21 -> .autotmp_14) -> Move(.autotmp_14 -> f)
				// Or: Store(.autotmp_21) -> Move(.autotmp_21 -> f)
				var selectNValue *Value
				currentAddr := tempAddr
				maxMoves := 5 // Limit to avoid infinite loops
				moveCount := 0

				// Trace back through Move operations to find the original Store
				foundNext := false
				for moveCount < maxMoves {
					// Search backwards from the Move operation
					moveBlock := moveToReturnLoc.Block
					foundNext = false

					for searchBlock := moveBlock; searchBlock != nil; {
						startIdx := len(searchBlock.Values)
						if searchBlock == moveBlock && moveCount == 0 {
							// In the same block, only search before the Move
							for i, v := range searchBlock.Values {
								if v == moveToReturnLoc {
									startIdx = i
									break
								}
							}
						}

						for i := startIdx - 1; i >= 0; i-- {
							v := searchBlock.Values[i]

							// Check for Store operation
							if v.Op == OpStore && len(v.Args) >= 3 {
								storeAddr := v.Args[0]
								storeValue := v.Args[1]

								// Check if this Store writes to currentAddr
								if storeAddr == currentAddr {
									// Check if storeValue is SelectN
									if storeValue.Op == OpSelectN {
										// Only optimize if SelectN is only used by this Store
										// (Uses == 1), so that when we replace Dereference with SelectN,
										// the Store will become unused and can be removed by deadcode.
										// If SelectN has multiple uses, expand_calls will panic.
										if storeValue.Uses == 1 {
											selectNValue = storeValue
											foundNext = true
											break
										}
									}
								}
							}

							// Check for Move operation that might be in the chain
							if v.Op == OpMove && len(v.Args) >= 2 {
								dst := v.Args[0]
								src := v.Args[1]
								// If this Move writes to currentAddr, trace back to its source
								if dst == currentAddr && src.Op == OpLocalAddr {
									currentAddr = src
									moveCount++
									foundNext = true
									break
								}
							}
						}

						if selectNValue != nil {
							break
						}

						if foundNext {
							// Found a Move in the chain, continue searching in the same block
							break
						}

						// Only search one predecessor to avoid complexity
						if len(searchBlock.Preds) == 0 {
							break
						}
						searchBlock = searchBlock.Preds[0].b
					}

					if selectNValue != nil {
						break
					}

					// If we didn't find a Store or another Move, give up
					if !foundNext {
						break
					}
				}

				if selectNValue != nil {
					// Found the pattern! Before replacing Dereference with SelectN,
					// we need to mark the Store that uses this SelectN as OpInvalid,
					// so that when expand_calls runs, it won't see the SelectN as having multiple uses.
					// We need to find the Store that uses this SelectN and update the memory chain.
					isFieldByName := f.Name == "(*structType).FieldByName" || f.Name == "(*structType).FieldByNameFunc"
					if isFieldByName {
						fmt.Printf("[nilcheck] %s: Found SelectN v%d (Uses=%d)\n", f.Name, selectNValue.ID, selectNValue.Uses)
					}
					var storeOp *Value
					for _, b2 := range f.Blocks {
						for _, v2 := range b2.Values {
							if v2.Op == OpStore && len(v2.Args) >= 3 && v2.Args[1] == selectNValue {
								storeOp = v2
								if isFieldByName {
									fmt.Printf("[nilcheck] %s: Found Store v%d (Uses=%d) using SelectN\n", f.Name, storeOp.ID, storeOp.Uses)
								}
								break
							}
						}
						if storeOp != nil {
							break
						}
					}

					if storeOp != nil {
						// Get the input memory (last argument for Store)
						inputMem := storeOp.Args[2]

						// Find all values that use this Store's memory output
						// and replace them with the input memory
						// Note: SetArg automatically updates Uses counts, so we don't need to manually adjust them
						replacedCount := 0
						for _, b3 := range f.Blocks {
							for _, v3 := range b3.Values {
								// Check all arguments for memory uses
								for i, a := range v3.Args {
									if a == storeOp && a.Type.IsMemory() {
										// Replace with input memory
										// SetArg will automatically:
										// - decrease storeOp.Uses (because it's no longer an arg)
										// - increase inputMem.Uses (because it's now an arg)
										v3.SetArg(i, inputMem)
										replacedCount++
									}
								}
							}
						}
						if isFieldByName {
							fmt.Printf("[nilcheck] %s: Replaced %d memory uses, Store v%d.Uses=%d\n", f.Name, replacedCount, storeOp.ID, storeOp.Uses)
						}

						// Store's Uses includes:
						// 1. Memory output uses (we just replaced these)
						// 2. Argument uses (SelectN in Args[1], inputMem in Args[2], etc.)
						// After replacing memory uses, if Store.Uses equals the number of argument uses,
						// then Store has no memory output uses left, and we can mark it as OpInvalid.
						// For Store, it has 3 args: [0]=addr, [1]=value (SelectN), [2]=inputMem
						// So if Uses == 2 (SelectN + inputMem), then no memory output uses remain.
						// But actually, we need to check: if Store is only used by its arguments, then
						// it has no "real" uses (memory output uses), so we can invalidate it.

						// Actually, the correct approach is: Store.Uses should only count memory output uses.
						// Argument uses are tracked separately. So if we've replaced all memory uses,
						// Store.Uses should be 0 (or negative if we over-decremented).
						// Let's just check if Store.Uses <= 0, meaning no memory output uses remain.
						if isFieldByName {
							fmt.Printf("[nilcheck] %s: Checking Store v%d.Uses=%d (<=0? %v)\n", f.Name, storeOp.ID, storeOp.Uses, storeOp.Uses <= 0)
						}
						if storeOp.Uses <= 0 {
							if isFieldByName {
								fmt.Printf("[nilcheck] %s: Marking Store v%d as OpInvalid, SelectN v%d Uses: %d -> ", f.Name, storeOp.ID, selectNValue.ID, selectNValue.Uses)
							}
							storeOp.Op = OpInvalid
							storeOp.resetArgs()
							if isFieldByName {
								fmt.Printf("%d\n", selectNValue.Uses)
							}
							// After resetArgs(), SelectN's Uses has been decreased by 1
							// (from 1 to 0, since Store was the only user)
						} else {
							if isFieldByName {
								fmt.Printf("[nilcheck] %s: ERROR: Store v%d still has %d uses (not <= 0)!\n", f.Name, storeOp.ID, storeOp.Uses)
							}
						}
					}

					// Now replace Dereference with SelectN
					// When we later call v.AddArgs(newArgs...), SelectN's Uses will increase back to 1
					if isFieldByName {
						fmt.Printf("[nilcheck] %s: Adding SelectN v%d to MakeResult, Uses=%d\n", f.Name, selectNValue.ID, selectNValue.Uses)
					}
					newArgs = append(newArgs, selectNValue)
					optimized = true

					// Note: We don't mark Move operations as OpInvalid here
					// because they are memory operations and their memory outputs may still be used.
					// The deadcode pass will remove them if they become unused.
				} else {
					newArgs = append(newArgs, arg)
				}
			}

			// If we optimized, update MakeResult arguments
			if optimized {
				// We need to find the correct memory argument to use
				// Since we're replacing Dereference with SelectN, we should use
				// the memory from the SelectN's call, not from Dereference
				var memArg *Value
				var sourceCall *Value

				// Find the source call from the SelectN values
				for _, newArg := range newArgs {
					if newArg.Op == OpSelectN && !newArg.Type.IsMemory() && len(newArg.Args) > 0 {
						sourceCall = newArg.Args[0]
						break
					}
				}

				// Find the memory SelectN for this call
				if sourceCall != nil {
					// Search in the same block first (most common case)
					for _, v2 := range b.Values {
						if v2.Op == OpSelectN && v2.Type.IsMemory() && len(v2.Args) > 0 && v2.Args[0] == sourceCall {
							memArg = v2
							break
						}
					}
					// If not found in the same block, search in other blocks
					if memArg == nil {
						for _, b2 := range f.Blocks {
							if b2 == b {
								continue // already searched
							}
							for _, v2 := range b2.Values {
								if v2.Op == OpSelectN && v2.Type.IsMemory() && len(v2.Args) > 0 && v2.Args[0] == sourceCall {
									memArg = v2
									break
								}
							}
							if memArg != nil {
								break
							}
						}
					}
				}

				// If we didn't find memory from SelectN, use the original memory argument
				if memArg == nil {
					memArg = v.Args[len(v.Args)-1]
				}
				v.resetArgs()
				v.AddArgs(newArgs...)
				v.AddArg(memArg)
			}
		}
	}
}

// All platforms are guaranteed to fault if we load/store to anything smaller than this address.
//
// This should agree with minLegalPointer in the runtime.
const minZeroPage = 4096

// faultOnLoad is true if a load to an address below minZeroPage will trigger a SIGSEGV.
var faultOnLoad = buildcfg.GOOS != "aix"

// nilcheckelim2 eliminates unnecessary nil checks.
// Runs after lowering and scheduling.
func nilcheckelim2(f *Func) {
	unnecessary := f.newSparseMap(f.NumValues()) // map from pointer that will be dereferenced to index of dereferencing value in b.Values[]
	defer f.retSparseMap(unnecessary)

	pendingLines := f.cachedLineStarts // Holds statement boundaries that need to be moved to a new value/block

	for _, b := range f.Blocks {
		// Walk the block backwards. Find instructions that will fault if their
		// input pointer is nil. Remove nil checks on those pointers, as the
		// faulting instruction effectively does the nil check for free.
		unnecessary.clear()
		pendingLines.clear()
		// Optimization: keep track of removed nilcheck with smallest index
		firstToRemove := len(b.Values)
		for i := len(b.Values) - 1; i >= 0; i-- {
			v := b.Values[i]
			if opcodeTable[v.Op].nilCheck && unnecessary.contains(v.Args[0].ID) {
				if f.fe.Debug_checknil() && v.Pos.Line() > 1 {
					f.Warnl(v.Pos, "removed nil check")
				}
				// For bug 33724, policy is that we might choose to bump an existing position
				// off the faulting load in favor of the one from the nil check.

				// Iteration order means that first nilcheck in the chain wins, others
				// are bumped into the ordinary statement preservation algorithm.
				u := b.Values[unnecessary.get(v.Args[0].ID)]
				if !u.Type.IsMemory() && !u.Pos.SameFileAndLine(v.Pos) {
					if u.Pos.IsStmt() == src.PosIsStmt {
						pendingLines.add(u.Pos)
					}
					u.Pos = v.Pos
				} else if v.Pos.IsStmt() == src.PosIsStmt {
					pendingLines.add(v.Pos)
				}

				v.reset(OpUnknown)
				firstToRemove = i
				continue
			}
			if v.Type.IsMemory() || v.Type.IsTuple() && v.Type.FieldType(1).IsMemory() {
				if v.Op == OpVarLive || (v.Op == OpVarDef && !v.Aux.(*ir.Name).Type().HasPointers()) {
					// These ops don't really change memory.
					continue
					// Note: OpVarDef requires that the defined variable not have pointers.
					// We need to make sure that there's no possible faulting
					// instruction between a VarDef and that variable being
					// fully initialized. If there was, then anything scanning
					// the stack during the handling of that fault will see
					// a live but uninitialized pointer variable on the stack.
					//
					// If we have:
					//
					//   NilCheck p
					//   VarDef x
					//   x = *p
					//
					// We can't rewrite that to
					//
					//   VarDef x
					//   NilCheck p
					//   x = *p
					//
					// Particularly, even though *p faults on p==nil, we still
					// have to do the explicit nil check before the VarDef.
					// See issue #32288.
				}
				// This op changes memory.  Any faulting instruction after v that
				// we've recorded in the unnecessary map is now obsolete.
				unnecessary.clear()
			}

			// Find any pointers that this op is guaranteed to fault on if nil.
			var ptrstore [2]*Value
			ptrs := ptrstore[:0]
			if opcodeTable[v.Op].faultOnNilArg0 && (faultOnLoad || v.Type.IsMemory()) {
				// On AIX, only writing will fault.
				ptrs = append(ptrs, v.Args[0])
			}
			if opcodeTable[v.Op].faultOnNilArg1 && (faultOnLoad || (v.Type.IsMemory() && v.Op != OpPPC64LoweredMove)) {
				// On AIX, only writing will fault.
				// LoweredMove is a special case because it's considered as a "mem" as it stores on arg0 but arg1 is accessed as a load and should be checked.
				ptrs = append(ptrs, v.Args[1])
			}

			for _, ptr := range ptrs {
				// Check to make sure the offset is small.
				switch opcodeTable[v.Op].auxType {
				case auxSym:
					if v.Aux != nil {
						continue
					}
				case auxSymOff:
					if v.Aux != nil || v.AuxInt < 0 || v.AuxInt >= minZeroPage {
						continue
					}
				case auxSymValAndOff:
					off := ValAndOff(v.AuxInt).Off()
					if v.Aux != nil || off < 0 || off >= minZeroPage {
						continue
					}
				case auxInt32:
					// Mips uses this auxType for atomic add constant. It does not affect the effective address.
				case auxInt64:
					// ARM uses this auxType for duffcopy/duffzero/alignment info.
					// It does not affect the effective address.
				case auxNone:
					// offset is zero.
				default:
					v.Fatalf("can't handle aux %s (type %d) yet\n", v.auxString(), int(opcodeTable[v.Op].auxType))
				}
				// This instruction is guaranteed to fault if ptr is nil.
				// Any previous nil check op is unnecessary.
				unnecessary.set(ptr.ID, int32(i))
			}
		}
		// Remove values we've clobbered with OpUnknown.
		i := firstToRemove
		for j := i; j < len(b.Values); j++ {
			v := b.Values[j]
			if v.Op != OpUnknown {
				if !notStmtBoundary(v.Op) && pendingLines.contains(v.Pos) { // Late in compilation, so any remaining NotStmt values are probably okay now.
					v.Pos = v.Pos.WithIsStmt()
					pendingLines.remove(v.Pos)
				}
				b.Values[i] = v
				i++
			}
		}

		if pendingLines.contains(b.Pos) {
			b.Pos = b.Pos.WithIsStmt()
		}

		b.truncateValues(i)

		// TODO: if b.Kind == BlockPlain, start the analysis in the subsequent block to find
		// more unnecessary nil checks.  Would fix test/nilptr3.go:159.
	}
}
