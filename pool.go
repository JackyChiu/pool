package pool

import (
	"context"
	"log"
	"sync"
)

type taskFunc func() error

// Pool should be used once. Once Wait() gets called it's goroutines are reaped.
// Consider writting a Reset to use the pool again.
// Should the pool be able to be used concurrently?
type Pool struct {
	ctx           context.Context
	tasksBuffered chan taskFunc
	tasks         chan taskFunc
	errGroup      errorPool

	// cap is the capacity of the pool
	cap int
	// the current size of the pool
	size int

	lock sync.RWMutex
}

func New(ctx context.Context, poolSize int) (*Pool, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &Pool{
		errGroup: errorPool{
			cancel: cancel,
		},
		ctx:   ctx,
		tasks: make(chan taskFunc),
		// taskBufferedChan can be buffered up to a equalivient size of the pool.
		// If the pool is ever filled then the allocated poolSize is too small or the task is too large.
		tasksBuffered: make(chan taskFunc, poolSize),

		cap: poolSize,
	}, ctx
}

func (p *Pool) Go(task taskFunc) {
	// how do you know if the goroutines are backed up or if you should create another goroutine?
	// by trying to send a task and then selecting
	// Ahh would've worked well if the task chan wasn't buffered
	// Requiring cap + 1 tasks to start spinning up more goroutines is unexpected behaviour
	if p.Size() >= int(p.cap) {
		select {
		case <-p.ctx.Done():
			// don't block when context is cancelled
		case p.tasksBuffered <- task:
		}
		return
	}

	select {
	case p.tasks <- task:
		log.Println("send successful")
		return
	case <-p.ctx.Done():
	default:
	}
	p.startWorker()
	log.Println("starting wait")

	select {
	case p.tasks <- task:
	case <-p.ctx.Done():
		// don't block when context is cancelled
	}
}

func (p *Pool) startWorker() {
	p.errGroup.Go(func() error {
		for {
			select {
			case task, ok := <-p.tasks:
				if !ok {
					return nil
				}
				p.errGroup.execute(task)
			case task, ok := <-p.tasksBuffered:
				if !ok {
					return nil
				}
				p.errGroup.execute(task)
			case <-p.ctx.Done():
				return p.ctx.Err()
			}
		}
	})
	p.incrementSize()
}

func (p *Pool) Wait() error {
	close(p.tasks)
	close(p.tasksBuffered)
	return p.errGroup.wait()
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
	if e.cancel != nil {
		e.cancel()
	}
	return e.err
}

func (e *errorPool) Go(task taskFunc) {
	e.wg.Add(1)

	go func() {
		defer e.wg.Done()
		log.Println("spun up")
		e.execute(task)
		log.Println("spun down")
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
