package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main_bad() {
	var counter int
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter++
		}()
	}
	wg.Wait()
	fmt.Println(counter)
}

// counter++ operation is non-atomic, leading to a race condition
// where multiple goroutines concurrently perform
// read-modify-write cycles on the same memory address

func main1() {

	var counter int
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock() // only one goroutine
			counter++
			mu.Unlock()
		}()
	}

	wg.Wait()

	fmt.Println(counter)

}

func main() {

	var counter int64 // must be int64/32 for atomic ops
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddInt64(&counter, 1)
		}()
	}

	wg.Wait()

	fmt.Println(counter)
}
