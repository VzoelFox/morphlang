package scheduler

import (
	"sync"
	"testing"

	"github.com/VzoelFox/morphlang/pkg/memory"
)

func TestQueueFIFO(t *testing.T) {
	memory.InitCabinet()

	q, err := NewQueue()
	if err != nil {
		t.Fatalf("NewQueue failed: %v", err)
	}

	// Create items
	item1, _ := memory.AllocInteger(100)
	item2, _ := memory.AllocInteger(200)
	item3, _ := memory.AllocInteger(300)

	// Enqueue
	if err := q.Enqueue(item1); err != nil {
		t.Fatal(err)
	}
	if err := q.Enqueue(item2); err != nil {
		t.Fatal(err)
	}
	if err := q.Enqueue(item3); err != nil {
		t.Fatal(err)
	}

	// Dequeue
	val1, err := q.Dequeue()
	if err != nil {
		t.Fatal(err)
	}
	if val1 != item1 {
		t.Errorf("expected item1, got %v", val1)
	}

	val2, err := q.Dequeue()
	if err != nil {
		t.Fatal(err)
	}
	if val2 != item2 {
		t.Errorf("expected item2, got %v", val2)
	}

	val3, err := q.Dequeue()
	if err != nil {
		t.Fatal(err)
	}
	if val3 != item3 {
		t.Errorf("expected item3, got %v", val3)
	}

	// Empty check
	valEmpty, err := q.Dequeue()
	if err != nil {
		t.Fatal(err)
	}
	if valEmpty != memory.NilPtr {
		t.Errorf("expected NilPtr, got %v", valEmpty)
	}
}

func TestQueueConcurrency(t *testing.T) {
	memory.InitCabinet()
	q, _ := NewQueue()

	var wg sync.WaitGroup
	count := 100

	// Producers
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			ptr, _ := memory.AllocInteger(int64(val))
			q.Enqueue(ptr)
		}(i)
	}

	wg.Wait()

	// Consumers
	// We expect 100 items. Order is not deterministic between producers, but FIFO implies we get all.
	itemsFound := 0
	for i := 0; i < count+10; i++ { // Try a bit more to be sure
		val, _ := q.Dequeue()
		if val == memory.NilPtr {
			break
		}
		itemsFound++
	}

	if itemsFound != count {
		t.Errorf("expected %d items, got %d", count, itemsFound)
	}
}
