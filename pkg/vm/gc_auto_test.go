package vm

import (
	"fmt"
	"testing"

	"github.com/VzoelFox/morphlang/pkg/compiler"
	"github.com/VzoelFox/morphlang/pkg/memory"
	"github.com/VzoelFox/morphlang/pkg/object"
)

func TestAutoGC_Trigger(t *testing.T) {
	// 1. Setup minimal VM
	instructions := compiler.Instructions{}
	bytecode := &compiler.Bytecode{Instructions: instructions, Constants: []object.Object{}}

	vm := New(bytecode)

	// 2. Simulate "Running" State
	GlobalVMLock.RLock()
	defer GlobalVMLock.RUnlock()

	// 3. Keep a persistent object
	keepPtr, err := memory.AllocInteger(12345)
	if err != nil {
		t.Fatalf("Initial allocation failed: %v", err)
	}
	vm.stack[0] = keepPtr
	vm.sp = 1

	// 4. Stress Test Loop
	// Tray Limit is 64KB. We alloc 40KB.
	// Drawer Limit is 128MB (Total Virtual).
	// 128MB / 40KB = ~3200 allocations.
	// We run 4000 to be sure we cross the limit.

	fmt.Println("Starting GC Stress Test (4000 x 40KB)...")

	chunkSize := 40 * 1024

	for i := 0; i < 4000; i++ {
		// Allocate garbage
		_, err := memory.Lemari.Alloc(chunkSize)
		if err != nil {
			t.Fatalf("Iteration %d: Allocation failed with error: %v", i, err)
		}

		if i > 0 && i%500 == 0 {
			// Verify our keepPtr is still valid
			val, err := memory.ReadInteger(keepPtr)
			if err != nil {
				t.Fatalf("Iteration %d: keepPtr corrupted: %v", i, err)
			}
			if val != 12345 {
				t.Fatalf("Iteration %d: keepPtr value changed: %d", i, val)
			}
			fmt.Printf("Iteration %d: OK (keepPtr valid)\n", i)
		}
	}

	fmt.Println("GC Stress Test Passed: Handled OOM automatically.")
}
