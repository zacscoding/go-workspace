package main

import (
	"log"
	"math/rand"
	"time"
)

func main() {
	runTicker()
}

func runTicker() {
	doFunc := func() {
		log.Println(">Start doFunc..")
		sleep := rand.Intn(5)
		log.Println(">>Sleep ", sleep)
		time.Sleep(time.Duration(sleep) * time.Second)
		log.Println(">>> complete")
	}

	interval := time.Second * 3

	var (
		nextBatch = time.Now().Add(interval)
		ticker    = time.NewTicker(interval / 10)
		tickerCh  = ticker.C
	)
	defer ticker.Stop()

	for {
		select {
		case <-tickerCh:
			if time.Now().After(nextBatch) {
				log.Println("Ticker")
				nextBatch = time.Now().Add(interval)
				doFunc()
			}
		}
	}
}