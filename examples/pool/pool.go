package main

import (
	"context"
	"crypto/md5"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/JackyChiu/pool"
)

// walkFiles starts a goroutine to walk the directory tree at root and send the
// path of each regular file on the string channel.  It sends the result of the
// walk on the error channel.  If done is closed, walkFiles abandons its work.
func walkFileTree(ctx context.Context, root string, paths chan<- string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		select {
		case paths <- path:
		case <-ctx.Done():
			return errors.New("walk canceled")
		}
		return nil
	})
}

// A result is the product of reading and summing a file using MD5.
type result struct {
	path string
	sum  [md5.Size]byte
}

// readAndSumFile reads path names from paths and sends digests of the corresponding
// files on results until either paths or done is closed.
func readAndSumFile(ctx context.Context, path string, c chan<- result) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	select {
	case c <- result{path, md5.Sum(data)}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// buildResultMap results blocks until all results are read and the producer has closed the channel.
func buildResultMap(ctx context.Context, results <-chan result) map[string][md5.Size]byte {
	m := make(map[string][md5.Size]byte)
	for {
		select {
		case r, ok := <-results:
			if !ok {
				return m
			}
			m[r.path] = r.sum
		case <-ctx.Done():
			return m
		}
	}
}

// MD5All reads all the files in the file tree rooted at root and returns a map
// from file path to the MD5 sum of the file's contents.  If the directory walk
// fails or any read operation fails, MD5All returns an error. In that case,
// MD5All does not wait for inflight read operations to complete.
func MD5All(root string) (map[string][md5.Size]byte, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Allow a fixed number of goroutines to read and digest files.
	pool, ctx := pool.New(ctx, 20)

	paths := make(chan string, 10)
	pool.Go(func() error {
		defer close(paths)
		return walkFileTree(ctx, root, paths)
	})

	results := make(chan result, 10)
	go func() {
		defer close(results)
		for filepath := range paths {
			filepath := filepath
			pool.Go(func() error {
				return readAndSumFile(ctx, filepath, results)
			})
		}
		pool.Wait()
	}()

	m := buildResultMap(ctx, results)
	if err := pool.Wait(); err != nil {
		return nil, err
	}
	return m, nil
}

func main() {
	// Calculate the MD5 sum of all files under the specified directory.
	start := time.Now()
	m, err := MD5All(os.Args[1])
	if err != nil {
		panic(err)
	}
	log.Printf("Took %v to MD5 %v files", time.Now().Sub(start), len(m))
}
