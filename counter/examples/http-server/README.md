# HTTP Server Example

## What it does

This example creates a simple HTTP server with one endpoint `/count` that responds with a fixed counter value.

## Why it matters

Serving over HTTP is a common way to expose your application's functionality. Go makes it easy with the `net/http` package.

## How it’s used in the project

In `go-counter`, the HTTP server exposes a `/count` endpoint so other systems or users can fetch the current counter value via an API.
