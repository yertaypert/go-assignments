package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func startServer(ctx context.Context, name string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(rand.Intn(500)) * time.Millisecond):
				out <- fmt.Sprintf("[%s] metric: %d", name, rand.Intn(100))
			}
		}
	}()
	return out
}

func FanIn(ctx context.Context, sources ...<-chan string) <-chan string {
	merged := make(chan string)
	var wg sync.WaitGroup

	// single source into merged
	forward := func(ch <-chan string) {
		defer wg.Done()
		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					return // source channel closed
				}
				select {
				case merged <- msg:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}

	for _, ch := range sources {
		wg.Add(1)
		go forward(ch)
	}

	go func() {
		wg.Wait()
		close(merged)
	}()

	return merged
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	alpha := startServer(ctx, "Alpha")
	beta := startServer(ctx, "Beta")
	gamma := startServer(ctx, "Gamma")

	merged := FanIn(ctx, alpha, beta, gamma)

	for msg := range merged {
		fmt.Println(msg)
	}

	fmt.Println("all servers shut down")
}
