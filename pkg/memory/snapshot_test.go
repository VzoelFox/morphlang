package memory

import (
	"os"
	"testing"
)

func TestSnapshotRewind(t *testing.T) {
	InitCabinet()
	defer Swap.FreeCache()
	defer os.Remove("test_snapshot.z")

	// 1. Alloc Setup
	p1, err := AllocInteger(100)
	if err != nil { t.Fatal(err) }

	p2, err := AllocInteger(200)
	if err != nil { t.Fatal(err) }

	// 2. Snapshot
	if err := Snapshot("test_snapshot.z"); err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	// 3. Modify State
	// Change p1 to 999
	if err := WriteInteger(p1, 999); err != nil {
		t.Fatal(err)
	}

	// Verify modification
	v1, _ := ReadInteger(p1)
	if v1 != 999 {
		t.Fatalf("Expected 999, got %d", v1)
	}

	// Alloc new stuff (p3) - Should be in same drawer
	p3, err := AllocInteger(300)
	if err != nil { t.Fatal(err) }
	v3, _ := ReadInteger(p3)
	if v3 != 300 { t.Fatal("p3 alloc failed") }

	// 4. Restore/Rewind
	if err := Restore("test_snapshot.z"); err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// 5. Verify State
	// p1 should be 100
	v1_restored, err := ReadInteger(p1)
	if err != nil { t.Fatal(err) }
	if v1_restored != 100 {
		t.Errorf("Rewind failed for p1. Expected 100, got %d", v1_restored)
	}

	// p2 should be 200
	v2_restored, err := ReadInteger(p2)
	if err != nil { t.Fatal(err) }
	if v2_restored != 200 {
		t.Errorf("Rewind failed for p2. Expected 200, got %d", v2_restored)
	}

	// p3 should be 0 (garbage/zeroed because it didn't exist in snapshot)
	v3_restored, err := ReadInteger(p3)
	if err != nil {
		t.Logf("p3 access failed as expected (or not?): %v", err)
	} else {
		// If it succeeds, likely value is 0 because Header Type is 0
		// But ReadInteger doesn't check header type strictly.
		// It just reads value offset.
		if v3_restored != 0 {
			t.Errorf("p3 should be 0 (unallocated in snapshot), got %d", v3_restored)
		}
	}
}
