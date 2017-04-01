package main

import (
	"container/heap"
	"fmt"
	"strconv"
)

type Pool []*Worker

func NewPool(workers int, done chan *Worker) *Pool {
	var pool Pool
	for i := 0; i < workers; i++ {
		requests := make(chan Request)
		worker := Worker{requests, 0, i}
		go worker.Work(done)
		pool = append(pool, &worker)
	}
	heap.Init(&pool)
	return &pool
}

func (p Pool) Len() int {
	return len(p)
}

func (p Pool) Less(i, j int) bool {
	return p[i].pending < p[j].pending
}

func (p Pool) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p *Pool) Push(x interface{}) {
	*p = append(*p, x.(*Worker))
}

func (p *Pool) Pop() interface{} {
	prev := *p
	last := len(prev) - 1
	elem := prev[last]
	*p = prev[:last]
	return elem
}

func (p Pool) String() string {
	// TODO: Add str devation
	var (
		workers string
		sum     int
		avg     float32
	)
	for _, worker := range p {
		sum += worker.pending
		workers += strconv.Itoa(worker.pending) + " "
	}
	avg = float32(sum / len(p))
	return fmt.Sprintf("Workers: %v, Avg Load: %v", workers, avg)
}
