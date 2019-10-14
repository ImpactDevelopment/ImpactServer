package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDoLater(t *testing.T) {
	runs := 0
	start := time.Now()
	delay := 2 * time.Millisecond

	// Test in the goroutine for epic memes, sleep on the main thread to give it time to iterate a bit
	DoLater(delay, callback(t, start, delay, &runs))

	time.Sleep(100 * delay)
	assert.Equal(t, 1, runs, "Should only run once")

	// Reset the test to iterate over DoLater a few times
	runs = 0
	var i int64
	for i = 1; i < 100; i++ {
		// Give each iteration its own time.Now() just in case this loop crosses into a new ms
		// Use an offset delay so they don't all run at once
		offsetDelay := time.Duration(i*ms(delay)) * time.Millisecond
		DoLater(offsetDelay, callback(t, time.Now(), delay, &runs))
	}
	time.Sleep(time.Duration(i*ms(delay)) * time.Millisecond)
	assert.Equal(t, i-1, int64(runs), "Expected the number of runs to match the number of loop iterations")
}

func TestDoRepeatedly(t *testing.T) {
	runs := 0
	start := time.Now()
	interval := 2 * time.Millisecond

	// Test in the goroutine for epic memes, sleep on the main thread to give it time to iterate a bit
	quit := DoRepeatedly(interval, callback(t, start, interval, &runs))

	// Let the test run for approx 100 iterations
	// Slow, but ensures the timer doesn't go out of sync after a few iterations
	time.Sleep(100 * interval)

	// Test closing channel
	close(quit)
	timesRun := runs
	time.Sleep(50 * interval)
	assert.Equal(t, timesRun, runs, "Should not run again after closing channel")

	// Reset the test to do a couple other bits
	runs = 0
	timesRun = 0
	quit = DoRepeatedly(interval/2, func() {
		runs++
	})

	assert.Equal(t, 0, runs, "Should not run immediately")
	time.Sleep(2 * interval)
	assert.Greater(t, runs, 1, "Should have run more than once")

	// Quit by sending to channel
	quit <- struct{}{}
	timesRun = runs

	time.Sleep(50 * interval)
	assert.Equal(t, timesRun, runs, "Should not run again sending to channel")
}

// callback generates a callback function that tests it is run
// at the expected interval(s) after the start time and  increments
// `runs` each time the callback is invoked
func callback(t *testing.T, start time.Time, interval time.Duration, runs *int) func() {
	return func() {
		*runs++
		delay := ms(time.Now().Sub(start))
		expect := int64(*runs) * ms(interval)
		assert.Equal(t, expect, delay, "Expected %d delay after %d runs", expect, *runs)
	}
}

// ms returns the duration as an integer millisecond count.
// taken from go 1.13, https://go-review.googlesource.com/c/go/+/167387/2/src/time/time.go
func ms(d time.Duration) int64 {
	return int64(d) / 1e6
}
