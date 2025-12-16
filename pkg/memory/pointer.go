package memory

import "unsafe"

// Ptr represents a virtual address in our Arena.
// It is essentially an offset from the Arena base address.
type Ptr uint64

// NilPtr represents a null pointer (0 offset? Or -1? Let's use 0 and reserve address 0).
const NilPtr Ptr = 0

// ToUnsafe converts our virtual Ptr to a real Go unsafe.Pointer.
// WARNING: This is valid only as long as `RAM` does not move (which it won't, as it's a global array).
func (p Ptr) ToUnsafe() unsafe.Pointer {
	if p == NilPtr {
		return nil
	}
	base := uintptr(RAM.BasePointer())
	return unsafe.Pointer(base + uintptr(p))
}

// FromUnsafe creates a Ptr from a real pointer.
// Use with extreme caution.
func FromUnsafe(p unsafe.Pointer) Ptr {
	if p == nil {
		return NilPtr
	}
	base := uintptr(RAM.BasePointer())
	addr := uintptr(p)
	return Ptr(addr - base)
}
