package pool

import (
	"context"
	"sync"
)

type Pool struct {
	cancel   func()
	workChan chan func() error

	wg      sync.WaitGroup
	errOnce sync.Once
	err     error
}

func New(poolSize int, ctx context.Context) (*Pool, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	p := &Pool{
		cancel: cancel,
	}
	for i := 0; i <= poolSize; i++ {
		p.wg.Add(1)
		go p.work()
	}
	return p, ctx
}

func (p *Pool) Wait() error {
	close(p.workChan)
	p.wg.Wait()
	if p.cancel != nil {
		p.cancel()
	}
	return p.err
}

func (p *Pool) Queue(fn func() error) {
	p.workChan <- fn
}

func (p *Pool) work() {
	defer p.wg.Done()
	for fn := range p.workChan {
		if err := fn(); err != nil {
			p.errOnce.Do(func() {
				p.err = err
				if p.cancel != nil {
					p.cancel()
				}
			})
		}
	}
}
