package aml

import (
	"gopheros/kernel/kfmt"
	"io"
)

type resolver interface {
	Resolve(io.Writer, ScopeEntity) bool
}

type Entity interface {
	getOpcode() opcode
	setOpcode(opcode)
	Name() string
	Parent() ScopeEntity
	setParent(ScopeEntity)
	getArgs() []interface{}
	setArg(uint8, interface{}) bool
}

type ScopeEntity interface {
	Entity

	Children() []Entity
	Append(Entity) bool
	removeLastChild() Entity
	lastChild() Entity
}

// unnamedEntity defines an unnamed entity that can be attached to a parent scope.
type unnamedEntity struct {
	op     opcode
	args   []interface{}
	parent ScopeEntity
}

func (ent *unnamedEntity) getOpcode() opcode            { return ent.op }
func (ent *unnamedEntity) setOpcode(op opcode)          { ent.op = op }
func (ent *unnamedEntity) Name() string                 { return "" }
func (ent *unnamedEntity) Parent() ScopeEntity          { return ent.parent }
func (ent *unnamedEntity) setParent(parent ScopeEntity) { ent.parent = parent }
func (ent *unnamedEntity) getArgs() []interface{}       { return ent.args }
func (ent *unnamedEntity) setArg(_ uint8, arg interface{}) bool {
	ent.args = append(ent.args, arg)
	return true
}

// namedEntity is a named entity that can be attached to the parent scope. The
// setArg() implementation for this type expects arg at index 0 to contain the
// entity name.
type namedEntity struct {
	op     opcode
	args   []interface{}
	parent ScopeEntity

	name string
}

func (ent *namedEntity) getOpcode() opcode            { return ent.op }
func (ent *namedEntity) setOpcode(op opcode)          { ent.op = op }
func (ent *namedEntity) Name() string                 { return ent.name }
func (ent *namedEntity) Parent() ScopeEntity          { return ent.parent }
func (ent *namedEntity) setParent(parent ScopeEntity) { ent.parent = parent }
func (ent *namedEntity) getArgs() []interface{}       { return ent.args }
func (ent *namedEntity) setArg(argIndex uint8, arg interface{}) bool {
	// arg 0 is the entity name
	if argIndex == 0 {
		var ok bool
		ent.name, ok = arg.(string)
		return ok
	}

	ent.args = append(ent.args, arg)
	return true
}

// constEntity is an unnamedEntity which always evaluates to a constant value.
// Calls to setArg for argument index 0 will memoize the argument value that is
// stored inside this entity.
type constEntity struct {
	op     opcode
	args   []interface{}
	parent ScopeEntity

	val interface{}
}

func (ent *constEntity) getOpcode() opcode { return ent.op }
func (ent *constEntity) setOpcode(op opcode) {
	ent.op = op

	// special const opcode cases that have an implicit value
	switch ent.op {
	case opZero:
		ent.val = uint64(0)
	case opOne:
		ent.val = uint64(1)
	case opOnes:
		ent.val = uint64(1<<64 - 1)
	}
}
func (ent *constEntity) Name() string                 { return "" }
func (ent *constEntity) Parent() ScopeEntity          { return ent.parent }
func (ent *constEntity) setParent(parent ScopeEntity) { ent.parent = parent }
func (ent *constEntity) getArgs() []interface{}       { return ent.args }
func (ent *constEntity) setArg(argIndex uint8, arg interface{}) bool {
	ent.val = arg
	return argIndex == 0
}

// scopeEntity is an optionally named entity that defines a scope.
type scopeEntity struct {
	op     opcode
	args   []interface{}
	parent ScopeEntity

	name     string
	children []Entity
}

func (ent *scopeEntity) getOpcode() opcode            { return ent.op }
func (ent *scopeEntity) setOpcode(op opcode)          { ent.op = op }
func (ent *scopeEntity) Name() string                 { return ent.name }
func (ent *scopeEntity) Parent() ScopeEntity          { return ent.parent }
func (ent *scopeEntity) setParent(parent ScopeEntity) { ent.parent = parent }
func (ent *scopeEntity) getArgs() []interface{}       { return ent.args }
func (ent *scopeEntity) setArg(argIndex uint8, arg interface{}) bool {
	// arg 0 *may* be the entity name. If it's not a string just add it to
	// the arg list.
	if argIndex == 0 {
		var ok bool
		if ent.name, ok = arg.(string); ok {
			return true
		}
	}

	ent.args = append(ent.args, arg)
	return true
}
func (ent *scopeEntity) Children() []Entity { return ent.children }
func (ent *scopeEntity) Append(child Entity) bool {
	child.setParent(ent)
	ent.children = append(ent.children, child)
	return true
}
func (ent *scopeEntity) lastChild() Entity { return ent.children[len(ent.children)-1] }
func (ent *scopeEntity) removeLastChild() Entity {
	lastIndex := len(ent.children) - 1
	child := ent.children[lastIndex]
	ent.children = ent.children[:lastIndex]
	return child
}

// bufferEntity defines a buffer object.
type bufferEntity struct {
	unnamedEntity

	size interface{}
	data []byte
}

func (ent *bufferEntity) setArg(argIndex uint8, arg interface{}) bool {
	switch argIndex {
	case 0: // size
		ent.size = arg
		return true
	case 1: // data
		if byteSlice, ok := arg.([]byte); ok {
			ent.data = byteSlice
			return true
		}
	}

	return false
}

// bufferFieldEntity describes a bit/byte/word/dword/qword or arbitrary length
// buffer field.
type bufferFieldEntity struct {
	namedEntity
}

func (ent *bufferFieldEntity) setArg(argIndex uint8, arg interface{}) bool {
	// opCreateField specifies the name using the arg at index 3 while
	// opCreateXXXField (byte, word e.t.c) specifies the name using the
	// arg at index 2
	if (ent.op == opCreateField && argIndex == 3) || argIndex == 2 {
		var ok bool
		ent.name, ok = arg.(string)
		return ok
	}
	ent.args = append(ent.args, arg)
	return true
}

type regionSpace uint8

const (
	regionSpaceSystemMemory regionSpace = iota
	regionSpaceSystemIO
	regionSpacePCIConfig
	regionSpaceEmbeddedControl
	regionSpaceSMBus
	regionSpacePCIBarTarget
	regionSpaceIPMI
)

// regionEntity defines a region located at a particular space (e.g in memory,
// an embedded controller, the SMBus e.t.c).
type regionEntity struct {
	namedEntity
}

type fieldAccessType uint8

const (
	fieldAccessTypeAny fieldAccessType = iota
	fieldAccessTypeByte
	fieldAccessTypeWord
	fieldAccessTypeDword
	fieldAccessTypeQword
	fieldAccessTypeBuffer
)

type fieldUpdateRule uint8

const (
	fieldUpdateRulePreserve fieldUpdateRule = iota
	fieldUpdateRuleWriteAsOnes
	fieldUpdateRuleWriteAsZeros
)

type fieldAccessAttrib uint8

const (
	fieldAccessAttribQuick            fieldAccessAttrib = 0x02
	fieldAccessAttribSendReceive                        = 0x04
	fieldAccessAttribByte                               = 0x06
	fieldAccessAttribWord                               = 0x08
	fieldAccessAttribBlock                              = 0x0a
	fieldAccessAttribBytes                              = 0x0b // byteCount contains the number of bytes
	fieldAccessAttribProcessCall                        = 0x0c
	fieldAccessAttribBlockProcessCall                   = 0x0d
	fieldAccessAttribRawBytes                           = 0x0e // byteCount contains the number of bytes
	fieldAccessAttribRawProcessBytes                    = 0x0f // byteCount contains the number of bytes
)

// fieldEntity is a named object that encapsulates the data shared between regular
// fields and index fields.
type fieldEntity struct {
	namedEntity

	bitOffset uint32
	bitWidth  uint32

	lock       bool
	updateRule fieldUpdateRule

	// accessAttrib is valid if accessType is BufferAcc
	// for the SMB or GPIO OpRegions.
	accessAttrib fieldAccessAttrib
	accessType   fieldAccessType

	// byteCount is valid when accessAttrib is one of:
	// Bytes, RawBytes or RawProcessBytes
	byteCount uint8
}

// fieldUnitEntity is a field defined inside an operating region.
type fieldUnitEntity struct {
	fieldEntity

	// The connection which this field references.
	connectionName     string
	resolvedConnection Entity

	// The region which this field references.
	regionName     string
	resolvedRegion *regionEntity
}

func (ent *fieldUnitEntity) Resolve(errWriter io.Writer, rootNs ScopeEntity) bool {
	var ok bool
	if ent.connectionName != "" && ent.resolvedConnection == nil {
		if ent.resolvedConnection = scopeFind(ent.parent, rootNs, ent.connectionName); ent.resolvedConnection == nil {
			kfmt.Fprintf(errWriter, "[field %s] could not resolve connection reference: %s\n", ent.name, ent.connectionName)
			return false
		}
	}

	if ent.resolvedRegion == nil {
		if ent.resolvedRegion, ok = scopeFind(ent.parent, rootNs, ent.regionName).(*regionEntity); !ok {
			kfmt.Fprintf(errWriter, "[field %s] could not resolve referenced region: %s\n", ent.name, ent.regionName)
		}
	}

	return ent.resolvedRegion != nil
}

// indexFieldEntity is a special field that groups together two field units so a
// index/data register pattern can be implemented. To write a value to an
// indexField, the interpreter must first write the appropriate offset to
// the indexRegister (using the alignment specifid by accessType) and then
// write the actual value to the dataRegister.
type indexFieldEntity struct {
	fieldEntity

	// The connection which this field references.
	connectionName     string
	resolvedConnection Entity

	indexRegName string
	indexReg     *fieldUnitEntity

	dataRegName string
	dataReg     *fieldUnitEntity
}

func (ent *indexFieldEntity) Resolve(errWriter io.Writer, rootNs ScopeEntity) bool {
	var ok bool
	if ent.connectionName != "" && ent.resolvedConnection == nil {
		if ent.resolvedConnection = scopeFind(ent.parent, rootNs, ent.connectionName); ent.resolvedConnection == nil {
			kfmt.Fprintf(errWriter, "[field %s] could not resolve connection reference: %s\n", ent.name, ent.connectionName)
			return false
		}
	}

	if ent.indexReg == nil {
		if ent.indexReg, ok = scopeFind(ent.parent, rootNs, ent.indexRegName).(*fieldUnitEntity); !ok {
			kfmt.Fprintf(errWriter, "[indexField %s] could not resolve referenced index register: %s\n", ent.name, ent.indexRegName)
		}
	}

	if ent.dataReg == nil {
		if ent.dataReg, ok = scopeFind(ent.parent, rootNs, ent.dataRegName).(*fieldUnitEntity); !ok {
			kfmt.Fprintf(errWriter, "[dataField %s] could not resolve referenced data register: %s\n", ent.name, ent.dataRegName)
		}
	}

	return ent.indexReg != nil && ent.dataReg != nil
}

// namedReference holds a named reference to an AML symbol. The spec allows
// the symbol not to be defined at the time when the reference is parsed. In
// such a case (forward reference) it will be resolved after the entire AML
// stream has successfully been parsed.
type namedReference struct {
	unnamedEntity

	targetName string
	target     Entity
}

func (ref *namedReference) Resolve(errWriter io.Writer, rootNs ScopeEntity) bool {
	if ref.target = scopeFind(ref.parent, rootNs, ref.targetName); ref.target == nil {
		kfmt.Fprintf(errWriter, "could not resolve referenced symbol: %s (parent: %s)\n", ref.targetName, ref.parent.Name())
		return false
	}

	return true
}

// methodInvocationEntity describes an AML method invocation.
type methodInvocationEntity struct {
	unnamedEntity

	methodDef *Method
}

// Method defines an invocable AML method.
type Method struct {
	scopeEntity

	argCount   uint8
	serialized bool
	syncLevel  uint8
}

func (m *Method) getOpcode() opcode { return opMethod }

// Device defines a device.
type Device struct {
	scopeEntity

	// The methodMap keeps track of all methods exposed by this device.
	methodMap map[string]*Method
}

func (d *Device) getOpcode() opcode { return opDevice }

// mutexEntity represents a named mutex object
type mutexEntity struct {
	parent ScopeEntity

	// isGlobal is set to true for the pre-defined global mutex (\_GL object)
	isGlobal bool

	name      string
	syncLevel uint8
}

func (ent *mutexEntity) getOpcode() opcode            { return opMutex }
func (ent *mutexEntity) setOpcode(op opcode)          {}
func (ent *mutexEntity) Name() string                 { return ent.name }
func (ent *mutexEntity) Parent() ScopeEntity          { return ent.parent }
func (ent *mutexEntity) setParent(parent ScopeEntity) { ent.parent = parent }
func (ent *mutexEntity) getArgs() []interface{}       { return nil }
func (ent *mutexEntity) setArg(argIndex uint8, arg interface{}) bool {
	// arg 0 is the mutex name
	if argIndex == 0 {
		var ok bool
		ent.name, ok = arg.(string)
		return ok
	}

	// arg1 is the sync level (bits 0:3)
	syncLevel, ok := arg.(uint64)
	ent.syncLevel = uint8(syncLevel) & 0xf
	return ok
}

// eventEntity represents a named ACPI sync event.
type eventEntity struct {
	namedEntity
}
