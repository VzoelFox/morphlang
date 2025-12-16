package memory

import "unsafe"

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
const HeaderSize = int(unsafe.Sizeof(Header{}))

// GetHeader reads the header at the given pointer
func GetHeader(ptr Ptr) *Header {
	raw := ptr.ToUnsafe()
	return (*Header)(raw)
}
