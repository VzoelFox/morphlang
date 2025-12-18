package memory

import "unsafe"

// AllocFloat allocates a raw Float object in the Cabinet.
// Layout: [Header][float64 Value]
func AllocFloat(value float64) (Ptr, error) {
	// Size = Header + float64
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
	header.Type = TagFloat
	header.Size = uint32(totalSize)

	// 2. Write Value
	headerPtr := unsafe.Pointer(header)
	valuePtr := unsafe.Pointer(uintptr(headerPtr) + uintptr(HeaderSize))
	*(*float64)(valuePtr) = value

	return ptr, nil
}

// ReadFloat reads the value of a raw Float object.
func ReadFloat(ptr Ptr) (float64, error) {
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
	return *(*float64)(valuePtr), nil
}

// WriteFloat updates the value of an existing Float object.
func WriteFloat(ptr Ptr, value float64) error {
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
	*(*float64)(valuePtr) = value
	return nil
}
