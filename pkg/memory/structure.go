package memory

// Constants for sizing
const (
	TRAY_SIZE   = 64 * 1024       // 64KB per Tray
	DRAWER_SIZE = TRAY_SIZE * 2   // 128KB per Drawer (2 Trays)

	// PHYSICAL_SLOTS is the actual RAM capacity (16 slots of 128KB = 2MB)
	PHYSICAL_SLOTS = 16

	// MAX_VIRTUAL_DRAWERS is the limit of our "virtual" cabinet
	MAX_VIRTUAL_DRAWERS = 1024
)

// Tray represents a semi-space (Nampan).
// It's a contiguous block of memory where we bump-allocate objects.
type Tray struct {
	Start   Ptr    // Start offset in the Arena (relative to Drawer base if we used relative, but here it's absolute Ptr)
	End     Ptr    // End offset
	Current Ptr    // Current bump pointer
}

// Drawer represents a memory region (Laci).
// It contains two Trays for copying garbage collection (FromSpace/ToSpace).
type Drawer struct {
	ID           int

	// State for Swap/Draft
	IsSwapped    bool
	SwapOffset   int64

	// Physical Mapping
	PhysicalSlot int // -1 if swapped out

	PrimaryTray  Tray
	SecondaryTray Tray
	IsPrimaryActive bool // Which tray is currently receiving allocations?
}

// Cabinet represents the entire Heap (Lemari).
type Cabinet struct {
	// Virtual Drawers
	Drawers []Drawer

	// Physical RAM Slots (Map SlotIndex -> DrawerID)
	// -1 means empty slot
	RAMSlots [PHYSICAL_SLOTS]int

	ActiveDrawerIndex int
}

// Global Cabinet instance
var Lemari Cabinet

// Initialize the Cabinet structure (carving up the Arena)
func InitCabinet() {
	RAM.Reset()
	InitSwap() // Initialize swap file

	Lemari.Drawers = make([]Drawer, 0, MAX_VIRTUAL_DRAWERS)

	// Reset RAM Slots
	for i := 0; i < PHYSICAL_SLOTS; i++ {
		Lemari.RAMSlots[i] = -1
	}

	// Create initial drawers filling the RAM
	for i := 0; i < PHYSICAL_SLOTS; i++ {
		CreateDrawer()
	}

	Lemari.ActiveDrawerIndex = 0
}

// CreateDrawer creates a new virtual drawer and tries to assign a physical slot
func CreateDrawer() *Drawer {
	id := len(Lemari.Drawers)
	if id >= MAX_VIRTUAL_DRAWERS {
		return nil // Limit reached
	}

	// Find free RAM slot
	slot := -1
	for i := 0; i < PHYSICAL_SLOTS; i++ {
		if Lemari.RAMSlots[i] == -1 {
			slot = i
			break
		}
	}

	drawer := Drawer{
		ID:           id,
		PhysicalSlot: slot,
		IsPrimaryActive: true,
	}

	if slot != -1 {
		// Assign RAM
		Lemari.RAMSlots[slot] = id
		SetupTrayPointers(&drawer, slot)
	} else {
		// Created in "Swapped" state (or empty state waiting for swap in)
		// For simplicity, we assume new drawers start in RAM.
		// If RAM is full, we must Evict someone else first.
		// But here we just create struct. Allocator handles eviction.
		drawer.IsSwapped = true // Temporarily mark as swapped (empty)
	}

	Lemari.Drawers = append(Lemari.Drawers, drawer)
	return &Lemari.Drawers[id]
}

// SetupTrayPointers calculates absolute Ptrs based on PhysicalSlot
func SetupTrayPointers(d *Drawer, slot int) {
	// Base address logic
	// Slot 0 starts at 8 (reserved null)
	baseOffset := uintptr(8) + uintptr(slot)*uintptr(DRAWER_SIZE)
	base := Ptr(baseOffset)

	d.PrimaryTray.Start = base
	d.PrimaryTray.Current = base
	d.PrimaryTray.End = base + Ptr(TRAY_SIZE)

	base += Ptr(TRAY_SIZE)

	d.SecondaryTray.Start = base
	d.SecondaryTray.Current = base
	d.SecondaryTray.End = base + Ptr(TRAY_SIZE)
}

func (t *Tray) Remaining() int {
	return int(t.End - t.Current)
}
