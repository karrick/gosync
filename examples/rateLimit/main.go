package main

import (
	"fmt"

	"github.com/karrick/gosync"
)

const MAX_WORKERS = 1000
const LIMIT = 25

func main() {
	fmt.Println("starting")

	// Create a semaphore that limits the number of in progress workers to a
	// specified value.
	sem := gosync.NewSemaphore(LIMIT)

	for i := 0; i < MAX_WORKERS; i++ {
		go func(limit *gosync.Semaphore) {
			limit.Wait()
			defer limit.Signal()

			// Do something...
		}(sem)
	}

	fmt.Println("complete")
}
