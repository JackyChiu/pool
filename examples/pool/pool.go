package main

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/JackyChiu/pool"
)

// walkFiles starts a goroutine to walk the directory tree at root and send the
// path of each regular file on the string channel.  It sends the result of the
// walk on the error channel.  If done is closed, walkFiles abandons its work.
func walkFiles(ctx context.Context, root string, paths chan<- string) error {
	// PRODUCER
	// No select needed for this send, since errc is buffered.
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

// digester reads path names from paths and sends digests of the corresponding
// files on c until either paths or done is closed.
func digester(ctx context.Context, path string, c chan<- result) error {
	// WORKER
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	select {
	case c <- result{path, md5.Sum(data)}:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

// MD5All reads all the files in the file tree rooted at root and returns a map
// from file path to the MD5 sum of the file's contents.  If the directory walk
// fails or any read operation fails, MD5All returns an error.  In that case,
// MD5All does not wait for inflight read operations to complete.
func MD5All(root string) (map[string][md5.Size]byte, error) {
	// MD5All closes the done channel when it returns; it may do so before
	// receiving all the values from c and errc.

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool, ctx := pool.New(ctx, 25)

	paths := make(chan string)
	pool.Go(func() error {
		// Close the paths channel after Walk returns.
		defer close(paths)
		return walkFiles(ctx, root, paths)
	})

	// Start a fixed number of goroutines to read and digest files.
	c := make(chan result)
	// CONSUMER
	// DEADLOCK happens because we haven't exited the producer of results
	// no consumers are able to read!
	pool.Go(func() error {
		for path := range paths {
			path := path
			pool.Go(func() error {
				// FINISH when md5'd file
				return digester(ctx, path, c)
			})
		}
		return nil
		// FINISH when producer is done
	})

	// Closing goroutine
	go func() {
		defer close(c)
		pool.Wait()
	}()

	m := make(map[string][md5.Size]byte)
	for r := range c {
		m[r.path] = r.sum
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}
	// Check whether the Walk failed.
	// The error can still be recieved here even tho it was called already in the closing goroutine
	if err := pool.Wait(); err != nil {
		return nil, err
	}
	return m, nil
}

func main() {
	// Calculate the MD5 sum of all files under the specified directory,
	// then print the results sorted by path name.
	start := time.Now()
	m, err := MD5All(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Printf("Took %v", time.Now().Sub(start))

	var paths []string
	for path := range m {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	log.Printf("MD5'd %v files", len(paths))

	//for _, path := range paths {
	//	fmt.Printf("%x  %s\n", m[path], path)
	//}
}
