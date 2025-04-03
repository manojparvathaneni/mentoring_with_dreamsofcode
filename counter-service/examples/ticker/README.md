# Ticker Example

## What it does

This example uses `time.Ticker` to run a loop every second. It prints a message each time the ticker ticks.

## Why it matters

Tickers are useful for scheduling recurring background tasks, like saving to disk or refreshing a cache.

## How it’s used in the project

In `go-counter`, a ticker periodically writes the current counter to a file, ensuring data is saved regularly without blocking the main logic.
