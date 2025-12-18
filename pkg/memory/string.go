package memory

import "unsafe"

// AllocString allocates a String object in the Cabinet.
// Layout: [Header][int32 Length][Bytes...]
func AllocString(s string) (Ptr, error) {
	length := len(s)
	// Payload: Length(4) + Bytes(N)
	payloadSize := 4 + length
	totalSize := HeaderSize + payloadSize

	// Align to 8 bytes
	allocSize := (totalSize + 7) & ^7

	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	ptr, err := Lemari.alloc(allocSize)
	if err != nil {
		return NilPtr, err
	}

	raw, err := Lemari.resolve(ptr)
	if err != nil {
		return NilPtr, err
	}

	// 1. Write Header
	header := (*Header)(raw)
	header.Type = TagString
	header.Size = uint32(allocSize)

	// 2. Write Length
	lenPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	*(*int32)(lenPtr) = int32(length)

	// 3. Write Bytes
	if length > 0 {
		dataPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize) + 4)
		// Unsafe copy
		dst := unsafe.Slice((*byte)(dataPtr), length)
		copy(dst, s)
	}

	return ptr, nil
}

// ReadString reads the string value.
func ReadString(ptr Ptr) (string, error) {
	if ptr == NilPtr {
		return "", nil
	}

	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(ptr)
	if err != nil {
		return "", err
	}

	// Read Length
	lenPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	length := int(*(*int32)(lenPtr))

	if length == 0 {
		return "", nil
	}

	// Read Bytes
	dataPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize) + 4)
	src := unsafe.Slice((*byte)(dataPtr), length)

	return string(src), nil
}
