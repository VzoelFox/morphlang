package memory

import (
	"sync"
	"time"
)

// Constants for sizing
const (
	TRAY_SIZE   = 64 * 1024       // 64KB per Tray
	DRAWER_SIZE = TRAY_SIZE * 2   // 128KB per Drawer (2 Trays)

	// PHYSICAL_SLOTS is the actual RAM capacity (16 slots of 128KB = 2MB)
	PHYSICAL_SLOTS = 16

	// MAX_VIRTUAL_DRAWERS is the limit of our "virtual" cabinet
	MAX_VIRTUAL_DRAWERS = 1024
)

// DrawerLease represents exclusive ownership of a drawer by a unit
type DrawerLease struct {
	DrawerID   int
	UnitID     int
	SnapshotID int64 // Pointer to the snapshot taken at acquire time
	IsActive   bool
}

// Tray represents a semi-space (Nampan).
// It's a contiguous block of memory where we bump-allocate objects.
type Tray struct {
	Start   Ptr    // Virtual Start Pointer
	End     Ptr    // Virtual End Pointer
	Current Ptr    // Current Virtual allocation pointer
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

	// GC Metadata
	AccessCount int64 // LFU Tracking

	// Lease System
	Lease *DrawerLease
}

// Cabinet represents the entire Heap (Lemari).
type Cabinet struct {
	// TODO: PERFORMANCE BOTTLENECK
	// Currently using a global Mutex for simplicity during Phase X.
	// This serializes all memory access and will be a contention point for multi-threaded workloads.
	// FUTURE OPTIMIZATION:
	// 1. Upgrade to sync.RWMutex to allow concurrent reads (resolve).
	// 2. Implement per-Drawer locking for finer granularity.
	// 3. Separate "Fast Path" (Resident Read) from "Slow Path" (Swap In Write).
	mu sync.RWMutex

	// Virtual Drawers
	Drawers []Drawer

	// Physical RAM Slots (Map SlotIndex -> DrawerID)
	// -1 means empty slot
	RAMSlots [PHYSICAL_SLOTS]int

	ActiveDrawerIndex int

	// Snapshot Store (Simple In-Memory Map for Phase X.1)
	Snapshots      map[int64][]byte
	NextSnapshotID int64

	// GC Interface
	RootProvider func() []*Ptr
	IsGCRunning  bool
}

// Global Cabinet instance
var Lemari Cabinet

// Initialize the Cabinet structure (carving up the Arena)
func InitCabinet() {
	RAM.Reset()
	InitSwap() // Initialize swap file

	Lemari.Drawers = make([]Drawer, 0, MAX_VIRTUAL_DRAWERS)
	Lemari.Snapshots = make(map[int64][]byte)
	Lemari.NextSnapshotID = 1

	// Reset RAM Slots
	for i := 0; i < PHYSICAL_SLOTS; i++ {
		Lemari.RAMSlots[i] = -1
	}

	// Create initial drawers filling the RAM
	for i := 0; i < PHYSICAL_SLOTS; i++ {
		CreateDrawer()
	}

	Lemari.ActiveDrawerIndex = 0

	// Start Background GC (Daemon)
	// Runs every 1 second to apply aging and free up slots.
	StartGC(1 * time.Second)
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

	SetupTrayPointers(&drawer)

	if slot != -1 {
		// Assign RAM
		Lemari.RAMSlots[slot] = id
	} else {
		// Created in "Swapped" state (or empty state waiting for swap in)
		drawer.IsSwapped = true
	}

	Lemari.Drawers = append(Lemari.Drawers, drawer)
	return &Lemari.Drawers[id]
}

// SetupTrayPointers sets up the VIRTUAL pointers for the drawer.
// These are invariant of the physical location.
func SetupTrayPointers(d *Drawer) {
	// Reserve address 0 (NilPtr) if this is Drawer 0
	baseOffset := uint32(0)
	if d.ID == 0 {
		baseOffset = 8
	}

	start := NewPtr(d.ID, baseOffset)
	mid   := NewPtr(d.ID, uint32(TRAY_SIZE))
	end   := NewPtr(d.ID, uint32(DRAWER_SIZE))

	d.PrimaryTray.Start = start
	d.PrimaryTray.Current = start
	d.PrimaryTray.End = mid

	d.SecondaryTray.Start = mid
	d.SecondaryTray.Current = mid
	d.SecondaryTray.End = end
}

func (t *Tray) Remaining() int {
	return int(t.End) - int(t.Current)
}
