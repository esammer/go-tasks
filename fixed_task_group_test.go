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

	<-tg.DoneC()
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
	<-tg.DoneC()
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

	<-tg.DoneC()
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

	for done := false; !done; {
		select {
		case err, ok := <-tg.ErrC():
			if ok {
				fmt.Printf("Error: %s\n", err)
				// Cancel the task group if any task produces an error.
				tg.Cancel()
			}
		case <-tg.DoneC():
			fmt.Printf("All tasks complete\n")
			done = true
		}
	}
}
