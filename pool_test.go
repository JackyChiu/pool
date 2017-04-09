package main

import (
	"container/heap"
	"math"
	"reflect"
	"testing"
)

var (
	ch    chan Request = make(chan Request)
	pools []Pool       = []Pool{
		Pool{
			&Worker{ch, 0, 0},
			&Worker{ch, 1, 1},
			&Worker{ch, 2, 2},
			&Worker{ch, 3, 3},
		},
		Pool{
			&Worker{ch, 3, 0},
			&Worker{ch, 2, 1},
			&Worker{ch, 1, 2},
			&Worker{ch, 0, 3},
		},
		Pool{
			&Worker{ch, 0, 0},
		},
		Pool{
			&Worker{ch, 5, 0},
			&Worker{ch, 7, 1},
			&Worker{ch, 2, 2},
			&Worker{ch, 9, 3},
		},
		Pool{
			&Worker{ch, 5, 0},
			&Worker{ch, 7, 1},
			&Worker{ch, 11, 2},
			&Worker{ch, 2, 3},
			&Worker{ch, 9, 4},
			&Worker{ch, 2, 5},
		},
	}
)

func TestNewPool(t *testing.T) {
	ch := make(chan *Worker)
	pool := NewPool(5, ch)
	if pool == nil {
		t.Errorf("got %+v, expected: %+v", pool, "not nil")
	}
}

var poolPopTests = []struct {
	pool     Pool
	expected *Worker
}{
	{pools[0], &Worker{ch, 3, 3}},
	{pools[1], &Worker{ch, 0, 3}},
	{pools[2], &Worker{ch, 0, 0}},
	{pools[3], &Worker{ch, 9, 3}},
	{pools[4], &Worker{ch, 2, 5}},
}

func TestPool_Pop(t *testing.T) {
	for _, test := range poolPopTests {
		worker := test.pool.Pop().(*Worker)
		if !reflect.DeepEqual(worker, test.expected) {
			t.Errorf("got %+v, expected: %+v", worker, test.expected)
		}
	}
}

var poolStatTests = []struct {
	pool   Pool
	mean   float64
	stdDev float64
}{
	{pools[0], 1.5, 1.12},
	{pools[1], 1.5, 1.12},
	{pools[2], 0, 0},
	{pools[3], 5.75, 2.59},
	{pools[4], 6.0, 3.37},
}

func TestPool_stats(t *testing.T) {
	decimalPoint := 0.0001
	for _, test := range poolStatTests {
		mean, stdDev := test.pool.stats()
		meanDiff := float64(mean - test.mean)
		stdDevDiff := float64(stdDev - test.stdDev)

		if meanDiff != 0 && math.Abs(meanDiff) < decimalPoint {
			t.Errorf("got %+v, expected: %+v", mean, test.mean)
		}
		if stdDevDiff != 0 && math.Abs(stdDevDiff) < decimalPoint {
			t.Errorf("got %+v, expected: %+v", stdDev, test.stdDev)
		}
	}
}

var heapPopPoolTests = []struct {
	pool     Pool
	expected *Worker
}{
	{pools[0], &Worker{ch, 0, len(pools[0]) - 1}},
	{pools[1], &Worker{ch, 0, len(pools[1]) - 1}},
	{pools[2], &Worker{ch, 0, len(pools[2]) - 1}},
	{pools[3], &Worker{ch, 2, len(pools[3]) - 1}},
	{pools[4], &Worker{ch, 2, len(pools[4]) - 1}},
}

func TestHeap_Pop_Pool(t *testing.T) {
	for _, test := range heapPopPoolTests {
		heap.Init(&test.pool)
		worker := heap.Pop(&test.pool).(*Worker)
		if !reflect.DeepEqual(worker, test.expected) {
			t.Errorf("got %+v, expected: %+v", worker, test.expected)
		}
	}
}
