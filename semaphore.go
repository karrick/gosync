package gosync

import "sync"

// Semaphore implements a basic concurrency primitive known as a semaphore.
// Where mutexes are designed such that the same thread of execution acquires
// the mutex, performs some critical code, then releases the mutex, semaphores
// are designed such that one thread of execution waits until some condition is
// met, and a different thread of execution signals that the condition is met.
//
// This implementation will neither create channels, spawn go-routines, nor uses
// any library that does.  It is built upon `sync.Cond` from Go's standard
// library, which itself uses a private semaphore data structure in Go's runtime
// that interfaces directly with the Go scheduler.
type Semaphore struct {
	noCopy noCopy

	cv    *sync.Cond
	value uint32
}

// NewSemaphore returns a Semaphore initialized with some specified initial
// value.
//
//     func ExampleSemaphore() {
//         stage1Ready := gosync.NewSemaphore(0)
//         stage2Ready := gosync.NewSemaphore(0)
//         stage3Ready := gosync.NewSemaphore(0)
//
//         var n int
//
//         go func() {
//             stage2Ready.Wait()
//             n *= 3
//             stage3Ready.Signal()
//         }()
//
//         go func() {
//             stage1Ready.Wait()
//             n += 2
//             stage2Ready.Signal()
//         }()
//
//         stage1Ready.Signal() // <-- this step triggers the process
//         stage3Ready.Wait()   // <-- and this step waits for it to all complete
//
//         fmt.Println("n", n)
//         // Output: n 6
//     }
func NewSemaphore(initialValue uint32) *Semaphore {
	return &Semaphore{
		cv:    &sync.Cond{L: new(sync.Mutex)},
		value: initialValue,
	}
}

// Signal is used to signal that a different thread of execution may proceed.
// This method never blocks, but it does allow a different thread, blocked on
// this semaphore, to proceed.
func (s *Semaphore) Signal() {
	s.cv.L.Lock()
	s.value++
	s.cv.L.Unlock()
	s.cv.Signal()
}

// Wait is used to block the current thread of execution until a signal is
// received.  Wait will not block if the semaphore value is greater than 0.
func (s *Semaphore) Wait() {
	s.cv.L.Lock()
	for s.value == 0 {
		s.cv.Wait()
	}
	s.value--
	s.cv.L.Unlock()
}

// noCopy may be embedded into structs which must not be copied after the first
// use.
//
// See https://golang.org/issues/8005#issuecomment-190753527 for details.
type noCopy struct{}

// Lock is a no-op used by -copylocks checker from `go vet`.
func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
