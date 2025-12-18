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
	TagFloat   TypeTag = 4
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

// ReadHeader reads the header at the given pointer safely.
// It returns a copy of the Header struct.
func ReadHeader(ptr Ptr) (Header, error) {
	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(ptr)
	if err != nil {
		return Header{}, err
	}
	return *(*Header)(raw), nil
}
