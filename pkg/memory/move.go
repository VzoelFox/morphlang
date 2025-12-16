package memory

import (
	"fmt"
	"unsafe"
)

// MemCpy copies `size` bytes from `src` to `dst`.
// This represents moving the glass from Nampan A to Nampan B.
// Thread-safe.
func MemCpy(src Ptr, dst Ptr, size int) error {
	if src == NilPtr || dst == NilPtr {
		return fmt.Errorf("memcpy: nil pointer access")
	}

	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	srcRaw, err := Lemari.resolve(src)
	if err != nil {
		return err
	}

	dstRaw, err := Lemari.resolve(dst)
	if err != nil {
		return err
	}

	srcSlice := unsafe.Slice((*byte)(srcRaw), size)
	dstSlice := unsafe.Slice((*byte)(dstRaw), size)

	copy(dstSlice, srcSlice)
	return nil
}

// MoveObject simulates the "Owner Tanpa Borrowing" concept.
// It allocates space in the destination tray (implicitly handled by Alloc in higher logic)
// and copies the data, essentially cloning it.
func MoveObject(src Ptr, size int) (Ptr, error) {
	// Allocate new space (e.g., in ToSpace)
	// For this simulation, we just alloc from the Cabinet (which points to current active tray)
	// Alloc handles its own locking.
	newPtr, err := Lemari.Alloc(size)
	if err != nil {
		return NilPtr, err
	}

	// MemCpy handles its own locking.
	err = MemCpy(src, newPtr, size)
	if err != nil {
		return NilPtr, err
	}

	return newPtr, nil
}
