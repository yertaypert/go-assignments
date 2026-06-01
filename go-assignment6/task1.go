package main

import (
	"fmt"
	"sync"
)

func main() {
	var safeMap sync.Map
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(key int) {
			defer wg.Done()
			safeMap.Store("key", key) // Write
		}(i)
	}

	wg.Wait()

	value, ok := safeMap.Load("key") // Read

	if ok {
		fmt.Printf("Value: %v\n", value)
	}
}

func _main() {
	var mu sync.RWMutex
	safeMap := make(map[string]int)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(key int) {
			defer wg.Done()
			mu.Lock() // write lock
			safeMap["key"] = key
			mu.Unlock()
		}(i)
	}

	wg.Wait()
	mu.RLock() // multiple readers
	value := safeMap["key"]
	mu.RUnlock()

	fmt.Printf("Value: %d\n", value)

}
