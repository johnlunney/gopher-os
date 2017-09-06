// DSDT-parser-testsuite
//
// This file contains various ASL constructs to ensure that the AML parser 
// properly handles all possible ASL opcodes it may encounter. This test file 
// is used in addition to the DSDT.aml file obtained by running acpidump inside 
// virtualbox.
DefinitionBlock ("DSDT-parser-testsuite.aml", "DSDT", 2, "GOPHER", "GOPHEROS", 0x00000002)
{
    OperationRegion (DBG0, SystemIO, 0x3000, 0x04)
    Field (DBG0, ByteAcc, NoLock, Preserve)
    {
        DHE1,   8
    }

    Device (DRV0)
    {
	Name (_ADR, Ones)

	// named entity containing qword const
	Name (H15F, 0xBADC0FEEDEADC0DE)
	Method (_GTF, 0, NotSerialized)  // _GTF: Get Task File
	{
	    Return (H15F)
	}
    }
    
    // example from p. 268 of ACPI 6.2 spec
    Scope(\_SB){
	OperationRegion(TOP1, GenericSerialBus, 0x00, 0x100) // GenericSerialBus device at command offset 0x00

	Name (SDB0, ResourceTemplate() {})
	Field(TOP1, BufferAcc, NoLock, Preserve){
	    Connection(SDB0), // Use the Resource Descriptor defined above
	    AccessAs(BufferAcc, AttribWord),
	    FLD0, 8,
	    FLD1, 8
	}
	
	Field(TOP1, BufferAcc, NoLock, Preserve){
	    Connection(I2cSerialBus(0x5b,,100000,, "\\_SB",,,,RawDataBuffer(){3,9})),
	    AccessAs(BufferAcc, AttribBytes(4)),
	    FLD2, 8,
	    AccessAs(BufferAcc, AttribRawBytes(3)),
	    FLD3, 8,
	    AccessAs(BufferAcc, AttribRawProcessBytes(2)),
	    FLD4, 8
	}
    }

    // Other executable bits
    Method (EXE0, 0, Serialized)
    {
	    Breakpoint
	    Debug = "test"
	    Fatal(0xf0, 0xdeadc0de, 1)

	    // Mutex support 
	    Mutex(MUT0, 1)
	    Acquire(MUT0, 0xffff) // no timeout
	    Release(MUT0)

	    // Signal/Wait
	    Signal(HLO0)
	    Wait(HLO0, 0xffff)

	    // Get monotonic timer value
	    Local0 = Timer
	    return(Local0)
    }

    // Other entity types 
    Event(HLO0)
}
