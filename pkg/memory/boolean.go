package memory

import "unsafe"

// AllocBoolean allocates a Boolean object in the Cabinet.
// Layout: [Header][int8 Value]
func AllocBoolean(value bool) (Ptr, error) {
	payloadSize := 1
	totalSize := HeaderSize + payloadSize
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
	header.Type = TagBoolean
	header.Size = uint32(allocSize)

	// 2. Write Value
	valPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	var v int8
	if value {
		v = 1
	}
	*(*int8)(valPtr) = v

	return ptr, nil
}

// ReadBoolean reads the boolean value.
func ReadBoolean(ptr Ptr) (bool, error) {
	if ptr == NilPtr {
		return false, nil // Default to false? Or error?
	}

	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(ptr)
	if err != nil {
		return false, err
	}

	valPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	v := *(*int8)(valPtr)
	return v == 1, nil
}
