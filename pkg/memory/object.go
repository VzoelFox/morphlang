package memory

import (
	"unsafe"
)

// ObjectType tag (byte)
type TypeTag uint8

const (
	TagInteger TypeTag = 1
	TagBoolean TypeTag = 2
	TagString  TypeTag = 3
	// ... add others later
)

// Header is the metadata for every object in our heap.
// Aligned to 8 bytes.
type Header struct {
	Type TypeTag
	Size uint32 // Total size including header
	// Padding/Forwarding pointer for GC could go here
}

// Size of header
const HeaderSize = int(unsafe.Sizeof(Header{})) // Should be 8 usually (1 byte + 4 bytes + padding)

// GetHeader reads the header at the given pointer.
// It resolves the virtual pointer.
func GetHeader(ptr Ptr) (*Header, error) {
	raw, err := Lemari.Resolve(ptr)
	if err != nil {
		return nil, err
	}
	return (*Header)(raw), nil
}
