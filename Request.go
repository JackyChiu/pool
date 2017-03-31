package main

import "time"

type Request struct {
	job    func() int
	result chan int
}

func requester(work chan<- Request) {
	result := make(chan int)
	for {
		// TODO: Make random sleep
		time.Sleep(2 * time.Second)
		work <- Request{job, result}
		<-result
	}
}

func job() int {
	time.Sleep(1 * time.Second)
	return 1
}
