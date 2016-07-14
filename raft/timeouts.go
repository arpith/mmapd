package main

import "time"

type timeout struct {
	period time.Duration
	ticker <-chan time.Time
}

func (t *timeout) resetTimeout() {
	t.ticker = time.Tick(t.period)
}

func createTimeout(period time.Duration) *timeout {
	return &timeout{period, time.Tick(period)}
}
