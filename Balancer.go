package main

import "container/heap"

type Balancer struct {
	pool Pool
	done chan *Worker
}

func (b *Balancer) Balance(work <-chan Request) {
	for {
		select {
		case req := <-work:
			b.dispatch(req)
		case worker := <-b.done:
			b.complete(worker)
		}
	}
}

func (b *Balancer) dispatch(req Request) {
	w := heap.Pop(&b.pool).(*Worker)
	w.requests <- req
	w.pending += 1
	heap.Push(&b.pool, w)
}

func (b Balancer) complete(worker *Worker) {
	worker.pending -= 1
	heap.Fix(&b.pool, worker.index)
}
