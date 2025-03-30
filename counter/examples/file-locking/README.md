# File Locking Example

## What it does

This example opens a file and uses `syscall.Flock` to acquire an exclusive lock. It prevents other processes from acquiring the same lock at the same time.

## Why it matters

When multiple processes might read or write the same file, you need a lock to avoid data corruption. File locks coordinate access and prevent simultaneous writes.

## How it’s used in the project

In `go-counter`, file locking ensures that only one instance of the app writes to the counter file at a time. This prevents concurrent writes that could overwrite each other.
