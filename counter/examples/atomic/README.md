# Atomic Counter Example

## What it does

This example uses the `sync/atomic` package to safely increment a counter from multiple goroutines. It demonstrates how to use `atomic.AddInt64` to update a shared `int64` variable without using a mutex.

## Why it matters

When multiple goroutines access the same variable, there's a risk of a **data race** — where reads and writes overlap unpredictably. The `sync/atomic` package provides low-level atomic memory primitives that avoid this issue.

Using atomic operations is faster and lighter than using a `sync.Mutex` when you only need simple updates, like incrementing a number.

## How it’s used in the project

In the `go-counter` project, atomic counters are used to track hits or counts in a high-concurrency environment (e.g., HTTP handlers receiving frequent requests). This ensures that increments are safe and accurate even when many goroutines are running concurrently.
