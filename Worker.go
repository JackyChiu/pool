package main

type Worker struct {
	requests chan Request
	pending  int
	index    int
}
