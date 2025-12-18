package memory

import (
	"sync/atomic"
	"unsafe"
)

// CompareAndSwapPtr performs an atomic compare-and-swap operation on a Ptr value in memory.
// It checks if the value at 'addr' is equal to 'old'. If so, it sets it to 'new' and returns true.
// Optimization: Uses RLock (Optimistic) -> Lock (Page Fault) strategy.
func CompareAndSwapPtr(addr Ptr, old, new Ptr) (bool, error) {
	// Optimistic Path: RLock
	Lemari.mu.RLock()
	raw, err := Lemari.resolveFast(addr)
	if err == nil {
		defer Lemari.mu.RUnlock()
		// Use Hardware CAS. Safe because RLock prevents eviction.
		target := (*uint64)(unsafe.Pointer(raw))
		swapped := atomic.CompareAndSwapUint64(target, uint64(old), uint64(new))
		return swapped, nil
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

		target := (*uint64)(unsafe.Pointer(raw))
		// We hold exclusive Lock, so simple check-set is atomic regarding other operations
		// BUT mixed access (one holding Lock, one holding RLock+Atomic) is tricky?
		// No, Lock excludes RLock. So no one else can be accessing via Optimistic Path.
		// So simple logic is fine.
		// But let's use atomic for consistency.
		swapped := atomic.CompareAndSwapUint64(target, uint64(old), uint64(new))
		return swapped, nil
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
		target := (*uint64)(unsafe.Pointer(raw))
		val := atomic.LoadUint64(target)
		return Ptr(val), nil
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
		target := (*uint64)(unsafe.Pointer(raw))
		// Even under Lock, reading should be atomic to avoid tearing if architecture requires it
		val := atomic.LoadUint64(target)
		return Ptr(val), nil
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
		target := (*uint64)(unsafe.Pointer(raw))
		atomic.StoreUint64(target, uint64(val))
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
		target := (*uint64)(unsafe.Pointer(raw))
		atomic.StoreUint64(target, uint64(val))
		return nil
	}

	return err
}
