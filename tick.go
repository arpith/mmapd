package main

import "fmt"
import "time"

type ticker struct {
	period time.Duration
	ticker <-chan time.Time
}

type server struct {
	doneChan chan bool
	tickerA  ticker
	tickerB  ticker
}

func createTicker(period time.Duration) *ticker {
	return &ticker{period, time.Tick(period)}
}

func (t *ticker) resetTicker() {
	t.ticker = time.Tick(t.period)
}

func (s *server) listener() {
	start := time.Now()
	for {
		select {
		case <-s.tickerA.ticker:
			elapsed := time.Since(start)
			fmt.Println("Elapsed: ", elapsed, " Ticker A")
		case <-s.tickerB.ticker:
			s.tickerA.resetTicker()
			elapsed := time.Since(start)
			fmt.Println("Elapsed: ", elapsed, " Ticker B - Going to reset ticker A")
		}
	}
	s.doneChan <- true
}

func main() {
	doneChan := make(chan bool)
	tickerA := createTicker(2 * time.Second)
	tickerB := createTicker(5 * time.Second)
	s := &server{doneChan, *tickerA, *tickerB}
	go s.listener()
	<-doneChan
}
