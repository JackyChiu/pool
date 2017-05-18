package main

import (
	"container/heap"
	"fmt"
	"math"
	"strconv"
)

// Pool is used as the worker pool
type Pool []*Worker

// NewPool creates a new pool
func NewPool(workers int, done chan *Worker) *Pool {
	var pool Pool
	for i := 0; i < workers; i++ {
		requests := make(chan Request, 25)
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
	p[i].index = i
	p[j].index = j
}

func (p *Pool) Push(x interface{}) {
	worker := x.(*Worker)
	worker.index = len(*p)
	*p = append(*p, worker)
}

func (p *Pool) Pop() interface{} {
	prev := *p
	last := len(prev) - 1
	elem := prev[last]
	*p = prev[:last]
	return elem
}

// stats returns the mean and stdDev values of the pool
func (p Pool) stats() (mean float64, stdDev float64) {
	length := float64(len(p))

	for _, worker := range p {
		mean += float64(worker.pending)
	}
	mean /= length

	for _, worker := range p {
		stdDev += math.Pow((float64(worker.pending) - mean), 2)
	}
	stdDev = math.Sqrt((1 / length) * stdDev)

	return mean, stdDev
}

func (p Pool) String() string {
	var workers string
	for _, worker := range p {
		workers += strconv.Itoa(worker.pending) + " "
	}

	mean, stdDev := p.stats()
	return fmt.Sprintf("Workers: %v| Avg Load: %.2f | Std Dev: %.2f", workers, mean, stdDev)
}
