package main

import "fmt"

type Pool []*Worker

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
	var workers string
	sum := 0
	for _, worker := range p {
		sum += worker.pending
		workers += string(worker.pending) + " "
	}
	avg := sum / len(p)
	return fmt.Sprintf("Workers: %v, Avg Load: %v", workers, avg)
}
