package util

import (
	"time"
)

// DoLater runs a callback function once, after the specified delay
func DoLater(delay time.Duration, f func()) {
	go func() {
		time.Sleep(delay)
		f()
	}()
}

// DoRepeatedly runs a callback function after each interval of delay
// it continues until quit is closed
func DoRepeatedly(interval time.Duration, f func()) (quit chan struct{}) {
	quit = make(chan struct{})
	go func() {
		ticker := time.NewTicker(interval)
		for range ticker.C {
			select {
			case <-quit:
				return
			default:
				f()
			}
		}
	}()
	return
}
