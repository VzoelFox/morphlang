package memory

// AllocNull allocates a Null object in the Cabinet.
// Layout: [Header] (No payload)
func AllocNull() (Ptr, error) {
	totalSize := HeaderSize
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
	header.Type = TagNull
	header.Size = uint32(allocSize)

	return ptr, nil
}
