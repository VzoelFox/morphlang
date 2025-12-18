package memory

import "unsafe"

// AllocHash allocates a Hash map container.
// For MVP Phase X, this is a linear list of Key-Value pairs.
// Layout: [Header][int32 Count][Pair0_Key][Pair0_Value]...
func AllocHash(count int) (Ptr, error) {
	// Each pair is 2 Ptrs (16 bytes)
	payloadSize := 4 + (count * 16)
	totalSize := HeaderSize + payloadSize
	allocSize := (totalSize + 7) & ^7

	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	ptr, err := Lemari.alloc(allocSize)
	if err != nil { return NilPtr, err }

	raw, err := Lemari.resolve(ptr)
	if err != nil { return NilPtr, err }

	header := (*Header)(raw)
	header.Type = TagHash
	header.Size = uint32(allocSize)

	countPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	*(*int32)(countPtr) = int32(count)

	return ptr, nil
}

func WriteHashPair(hashPtr Ptr, index int, key, value Ptr) error {
	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(hashPtr)
	if err != nil { return err }

	// Offset = Header + Count(4) + (Index * 16)
	base := uintptr(raw) + uintptr(HeaderSize) + 4
	offset := uintptr(index) * 16

	keyPtr := unsafe.Pointer(base + offset)
	valPtr := unsafe.Pointer(base + offset + 8)

	*(*uint64)(keyPtr) = uint64(key)
	*(*uint64)(valPtr) = uint64(value)

	return nil
}

func ReadHashPair(hashPtr Ptr, index int) (Ptr, Ptr, error) {
	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(hashPtr)
	if err != nil { return NilPtr, NilPtr, err }

	base := uintptr(raw) + uintptr(HeaderSize) + 4
	offset := uintptr(index) * 16

	keyPtr := unsafe.Pointer(base + offset)
	valPtr := unsafe.Pointer(base + offset + 8)

	return Ptr(*(*uint64)(keyPtr)), Ptr(*(*uint64)(valPtr)), nil
}

func ReadHashCount(hashPtr Ptr) (int, error) {
	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(hashPtr)
	if err != nil { return 0, err }

	countPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	return int(*(*int32)(countPtr)), nil
}
