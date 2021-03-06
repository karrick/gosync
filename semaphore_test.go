package gosync_test

import (
	"fmt"

	"github.com/karrick/gosync"
)

func ExampleSemaphore() {
	stage1Ready := gosync.NewSemaphore(0)
	stage2Ready := gosync.NewSemaphore(0)
	stage3Ready := gosync.NewSemaphore(0)

	var n int

	go func() {
		stage2Ready.Wait()
		n *= 3
		stage3Ready.Signal()
	}()

	go func() {
		stage1Ready.Wait()
		n += 2
		stage2Ready.Signal()
	}()

	stage1Ready.Signal() // <-- this step triggers the process
	stage3Ready.Wait()   // <-- and this step waits for it to all complete

	fmt.Println("n", n)
	// Output: n 6
}
