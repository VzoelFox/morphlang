package memory

// Ptr represents a VIRTUAL address in our Morph Memory System.
// Unlike a standard pointer, it does not point to a physical memory address directly.
// Instead, it acts like a handle:
//
//   [  Drawer ID (32 bits)  ] [   Offset (32 bits)    ]
//   Total 64 bits.
//
// This indirection allows us to move "Drawers" (Memory Pages) between RAM and Disk
// (Swapping/Draft Otomatis) without breaking the pointers held by the program.
//
// To access the data, you must call `Resolve()` which translates this Virtual Ptr
// to a Physical unsafe.Pointer (and potentially triggers a page fault/swap-in).
type Ptr uint64

const (
	NilPtr Ptr = 0
)

// Bit masks and shifts
const (
	OffsetMask = 0xFFFFFFFF
	DrawerShift = 32
)

// NewPtr creates a Virtual Pointer from Drawer ID and Offset.
func NewPtr(drawerID int, offset uint32) Ptr {
	return Ptr(uint64(drawerID)<<DrawerShift | uint64(offset))
}

// DrawerID extracts the Drawer ID from the virtual pointer.
func (p Ptr) DrawerID() int {
	return int(uint64(p) >> DrawerShift)
}

// Offset extracts the intra-drawer offset from the virtual pointer.
func (p Ptr) Offset() uint32 {
	return uint32(uint64(p) & OffsetMask)
}

// Add adds an offset to the pointer.
// Note: This does not handle crossing Drawer boundaries!
// If an object spans across drawers, we are in trouble.
// For now, we assume objects fit in a single Drawer (128KB).
func (p Ptr) Add(offset uint32) Ptr {
	newOffset := p.Offset() + offset
	// We could check for overflow here
	return NewPtr(p.DrawerID(), newOffset)
}
