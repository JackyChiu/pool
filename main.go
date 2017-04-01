package main

import (
	"container/heap"
	"fmt"
	"runtime"
)

func main() {
	var (
		pool     Pool
		balancer Balancer
		requests chan Request = make(chan Request)
		done     chan *Worker = make(chan *Worker)
	)

	fmt.Println("Number of CPUs avaiable: ", runtime.NumCPU())

	for i := 0; i < runtime.NumCPU(); i++ {
		requests := make(chan Request)
		worker := Worker{requests, 0, i}
		go worker.Work(done)
		pool = append(pool, &worker)
	}
	heap.Init(&pool)

	balancer = Balancer{pool, done}
	go balancer.Balance(requests)

	requester(requests)
}
