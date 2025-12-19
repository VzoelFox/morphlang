package memory

import (
	"unsafe"
)

// AllocModule allocates a Module object.
// Layout: [Header][Ptr Init][Ptr Exports]
func AllocModule(initFn Ptr, exports Ptr) (Ptr, error) {
	size := HeaderSize + 16

	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	ptr, err := Lemari.alloc(size)
	if err != nil { return NilPtr, err }

	raw, err := Lemari.resolve(ptr)
	if err != nil { return NilPtr, err }

	// Header
	header := (*Header)(raw)
	header.Type = TagModule
	header.Size = uint32(size)

	// Init Function Ptr (Offset 0 after Header)
	initPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	*(*Ptr)(initPtr) = initFn

	// Exports Ptr (Offset 8 after Header)
	expPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize) + 8)
	*(*Ptr)(expPtr) = exports

	return ptr, nil
}

func ReadModule(ptr Ptr) (Ptr, Ptr, error) {
	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(ptr)
	if err != nil { return NilPtr, NilPtr, err }

	initPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	expPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize) + 8)

	return *(*Ptr)(initPtr), *(*Ptr)(expPtr), nil
}

func WriteModuleExports(ptr Ptr, exports Ptr) error {
	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(ptr)
	if err != nil { return err }

	expPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize) + 8)
	*(*Ptr)(expPtr) = exports
	return nil
}

func WriteModuleInit(ptr Ptr, initFn Ptr) error {
	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(ptr)
	if err != nil { return err }

	initPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	*(*Ptr)(initPtr) = initFn
	return nil
}
