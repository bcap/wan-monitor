package main

import (
	"context"
	"sync"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		monitorLatency(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		monitorBandwidth(ctx)
	}()

	wg.Wait()
}
