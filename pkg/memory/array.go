package memory

import (
	"fmt"
	"unsafe"
)

// AllocArray allocates an Array object in the Cabinet.
// Layout: [Header][int32 Capacity][int32 Length][Ptr... elements]
func AllocArray(length int, capacity int) (Ptr, error) {
	if capacity < length {
		capacity = length
	}

	// 4 bytes Capacity + 4 bytes Length = 8 bytes
	// capacity * 8 bytes for Pointers
	payloadSize := 8 + (capacity * 8)
	totalSize := HeaderSize + payloadSize

	// Already aligned to 8 bytes if HeaderSize is 8
	allocSize := totalSize

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
	header.Type = TagArray
	header.Size = uint32(allocSize)

	// 2. Write Capacity and Length
	// Offset 8: Capacity
	capPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	*(*int32)(capPtr) = int32(capacity)

	// Offset 12: Length
	lenPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize) + 4)
	*(*int32)(lenPtr) = int32(length)

	// 3. Zero out the elements (Safety)
	elementsStart := uintptr(raw) + uintptr(HeaderSize) + 8
	for i := 0; i < capacity; i++ {
		elemPtr := unsafe.Pointer(elementsStart + uintptr(i*8))
		*(*Ptr)(elemPtr) = NilPtr
	}

	return ptr, nil
}

// WriteArrayElement updates the pointer at the given index.
func WriteArrayElement(arrayPtr Ptr, index int, valuePtr Ptr) error {
	if arrayPtr == NilPtr {
		return fmt.Errorf("nil array pointer")
	}

	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(arrayPtr)
	if err != nil {
		return err
	}

	// Read Length to check bounds
	lenPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize) + 4)
	length := int(*(*int32)(lenPtr))

	if index < 0 || index >= length {
		return fmt.Errorf("index out of bounds: %d (len %d)", index, length)
	}

	// Write Pointer
	elementsStart := uintptr(raw) + uintptr(HeaderSize) + 8
	elemPtr := unsafe.Pointer(elementsStart + uintptr(index*8))
	*(*Ptr)(elemPtr) = valuePtr

	return nil
}

// ReadArrayElement reads the pointer at the given index.
func ReadArrayElement(arrayPtr Ptr, index int) (Ptr, error) {
	if arrayPtr == NilPtr {
		return NilPtr, nil
	}

	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(arrayPtr)
	if err != nil {
		return NilPtr, err
	}

	// Read Length
	lenPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize) + 4)
	length := int(*(*int32)(lenPtr))

	if index < 0 || index >= length {
		return NilPtr, fmt.Errorf("index out of bounds: %d (len %d)", index, length)
	}

	// Read Pointer
	elementsStart := uintptr(raw) + uintptr(HeaderSize) + 8
	elemPtr := unsafe.Pointer(elementsStart + uintptr(index*8))
	return *(*Ptr)(elemPtr), nil
}
