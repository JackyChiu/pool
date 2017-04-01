package main

import (
	"math/rand"
	"time"
)

type Request struct {
	job    func() int
	result chan int
}

func requester(requests chan<- Request) {
	result := make(chan int)
	for {
		randDuration := time.Duration(rand.Intn(1))
		time.Sleep(randDuration * time.Second)
		requests <- Request{job, result}
		<-result
	}
}

func job() int {
	randDuration := time.Duration(rand.Intn(3))
	time.Sleep(randDuration * 5 * time.Second)
	return 1
}
