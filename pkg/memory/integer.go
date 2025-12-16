package memory

import "unsafe"

// AllocInteger allocates a raw Integer object in the Cabinet.
// Layout: [Header][int64 Value]
func AllocInteger(value int64) (Ptr, error) {
	// Size = Header + int64
	payloadSize := int(unsafe.Sizeof(value))
	totalSize := HeaderSize + payloadSize

	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	ptr, err := Lemari.alloc(totalSize)
	if err != nil {
		return NilPtr, err
	}

	// Resolve for writing
	raw, err := Lemari.resolve(ptr)
	if err != nil {
		return NilPtr, err
	}

	// 1. Write Header
	header := (*Header)(raw)
	header.Type = TagInteger
	header.Size = uint32(totalSize)

	// 2. Write Value
	headerPtr := unsafe.Pointer(header)
	valuePtr := unsafe.Pointer(uintptr(headerPtr) + uintptr(HeaderSize))
	*(*int64)(valuePtr) = value

	return ptr, nil
}

// ReadInteger reads the value of a raw Integer object.
func ReadInteger(ptr Ptr) (int64, error) {
	if ptr == NilPtr {
		return 0, nil
	}

	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	// Resolve pointer
	raw, err := Lemari.resolve(ptr)
	if err != nil {
		return 0, err
	}

	valuePtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	return *(*int64)(valuePtr), nil
}

// WriteInteger updates the value of an existing Integer object.
func WriteInteger(ptr Ptr, value int64) error {
	if ptr == NilPtr {
		return nil
	}

	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(ptr)
	if err != nil {
		return err
	}

	valuePtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	*(*int64)(valuePtr) = value
	return nil
}
