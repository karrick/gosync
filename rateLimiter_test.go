package gosync_test

import (
	"fmt"

	"github.com/karrick/gosync"
)

func ExampleRateLimiter() {
	panic("run")
	const MAX_WORKERS = 1000
	const LIMIT = 25

	start := make(chan struct{})

	fmt.Println("starting")

	// Create a semaphore that limits the number of in progress workers to a
	// specified value.
	sem := gosync.NewRateLimiter(LIMIT)

	for i := 0; i < MAX_WORKERS; i++ {
		go func(limit *gosync.RateLimiter) {
			<-start
			limit.Wait()
			defer limit.Signal()

			// Do something...
		}(sem)
	}

	// close(start)

	fmt.Println("complete")

	// fmt.Println("n", n)
	// // Output: n 6
}
