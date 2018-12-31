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
	ctx      context.Context
	taskChan chan taskFunc
	errGroup errGroup

	// cap is the capacity of the pool
	cap int
	// the current size of the pool
	size int

	lock sync.RWMutex
}

func New(ctx context.Context, poolSize int) (*Pool, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &Pool{
		errGroup: errGroup{
			cancel: cancel,
		},
		ctx: ctx,
		// taskChan can be buffered up to a equalivient size of the pool.
		// If the pool is ever filled then the allocated poolSize is too small or the task is too large.
		taskChan: make(chan taskFunc, poolSize),

		cap: poolSize,
	}, ctx
}

func (p *Pool) Wait() error {
	close(p.taskChan)
	//return p.wait()
	return nil
}

func (p *Pool) Go(task taskFunc) {
	// how do you know if the goroutines are backed up or if you should create another goroutine?
	// by trying to send a task and then selecting
	// Ahh would've worked well if the task chan wasn't buffered
	select {
	case p.taskChan <- task:
		return
	default:
	}
	if p.Size() < int(p.cap) {
		p.startWorker()
	}
	p.taskChan <- task
}

func (p *Pool) startWorker() {
	p.incrementSize()
	p.errGroup.Go(func() error {
		for task := range p.taskChan {
			select {
			case <-p.ctx.Done():
				return nil
			default:
				p.errGroup.execute(task)
			}
		}
		return nil
	})
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

type errGroup struct {
	cancel context.CancelFunc
	wg     sync.WaitGroup

	errOnce sync.Once
	err     error
}

func (e *errGroup) wait() error {
	e.wg.Wait()
	if e.cancel != nil {
		e.cancel()
	}
	return e.err
}

func (e *errGroup) Go(task taskFunc) {
	e.wg.Add(1)

	go func() {
		defer e.wg.Done()
		e.execute(task)
	}()
}

// execute runs the task and records the first error that occurs.
// This in turn cancels any other tasks.
func (e *errGroup) execute(task taskFunc) {
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
