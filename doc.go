// Package go_tasks is a small, zero dependency, library that provides a production-ready implementation of common
// concurrent task management patterns. It's suitable for implementing fine-grained worker pools. It's meant to make
// task implementations easily testable without adding a ton of framework or concurrency infrastructure.
//
// The primary abstraction is the Task, which is simply a function that takes a cancellation channel.
//
// FixedTaskGroup implements a worker group of a known-size that shares a lifecycle. It's extremely useful for cases
// where you have a group of tasks that operate together.
package go_tasks
