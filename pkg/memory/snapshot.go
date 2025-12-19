package memory

import (
	"encoding/gob"
	"fmt"
	"os"
	"unsafe"
)

// Snapshot saves the entire memory state to a file.
// Format: Gob Stream (Cabinet Metadata, then Drawer 0 Blob, Drawer 1 Blob, ...)
func Snapshot(filename string) error {
	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := gob.NewEncoder(f)

	// 1. Save Cabinet Metadata
	if err := enc.Encode(&Lemari); err != nil {
		return err
	}

	// 2. Save Drawer Contents
	for i := range Lemari.Drawers {
		d := &Lemari.Drawers[i]

		var data []byte

		if d.PhysicalSlot != -1 {
			// Resident in RAM
			ptr := unsafe.Add(RAM.BasePointer(), uintptr(d.PhysicalSlot)*uintptr(DRAWER_SIZE))
			data = unsafe.Slice((*byte)(ptr), DRAWER_SIZE)
		} else if d.IsSwapped && d.SwapOffset > 0 {
			// In Swap File
			data = make([]byte, DRAWER_SIZE)
			if err := Swap.Restore(d.SwapOffset, data); err != nil {
				return err
			}
		} else {
			// Empty / Zero
			data = make([]byte, DRAWER_SIZE)
		}

		// Write to snapshot using Gob
		if err := enc.Encode(data); err != nil {
			return err
		}
	}

	return nil
}

// SnapshotDrawer creates an in-memory snapshot of a specific drawer.
func (c *Cabinet) SnapshotDrawer(drawerID int) (int64, error) {
	if drawerID < 0 || drawerID >= len(c.Drawers) {
		return 0, fmt.Errorf("invalid drawer ID: %d", drawerID)
	}

	d := &c.Drawers[drawerID]

	// Data buffer
	data := make([]byte, DRAWER_SIZE)

	if d.PhysicalSlot != -1 {
		// Copy from RAM
		ptr := unsafe.Add(RAM.BasePointer(), uintptr(d.PhysicalSlot)*uintptr(DRAWER_SIZE))
		src := unsafe.Slice((*byte)(ptr), DRAWER_SIZE)
		copy(data, src)
	} else if d.IsSwapped && d.SwapOffset > 0 {
		// Copy from Swap
		if err := Swap.Restore(d.SwapOffset, data); err != nil {
			return 0, err
		}
	}
	// Else: Empty drawer, zero buffer

	id := c.NextSnapshotID
	c.NextSnapshotID++
	c.Snapshots[id] = data

	return id, nil
}

// RestoreDrawer reverts a drawer to the state stored in snapshotID.
func (c *Cabinet) RestoreDrawer(drawerID int, snapshotID int64) error {
	if drawerID < 0 || drawerID >= len(c.Drawers) {
		return fmt.Errorf("invalid drawer ID: %d", drawerID)
	}

	data, ok := c.Snapshots[snapshotID]
	if !ok {
		return fmt.Errorf("snapshot %d not found", snapshotID)
	}

	d := &c.Drawers[drawerID]

	if d.PhysicalSlot != -1 {
		// Restore to RAM
		ptr := unsafe.Add(RAM.BasePointer(), uintptr(d.PhysicalSlot)*uintptr(DRAWER_SIZE))
		dst := unsafe.Slice((*byte)(ptr), DRAWER_SIZE)
		copy(dst, data)
	} else {
		// Restore to Swap (Spill new version)
		offset, err := Swap.Spill(data)
		if err != nil {
			return err
		}
		d.SwapOffset = offset
		d.IsSwapped = true
	}

	return nil
}

// Restore loads memory state from a file.
// It effectively rewinds the memory to the snapshot state.
// Strategy: Load all data into the Swap File and mark all Drawers as "Swapped".
// This minimizes RAM usage usage initially and relies on Demand Paging.
func Restore(filename string) error {
	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	dec := gob.NewDecoder(f)

	// 0. Reset RAM and Swap
	RAM.Reset()
	Swap.FreeCache()
	if err := InitSwap(); err != nil {
		return err
	}

	// 1. Load Cabinet Metadata
	if err := dec.Decode(&Lemari); err != nil {
		return err
	}

	// 2. Reset RAM Slots
	for i := 0; i < PHYSICAL_SLOTS; i++ {
		Lemari.RAMSlots[i] = -1
	}

	// 3. Reconstruct Data & Update Drawers
	for i := range Lemari.Drawers {
		d := &Lemari.Drawers[i]

		var buffer []byte // Gob will allocate new slice
		if err := dec.Decode(&buffer); err != nil {
			return err
		}

		// Write to new Swap File
		offset, err := Swap.Spill(buffer)
		if err != nil {
			return err
		}

		// Update Drawer State to "Swapped"
		d.IsSwapped = true
		d.SwapOffset = offset
		d.PhysicalSlot = -1
	}

	return nil
}
