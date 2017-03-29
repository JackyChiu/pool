package main

type Balancer struct {
	pool Pool
	done chan *Worker
}
