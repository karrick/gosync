package gosync_test

import (
	"fmt"
	"sync"

	"github.com/karrick/gosync"
)

// Mutex implements a basic concurrency primitive using a semaphore, and is
// shown here for illustrative purposes only.
type Mutex struct {
	sema *gosync.Semaphore
}

func NewMutex() *Mutex {
	return &Mutex{sema: gosync.NewSemaphore(1)}
}

func (m *Mutex) Lock() {
	m.sema.Wait()
}

func (m *Mutex) Unlock() {
	m.sema.Signal()
}

func ExampleMutex() {
	mutex := NewMutex()
	var n int

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		mutex.Lock()
		n++
		mutex.Unlock()
		wg.Done()
	}()

	go func() {
		defer wg.Done()
		mutex.Lock()
		defer mutex.Unlock()
		n++
	}()

	wg.Wait()
	fmt.Println("n", n)
	// Output: n 2
}
