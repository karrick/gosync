# gosync

## Description

Semaphore concurrency primitive library for Go.

## Example

```Go
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

    stage1Ready.Signal()
    stage3Ready.Wait()

    fmt.Println("n", n)
    // Output: n 6
}
```

## Why?

Semaphores are powerful synchronization primitives that allow
implementing other synchronization primitives, such as mutexes and
condition variables.  To be clear, each of these concurrency
primitives can be implemented using either of the other two, and has
been demonstrated to be equally powerful.

Each of these concurrency primitives requires hardware support of
atomic operations, commonly Test-And-Set or Compare-And-Swap.
Additionally, in order to be implemented without polling, each of
these must be implemented with coordination from the scheduler.  For
instance, operating system (OS) concurrency primitives must be
implemented to coordinate with the OS scheduler, and Go concurrency
primitives must be implemented to coordinate with the go-routine
scheduler.

Go provides and champions channels as its high level method of
designing and programming concurrent algorithms.  Channels are
supposed to be simple to use, reliable, and fast.  However, numerous
benchmarks demonstrate that channels got the first two priorities
right, but are still lagging behind in the performance category, at
least when benchmarked against the provided mid-level concurrency
primitives.  Furthermore, more and more of the standard library is
leveraging channels in their implementation, despite the fact that
they are significantly less performant than these mid-level
primitives.  Basic concurrency primitives are far superior in terms of
performance, which is not surprising.  The term primitive itself
denotes a simple thing upon which more complex things are fashioned.
In other words channels are great, but more complicated and less
performant than mid-level alternatives.

While Go encourages developers to stick to using channels, it also
provides `sync.Mutex`, `sync.RWMutex`, `sync.Cond`, `sync.WaitGroup`,
and similar abstractions as mid-level concurrency primitives, and even
provides atomic low-level primitives in the standard library in the
`sync/atomic` package.  However, semaphores are notably absent.  Why
is this?  Are semaphores less efficient than mutexes or condition
variables?  Maybe, depending on their implementation.  However, the Go
language implements semaphores in its runtime with coordination of its
scheduler.  Oddly, that semaphore implementation is private and not
usable by code written outside the scope of the language runtime and
standard library.  The `sync.Cond` structure in Go's standard library
is implemented using both the `sync.Locker` interface, making use of
the `sync.Mutex` structure, and this private runtime semaphore
implementation.  Clearly semaphores are performant and useful, so why
not expose them for others to leverage?

People want semaphores in Go because it's a general case and
performant synchronization primitive.  There is a Go library providing
semaphores, https://github.com/golang/sync, but oddly it builds
semaphores on top of channels rather than on either of the other two
more performant concurrency primitives available in Go.  Channels are
a beutiful abstraction, but the fact remains that at least in my
benchmarks, they are big, heavy, and slow.

To recap, at the bottom of the concurrency dependency tree are atomic
operations and integration with the scheduler.  Above that are
condition variables, and mutexes, both available to Go programs, and
semaphores, only avaiable to the Go runtime and standard library.
Remember that condition variables are implemented in Go using the
public mutex library and private semaphore library.  Above these
abstactions come channels, standing firmly built on top of many other
primitives.  And to top off this summary, above channels is the
`Weighted` structure from `github.com/golang/sync`, implementing a
weighted semaphore on top of channels, `context.Context`, `sync.Lock`,
and `List` from `container/list`.  Seems like a lot of layers of
complexity all on top of runtime semaphroes, just to have user space
semaphores.

Why not expose the already excellent runtime semaphore code to users
by including hooks to it in the standard library?  That would allow
users to design software using semaphores without the additional
performance hit of additional packages, libraries, and structures that
are not needed for their domain space.

While I chew on that question for a while, this is my attempt to
provide semaphores using the bare minimum but already avaiable
structures in Go's standard library.
