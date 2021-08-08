package go_tasks

// Task represents an independent worker.
//
// Tasks should respect requests for cancellation which is indicated by cancelC becoming readable. Additionally, tasks
// must not close the cancelC channel. Errors returned by the task are made available to clients via the FixedTaskGroup's
// ErrC. In most cases, you'll close over existing functions and blocks of code with an anonymous function to create
// a Task.
//
// Example:
//   inputC := make(chan interface{})
//   task := func(cancelC <-chan struct{}) error {
//   	for {
//   		select {
//  		case <-cancelC:
//  			return nil
//  		case input := <-inputC:
//   			if err := process(input); err != nil {
//   				return err
//   			}
//  		}
//  	}
//   }
type Task func(cancelC <-chan struct{}) error
