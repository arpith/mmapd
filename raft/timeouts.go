package raft

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

type timeout struct {
	period time.Duration
	ticker time.Ticker
}

func (t *timeout) reset() {
	t.ticker.Stop()
	t.ticker = *time.NewTicker(t.period)
}

func createTimeout(period time.Duration) *timeout {
	return &timeout{period, *time.NewTicker(period)}
}

func generateRandomInt(lower int, upper int) int {
	l := int64(lower)
	u := int64(upper)
	max := big.NewInt(u - l)
	r, err := rand.Int(rand.Reader, max)
	if err != nil {
		fmt.Println("Couldn't generate random int!")
	}
	return int(l + r.Int64())
}

func createRandomTimeout(lower int, upper int, period time.Duration) *timeout {
	randomInt := generateRandomInt(lower, upper)
	period = time.Duration(randomInt) * period
	return createTimeout(period)
}
