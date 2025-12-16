package memory

// Allocator defines the contract for our memory managers.
type Allocator interface {
	// Alloc allocates a block of `size` bytes and returns its address.
	Alloc(size int) (Ptr, error)

	// Free releases the memory block pointed to by `ptr`.
	Free(ptr Ptr) error
}
