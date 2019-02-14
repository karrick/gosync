package gosync

import "sync"

// Semaphore implements a semaphore a basic concurrency primitive.  This
// implementation uses `sync.Cond`, which uses `sync.Mutex` and the Go private
// runtime semaphore data structure, which interfaces directly with the Go
// scheduler.
type Semaphore struct {
	cv    *sync.Cond
	value uint32
}

// NewSemaphore returns a Semaphore initialized with some specified initial
// value.
func NewSemaphore(initialValue uint32) *Semaphore {
	return &Semaphore{
		cv:    &sync.Cond{L: new(sync.Mutex)},
		value: initialValue,
	}
}

// Signal is used to increment the semaphore, allowing some different thread of
// execution that is waiting on this semaphore to continue.
func (s *Semaphore) Signal() {
	s.cv.L.Lock()
	s.value++
	s.cv.L.Unlock()
	s.cv.Signal()
}

// Wait is used to block the current thread of execution until a semaphore is
// greater than 0, namely until some other thread of execution invokes the
// Signal method for this semaphore.
func (s *Semaphore) Wait() {
	s.cv.L.Lock()
	for s.value == 0 {
		s.cv.Wait()
	}
	s.value--
	s.cv.L.Unlock()
}
