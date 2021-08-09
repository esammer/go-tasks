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

// Package go_tasks is a small, zero dependency, library that provides a production-ready implementation of common
// concurrent task management patterns. It's suitable for implementing fine-grained worker pools. It's meant to make
// task implementations easily testable without adding a ton of framework or concurrency infrastructure.
//
// The primary abstraction is the Task, which is simply a function that takes a cancellation channel.
//
// FixedTaskGroup implements a worker group of a known-size that shares a lifecycle. It's extremely useful for cases
// where you have a group of tasks that operate together.
package go_tasks
