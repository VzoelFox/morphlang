package memory

import (
	"bytes"
	"fmt"
	"testing"
)

func TestSwapping(t *testing.T) {
	InitCabinet()

	// RAM has 16 slots.
	// Each Drawer takes 1 slot.
	// We want to create > 16 drawers.
	// A drawer is created when the current one is full.
	// Tray size is 64KB.
	// So if we allocate 60KB, the next alloc will force a new drawer.

	chunkSize := 60 * 1024 // 60KB
	numChunks := 20        // 20 drawers > 16 slots

	ptrs := make([]Ptr, numChunks)

	// 1. Fill RAM and Overflow to Swap
	for i := 0; i < numChunks; i++ {
		// Unique pattern for verification
		pattern := byte(i + 1)
		data := bytes.Repeat([]byte{pattern}, chunkSize)

		ptr, err := Lemari.Alloc(chunkSize)
		if err != nil {
			t.Fatalf("Alloc failed at chunk %d: %v", i, err)
		}
		ptrs[i] = ptr

		err = Write(ptr, data)
		if err != nil {
			t.Fatalf("Write failed at chunk %d: %v", i, err)
		}

		// t.Logf("Allocated chunk %d at %v (Drawer %d)", i, ptr, ptr.DrawerID())
	}

	// At this point, Drawers 0-3 should be evicted (Swapped Out) to make room for 16-19.
	// Check Drawer 0 status
	Lemari.mu.Lock()
	d0 := &Lemari.Drawers[0]
	if !d0.IsSwapped {
		t.Error("Drawer 0 should be swapped out")
	}
	Lemari.mu.Unlock()

	// 2. Verify Data (Triggers Restore)
	// Read Chunk 0 (should bring Drawer 0 back to RAM, evicting someone else)
	fmt.Println("Verifying Chunk 0 (Swap Restore)...")
	readData, err := Read(ptrs[0], chunkSize)
	if err != nil {
		t.Fatalf("Read failed for chunk 0: %v", err)
	}

	expectedPattern := byte(1)
	expectedData := bytes.Repeat([]byte{expectedPattern}, chunkSize)
	if !bytes.Equal(readData, expectedData) {
		t.Errorf("Data corruption in Chunk 0! First byte: %d", readData[0])
	}

	// 3. Verify Chunk 19 (should still be in RAM or swapped out depending on LRU?)
	// Our eviction policy is simple: Find free slot, or evict RAMSlots[0]?
	// Allocator `bringToRAM`:
	// if freeSlot == -1 { victimID := c.RAMSlots[0] ... }
	// It evicts index 0 of RAMSlots.
	// So it's a Ring Buffer / FIFO eviction?
	// Let's check `allocator.go` logic again if needed.
	// But as long as Read works, the system works.

	fmt.Println("Verifying Chunk 19...")
	readData19, err := Read(ptrs[19], chunkSize)
	if err != nil {
		t.Fatalf("Read failed for chunk 19: %v", err)
	}
	if readData19[0] != byte(20) {
		t.Errorf("Data corruption in Chunk 19!")
	}
}
