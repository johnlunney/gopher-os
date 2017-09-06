package aml

import "testing"

func TestOpcodeToString(t *testing.T) {
	if exp, got := "Acquire", opAcquire.String(); got != exp {
		t.Fatalf("expected opAcquire.toString() to return %q; got %q", exp, got)
	}

	if exp, got := "unknown", opcode(0xffff).String(); got != exp {
		t.Fatalf("expected opcode.String() to return %q; got %q", exp, got)
	}
}

func TestOpcodeIsX(t *testing.T) {
	specs := []struct {
		op     opcode
		testFn func(opcode) bool
		want   bool
	}{
		// opIsLocalArg
		{opLocal0, opIsLocalArg, true},
		{opLocal1, opIsLocalArg, true},
		{opLocal2, opIsLocalArg, true},
		{opLocal3, opIsLocalArg, true},
		{opLocal4, opIsLocalArg, true},
		{opLocal5, opIsLocalArg, true},
		{opLocal6, opIsLocalArg, true},
		{opLocal7, opIsLocalArg, true},
		{opArg0, opIsLocalArg, false},
		{opDivide, opIsLocalArg, false},
		// opIsArg
		{opArg0, opIsArg, true},
		{opArg1, opIsArg, true},
		{opArg2, opIsArg, true},
		{opArg3, opIsArg, true},
		{opArg4, opIsArg, true},
		{opArg5, opIsArg, true},
		{opArg6, opIsArg, true},
		{opLocal7, opIsArg, false},
		{opIf, opIsArg, false},
		// opIsArithmetic
		{opShiftLeft, opIsArithmetic, true},
		{opShiftRight, opIsArithmetic, true},
		{opAnd, opIsArithmetic, true},
		{opOr, opIsArithmetic, true},
		{opNand, opIsArithmetic, true},
		{opNor, opIsArithmetic, true},
		{opXor, opIsArithmetic, true},
		{opNot, opIsArithmetic, true},
		{opIncrement, opIsArithmetic, true},
		{opDecrement, opIsArithmetic, true},
		{opAdd, opIsArithmetic, true},
		{opSubtract, opIsArithmetic, true},
		{opMultiply, opIsArithmetic, true},
		{opMod, opIsArithmetic, true},
		{opDivide, opIsArithmetic, true},
		{opFindSetLeftBit, opIsArithmetic, true},
		{opFindSetRightBit, opIsArithmetic, true},
		{opLocal7, opIsArithmetic, false},
		{opLand, opIsArithmetic, false},
		// opIsLogic
		{opLEqual, opIsLogic, true},
		{opLLess, opIsLogic, true},
		{opLGreater, opIsLogic, true},
		{opLand, opIsLogic, true},
		{opLor, opIsLogic, true},
		{opLnot, opIsLogic, true},
		{opSubtract, opIsLogic, false},
		{opMultiply, opIsLogic, false},
	}

	for specIndex, spec := range specs {
		if got := spec.testFn(spec.op); got != spec.want {
			t.Errorf("[spec %d] opcode %q: expected to get %t; got %t", specIndex, spec.op, spec.want, got)
		}
	}
}

func TestOpArgFlagToString(t *testing.T) {
	specs := map[opArgFlag]string{
		opArgTermList:   "opArgTermList",
		opArgTermObj:    "opArgTermObj",
		opArgByteList:   "opArgByteList",
		opArgPackage:    "opArgPackage",
		opArgString:     "opArgString",
		opArgByteData:   "opArgByteData",
		opArgWord:       "opArgWord",
		opArgDword:      "opArgDword",
		opArgQword:      "opArgQword",
		opArgNameString: "opArgNameString",
		opArgSuperName:  "opArgSuperName",
		opArgSimpleName: "opArgSimpleName",
		opArgDataRefObj: "opArgDataRefObj",
		opArgTarget:     "opArgTarget",
		opArgFieldList:  "opArgFieldList",
		opArgFlag(0xff): "",
	}

	for flag, want := range specs {
		if got := flag.String(); got != want {
			t.Errorf("expected %q; got %q", want, got)
		}
	}
}

// TestFindUnmappedOpcodes is a helper test that pinpoints opcodes that have
// not yet been mapped via an opcode table. This test will be removed once all
// opcodes are supported.
func TestFindUnmappedOpcodes(t *testing.T) {
	//t.SkipNow()
	for opIndex, opRef := range opcodeMap {
		if opRef != badOpcode {
			continue
		}

		for tabIndex, info := range opcodeTable {
			if uint16(info.op) == uint16(opIndex) {
				t.Errorf("set opcodeMap[0x%02x] = 0x%02x // %s\n", opIndex, tabIndex, info.op.String())
				break
			}
		}
	}

	for opIndex, opRef := range extendedOpcodeMap {
		// 0xff (opOnes) is defined in opcodeTable
		if opRef != badOpcode || opIndex == 0 {
			continue
		}

		opIndex += 0xff
		for tabIndex, info := range opcodeTable {
			if uint16(info.op) == uint16(opIndex) {
				t.Errorf("set extendedOpcodeMap[0x%02x] = 0x%02x // %s\n", opIndex-0xff, tabIndex, info.op.String())
				break
			}
		}
	}
}
