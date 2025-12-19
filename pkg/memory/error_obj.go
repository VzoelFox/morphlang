package memory

import "unsafe"

// AllocError allocates an Error object.
// Layout: [Header][PtrMessage][PtrCode][Line(4)][Col(4)]
// We store Message and Code as Strings (Ptr).
func AllocError(message Ptr, code Ptr, line, col int) (Ptr, error) {
	payloadSize := 8 + 8 + 4 + 4
	totalSize := HeaderSize + payloadSize
	allocSize := (totalSize + 7) & ^7

	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	ptr, err := Lemari.alloc(allocSize)
	if err != nil { return NilPtr, err }

	raw, err := Lemari.resolve(ptr)
	if err != nil { return NilPtr, err }

	header := (*Header)(raw)
	header.Type = TagError
	header.Size = uint32(allocSize)

	base := uintptr(raw) + uintptr(HeaderSize)
	*(*uint64)(unsafe.Pointer(base)) = uint64(message)
	*(*uint64)(unsafe.Pointer(base + 8)) = uint64(code)
	*(*int32)(unsafe.Pointer(base + 16)) = int32(line)
	*(*int32)(unsafe.Pointer(base + 20)) = int32(col)

	return ptr, nil
}

func ReadError(ptr Ptr) (Ptr, Ptr, int, int, error) {
	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(ptr)
	if err != nil { return NilPtr, NilPtr, 0, 0, err }

	base := uintptr(raw) + uintptr(HeaderSize)
	msgPtr := Ptr(*(*uint64)(unsafe.Pointer(base)))
	codePtr := Ptr(*(*uint64)(unsafe.Pointer(base + 8)))
	line := int(*(*int32)(unsafe.Pointer(base + 16)))
	col := int(*(*int32)(unsafe.Pointer(base + 20)))

	return msgPtr, codePtr, line, col, nil
}
