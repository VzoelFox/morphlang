package memory

import "unsafe"

// Size of our Virtual RAM (e.g., 10 MB for starter)
const HEAP_SIZE = 10 * 1024 * 1024

// Arena represents our raw memory block.
// In the spirit of Graydon Hoare, we take control of the bytes.
type Arena struct {
	Memory [HEAP_SIZE]byte
	Offset uintptr // Points to the next free byte (simple bump pointer for now)
}

// Global "RAM" instance
var RAM Arena

// Reset completely wipes the memory (dangerous!)
func (a *Arena) Reset() {
	a.Offset = 0
	// Optional: Zero out memory to debug issues later
	for i := range a.Memory {
		a.Memory[i] = 0
	}
}

// GetPointer returns the raw unsafe pointer to the start of our heap
func (a *Arena) BasePointer() unsafe.Pointer {
	return unsafe.Pointer(&a.Memory[0])
}
