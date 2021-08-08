package go_tasks

import "sync"

// FixedTaskGroup is a fixed number of tasks that share a lifecycle.
//
// Use a task group when you have a fixed number of tasks that you want to start and cancel together. The group is
// considered done when all tasks have completed. Additionally, the group may be cancelled as a unit. The number of
// tasks is fixed at creation time making termination simple (i.e. "done" is when all tasks reach completion).
//
// If DoneC is readable, you can be sure that there are no resource leaks.
type FixedTaskGroup struct {
	tasks []Task
	wg    *sync.WaitGroup
	doneC chan struct{}
	errC  chan error

	cancelC    chan struct{}
	cancelOnce *sync.Once
}

// NewFixedTaskGroup creates and starts a group of tasks that share a lifecycle.
func NewFixedTaskGroup(tasks ...Task) *FixedTaskGroup {
	tg := &FixedTaskGroup{
		tasks: tasks,
		wg:    &sync.WaitGroup{},
		doneC: make(chan struct{}),
		// NB: We allocate just enough space for each task to produce an error. If more tasks than capacity exist _and_
		// the caller doesn't consume the error channel completely it's possible to leak go routines, and we wouldn't
		// make good on our promise to not leak resources when DoneC() is readable.
		errC: make(chan error, len(tasks)),

		cancelC:    make(chan struct{}),
		cancelOnce: &sync.Once{},
	}

	for _, task := range tasks {
		tg.startTask(task)
	}

	// Close doneC and errC when all tasks are complete.
	go func() {
		defer close(tg.errC)
		defer close(tg.doneC)
		tg.wg.Wait()
	}()

	return tg
}

// Adds and starts a single task to the group.
//
// Note that calling startTask() after someone has starting listening
func (g *FixedTaskGroup) startTask(task Task) {
	g.tasks = append(g.tasks, task)
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()

		if err := task(g.cancelC); err != nil {
			g.errC <- err
		}
	}()
}

// Cancel signals all tasks to stop.
//
// This method is asynchronous and may return before all tasks have exited. Note that cancellation is cooperative; it is
// up to the task implementation to respect the request for cancellation. If you want to wait until all tasks have
// completed, you may read from DoneC() after calling this method.
//
// It is safe to call this method multiple times and from multiple threads, although it has no effect after the first
// call.
func (g *FixedTaskGroup) Cancel() {
	g.cancelOnce.Do(func() {
		close(g.cancelC)
	})
}

// ErrC returns a channel that will contain any task errors.
//
// Keep in mind that reading from this channel will succeed when it's closed. Use the two variable read form to detect
// when the channel is closed.
//
// Example:
//   err, ok := <-tg.ErrC()
//   if err != nil && ok {
//   	// ...
//   }
//
// This method always returns the same channel.
func (g *FixedTaskGroup) ErrC() <-chan error {
	return g.errC
}

// DoneC returns a channel that is closed when all tasks are complete.
//
// Task completion may be the result of cancellation or natural termination. This method always returns the same
// channel.
func (g *FixedTaskGroup) DoneC() <-chan struct{} {
	return g.doneC
}
