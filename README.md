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

    stage1Ready.Signal() // <-- this step triggers the process
    stage3Ready.Wait()   // <-- and this step waits for it to all complete

    fmt.Println("n", n)
    // Output: n 6
}
```

## Why?

Semaphores are powerful and elegant synchronization primitives,
equally powerful to mutexes and condition variables.  While the Go
Programming Language release includes support for mutex and condition
variables, there is no built in support for semaphores.

### If Condition Variables and Mutexes Are Just As Powerful, Who Cares?

While it has been demonstrated that all three of these concurrency
primitives can be used to implement either of the other two, sometimes
a particular problem can be more succinctly expressed using one of
them, and sometimes using one of the others.  To put it a different
way--although admittedly a bit of a stretch--sometimes a programmer
wants to use `Print`, sometimes `Printf`, and sometimes `Println`.

### Fundamental Requirements for Synchronizing Concurrent Algorithms

Each of these concurrency primitives requires hardware support of
atomic operations.  On uniprocessor architectures, suspending hardware
interrupts was done to effect atomic operations.  On multiprocessor
architectures, however, each processor is working concurrently, and
can independently read data from memory, modify that data, and write
it back to memory, without regard to what the other processors are
doing.  This requires additional coordination to achieve atomicity,
necessessitating special new atomic instructions being added to the
CPU that enforce atomicity in a multiprocessor environment.  Most CPUs
provides either Test-And-Set or Compare-And-Swap instructions.
Although different in how they work, both allow the CPU to provide
Read-Modify-Write (RMW) atomicity in a multiprocessor environment.

Go provides access to low-level atomic RMW operations through its
Compare-And-Swap functions in the `sync/atomic` library.  These
functions allow our programs to read, modify, and write data with the
same level of atomicity as the operating system or Go runtime have.
However, these operations are extremely low-level, and they are not
enough to synchronize a concurrent algorithm.

For the sake of simplicity, for the extent of this document I will
refer to go-routines as threads of execution, or simply as a thread.
For quite some time the Go scheduler has scheduled M go-routines to
run on N operating system (OS) threads, so it's a fitting
simplification of terminology.

When designing a concurrent algorithm, there are moments when one
thread of execution must wait for an event to take place in a
different thread before it may continue.  In terms of the classic
producer-consumer problem, a fixed size buffer can only hold so many
elements, and a producer might need to wait for vacancies in the
buffer before placing additional elements in the buffer.  Similarly a
consumer thread must wait while there are no elements on that buffer,
and when there is one or more elements, the consumer can consume those
elements from the buffer.

When low-level atomic operations are used to implement these in Go
code, without coordination from the scheduler, the results are less
than optimal.  Either the threads will spin in a loop trying to
acquire the lock, commonly called a spinlock, or they will sleep for a
brief moment in time and try the lock again.  The problem with this
approach is there is no way for our code to tell the scheduler which
thread is runnable and which thread is blocked based on how our
algorithm should work.  The consumer thread will be given 50
milliseconds to CPU runtime, even when there are no elements in its
queue to consume.  Similarly the scheduler might schedule 50
milliseconds of CPU time for the producer while the queue is already
full.  In both of these circumstances, the scheduled thread cannot
make forward process, and ends up blocking progress of the entire
algorithm while it spinlocks waiting for algorithmic eligibility to
run.  Ironically, it is consuming the very CPU resources that could be
used by its complementing thread to unblock it.  The result is that
performance degrades very rapidly in a high contension environment.

In order to build concurrent algorithms that work properly, it is
imperative to tell the scheduler when a given thread should not be
scheduled because it is algorithmically blocked, and when a thread is
not algorithmically blocked and is eligible for scheduling.  In other
words, when working with OS threads, we need to have concurrency
primitives that work with the OS scheduler, and when working with
go-routines, we need to have concurrency primitives that work with the
Go scheduler.  This is an important point.  Just because all
concurrency primitives require atomic operations, which we have access
to via the `atomic` Go standard library, one still needs to work with
the scheduler in order to write concurrent software that does not
poll.  How do we interact with the scheduler to notify it that a
thread is either blocked or eligible according to our algorithm?  We
use mutexes, condition variables, and semaphores.

## Go Concurrency

Go provides access to mutexes and condition variables that both work
with the scheduler to prevent polling.  But it does not provide access
to semaphores.  Interestingly, Go's implementation of condition
variables uses semaphores in Go's runtime, which does in fact work
with the Go scheduler to mark go-routines as elibible to run or
blocked.  But this internal implementation of semaphores in Go is not
accessible to programmers not working in Go's runtime or standard
library.

Many in the Go community discourage use of the provided mutexes and
condition variables, not because the implementations are bad.  Quite
the reveerse: Go's mutexes and condition variables are very well
implemented.  They are efficient and well integrated with the runtime
scheduler.  However, many people have stated that because writing
concurrent software is difficult, all application developers should
write software that uses concurrency primitives at a higher-level of
abstraction than mutexes, condition variables, and semaphores.  For
this reason, the Go community champions channels as its high level
method of designing and programming concurrent algorithms.  Despite
the availability of channels since Go's initial public release,
mutexes have been provided in Go as long as I have worked with it, and
condition variables were more recently added to the `sync` standard
library.  This might suggest that the intention is to have application
developers use channels for their concurrency needs, and eschew
low-level and mid-level concurrency primitives, leaving those to the
realm of the language implementors rather than the application
developers.  Under the covers, Go provides a very easy to use
concurrency primitive called channels, that itself will be written
with the required low-level and mid-level concurrency primitives.

Channels are supposed to be simple to use, reliable, and fast.  While
they do a great job at accomplishing these goals, numerous benchmarks
demonstrate that channels are still lagging behind in the performance
category.  This is especially true when benchmarked against mutexes
and condition variables.  Furthermore, more and more of the standard
library is being updated and leveraging channels in their
implementation.  This is unfortunate, because channels are
demonstratively less performant than mid-level concurrency primitives.
The performance difference is not necessarily surprising.  If channels
are implented at a higher level of abstraction than the mid-level
primitives, then it stands to reason that channels can be no faster
than those primitives they are built upon.

Which brings me to the crux of the problem.  Go does not only attract
the novice programmer, but advanced programmers as well.  This should
not be a surprise as Go declares itself as a systems programming
language.  However, when a programmer implements an algorithm using
channels, and it's one quarter or one tenth the speed of the same
algorithm implmented using locks, then the programmer will likely
abandon the algorithm that uses channels in deference to the faster
code.

Go encourages developers to stick to using channels, but it concedes
the utility of mid-level concurrency primitives and provides
`sync.Mutex`, `sync.RWMutex`, `sync.WaitGroup`, `sync.Cond`, and
similar concurrency primitives.  Go also provides atomic low-level
primitives in the standard library in the `sync/atomic` package.
Semaphores are notably absent, but why?  Are semaphores less efficient
than mutexes or condition variables?  Maybe, depending on their
implementation.  However, the Go language itself has an implementation
of semaphores in its runtime, and that runtime implmentation of
semaphores is used to implement both `sync.Mutex` and `sync.Cond` in
the standard library.  This private semaphore code is the basis for
other Go concurrency primitives because it is elegant, fast, and
reliable.  However, the runtime implementation of semaphores is
private and not usable by code written outside the scope of the
language runtime and standard library.  So why not expose it for
others to leverage?

People want semaphores in Go because it's a general case and
performant synchronization tool.  There is even a Go library providing
semaphores, https://github.com/golang/sync, but surprisingly it builds
semaphores on top of channels rather than on either of the other two
more performant concurrency primitives available in Go.  Channels are
a beutiful abstraction, but the fact remains that at least in my
benchmarks, they are far slower than mutexes and condition variables.

To recap, at the bottom of the concurrency dependency tree are
hardware atomic operations, then integration with the thread
scheduler.  With these two prerequisites are built mutexes, condition
variables, and semaphores.  In the Go runtime, there is private
semaphore and mutex structures that use the low-level atomic
primitives.  On top of these two private structures the Go standard
library implements `sync.Mutex`, `sync.Cond`, giving application
developers access to mutexes and condition variables, and channels,
giving application developers access to high-level synchronization for
concurrent software development.

The semaphore library available at https://github.com/golang/sync is
built on top of channels, `context.Context`, `sync.Locker`, and
`container/list.List`.  `context.Context` itself is built on channels,
spawning new go-routines with channels to monitor `context.Done`
status at various points in execution.  This library is designed to be
used at a very high-level of abstraction, and is suitable for many
uses.  However, if you just want a simple semaphore, which is already
present in the Go runtime, when you use this library, the resultant
code will spawn go-routines and creates channels at multiple places,
when none of it might be needed by your algorithm.

This library is a minimal semaphore library written on nothing but
`sync.Cond`, which use the Go runtime semaphore code that is
integrated with the scheduler.  There are no channels created and no
extra go-routines spawned to use semaphores, other than those created
by the application code that might use this library.
