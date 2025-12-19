package memory

import "unsafe"

// AllocUpvalue allocates an Upvalue object.
// Layout: [Header][ValuePtr(8)][StackIdx(8)][IsOpen(8)]
func AllocUpvalue(stackIdx int) (Ptr, error) {
	size := HeaderSize + 24

	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	ptr, err := Lemari.alloc(size)
	if err != nil { return NilPtr, err }

	raw, err := Lemari.resolve(ptr)
	if err != nil { return NilPtr, err }

	header := (*Header)(raw)
	header.Type = TagUpvalue
	header.Size = uint32(size)

	// ValuePtr (Offset 0 payload)
	valPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	*(*Ptr)(valPtr) = NilPtr

	// StackIdx (Offset 8 payload)
	stackIdxPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize) + 8)
	*(*int64)(stackIdxPtr) = int64(stackIdx)

	// IsOpen (Offset 16 payload)
	isOpenPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize) + 16)
	*(*int64)(isOpenPtr) = 1 // True

	return ptr, nil
}

func CloseUpvalue(ptr Ptr, val Ptr) error {
	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(ptr)
	if err != nil { return err }

	// Write Value
	valPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	*(*Ptr)(valPtr) = val

	// Set IsOpen = 0
	isOpenPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize) + 16)
	*(*int64)(isOpenPtr) = 0

	return nil
}

func ReadUpvalue(ptr Ptr) (val Ptr, stackIdx int64, isOpen bool, err error) {
	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(ptr)
	if err != nil { return NilPtr, 0, false, err }

	valPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	stackIdxPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize) + 8)
	isOpenPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize) + 16)

	val = *(*Ptr)(valPtr)
	stackIdx = *(*int64)(stackIdxPtr)
	isOpen = *(*int64)(isOpenPtr) == 1

	return
}
