package pkg

import (
	"log"
	"sync"
	"time"
)

// ProgressTracker tracks and displays progress of data generation
type ProgressTracker struct {
	total      int
	current    int
	startTime  time.Time
	lastUpdate time.Time
	mutex      sync.Mutex
	logger     *log.Logger
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(logger *log.Logger) *ProgressTracker {
	return &ProgressTracker{
		logger: logger,
	}
}

// Start initializes the progress tracker
func (pt *ProgressTracker) Start(total int) {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()

	pt.total = total
	pt.current = 0
	pt.startTime = time.Now()
	pt.lastUpdate = pt.startTime

	pt.logger.Printf("Starting progress tracking for %d items", total)
}

// Update increments the progress counter
func (pt *ProgressTracker) Update(increment int) {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()

	pt.current += increment
	now := time.Now()

	// Log progress every 10% or every 30 seconds
	percentage := float64(pt.current) / float64(pt.total) * 100
	timeSinceLastUpdate := now.Sub(pt.lastUpdate)

	if timeSinceLastUpdate > 30*time.Second || pt.current%max(1, pt.total/10) == 0 {
		elapsed := now.Sub(pt.startTime)
		rate := float64(pt.current) / elapsed.Seconds()
		remaining := time.Duration(float64(pt.total-pt.current) / rate * float64(time.Second))

		pt.logger.Printf("Progress: %d/%d (%.1f%%) - Rate: %.1f items/sec - ETA: %v",
			pt.current, pt.total, percentage, rate, remaining.Round(time.Second))

		pt.lastUpdate = now
	}
}

// Finish completes the progress tracking
func (pt *ProgressTracker) Finish() {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()

	elapsed := time.Since(pt.startTime)
	rate := float64(pt.current) / elapsed.Seconds()

	pt.logger.Printf("Completed: %d items in %v (%.1f items/sec)",
		pt.current, elapsed.Round(time.Second), rate)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
