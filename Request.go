package main

type Request struct {
	job    func() int
	result chan int
}
