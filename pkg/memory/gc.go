package memory

import (
	"fmt"
	"time"
	"unsafe"
)

var gcStop chan bool

// StartGC starts the background Garbage Collector daemon.
// It runs periodically to manage memory pressure and apply LFU aging.
func StartGC(interval time.Duration) {
	if gcStop != nil {
		close(gcStop) // Stop previous daemon to prevent race/decay storm
	}
	gcStop = make(chan bool)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				Lemari.LFUAging()
			case <-gcStop:
				return
			}
		}
	}()
}

// LFUAging performs background maintenance.
// 1. Aging: Decays AccessCounts to ensure recent usage is prioritized.
// 2. Preemptive Eviction: Frees up RAM slots if pressure is high.
func (c *Cabinet) LFUAging() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 1. Aging (Decay)
	// Iterate through all drawers to decay history
	for i := range c.Drawers {
		if c.Drawers[i].AccessCount > 0 {
			c.Drawers[i].AccessCount /= 2 // Halve the count (Right Shift)
		}
	}

	// 2. Check Pressure
	freeSlots := 0
	for _, id := range c.RAMSlots {
		if id == -1 {
			freeSlots++
		}
	}

	// Maintain a buffer of free slots (e.g., 2 slots)
	// If fewer than 2 slots are free, evict the coldest ones until we have breathing room.
	// Limit to 1 eviction per cycle to prevent stutter.
	if freeSlots < 2 {
		victimSlot := c.findVictimLFU()
		if victimSlot != -1 {
			victimID := c.RAMSlots[victimSlot]

			// Log for debugging (optional)
			// fmt.Printf("GC: Evicting Cold Drawer %d (Access: %d)\n", victimID, c.Drawers[victimID].AccessCount)

			err := c.evictToSwap(victimID)
			if err != nil {
				fmt.Printf("GC Error: Failed to evict drawer %d: %v\n", victimID, err)
			}
		}
	}
}

// findVictimLFU finds the RAM slot containing the drawer with the lowest AccessCount.
// Returns the Slot Index, or -1 if no suitable victim found.
// Assumes Lock held.
func (c *Cabinet) findVictimLFU() int {
	minScore := int64(-1)
	victimSlot := -1

	// Iterate over physical slots to find a resident victim
	for i, drawerID := range c.RAMSlots {
		if drawerID == -1 {
			continue
		}

		drawer := &c.Drawers[drawerID]

		// Initialize minScore or find smaller
		// We use <= to evict lower slots first in case of ties (FIFO fallback)
		if minScore == -1 || drawer.AccessCount < minScore {
			minScore = drawer.AccessCount
			victimSlot = i
		}
		// DEBUG
		// fmt.Printf("Slot %d (ID %d): Score %d. Current Victim: %d\n", i, drawerID, drawer.AccessCount, victimSlot)
	}

	return victimSlot
}

// MarkAndCompact performs a Stop-the-World, Copying Garbage Collection.
// 1. Flips the Active Trays (Semi-Space).
// 2. Evacuates Roots (Stack + Globals).
// 3. Scans and Evacuates reachable objects (Cheney's Algorithm).
// 4. Discards old Trays (Implicitly by reuse).
func (c *Cabinet) MarkAndCompact(roots []*Ptr) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.IsGCRunning {
		return fmt.Errorf("GC re-entry")
	}
	c.IsGCRunning = true
	defer func() { c.IsGCRunning = false }()

	// 1. Flip Trays
	for i := range c.Drawers {
		d := &c.Drawers[i]
		d.IsPrimaryActive = !d.IsPrimaryActive

		var activeTray *Tray
		if d.IsPrimaryActive {
			activeTray = &d.PrimaryTray
		} else {
			activeTray = &d.SecondaryTray
		}
		activeTray.Current = activeTray.Start
	}

	// Reset ActiveDrawerIndex to 0 to compact from the beginning
	c.ActiveDrawerIndex = 0

	// 2. Evacuate Roots
	for _, root := range roots {
		newPtr, err := c.evacuate(*root)
		if err != nil { return err }
		*root = newPtr
	}

	// 3. Scan (Cheney's Algorithm)
	scanIdx := 0
	scanTray := c.getTray(0)
	scanPtr := scanTray.Start

	for {
		// Move to next drawer if current tray exhausted
		if scanPtr >= scanTray.Current {
			scanIdx++
			if scanIdx > c.ActiveDrawerIndex {
				break // Done
			}
			scanTray = c.getTray(scanIdx)
			scanPtr = scanTray.Start
			continue
		}

		// Read Object at scanPtr
		raw, err := c.resolve(scanPtr)
		if err != nil { return err }
		header := (*Header)(raw)
		size := header.Size

		// Scan children
		children, err := Scan(scanPtr)
		if err != nil { return err }

		for _, child := range children {
			newPtr, err := c.evacuate(*child)
			if err != nil { return err }
			*child = newPtr
		}

		// Advance
		scanPtr = scanPtr.Add(uint32(size))
	}

	return nil
}

func (c *Cabinet) getTray(drawerIdx int) *Tray {
	d := &c.Drawers[drawerIdx]
	if d.IsPrimaryActive {
		return &d.PrimaryTray
	}
	return &d.SecondaryTray
}

// evacuate moves an object to the To-Space if not already moved.
// Returns the new pointer.
func (c *Cabinet) evacuate(oldPtr Ptr) (Ptr, error) {
	if oldPtr == NilPtr { return NilPtr, nil }

	// Resolve old object (From-Space)
	raw, err := c.resolve(oldPtr)
	if err != nil { return NilPtr, err }

	header := (*Header)(raw)

	// Check Forwarding
	if header.Forwarding != NilPtr {
		return header.Forwarding, nil
	}

	// Copy
	size := int(header.Size)
	newPtr, err := c.alloc(size) // Allocates in To-Space
	if err != nil { return NilPtr, err }

	newRaw, err := c.resolve(newPtr)
	if err != nil { return NilPtr, err }

	// Memcpy
	src := unsafe.Slice((*byte)(raw), size)
	dst := unsafe.Slice((*byte)(newRaw), size)
	copy(dst, src)

	// Update Forwarding in Old Object
	// Re-resolve because alloc might have caused page fault/eviction
	raw, _ = c.resolve(oldPtr)
	header = (*Header)(raw)
	header.Forwarding = newPtr

	return newPtr, nil
}
