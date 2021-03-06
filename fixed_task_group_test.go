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

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"sync/atomic"
	"testing"
	"time"
)

func TestFixedTaskGroup_TasksCalled(t *testing.T) {
	tasksCalled := int64(0)
	task := func(cancelC <-chan struct{}) error {
		atomic.AddInt64(&tasksCalled, 1)
		return nil
	}

	tg := NewFixedTaskGroup(task, task)
	require.NotNil(t, tg)

	require.NoError(t, <-tg.ErrC()) // Should not block, should produce nil.
	require.Equal(t, int64(2), atomic.LoadInt64(&tasksCalled))
}

func TestFixedTaskGroup_ErrorPropagation(t *testing.T) {
	tg := NewFixedTaskGroup(
		func(cancelC <-chan struct{}) error {
			return errors.New("task A")
		},
		func(cancelC <-chan struct{}) error {
			return errors.New("task B")
		},
	)

	require.NotNil(t, tg)

	for i := 0; i < 2; i++ {
		err := <-tg.ErrC()
		require.Error(t, err)

		if err.Error() != "task A" && err.Error() != "task B" {
			t.Fatalf("Unexpected error: %s\n", err)
		}
	}

	require.NoError(t, <-tg.ErrC())
}

func TestFixedTaskGroup_Cancel(t *testing.T) {
	cancelled := int64(0)
	tg := NewFixedTaskGroup(
		func(cancelC <-chan struct{}) error {
			select {
			case <-time.After(time.Minute):
				return errors.New("timed out")
			case <-cancelC:
				atomic.AddInt64(&cancelled, 1)
			}

			return nil
		},
	)

	tg.Cancel()
	require.NotPanics(t, func() {
		tg.Cancel() // This should be fine
	})

	require.NoError(t, <-tg.ErrC())
	require.Equal(t, int64(1), atomic.LoadInt64(&cancelled))
}

func ExampleFixedTaskGroup() {
	taskFact := func(taskId int, iters int) Task {
		return func(cancelC <-chan struct{}) error {
			done := false
			for i := 0; i < iters && !done; i++ {
				select {
				case <-cancelC:
					fmt.Printf("task: %d - cancel received\n", taskId)
					done = true
				default:
					fmt.Printf("task: %d - iter: %d\n", taskId, i)
					time.Sleep(time.Millisecond * 500)
				}
			}

			fmt.Printf("task: %d - complete\n", taskId)
			return nil
		}
	}

	tg := NewFixedTaskGroup(
		taskFact(1, 10),
		taskFact(2, 10),
		taskFact(3, 10),
		func(cancelC <-chan struct{}) error {
			timeC := time.After(time.Second * 3)
			<-timeC
			fmt.Printf("Producing an error\n")
			return errors.New("boom")
		},
	)

	for err := range tg.ErrC() {
		fmt.Printf("Error: %s\n", err)
		// Cancel the task group if any task produces an error.
		tg.Cancel()
	}
}
