package memory

import "unsafe"

// CompareAndSwapPtr performs an atomic compare-and-swap operation on a Ptr value in memory.
// It checks if the value at 'addr' is equal to 'old'. If so, it sets it to 'new' and returns true.
// Note: Currently uses the global lock (emulated atomicity) to ensure safety in the Hybrid model.
func CompareAndSwapPtr(addr Ptr, old, new Ptr) (bool, error) {
	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(addr)
	if err != nil {
		return false, err
	}

	// Assuming addr points to a Ptr-sized slot
	target := (*Ptr)(unsafe.Pointer(raw))

	if *target == old {
		*target = new
		return true, nil
	}

	return false, nil
}

// LoadPtr reads a Ptr atomically.
func LoadPtr(addr Ptr) (Ptr, error) {
	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(addr)
	if err != nil {
		return NilPtr, err
	}

	target := (*Ptr)(unsafe.Pointer(raw))
	return *target, nil
}

// StorePtr writes a Ptr atomically.
func StorePtr(addr Ptr, val Ptr) error {
	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(addr)
	if err != nil {
		return err
	}

	target := (*Ptr)(unsafe.Pointer(raw))
	*target = val
	return nil
}
