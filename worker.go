package main

// Worker is a worker that takes requests and processes them
type Worker struct {
	requests chan Request
	pending  int
	index    int
}

// Work preforms the request
func (w *Worker) Work(done chan *Worker) {
	for {
		request := <-w.requests
		request.result <- request.job()
		done <- w
	}
}
