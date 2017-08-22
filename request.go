package main

import (
	"math/rand"
	"time"
)

type Request struct {
	job    func() int
	result chan int
}

func job() int {
	randDuration := time.Duration(rand.Intn(4)) * time.Second
	time.Sleep(randDuration + time.Second)
	return 1
}

func requester(requests chan<- Request) {
	result := make(chan int)
	for {
		randDuration := time.Duration(rand.Intn(1)) * time.Second
		time.Sleep(randDuration + 250*time.Millisecond)
		select {
		case requests <- Request{job, result}:
			// Sent request
		case <-result:
			// Request came back
		}
	}
}
