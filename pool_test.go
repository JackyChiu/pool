package main

import (
	"container/heap"
	"testing"
)

func TestNewPool(t *testing.T) {
	pool := NewPool(5, make(chan *Worker))
	if pool == nil {
		t.Error("Expected pool pointer")
	}
}

func TestPool_Pop(t *testing.T) {
	c := make(chan Request)
	expected := 2

	pool := &Pool{
		&Worker{c, 5, 0},
		&Worker{c, 7, 1},
		&Worker{c, 2, 2},
		&Worker{c, 9, 3},
	}
	heap.Init(pool)

	worker, ok := heap.Pop(pool).(*Worker)

	if !ok {
		t.Error("Expected *Worker")
	}

	if worker.pending != expected {
		t.Errorf("Expected: %v, got: %v", expected, worker.pending)
	}
}
