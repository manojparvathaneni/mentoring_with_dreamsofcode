# HTTP Test Example

## What it does

This example uses `net/http/httptest` to test an HTTP handler without running a real server.

## Why it matters

Testing HTTP handlers without starting a full server is fast and reliable. It ensures your logic works as expected in isolation.

## How it’s used in the project

In `go-counter`, this technique is used to unit test the `/count` and `/inc` endpoints, verifying behavior without hitting the network.
