package memory

import "unsafe"

// CompareAndSwapPtr performs an atomic compare-and-swap operation on a Ptr value in memory.
// It checks if the value at 'addr' is equal to 'old'. If so, it sets it to 'new' and returns true.
// Optimization: Uses RLock (Optimistic) -> Lock (Page Fault) strategy.
func CompareAndSwapPtr(addr Ptr, old, new Ptr) (bool, error) {
	// Optimistic Path: RLock
	Lemari.mu.RLock()
	raw, err := Lemari.resolveFast(addr)
	if err == nil {
		defer Lemari.mu.RUnlock()
		target := (*Ptr)(unsafe.Pointer(raw))
		if *target == old {
			*target = new
			return true, nil
		}
		return false, nil
	}
	Lemari.mu.RUnlock()

	// Slow Path: Page Fault (Lock)
	if err == ErrPageFault {
		Lemari.mu.Lock()
		defer Lemari.mu.Unlock()

		raw, err = Lemari.resolve(addr)
		if err != nil {
			return false, err
		}

		target := (*Ptr)(unsafe.Pointer(raw))
		if *target == old {
			*target = new
			return true, nil
		}
		return false, nil
	}

	return false, err
}

// LoadPtr reads a Ptr atomically.
func LoadPtr(addr Ptr) (Ptr, error) {
	// Optimistic Path
	Lemari.mu.RLock()
	raw, err := Lemari.resolveFast(addr)
	if err == nil {
		defer Lemari.mu.RUnlock()
		target := (*Ptr)(unsafe.Pointer(raw))
		return *target, nil
	}
	Lemari.mu.RUnlock()

	// Slow Path
	if err == ErrPageFault {
		Lemari.mu.Lock()
		defer Lemari.mu.Unlock()

		raw, err = Lemari.resolve(addr)
		if err != nil {
			return NilPtr, err
		}
		target := (*Ptr)(unsafe.Pointer(raw))
		return *target, nil
	}

	return NilPtr, err
}

// StorePtr writes a Ptr atomically.
func StorePtr(addr Ptr, val Ptr) error {
	// Optimistic Path
	Lemari.mu.RLock()
	raw, err := Lemari.resolveFast(addr)
	if err == nil {
		defer Lemari.mu.RUnlock()
		target := (*Ptr)(unsafe.Pointer(raw))
		*target = val
		return nil
	}
	Lemari.mu.RUnlock()

	// Slow Path
	if err == ErrPageFault {
		Lemari.mu.Lock()
		defer Lemari.mu.Unlock()

		raw, err = Lemari.resolve(addr)
		if err != nil {
			return err
		}
		target := (*Ptr)(unsafe.Pointer(raw))
		*target = val
		return nil
	}

	return err
}
