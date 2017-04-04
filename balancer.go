package main

import (
	"container/heap"
	"fmt"
)

type Balancer struct {
	pool *Pool
	done chan *Worker
}

func (b *Balancer) Balance(requests <-chan Request) {
	for {
		select {
		case request := <-requests:
			b.dispatch(request)
			fmt.Println(b.pool)
		case worker := <-b.done:
			b.complete(worker)
		}
	}
}

func (b *Balancer) dispatch(request Request) {
	w := heap.Pop(b.pool).(*Worker)
	w.requests <- request
	w.pending += 1
	heap.Push(b.pool, w)
}

func (b Balancer) complete(worker *Worker) {
	worker.pending -= 1
	heap.Fix(b.pool, worker.index)
}
