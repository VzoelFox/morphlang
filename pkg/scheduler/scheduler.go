package scheduler

import (
	"fmt"
	"github.com/VzoelFox/morphlang/pkg/memory"
)

var Global *Scheduler

type Scheduler struct {
	Queue   *Queue
	Workers []*Worker
}

// Init initializes the global scheduler with a specific number of workers.
// The executor function is the callback that runs the task.
func Init(numWorkers int, executor func(memory.Ptr)) error {
	q, err := NewQueue()
	if err != nil {
		return err
	}

	s := &Scheduler{
		Queue:   q,
		Workers: make([]*Worker, numWorkers),
	}

	for i := 0; i < numWorkers; i++ {
		w := NewWorker(i, q, executor)
		s.Workers[i] = w
		w.Start()
	}

	Global = s
	return nil
}

// Submit adds a task to the global queue.
func Submit(task memory.Ptr) error {
	if Global == nil {
		return fmt.Errorf("scheduler not initialized")
	}
	return Global.Queue.Enqueue(task)
}
