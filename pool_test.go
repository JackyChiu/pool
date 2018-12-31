package pool

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JackyChiu/pool"
)

func TestPool(t *testing.T) {
	poolSize := 5
	pool := New(context.Background(), poolSize)

	results := make(chan int)

	for i := 0; i < poolSize; i++ {
		pool.Go(func() error {
			results <- i
			return nil
		})
	}

	for i := 0; i < poolSize; i++ {
		<-results
	}
}

func TestPool_limits_goroutines(t *testing.T) {
	// TODO: how can you actually test this?
	// can use size var in pool
	// spin up a bunch of "heavy" tasks
}

func TestPool_lazily_loads_goroutines(t *testing.T) {
	// TODO: how can you actually test this?
	// can use size var in pool
	// spin up a bunch of "light" tasks
}

func TestPool_exits_when_context_is_cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	pool := New(ctx, 5)
	var events int32

	cancel()

	for i := 0; i < 100; i++ {
		pool.Go(func() error {
			atomic.AddInt32(&events, 1)
			return nil
		})
	}

	if err := pool.Wait(); err != context.Canceled {
		t.Errorf("expected error to be context.Canceled, got: %v", err)
	}

	if events != 0 {
		t.Error("expected only 1 go routine to run and error")
	}
}

func TestPool_exits_on_error(t *testing.T) {
	expectedErr := errors.New("error")
	pool := New(context.Background(), 5)
	var events int32

	for i := 0; i < 100; i++ {
		pool.Go(func() error {
			atomic.AddInt32(&events, 1)
			if events != 1 {
				time.Sleep(time.Millisecond)
			}
			return expectedErr
		})
	}

	err := pool.Wait()
	if expectedErr.Error() != err.Error() {
		t.Errorf("expected error to be %v, got: %v", expectedErr, err)
	}

	if events != 1 {
		t.Error("expected only 1 go routine to run and error")
	}
}

func TestZeroGroup(t *testing.T) {
	err1 := errors.New("pool_test: 1")
	err2 := errors.New("pool_test: 2")

	cases := []struct {
		errs []error
	}{
		{errs: []error{}},
		{errs: []error{nil}},
		{errs: []error{err1}},
		{errs: []error{err1, nil}},
		{errs: []error{err1, nil, err2}},
	}

	for _, tc := range cases {
		poolSize := 5
		pool := pool.New(context.Background(), poolSize)

		var firstErr error
		for i, err := range tc.errs {
			err := err
			pool.Go(func() error { return err })

			if firstErr == nil && err != nil {
				firstErr = err
			}

			if gErr := pool.Wait(); gErr != firstErr {
				t.Errorf("after %T.Go(func() error { return err }) for err in %v\n"+
					"g.Wait() = %v; want %v",
					pool, tc.errs[:i+1], err, firstErr)
			}
		}
	}
}

func TestWithContext(t *testing.T) {
	errDoom := errors.New("group_test: doomed")

	cases := []struct {
		errs []error
		want error
	}{
		{want: nil},
		{errs: []error{nil}, want: nil},
		{errs: []error{errDoom}, want: errDoom},
		{errs: []error{errDoom, nil}, want: errDoom},
	}

	for _, tc := range cases {
		ctx := context.Background()
		pool := pool.New(ctx, 5)

		for _, err := range tc.errs {
			err := err
			pool.Go(func() error { return err })
		}

		if err := pool.Wait(); err != tc.want {
			t.Errorf("after %T.Go(func() error { return err }) for err in %v\n"+
				"g.Wait() = %v; want %v",
				pool, tc.errs, err, tc.want)
		}

		canceled := false
		select {
		case <-ctx.Done():
			canceled = true
		default:
		}
		if !canceled {
			t.Errorf("after %T.Go(func() error { return err }) for err in %v\n"+
				"ctx.Done() was not closed",
				pool, tc.errs)
		}
	}
}
