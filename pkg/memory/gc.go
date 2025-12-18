package memory

import (
	"fmt"
	"time"
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
				Lemari.GC()
			case <-gcStop:
				return
			}
		}
	}()
}

// GC performs a garbage collection cycle.
// 1. Aging: Decays AccessCounts to ensure recent usage is prioritized.
// 2. Preemptive Eviction: Frees up RAM slots if pressure is high.
func (c *Cabinet) GC() {
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
