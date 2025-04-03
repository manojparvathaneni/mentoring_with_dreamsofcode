# JSON Marshalling Example

## What it does

This example demonstrates how to convert a Go struct into JSON (`Marshal`) and back (`Unmarshal`). It uses Go’s built-in `encoding/json` package.

## Why it matters

JSON is a common format for saving structured data (like counters) to files or APIs. Go makes it easy to work with JSON using struct tags.

## How it’s used in the project

In `go-counter`, counters are periodically written to and loaded from a JSON file to persist state between restarts. This ensures that the counter survives crashes or shutdowns.
