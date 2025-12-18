package memory

import "unsafe"

// AllocBuiltin allocates a wrapper for a Builtin function index.
// Layout: [Header][int32 Index]
func AllocBuiltin(index int) (Ptr, error) {
	payloadSize := 4
	totalSize := HeaderSize + payloadSize
	allocSize := (totalSize + 7) & ^7

	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	ptr, err := Lemari.alloc(allocSize)
	if err != nil { return NilPtr, err }

	raw, err := Lemari.resolve(ptr)
	if err != nil { return NilPtr, err }

	header := (*Header)(raw)
	header.Type = TagBuiltin
	header.Size = uint32(allocSize)

	idxPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	*(*int32)(idxPtr) = int32(index)

	return ptr, nil
}

// ReadBuiltin reads the index.
func ReadBuiltin(ptr Ptr) (int, error) {
	// Optimistic
	Lemari.mu.RLock()
	raw, err := Lemari.resolveFast(ptr)
	if err == nil {
		defer Lemari.mu.RUnlock()
		idxPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
		return int(*(*int32)(idxPtr)), nil
	}
	Lemari.mu.RUnlock()

	// Slow
	return readBuiltinLocked(ptr)
}

func readBuiltinLocked(ptr Ptr) (int, error) {
	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(ptr)
	if err != nil { return 0, err }

	idxPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	return int(*(*int32)(idxPtr)), nil
}
