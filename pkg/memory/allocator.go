package memory

import (
	"fmt"
	"unsafe"
)

// Alloc allocates memory. It handles "Draft Otomatis" (Swapping) if RAM is full.
// Thread-safe.
func (c *Cabinet) Alloc(size int) (Ptr, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.alloc(size)
}

// Internal recursive alloc (assumes lock held)
func (c *Cabinet) alloc(size int) (Ptr, error) {
	// Auto-initialize if needed (Lazy Init)
	// Safe because Lock is held by caller.
	if len(c.Drawers) == 0 {
		RAM.Reset()
		InitSwap()
		c.Drawers = make([]Drawer, 0, MAX_VIRTUAL_DRAWERS)
		c.Snapshots = make(map[int64][]byte)
		c.NextSnapshotID = 1
		for i := 0; i < PHYSICAL_SLOTS; i++ {
			c.RAMSlots[i] = -1
		}
		for i := 0; i < PHYSICAL_SLOTS; i++ {
			CreateDrawer()
		}
		c.ActiveDrawerIndex = 0
	}
	if size <= 0 {
		return NilPtr, fmt.Errorf("invalid allocation size: %d", size)
	}

	alignedSize := (size + 7) &^ 7

	if alignedSize > TRAY_SIZE {
		return NilPtr, fmt.Errorf("allocation size %d exceeds tray limit %d", alignedSize, TRAY_SIZE)
	}

	activeDrawer := &c.Drawers[c.ActiveDrawerIndex]

	// Ensure active drawer is in RAM (Resident)
	if activeDrawer.PhysicalSlot == -1 {
		if err := c.bringToRAM(c.ActiveDrawerIndex); err != nil {
			return NilPtr, err
		}
		// Re-fetch pointer
		activeDrawer = &c.Drawers[c.ActiveDrawerIndex]
	}

	var activeTray *Tray
	if activeDrawer.IsPrimaryActive {
		activeTray = &activeDrawer.PrimaryTray
	} else {
		activeTray = &activeDrawer.SecondaryTray
	}

	if activeTray.Remaining() < alignedSize {
		// Drawer Full. Check if next drawer exists (Reusable)
		nextID := c.ActiveDrawerIndex + 1
		if nextID < len(c.Drawers) {
			c.ActiveDrawerIndex = nextID
			// Ensure resident
			if c.Drawers[nextID].PhysicalSlot == -1 {
				if err := c.bringToRAM(nextID); err != nil {
					return NilPtr, err
				}
			}
			// Recurse
			return c.alloc(size)
		}

		// Create new drawer.
		newDrawer := CreateDrawer()
		if newDrawer == nil {
			return NilPtr, fmt.Errorf("virtual memory limit reached")
		}

		c.ActiveDrawerIndex = newDrawer.ID

		// Ensure new drawer is resident
		if newDrawer.PhysicalSlot == -1 {
			if err := c.bringToRAM(newDrawer.ID); err != nil {
				return NilPtr, err
			}
			newDrawer = &c.Drawers[c.ActiveDrawerIndex]
		}

		// Recurse
		return c.alloc(size)
	}

	ptr := activeTray.Current
	// Bump pointer
	activeTray.Current += Ptr(alignedSize)

	return ptr, nil
}

// bringToRAM ensures the drawer is in physical RAM.
// Assumes c.mu is Locked.
func (c *Cabinet) bringToRAM(drawerID int) error {
	target := &c.Drawers[drawerID]
	if target.PhysicalSlot != -1 {
		return nil // Already resident
	}

	// Find free slot
	freeSlot := -1
	for i := 0; i < PHYSICAL_SLOTS; i++ {
		if c.RAMSlots[i] == -1 {
			freeSlot = i
			break
		}
	}

	// If no free slot, EVICT victim
	if freeSlot == -1 {
		// Use LFU Strategy to find victim
		victimSlot := c.findVictimLFU()
		if victimSlot == -1 {
			return fmt.Errorf("OOM: No suitable victim found for eviction")
		}

		victimID := c.RAMSlots[victimSlot]
		if victimID == drawerID {
			return fmt.Errorf("deadlock: trying to evict self")
		}

		err := c.evictToSwap(victimID)
		if err != nil {
			return err
		}
		freeSlot = victimSlot
	}

	// Calculate Physical Address for this slot
	physAddr := unsafe.Add(RAM.BasePointer(), uintptr(freeSlot)*uintptr(DRAWER_SIZE))
	destSlice := unsafe.Slice((*byte)(physAddr), DRAWER_SIZE)

	// Load data if it was previously saved
	if target.SwapOffset > 0 {
		err := Swap.Restore(target.SwapOffset, destSlice)
		if err != nil {
			return err
		}
	} else {
		// Zero out memory for fresh drawer
		for i := range destSlice {
			destSlice[i] = 0
		}
	}

	c.RAMSlots[freeSlot] = drawerID
	target.PhysicalSlot = freeSlot
	// target.IsSwapped was used as "Is currently swapped out".
	// The struct comment said "IsSwapped bool". Let's use PhysicalSlot != -1 as the truth.
	// But `IsSwapped` might mean "Has backup on disk".
	// Let's interpret `IsSwapped` as "Currently NOT in RAM".
	target.IsSwapped = false

	return nil
}

// evictToSwap moves a drawer from RAM to Disk.
// Assumes c.mu is Locked.
func (c *Cabinet) evictToSwap(drawerID int) error {
	victim := &c.Drawers[drawerID]
	slot := victim.PhysicalSlot

	if slot == -1 {
		return nil // Already swapped
	}

	// 1. Get RAM content
	physAddr := unsafe.Add(RAM.BasePointer(), uintptr(slot)*uintptr(DRAWER_SIZE))
	srcSlice := unsafe.Slice((*byte)(physAddr), DRAWER_SIZE)

	// 2. Write to .z file
	fileOffset, err := Swap.Spill(srcSlice)
	if err != nil {
		return err
	}

	// 3. Update State
	victim.IsSwapped = true
	victim.SwapOffset = fileOffset
	victim.PhysicalSlot = -1

	// 4. Free RAM slot
	c.RAMSlots[slot] = -1

	return nil
}

// Write writes data to the pointer address.
// Thread-safe.
func Write(dst Ptr, data []byte) error {
	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	rawPtr, err := Lemari.resolve(dst)
	if err != nil {
		return err
	}

	size := len(data)
	targetSlice := unsafe.Slice((*byte)(rawPtr), size)
	copy(targetSlice, data)
	return nil
}

// Read reads data from the pointer address.
// Thread-safe.
func Read(src Ptr, size int) ([]byte, error) {
	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	rawPtr, err := Lemari.resolve(src)
	if err != nil {
		return nil, err
	}

	srcSlice := unsafe.Slice((*byte)(rawPtr), size)
	ret := make([]byte, size)
	copy(ret, srcSlice)
	return ret, nil
}
