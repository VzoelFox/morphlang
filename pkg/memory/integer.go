package memory

import "unsafe"

// AllocInteger allocates a raw Integer object in the Cabinet.
// Layout: [Header][int64 Value]
func AllocInteger(value int64) (Ptr, error) {
	// Size = Header + int64
	payloadSize := int(unsafe.Sizeof(value))
	totalSize := HeaderSize + payloadSize

	ptr, err := Lemari.Alloc(totalSize)
	if err != nil {
		return NilPtr, err
	}

	// 1. Write Header
	header := GetHeader(ptr)
	header.Type = TagInteger
	header.Size = uint32(totalSize)

	// 2. Write Value
	// Value is located immediately after Header
	valuePtr := unsafe.Pointer(uintptr(ptr.ToUnsafe()) + uintptr(HeaderSize))
	*(*int64)(valuePtr) = value

	return ptr, nil
}

// ReadInteger reads the value of a raw Integer object.
func ReadInteger(ptr Ptr) int64 {
	// Safety check (minimal)
	if ptr == NilPtr {
		return 0 // Panic?
	}

	// Check Tag?
	// header := GetHeader(ptr)
	// if header.Type != TagInteger { panic("Type mismatch") }

	valuePtr := unsafe.Pointer(uintptr(ptr.ToUnsafe()) + uintptr(HeaderSize))
	return *(*int64)(valuePtr)
}
