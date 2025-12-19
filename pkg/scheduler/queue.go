package scheduler

import (
	"github.com/VzoelFox/morphlang/pkg/memory"
)

// Offsets for Array Layout [Header(8)][Cap(4)][Len(4)][Elem0(8)][Elem1(8)]...
// We assume standard layout from pkg/memory/array.go
const (
	ArrayDataOffset = 24 // Header(16) + Cap(4) + Len(4)
	PtrSize         = 8
)

// Queue is a FIFO queue implemented in custom memory (Michael-Scott Algorithm).
type Queue struct {
	Address memory.Ptr
}

// NewQueue creates a new queue in memory.
func NewQueue() (*Queue, error) {
	// Queue Structure: [Head][Tail] (Array size 2)
	qPtr, err := memory.AllocArray(2, 2)
	if err != nil {
		return nil, err
	}

	// Sentinel Node: [Next][Value] (Array size 2)
	sentinel, err := memory.AllocArray(2, 2)
	if err != nil {
		return nil, err
	}

	// Init Head = Sentinel, Tail = Sentinel
	err = memory.WriteArrayElement(qPtr, 0, sentinel)
	if err != nil { return nil, err }
	err = memory.WriteArrayElement(qPtr, 1, sentinel)
	if err != nil { return nil, err }

	return &Queue{Address: qPtr}, nil
}

// Enqueue adds a value (pointer) to the tail of the queue.
func (q *Queue) Enqueue(value memory.Ptr) error {
	// 1. Create Node: [Next=Nil][Value=value]
	node, err := memory.AllocArray(2, 2)
	if err != nil {
		return err
	}
	// Index 0 (Next) is already NilPtr from AllocArray
	// Index 1 (Value) = value
	err = memory.WriteArrayElement(node, 1, value)
	if err != nil { return err }

	// 2. Loop until success
	for {
		// Read Tail (Index 1 of Queue Struct)
		tailAddr := q.Address.Add(ArrayDataOffset + PtrSize) // Offset 24
		tail, err := memory.LoadPtr(tailAddr)
		if err != nil { return err }

		// Read Tail.Next (Index 0 of Tail Node)
		tailNextAddr := tail.Add(ArrayDataOffset) // Offset 16
		next, err := memory.LoadPtr(tailNextAddr)
		if err != nil { return err }

		if next == memory.NilPtr {
			// Tail is last. Try to link node.
			// CAS(Tail.Next, Nil, Node)
			ok, err := memory.CompareAndSwapPtr(tailNextAddr, memory.NilPtr, node)
			if err != nil { return err }
			if ok {
				// Success. Try to advance Tail to Node.
				// CAS(Queue.Tail, OldTail, Node) - Failure is fine
				memory.CompareAndSwapPtr(tailAddr, tail, node)
				return nil
			}
		} else {
			// Tail not last. Help advance Tail.
			memory.CompareAndSwapPtr(tailAddr, tail, next)
		}
	}
}

// Dequeue removes and returns a value from the head of the queue.
// Returns NilPtr if empty.
func (q *Queue) Dequeue() (memory.Ptr, error) {
	for {
		// Read Head (Index 0 of Queue Struct)
		headAddr := q.Address.Add(ArrayDataOffset)
		head, err := memory.LoadPtr(headAddr)
		if err != nil { return 0, err }

		// Read Tail (Index 1 of Queue Struct)
		tailAddr := q.Address.Add(ArrayDataOffset + PtrSize)
		tail, err := memory.LoadPtr(tailAddr)
		if err != nil { return 0, err }

		// Read Head.Next (Index 0 of Head Node)
		headNextAddr := head.Add(ArrayDataOffset)
		next, err := memory.LoadPtr(headNextAddr)
		if err != nil { return 0, err }

		if head == tail {
			// Queue might be empty
			if next == memory.NilPtr {
				return memory.NilPtr, nil // Empty
			}
			// Tail is falling behind, help advance
			memory.CompareAndSwapPtr(tailAddr, tail, next)
		} else {
			// Read Value from Next Node (Index 1)
			val, err := memory.ReadArrayElement(next, 1)
			if err != nil { return 0, err }

			// Try to advance Head
			ok, err := memory.CompareAndSwapPtr(headAddr, head, next)
			if err != nil { return 0, err }
			if ok {
				// Success. Return value.
				// Note: Old Head is now freeable (if we had GC)
				return val, nil
			}
		}
	}
}
