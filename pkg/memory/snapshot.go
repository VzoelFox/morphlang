package memory

import (
	"encoding/gob"
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
