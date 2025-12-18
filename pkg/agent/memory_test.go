package agent

import (
	"os"
	"testing"
)

func TestMemoryChain(t *testing.T) {
	tmp := "test_memory.json"
	defer os.Remove(tmp)

	mem, err := LoadMemory(tmp)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	err = mem.Record("User1", "Assist1")
	if err != nil {
		t.Fatalf("Record 1 failed: %v", err)
	}

	err = mem.Record("User2", "Assist2")
	if err != nil {
		t.Fatalf("Record 2 failed: %v", err)
	}

	// Reload and Verify
	mem2, err := LoadMemory(tmp)
	if err != nil {
		t.Fatalf("Reload failed: %v", err)
	}

	if len(mem2.Interactions) != 2 {
		t.Fatalf("Expected 2 interactions, got %d", len(mem2.Interactions))
	}

	if mem2.Interactions[1].PrevHash != mem2.Interactions[0].Hash {
		t.Errorf("Chain broken")
	}
}
