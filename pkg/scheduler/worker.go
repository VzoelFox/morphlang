package scheduler

import (
	"fmt"
	"time"

	"github.com/VzoelFox/morphlang/pkg/memory"
)

// Worker represents a processing unit that consumes tasks from the queue.
type Worker struct {
	ID    int
	Queue *Queue
	Quit  chan bool
}

// NewWorker creates a new worker instance.
func NewWorker(id int, q *Queue) *Worker {
	return &Worker{
		ID:    id,
		Queue: q,
		Quit:  make(chan bool),
	}
}

// Start launches the worker loop in a goroutine.
func (w *Worker) Start() {
	go func() {
		for {
			select {
			case <-w.Quit:
				return
			default:
				// Attempt to dequeue a task
				ptr, err := w.Queue.Dequeue()
				if err != nil {
					fmt.Printf("Worker %d error: %v\n", w.ID, err)
					return
				}

				if ptr == memory.NilPtr {
					// Queue empty, wait briefly (Spin/Sleep strategy)
					time.Sleep(100 * time.Microsecond)
					continue
				}

				// Execute the task
				w.execute(ptr)
			}
		}
	}()
}

// execute processes the task pointed to by taskPtr.
// In the full system, this would load a Routine context and resume the VM.
func (w *Worker) execute(taskPtr memory.Ptr) {
	// Placeholder for Phase X execution logic
	// e.g., Resume(taskPtr)
}

// Stop signals the worker to exit.
func (w *Worker) Stop() {
	w.Quit <- true
}
