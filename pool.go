package pool

import (
	"context"
	"sync"
)

type taskFunc func() error

// Pool should be used once. Once Wait() gets called it's goroutines are reaped.
// Consider writting a Reset to use the pool again.
// Should the pool be able to be used concurrently?
type Pool struct {
	ctx context.Context

	tasks     chan taskFunc
	closeOnce sync.Once

	errPool errorPool

	// taskWg is used for task synconization in the Pool
	taskWg sync.WaitGroup

	// cap is the capacity of the pool
	cap int
	// the current size of the pool
	size int

	lock sync.RWMutex
}

func New(ctx context.Context, poolSize int) (*Pool, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &Pool{
		errPool: errorPool{
			cancel: cancel,
		},
		ctx:   ctx,
		tasks: make(chan taskFunc),
		cap:   poolSize,
	}, ctx
}

func (p *Pool) Go(task taskFunc) {
	// how do you know if the goroutines are backed up or if you should create another goroutine?
	// by trying to send a task and then selecting
	// Ahh would've worked well if the task chan wasn't buffered
	// Requiring cap + 1 tasks to start spinning up more goroutines is unexpected behaviour

	p.taskWg.Add(1)

	select {
	case p.tasks <- task:
		//log.Println("send successful")
		return
	case <-p.ctx.Done():
		// cancel out
		p.taskWg.Done()
		return
	default:
	}

	if p.Size() < int(p.cap) {
		p.startWorker()
	}
	//log.Println("starting wait")

	select {
	case p.tasks <- task:
	case <-p.ctx.Done():
		// cancel out
		p.taskWg.Done()
		// don't block when context is cancelled
	}
}

func (p *Pool) startWorker() {
	p.errPool.Go(func() error {
		for {
			select {
			case task, ok := <-p.tasks:
				if !ok {
					// task channel is closed, kill routine
					return nil
				}
				p.errPool.execute(task)
				// mark that task as done
				p.taskWg.Done()
			case <-p.ctx.Done():
				p.taskWg.Done()
				return p.ctx.Err()
			}
		}
	})
	p.incrementSize()
}

func (p *Pool) Wait() error {
	p.taskWg.Wait()
	//log.Println("task group done")

	p.closeOnce.Do(func() {
		close(p.tasks)
	})
	//close(p.tasksBuffered)

	return p.errPool.wait()
}

func (p *Pool) Size() int {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.size
}

func (p *Pool) incrementSize() {
	p.lock.Lock()
	p.size++
	p.lock.Unlock()
}

type errorPool struct {
	cancel context.CancelFunc
	wg     sync.WaitGroup

	errOnce sync.Once
	err     error
}

func (e *errorPool) wait() error {
	e.wg.Wait()
	//log.Println("err group done")
	if e.cancel != nil {
		e.cancel()
	}
	return e.err
}

func (e *errorPool) Go(task taskFunc) {
	e.wg.Add(1)

	go func() {
		defer e.wg.Done()
		//log.Println("spun up")
		e.execute(task)
		//log.Println("spun down")
	}()
}

// execute runs the task and records the first error that occurs.
// This in turn cancels any other tasks.
func (e *errorPool) execute(task taskFunc) {
	err := task()
	if err == nil {
		return
	}
	e.errOnce.Do(func() {
		e.err = err
		if e.cancel != nil {
			e.cancel()
		}
	})
}
