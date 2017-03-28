package multiboot

import "unsafe"

type tagType uint32

const (
	tagMbSectionEnd tagType = iota
	tagBootCmdLine
	tagBootLoaderName
	tagModules
	tagBasicMemoryInfo
	tagBiosBootDevice
	tagMemoryMap
	tagVbeInfo
	tagFramebufferInfo
	tagElfSymbols
	tagApmTable
)

// info describes the multiboot info section header.
type info struct {
	// Total size of multiboot info section.
	totalSize uint32

	// Always set to zero; reserved for future use
	reserved uint32
}

// tagHeader describes the header the preceedes each tag.
type tagHeader struct {
	// The type of the tag
	tagType tagType

	// The size of the tag including the header but *not* including any
	// padding. According to the spec, each tag starts at a 8-byte aligned
	// address.
	size uint32
}

// mmapHeader describes the header for a memory map specification.
type mmapHeader struct {
	// The size of each entry.
	entrySize uint32

	// The version of the entries that follow.
	entryVersion uint32
}

// MemoryEntryType defines the type of a MemoryMapEntry.
type MemoryEntryType uint32

const (
	// MemAvailable indicates that the memory region is available for use.
	MemAvailable MemoryEntryType = iota + 1

	// MemReserved indicates that the memory region is not available for use.
	MemReserved

	// MemAcpiReclaimable indicates a memory region that holds ACPI info that
	// can be reused by the OS.
	MemAcpiReclaimable

	// MemNvs indicates memory that must be preserved when hibernating.
	MemNvs

	// Any value >= memUnknown will be mapped to MemReserved.
	memUnknown
)

/// MemoryMapEntry describes a memory region entry, namely its physical address,
// its length and its type.
type MemoryMapEntry struct {
	// The physical address for this memory region.
	PhysAddress uint64

	// The length of the memory region.
	Length uint64

	// The type of this entry.
	Type MemoryEntryType
}

var (
	infoData uintptr
)

// SetInfoPtr updates the internal multiboot information pointer to the given
// value. This function must be invoked before invoking any other function
// exported by this package.
func SetInfoPtr(ptr uintptr) {
	infoData = ptr
}

// findTagByType scans the multiboot info data looking for the start of of the
// specified type. It returns a pointer to the tag contents start offset and
// the content length exluding the tag header.
//
// If the tag is not present in the multiboot info, findTagSection will return
// back (0,0).
func findTagByType(tagType tagType) (uintptr, uint32) {
	var ptrTagHeader *tagHeader

	curPtr := infoData + 8
	for ptrTagHeader = (*tagHeader)(unsafe.Pointer(curPtr)); ptrTagHeader.tagType != tagMbSectionEnd; ptrTagHeader = (*tagHeader)(unsafe.Pointer(curPtr)) {
		if ptrTagHeader.tagType == tagType {
			return curPtr + 8, ptrTagHeader.size - 8
		}

		// Tags are aligned at 8-byte aligned addresses
		curPtr += uintptr(int32(ptrTagHeader.size+7) & ^7)
	}

	return 0, 0
}
