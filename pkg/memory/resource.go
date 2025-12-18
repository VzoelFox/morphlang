package memory

import "unsafe"

// AllocResource allocates a wrapper for a Host Resource ID.
// Layout: [Header][int64 ResourceID]
func AllocResource(id int64) (Ptr, error) {
	payloadSize := 8
	totalSize := HeaderSize + payloadSize
	allocSize := (totalSize + 7) & ^7

	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	ptr, err := Lemari.alloc(allocSize)
	if err != nil { return NilPtr, err }

	raw, err := Lemari.resolve(ptr)
	if err != nil { return NilPtr, err }

	header := (*Header)(raw)
	header.Type = TagResource
	header.Size = uint32(allocSize)

	idPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	*(*int64)(idPtr) = id

	return ptr, nil
}

func ReadResource(ptr Ptr) (int64, error) {
	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(ptr)
	if err != nil { return 0, err }

	idPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	return *(*int64)(idPtr), nil
}
