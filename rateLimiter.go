package gosync

import "sync"

// RateLimiter implements a basic concurrency primitive known as a semaphore.
// Where mutexes are designed such that the same thread of execution acquires
// the mutex, performs some critical code, then releases the mutex, semaphores
// are designed such that one thread of execution waits until some condition is
// met, and a different thread of execution signals that the condition is met.
//
// This implementation will neither create channels, spawn go-routines, nor uses
// any library that does.  It is built upon `sync.Cond` from Go's standard
// library, which itself uses a private semaphore data structure in Go's runtime
// that interfaces directly with the Go scheduler.
type RateLimiter struct {
	noCopy noCopy

	cv           *sync.Cond
	value, limit uint32
}

// NewRateLimiter returns a RateLimiter initialized with some specified initial
// value.
//
//     func ExampleRateLimiter() {
//         stage1Ready := gosync.NewRateLimiter(0)
//         stage2Ready := gosync.NewRateLimiter(0)
//         stage3Ready := gosync.NewRateLimiter(0)
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
func NewRateLimiter(limit uint32) *RateLimiter {
	return &RateLimiter{
		cv: &sync.Cond{L: new(sync.Mutex)},
		// value: limit,
		limit: limit,
	}
}

// Signal is used to signal that a different thread of execution may proceed.
// This method never blocks, but it does allow a different thread, blocked on
// this semaphore, to proceed.
func (s *RateLimiter) Signal() {
	s.cv.L.Lock()
	s.value--
	s.cv.L.Unlock()
	s.cv.Signal()
}

// Wait is used to block the current thread of execution until a signal is
// received.  Wait will not block if the semaphore value is greater than 0.
func (s *RateLimiter) Wait() {
	s.cv.L.Lock()
	for s.value >= s.limit {
		s.cv.Wait()
	}
	s.value++
	s.cv.L.Unlock()
	s.cv.Signal()
}

// func (s *RateLimiter) UpdateLimit(limit uint32) {
// 	s.cv.L.Lock()

// 	increase := limit - s.limit // positive means new limit is larger

// 	if increase > 0 {
// 		// The new limit is larger than the existing limit.
// 		s.value += increase
// 		s.limit = limit

// 		// Done
// 		s.cv.L.Unlock()
// 		s.cv.Signal()
// 		return
// 	}

// 	// The new limit is smaller than existing limit.

// 	if s.value >= (-increase) { // FIXME sign management
// 		// Have enough room to shrink
// 		s.value += increase
// 		s.limit = limit

// 		// Done
// 		s.cv.L.Unlock()
// 		s.cv.Signal()
// 	}

// 	panic("hard case: do not have room to shrink")

// 	// Done
// 	s.cv.L.Unlock()
// 	s.cv.Signal()
// }
