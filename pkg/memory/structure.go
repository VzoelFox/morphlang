package memory

// Constants for sizing
const (
	TRAY_SIZE   = 64 * 1024       // 64KB per Tray
	DRAWER_SIZE = TRAY_SIZE * 2   // 128KB per Drawer (2 Trays)
	MAX_DRAWERS = 16              // Total 2MB for start
)

// Tray represents a semi-space (Nampan).
// It's a contiguous block of memory where we bump-allocate objects.
type Tray struct {
	Start   Ptr    // Start offset in the Arena
	End     Ptr    // End offset in the Arena
	Current Ptr    // Current bump pointer
}

// Drawer represents a memory region (Laci).
// It contains two Trays for copying garbage collection (FromSpace/ToSpace).
type Drawer struct {
	ID           int
	PrimaryTray  Tray
	SecondaryTray Tray
	IsPrimaryActive bool // Which tray is currently receiving allocations?
}

// Cabinet represents the entire Heap (Lemari).
type Cabinet struct {
	Drawers [MAX_DRAWERS]Drawer
	ActiveDrawerIndex int
}

// Global Cabinet instance
var Lemari Cabinet

// Initialize the Cabinet structure (carving up the Arena)
func InitCabinet() {
	RAM.Reset()
	base := Ptr(0)

	// We reserve address 0 as Nil
	base += 8

	for i := 0; i < MAX_DRAWERS; i++ {
		drawer := &Lemari.Drawers[i]
		drawer.ID = i
		drawer.IsPrimaryActive = true

		// Setup Primary Tray (Nampan A)
		drawer.PrimaryTray.Start = base
		drawer.PrimaryTray.Current = base
		drawer.PrimaryTray.End = base + Ptr(TRAY_SIZE)
		base += Ptr(TRAY_SIZE)

		// Setup Secondary Tray (Nampan B)
		drawer.SecondaryTray.Start = base
		drawer.SecondaryTray.Current = base
		drawer.SecondaryTray.End = base + Ptr(TRAY_SIZE)
		base += Ptr(TRAY_SIZE)
	}
	Lemari.ActiveDrawerIndex = 0
}

func (t *Tray) Remaining() int {
	return int(t.End - t.Current)
}
