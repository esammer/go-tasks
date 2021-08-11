// Copyright 2021 Eric Sammer
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package go_tasks

import "sync"

// FixedTaskGroup is a fixed number of tasks that share a lifecycle.
//
// Use a task group when you have a fixed number of tasks that you want to start and cancel together. The group is
// considered done when all tasks have completed. Additionally, the group may be cancelled as a unit. The number of
// tasks is fixed at creation time making termination simple (i.e. "done" is when all tasks reach completion).
//
// If ErrC is closed, you can be sure that there are no resource leaks.
type FixedTaskGroup struct {
	tasks []Task
	wg    *sync.WaitGroup
	errC  chan error

	cancelC    chan struct{}
	cancelOnce *sync.Once
}

// NewFixedTaskGroup creates and starts a group of tasks that share a lifecycle.
func NewFixedTaskGroup(tasks ...Task) *FixedTaskGroup {
	tg := &FixedTaskGroup{
		tasks: tasks,
		wg:    &sync.WaitGroup{},
		// NB: We allocate just enough space for each task to produce an error. If more tasks than capacity exist _and_
		// the caller doesn't consume the error channel completely it's possible to leak go routines, and we wouldn't
		// make good on our promise to not leak resources when ErrC() is closed.
		errC: make(chan error, len(tasks)),

		cancelC:    make(chan struct{}),
		cancelOnce: &sync.Once{},
	}

	for _, task := range tasks {
		tg.startTask(task)
	}

	// Close errC when all tasks are complete.
	go func() {
		tg.wg.Wait()
		defer close(tg.errC)
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
// completed, you may read from ErrC() after calling this method.
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
// This channel is closed when the task group is complete. Callers should consume from this channel until it is closed,
// typically using a for / range loop.
//
// Example:
//   tg := NewFixedTaskGroup(tasks)
//   for err := range tg.ErrC() {
//   	// Handle err. Optionally call tg.Cancel() to cause all tasks to return.
//   }
//
//   // Once we're here we know all tasks are complete, all errors have been seen,
//   // and all resources have been released.
//
// This method always returns the same channel.
func (g *FixedTaskGroup) ErrC() <-chan error {
	return g.errC
}
