package pool_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/JackyChiu/pool"
)

var (
	Web   = fakeSearch("web")
	Image = fakeSearch("image")
	Video = fakeSearch("video")
)

type Result string
type Search func(ctx context.Context, query string) (Result, error)

func fakeSearch(kind string) Search {
	return func(_ context.Context, query string) (Result, error) {
		return Result(fmt.Sprintf("%s result for %q", kind, query)), nil
	}
}

func fake() error {
	return nil
}

func ExamplePool_without() error {
	parts := make([]string, 20)
	partChan := make(chan string)
	errChan := make(chan error, 1)

	ctx, _ := context.WithTimeout(context.Background(), time.Second)

	// Bounded Consumers/Workers
	var wg sync.WaitGroup
	wg.Add(20)
	for i := 0; i < 20; i++ {
		go func() {
			defer wg.Done()
			for part := range partChan {
				if err := fake(); err != nil {
					errChan <- ctx.Err()
					return
				}

				select {
				case <-ctx.Done():
					errChan <- ctx.Err()
					return
				default:
				}
			}
		}()
	}

	// Producer in main goroutine
	for _, part := range parts {
		select {
		case partChan <- part:
		case <-ctx.Done():
			break
		}
	}
	close(partChan)
	wg.Wait()

	select {
	case err := <-errChan:
		// did error happen?
		return err
	default:
		// success
		return nil
	}
}

func ExamplePool_with() {
	parts := make([]string, 20)
	partChan := make(chan string)

	ctx, _ := context.WithTimeout(context.Background(), time.Second)

	// Singular Producer
	go func() {
		for _, part := range parts {
			select {
			case partChan <- part:
			case <-ctx.Done():
				break
			}
		}
		close(partChan)
	}()

	// Bounded workers
	pool, ctx := pool.New(ctx, 20)
	// Singular Consumers
	for part := range partChan {
		part := part // https://golang.org/doc/faq#closures_and_goroutines
		pool.Go(func() error {
			return fake(part)
		})
	}

	if err := pool.Wait(); err != nil {
		//return err
	}
}

func ExamplePool_withoutPool() {
	errChan := make(chan error)

	var urls = []string{
		"http://www.golang.org/",
		"http://www.google.com/",
		"http://www.somestupidname.com/",
	}
	for _, url := range urls {
		// Launch a goroutine to fetch the URL.
		url := url // https://golang.org/doc/faq#closures_and_goroutines
		go func() {
			// Fetch the URL.
			resp, err := http.Get(url)
			if err == nil {
				resp.Body.Close()
			}
			errChan <- err
		}()
	}

	// No wait available, must recieve all results
	for _ = range urls {
		if err := <-errChan; err != nil {
			// Error handle
		}
	}

	// Goroutines spwaned are boundless, could be spwaning thousands.
	//

}

// JustErrors illustrates the use of a Group in place of a sync.WaitGroup to
// simplify goroutine counting and error handling. This example is derived from
// the sync.WaitGroup example at https://golang.org/pkg/sync/#example_WaitGroup.
func ExamplePool_waitgroup_functionality() {
	var g pool.Pool
	var urls = []string{
		"http://www.golang.org/",
		"http://www.google.com/",
		"http://www.somestupidname.com/",
	}
	for _, url := range urls {
		// Launch a goroutine to fetch the URL.
		url := url // https://golang.org/doc/faq#closures_and_goroutines
		g.Go(func() error {
			// Fetch the URL.
			resp, err := http.Get(url)
			if err == nil {
				resp.Body.Close()
			}
			return err
		})
	}
	// Wait for all HTTP fetches to complete.
	if err := g.Wait(); err == nil {
		fmt.Println("Successfully fetched all URLs.")
	}
}

// Parallel illustrates the use of a Group for synchronizing a simple parallel
// task: the "Google Search 2.0" function from
// https://talks.golang.org/2012/concurrency.slide#46, augmented with a Context
// and error-handling.
func ExamplePool_parallel() {
	Google := func(ctx context.Context, query string) ([]Result, error) {
		g, ctx := pool.New(ctx, 1)

		searches := []Search{Web, Image, Video}
		results := make([]Result, len(searches))
		for i, search := range searches {
			i, search := i, search // https://golang.org/doc/faq#closures_and_goroutines
			g.Go(func() error {
				result, err := search(ctx, query)
				if err == nil {
					results[i] = result
				}
				return err
			})
		}
		if err := g.Wait(); err != nil {
			return nil, err
		}
		return results, nil
	}

	results, err := Google(context.Background(), "golang")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	for _, result := range results {
		fmt.Println(result)
	}

	// Output:
	// web result for "golang"
	// image result for "golang"
	// video result for "golang"
}
