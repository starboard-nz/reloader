package reloader_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/starboard-nz/reloader"
)

var i int32

func intLoad() (*int32, error) {
	n := i
	i++

	return &n, nil
}

func TestReloaderInt32(t *testing.T) {
	l := reloader.NewLoader[int32](intLoad, 100 * time.Millisecond)

	wg := &sync.WaitGroup{}

	// start 100 goroutines that read the same data
	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func() {
			var prev int32

			for {
				var d int32
				err := l.Do(func(n *int32) error {
					if n == nil {
						return fmt.Errorf("nil pointer")
					}

					d = *n

					return nil
				})

				if err != nil {
					t.Errorf("Error returned: %v", err)
				}

				if prev != d {
					fmt.Printf("Data = %d\n", d)
					prev = d
				}

				time.Sleep(time.Millisecond)

				if d >= 10 {
					wg.Done()
					return
				}
			}
		}()
	}

	// wait for all goroutines to return
	wg.Wait()
}

type Foo struct {
	m map[string]int32
}

var x int32

func fooLoad() (*Foo, error) {
	m := make(map[string]int32)
	m["hello"] = x
	x++

	return &Foo{m: m}, nil
}

func TestReloaderFoo(t *testing.T) {
	l := reloader.NewLoader[Foo](fooLoad, 100 * time.Millisecond)

	wg := &sync.WaitGroup{}

	// start 100 goroutines that read the same data
	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func() {
			var prev int32

			for {
				var d int32
				err := l.Do(func(n *Foo) error {
					if n == nil {
						return fmt.Errorf("nil pointer")
					}

					d = n.m["hello"]

					return nil
				})

				if err != nil {
					t.Errorf("Error returned: %v", err)
				}

				if prev != d {
					fmt.Printf("Data = %d\n", d)
					prev = d
				}

				time.Sleep(time.Millisecond)

				if d >= 10 {
					wg.Done()
					return
				}
			}
		}()
	}

	// wait for all goroutines to return
	wg.Wait()
}
