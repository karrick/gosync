package gosync_test

import (
	"fmt"
	"sync"
)

// WaitGroup implementation is very similar to a semaphore, and is shown here
// for illustrative purposes only.
type WaitGroup struct {
	cv    *sync.Cond
	value uint32
}

func NewWaitGroup() *WaitGroup {
	return &WaitGroup{
		cv: &sync.Cond{L: new(sync.Mutex)},
	}
}

func (wg *WaitGroup) Add(n uint32) {
	wg.cv.L.Lock()
	wg.value += n
	wg.cv.L.Unlock()
	wg.cv.Signal()
}

func (wg *WaitGroup) Done() {
	wg.cv.L.Lock()
	if wg.value == 0 {
		panic("negative condition variable")
	}
	wg.value--
	wg.cv.L.Unlock()
	wg.cv.Signal()
}

func (wg *WaitGroup) Wait() {
	wg.cv.L.Lock()
	for wg.value > 0 {
		wg.cv.Wait()
	}
	wg.cv.L.Unlock()
}

func ExampleWaitGroup() {
	var mutex sync.Mutex
	var n int

	wg := NewWaitGroup()
	wg.Add(2)

	go func() {
		defer wg.Done()
		mutex.Lock()
		n++
		mutex.Unlock()
	}()

	go func() {
		mutex.Lock()
		n++
		mutex.Unlock()
		wg.Done()
	}()

	wg.Wait()
	fmt.Println("n", n)
	// Output: n 2
}
