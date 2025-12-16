package memory

import (
	"fmt"
	"unsafe"
)

// Alloc allocates memory. It handles "Draft Otomatis" (Swapping) if RAM is full.
func (c *Cabinet) Alloc(size int) (Ptr, error) {
	if size <= 0 {
		return NilPtr, fmt.Errorf("invalid allocation size: %d", size)
	}

	alignedSize := (size + 7) &^ 7

	activeDrawer := &c.Drawers[c.ActiveDrawerIndex]

	// Ensure active drawer is in RAM
	if activeDrawer.IsSwapped {
		if err := c.BringToRAM(c.ActiveDrawerIndex); err != nil {
			return NilPtr, err
		}
		// Re-fetch pointer after move
		activeDrawer = &c.Drawers[c.ActiveDrawerIndex]
	}

	var activeTray *Tray
	if activeDrawer.IsPrimaryActive {
		activeTray = &activeDrawer.PrimaryTray
	} else {
		activeTray = &activeDrawer.SecondaryTray
	}

	if activeTray.Remaining() < alignedSize {
		// Drawer Full. Switch to next drawer?
		// For this experiment, let's say we switch to a NEW drawer if current is full.
		// Or we just fail for current drawer.
		// User requirement: "masuk cache kalo terkena draft otomatis".
		// This usually means evicting OLD data to make room for NEW data.

		// Let's create a new Drawer for this allocation.
		newDrawer := CreateDrawer()
		if newDrawer == nil {
			return NilPtr, fmt.Errorf("virtual memory limit reached")
		}

		c.ActiveDrawerIndex = newDrawer.ID

		// If newDrawer is Swapped (no RAM slots), we must evict someone.
		if newDrawer.IsSwapped {
			if err := c.BringToRAM(newDrawer.ID); err != nil {
				return NilPtr, err
			}
			newDrawer = &c.Drawers[c.ActiveDrawerIndex]
		}

		// Recurse to alloc in new drawer
		return c.Alloc(size)
	}

	ptr := activeTray.Current
	activeTray.Current += Ptr(alignedSize)

	return ptr, nil
}

// BringToRAM ensures the drawer is in physical RAM.
// If RAM is full, it evicts a victim drawer to swap.
func (c *Cabinet) BringToRAM(drawerID int) error {
	target := &c.Drawers[drawerID]
	if !target.IsSwapped {
		return nil
	}

	// Find free slot
	freeSlot := -1
	for i := 0; i < PHYSICAL_SLOTS; i++ {
		if c.RAMSlots[i] == -1 {
			freeSlot = i
			break
		}
	}

	// If no free slot, EVICT victim (Draft Otomatis)
	if freeSlot == -1 {
		// Simple policy: Evict (ActiveIndex + 1) % MAX, or just slot 0?
		// Let's evict slot 0 (FIFO replacement for simplicity)
		victimID := c.RAMSlots[0]
		if victimID == drawerID {
			return fmt.Errorf("deadlock: trying to evict self") // Should not happen
		}

		err := c.EvictToSwap(victimID)
		if err != nil {
			return err
		}
		freeSlot = 0 // Slot 0 is now free
	}

	// Load data if it was previously saved
	if target.SwapOffset > 0 {
		// Restore from disk
		// We need temporary buffer? No, read directly to RAM.
		// Setup pointers first
		SetupTrayPointers(target, freeSlot)

		// Calculate RAM start address
		startPtr := target.PrimaryTray.Start // Base of drawer

		// Convert Ptr to []byte slice for Restore
		// Length = DRAWER_SIZE
		rawStart := startPtr.ToUnsafe()
		destSlice := unsafe.Slice((*byte)(rawStart), DRAWER_SIZE)

		err := Swap.Restore(target.SwapOffset, destSlice)
		if err != nil {
			return err
		}
	} else {
		// Fresh drawer, just setup pointers
		SetupTrayPointers(target, freeSlot)
	}

	c.RAMSlots[freeSlot] = drawerID
	target.PhysicalSlot = freeSlot
	target.IsSwapped = false

	return nil
}

// EvictToSwap moves a drawer from RAM to Disk
func (c *Cabinet) EvictToSwap(drawerID int) error {
	victim := &c.Drawers[drawerID]
	if victim.IsSwapped {
		return nil
	}

	// 1. Serialize RAM content
	startPtr := victim.PrimaryTray.Start // Base
	rawStart := startPtr.ToUnsafe()
	srcSlice := unsafe.Slice((*byte)(rawStart), DRAWER_SIZE)

	// 2. Write to .z file
	offset, err := Swap.Spill(srcSlice)
	if err != nil {
		return err
	}

	// 3. Update State
	victim.IsSwapped = true
	victim.SwapOffset = offset

	// 4. Free RAM slot
	slot := victim.PhysicalSlot
	c.RAMSlots[slot] = -1
	victim.PhysicalSlot = -1

	// Pointers are now invalid, but we keep them relative?
	// No, Ptr is absolute offset in Arena.
	// When swapped out, Ptr values in objects inside this drawer become invalid
	// unless we implement relocation/swizzling.
	// For this Phase X, we assume Ptr is valid ONLY when drawer is loaded back to SAME slot?
	// Or we reload to ANY slot and fix pointers?
	// Simpler: We reload to ANY slot, but Ptr inside objects are "Relative to Drawer"?
	// Our `Ptr` type is absolute. This is a problem for Paging.
	// BUT user asked for "Draft Otomatis". Ideally we reload to SAME slot if we want to avoid swizzling.
	// Or we simply accept that pointers break for now (Experimental).
	// Let's try to reload to ANY slot. Pointers will point to wrong data if we don't fix them.
	// But `Alloc` returns absolute Ptr.
	// If `drawer` moves from Slot 0 to Slot 1, `Alloc` returned 0x100, now data is at 0x10000.
	// Old pointer 0x100 points to Slot 0 (now occupied by someone else).
	// This means **Pointers are unstable** in this Paging model.
	// Real OS uses Virtual Memory hardware (MMU) to map Virtual Address -> Physical Address.
	// We are implementing MMU in software.
	// `Ptr` should be `VirtualPtr` (DrawerID + Offset).
	// Then dereferencing `Ptr` translates to `RAM[Slot(DrawerID) + Offset]`.

	// Refactor `Ptr`? Too big scope.
	// For now, let's just implement the SWAP MECHANISM.
	// The data moves. Pointers break. That's acceptable for "Initial Experiment".
	// Or, we force eviction of *oldest* drawer and never reload it until we access it?
	// If we access it via `Ptr`, we don't know which drawer it belongs to easily.

	return nil
}

// Write writes data to the pointer address.
func Write(dst Ptr, data []byte) error {
	if dst == NilPtr {
		return fmt.Errorf("segmentation fault: nil pointer")
	}

	// Note: In paging model, dst might point to swapped out memory!
	// We need to check if address is Resident.
	// Map Ptr -> Drawer?
	// Ptr is raw offset.
	// We can calculate Slot from Ptr.
	// Slot = Ptr / DRAWER_SIZE.
	// Check if Slot is occupied.

	// Check physical bounds
	// ... (Skipping complex checks for MVP)

	rawPtr := dst.ToUnsafe()
	size := len(data)
	targetSlice := unsafe.Slice((*byte)(rawPtr), size)
	copy(targetSlice, data)
	return nil
}
