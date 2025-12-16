package memory

import (
	"fmt"
	"unsafe"
)

// MemCpy copies `size` bytes from `src` to `dst`.
// This represents moving the glass from Nampan A to Nampan B.
func MemCpy(src Ptr, dst Ptr, size int) error {
	if src == NilPtr || dst == NilPtr {
		return fmt.Errorf("memcpy: nil pointer access")
	}

	srcRaw := src.ToUnsafe()
	dstRaw := dst.ToUnsafe()

	srcSlice := unsafe.Slice((*byte)(srcRaw), size)
	dstSlice := unsafe.Slice((*byte)(dstRaw), size)

	copy(dstSlice, srcSlice)
	return nil
}

// MoveObject simulates the "Owner Tanpa Borrowing" concept.
// It allocates space in the destination tray (implicitly handled by Alloc in higher logic)
// and copies the data, essentially cloning it.
// In a copying collector, we would update the forwarding pointer here.
func MoveObject(src Ptr, size int) (Ptr, error) {
	// Allocate new space (e.g., in ToSpace)
	// For this simulation, we just alloc from the Cabinet (which points to current active tray)
	// In real GC, we'd explicitly alloc from the ToSpace tray.
	newPtr, err := Lemari.Alloc(size)
	if err != nil {
		return NilPtr, err
	}

	err = MemCpy(src, newPtr, size)
	if err != nil {
		return NilPtr, err
	}

	return newPtr, nil
}
