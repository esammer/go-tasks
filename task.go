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

import "errors"

// ErrTaskCancelled indicates a task was cancelled.
//
// Tasks that wish to explicitly indicate they were cancelled rather than completing normally may return this error
// value as a sentinel value.
var ErrTaskCancelled = errors.New("task cancelled")

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
