package memory

import (
	"fmt"
	"unsafe"
)

// BumpAllocate tries to allocate `size` bytes in the current active tray.
// If it fits, it returns the Ptr and advances the bump pointer.
// If it fails, it returns error (triggering GC/Drawer Switch in future).
func (c *Cabinet) Alloc(size int) (Ptr, error) {
	if size <= 0 {
		return NilPtr, fmt.Errorf("invalid allocation size: %d", size)
	}

	// Align to 8 bytes (64-bit machine standard)
	alignedSize := (size + 7) &^ 7

	activeDrawer := &c.Drawers[c.ActiveDrawerIndex]
	var activeTray *Tray

	if activeDrawer.IsPrimaryActive {
		activeTray = &activeDrawer.PrimaryTray
	} else {
		activeTray = &activeDrawer.SecondaryTray
	}

	if activeTray.Remaining() < alignedSize {
		// TODO: Implement Drawer Switch or Collection here.
		// For now, fail explicitly or try next drawer?
		// Let's keep it simple: Drawer Full!
		return NilPtr, fmt.Errorf("out of memory in current drawer %d", c.ActiveDrawerIndex)
	}

	ptr := activeTray.Current
	activeTray.Current += Ptr(alignedSize)

	return ptr, nil
}

// Write writes data to the pointer address.
// This is the "Manual Copy" part where we move data into the Tray.
func Write(dst Ptr, data []byte) error {
	if dst == NilPtr {
		return fmt.Errorf("segmentation fault: nil pointer")
	}

	rawPtr := dst.ToUnsafe()
	size := len(data)

	// Convert raw pointer to slice for copying
	// Safe for now as we control the Arena
	targetSlice := unsafe.Slice((*byte)(rawPtr), size)
	copy(targetSlice, data)

	return nil
}
