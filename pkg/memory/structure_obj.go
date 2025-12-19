package memory

import (
	"fmt"
	"unsafe"
)

// AllocSchema allocates a Schema object.
// Layout: [Header][Ptr Name][Ptr FieldNames(Array)]
func AllocSchema(name Ptr, fields Ptr) (Ptr, error) {
	// Header + 8 (Name) + 8 (Fields)
	payloadSize := 16
	totalSize := HeaderSize + payloadSize

	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	ptr, err := Lemari.alloc(totalSize)
	if err != nil { return NilPtr, err }

	raw, err := Lemari.resolve(ptr)
	if err != nil { return NilPtr, err }

	header := (*Header)(raw)
	header.Type = TagSchema
	header.Size = uint32(totalSize)

	// Write Name
	namePtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	*(*Ptr)(namePtr) = name

	// Write Fields
	fieldsPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize) + 8)
	*(*Ptr)(fieldsPtr) = fields

	return ptr, nil
}

// ReadSchema reads the schema components.
func ReadSchema(ptr Ptr) (name Ptr, fields Ptr, err error) {
	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(ptr)
	if err != nil { return NilPtr, NilPtr, err }

	namePtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	fieldsPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize) + 8)

	return *(*Ptr)(namePtr), *(*Ptr)(fieldsPtr), nil
}

// AllocStruct allocates a Struct instance.
// Layout: [Header][Ptr Schema][Ptr... Fields]
func AllocStruct(schema Ptr, fieldCount int) (Ptr, error) {
	payloadSize := 8 + (fieldCount * 8)
	totalSize := HeaderSize + payloadSize

	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	ptr, err := Lemari.alloc(totalSize)
	if err != nil { return NilPtr, err }

	raw, err := Lemari.resolve(ptr)
	if err != nil { return NilPtr, err }

	header := (*Header)(raw)
	header.Type = TagStruct
	header.Size = uint32(totalSize)

	// Write Schema
	schemaSlot := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	*(*Ptr)(schemaSlot) = schema

	// Zero out fields
	fieldsStart := uintptr(raw) + uintptr(HeaderSize) + 8
	for i := 0; i < fieldCount; i++ {
		slot := unsafe.Pointer(fieldsStart + uintptr(i*8))
		*(*Ptr)(slot) = NilPtr
	}

	return ptr, nil
}

// ReadStructSchema reads the schema pointer of a struct.
func ReadStructSchema(ptr Ptr) (Ptr, error) {
	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(ptr)
	if err != nil { return NilPtr, err }

	schemaSlot := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	return *(*Ptr)(schemaSlot), nil
}

// WriteStructField writes a value to a struct field by index.
func WriteStructField(ptr Ptr, index int, val Ptr) error {
	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(ptr)
	if err != nil { return err }

	// Calculate offset: Header + Schema(8) + Index*8
	offset := HeaderSize + 8 + (index * 8)

	// Bounds check based on Size in Header
	header := (*Header)(raw)
	if uint32(offset+8) > header.Size {
		return fmt.Errorf("struct field index out of bounds")
	}

	slot := unsafe.Pointer(uintptr(raw) + uintptr(offset))
	*(*Ptr)(slot) = val
	return nil
}

// ReadStructField reads a value from a struct field by index.
func ReadStructField(ptr Ptr, index int) (Ptr, error) {
	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(ptr)
	if err != nil { return NilPtr, err }

	offset := HeaderSize + 8 + (index * 8)

	header := (*Header)(raw)
	if uint32(offset+8) > header.Size {
		return NilPtr, fmt.Errorf("struct field index out of bounds")
	}

	slot := unsafe.Pointer(uintptr(raw) + uintptr(offset))
	return *(*Ptr)(slot), nil
}
