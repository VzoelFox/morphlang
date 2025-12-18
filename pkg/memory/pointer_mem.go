package memory

import "unsafe"

func AllocPointer(addr uint64) (Ptr, error) {
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
	header.Type = TagPointer
	header.Size = uint32(allocSize)

	valPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	*(*uint64)(valPtr) = addr

	return ptr, nil
}

func ReadPointer(ptr Ptr) (uint64, error) {
	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(ptr)
	if err != nil { return 0, err }

	valPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	return *(*uint64)(valPtr), nil
}
