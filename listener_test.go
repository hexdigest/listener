package listener

import (
	"fmt"
	"io/ioutil"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"net/http"

	"golang.org/x/sync/errgroup"
)

func TestListener(t *testing.T) {
	l, err := New("tcp", "localhost:0", 5)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	t.Logf(l.Addr().String())

	var counter int32
	var max int32
	var maxGoroutines int
	var mu sync.Mutex

	concurrencyLevel := 100

	go func() {
		err := http.Serve(l, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			t.Logf("Got request")
			c := atomic.AddInt32(&counter, 1)
			defer atomic.AddInt32(&counter, -1)

			numGorotines := runtime.NumGoroutine()

			mu.Lock()
			if c > max {
				max = c
			}
			if numGorotines > maxGoroutines {
				maxGoroutines = numGorotines
			}
			mu.Unlock()
			time.Sleep(10 * time.Millisecond)
		}))
		t.Logf("serve error: %v", err)
	}()

	var eg errgroup.Group
	for i := 0; i < concurrencyLevel; i++ {
		eg.Go(func() error {
			return req(l.Addr().String())
		})
	}

	if err := eg.Wait(); err != nil {
		t.Fatalf("eg.Weiat: %v", err)
	}

	if max != 5 {
		t.Fatalf("max = %d", max)
	} else {
		t.Logf("max = %d", max)
		t.Logf("maxGoroutines = %d", maxGoroutines-concurrencyLevel)
	}
}

func req(addr string) error {
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	defer c.Close()

	if _, err := c.Write([]byte("GET / HTTP/1.0\n\n")); err != nil {
		return fmt.Errorf("c.Write: %w", err)
	}

	if _, err := ioutil.ReadAll(c); err != nil {
		return fmt.Errorf("ioutil.ReadAll: %w", err)
	}

	return nil
}
