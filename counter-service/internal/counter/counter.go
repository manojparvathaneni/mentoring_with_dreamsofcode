package counter

import (
	"sync/atomic"
)

// Counter represents a thread-safe counter
type Counter struct {
	// Visits is the counter value
	Visits atomic.Int64
	
	// lastSaved is the last persisted value
	lastSaved atomic.Int64
	
	// dirty indicates if the counter has been modified since last save
	dirty atomic.Bool
}

// NewCounter creates a new counter with the given initial value
func NewCounter(initialValue int64) *Counter {
	counter := &Counter{}
	counter.Visits.Store(initialValue)
	counter.lastSaved.Store(initialValue)
	return counter
}

// Increment atomically increments the counter and returns the new value
func (c *Counter) Increment() int64 {
	// Increment counter
	newValue := c.Visits.Add(1)
	
	// Mark as dirty
	c.dirty.Store(true)
	
	return newValue
}

// GetValue returns the current counter value
func (c *Counter) GetValue() int64 {
	return c.Visits.Load()
}

// IsDirty returns true if the counter has been modified since last save
func (c *Counter) IsDirty() bool {
	return c.dirty.Load()
}

// MarkClean marks the counter as clean (not dirty)
func (c *Counter) MarkClean() {
	c.lastSaved.Store(c.Visits.Load())
	c.dirty.Store(false)
}