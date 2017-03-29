package main

type Worker struct {
	requests chan Request
	pending  int
	index    int
}

func (w *Worker) work(done chan *Worker) {
	for {
		request := <-w.requests
		request.result <- request.job()
		done <- w
	}
}
