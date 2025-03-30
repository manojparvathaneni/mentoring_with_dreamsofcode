# Graceful Shutdown Example

## What it does

This example listens for system interrupt signals (like Ctrl+C) and uses a `context.Context` to handle cleanup before exiting.

## Why it matters

Programs need a chance to clean up (e.g., save files, close connections) before exiting. A graceful shutdown gives you that opportunity instead of quitting abruptly.

## How it’s used in the project

In `go-counter`, the server shuts down cleanly when it receives an interrupt. It saves the current counter value and releases file locks safely.
