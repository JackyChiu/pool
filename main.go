package main

import (
	"fmt"
	"runtime"
)

func main() {
	fmt.Println("Number of CPUs avaiable: ", runtime.NumCPU())

	requests := make(chan Request)
	done := make(chan *Worker)

	pool := NewPool(runtime.NumCPU(), done)
	balancer := &Balancer{*pool, done}

	go balancer.Balance(requests)

	requester(requests)
}
