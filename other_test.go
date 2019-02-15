package gosync_test

// This file is included for illustrative purposes only, and the example code
// herein is meant to exemplify commonly found albeit non-optimal
// synchronization attempts that many programmers, including myself, have used
// over the years.  They are not written as complete stand alone functions, and
// indeed will likely do no work at all, because there is no wait groups, nor
// attempt to shut down the consumer when there is no more incoming data.  They
// are snippets of the algorithm that one might find, only illustrating the
// attempt to synchronize the producer and consumer.

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// example_spinlock illustrates a common yet flawed approach in building
// concurrency algorithms on low-level atomic operations.  Running algorithms
// like this will result in programs taking much more time to complete than they
// should.  The effect is more pronounced in a high contention environment.
func example_spinlock() {
	const bufsize = 10
	const n = bufsize << 3
	buf := make([]int, 0, bufsize)
	var mutex uint32

	// producer
	go func() {
		for i := 0; i < n; i++ {
			if atomic.CompareAndSwapUint32(&mutex, 0, 1) {
				if len(buf) < cap(buf) {
					buf = append(buf, rand.Intn(10))
				}
				atomic.StoreUint32(&mutex, 0)
			}
		}
	}()

	// consumer
	go func() {
		var i int
		for {
			if atomic.CompareAndSwapUint32(&mutex, 0, 1) {
				if len(buf) > 0 {
					i, buf = buf[0], buf[1:]
					fmt.Println(i)
				}
				atomic.StoreUint32(&mutex, 0)
			}
		}
	}()
}

// example_spinlock_with_sleep illustrates another common yet ultimately flawed
// approach in building concurrency algorithms on low-level atomic operations.
// They represent an improvement over spinlocking alone, because threads yield
// to the scheduler in the form of a sleep operation when they are not eligible
// to run algorithmically run.  Running this sort of algorithm is also going to
// take longer than necessary to complete, because threads are sleeping longer
// than they would need to.  While they are also adversely affected in high
// contention environments, they are less affected than spinlocks without the
// sleep operations.
func example_spinlock_with_sleep() {
	const bufsize = 10
	const n = bufsize << 3
	buf := make([]int, 0, bufsize)
	var mutex uint32

	// producer
	go func() {
		for i := 0; i < n; i++ {
			if atomic.CompareAndSwapUint32(&mutex, 0, 1) {
				if len(buf) < cap(buf) {
					buf = append(buf, rand.Intn(10))
				}
				atomic.StoreUint32(&mutex, 0)
			} else {
				time.Sleep(time.Millisecond)
			}
		}
	}()

	// consumer
	go func() {
		var i int
		for {
			if atomic.CompareAndSwapUint32(&mutex, 0, 1) {
				if len(buf) > 0 {
					i, buf = buf[0], buf[1:]
					fmt.Println(i)
				}
				atomic.StoreUint32(&mutex, 0)
			} else {
				time.Sleep(time.Millisecond)
			}
		}
	}()
}

// example_mutex illustrates one of several ways to implement the
// producer-consumer problem.  It is not optimal because it lock-steps the
// producers and consumers such that only one thread may ever work with the
// buffer at a given time.  This algorithm will only scale worse as more
// producers or consumers are added, because there will be higher contention for
// the locks, while still only one thread at a given moment in time can be
// working with the queue.  If the program can process 1000 elements per second
// with a single producer and single consumer, it will never be able to work
// faster with more producers and consumers.  This is opposite to the goal we
// have when we design a producer-consumer algorithm, hoping to have it scale
// horizontally with the number of producers and consumers working with the
// data.
func example_mutex() {
	const bufsize = 10
	const n = bufsize << 3
	buf := make([]int, 0, bufsize)
	var mutex sync.Mutex

	// producer
	go func() {
		for i := 0; i < n; i++ {
			mutex.Lock()
			if len(buf) < cap(buf) {
				buf = append(buf, rand.Intn(10))
			}
			mutex.Unlock()
		}
	}()

	// consumer
	go func() {
		var i int
		for {
			mutex.Lock()
			if len(buf) > 0 {
				i, buf = buf[0], buf[1:]
				fmt.Println(i)
			}
			mutex.Unlock()
		}
	}()
}
