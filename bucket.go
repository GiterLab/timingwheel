package timingwheel

import (
	"container/list"
	"sync"
	"sync/atomic"
	"unsafe"
)

// Timer represents a single event. When the Timer expires, the given
// task will be executed.
type Timer struct {
	expiration int64 // in milliseconds
	task       func(string, interface{})
	taskID     string      // task id, globally unique
	taskArgs   interface{} // task function parameters

	// The bucket that holds the list to which this timer's element belongs.
	//
	// NOTE: This field may be updated and read concurrently,
	// through Timer.Stop() and Bucket.Flush().
	b unsafe.Pointer // type: *bucket

	// The timer's element.
	element *list.Element
}

func (t *Timer) getBucket() *bucket {
	if t != nil {
		return (*bucket)(atomic.LoadPointer(&t.b))
	}
	return nil
}

func (t *Timer) setBucket(b *bucket) {
	if t != nil {
		atomic.StorePointer(&t.b, unsafe.Pointer(b))
	}
}

// Stop prevents the Timer from firing. It returns true if the call
// stops the timer, false if the timer has already expired or been stopped.
//
// If the timer t has already expired and the t.task has been started in its own
// goroutine; Stop does not wait for t.task to complete before returning. If the caller
// needs to know whether t.task is completed, it must coordinate with t.task explicitly.
func (t *Timer) Stop() bool {
	if t != nil {
		stopped := false
		for b := t.getBucket(); b != nil; b = t.getBucket() {
			// If b.Remove is called just after the timing wheel's goroutine has:
			//     1. removed t from b (through b.Flush -> b.remove)
			//     2. moved t from b to another bucket c (through b.Flush -> b.remove and c.Add)
			// this may fail to remove t due to the change of t's bucket.
			stopped = b.Remove(t)

			// Thus, here we re-get t's possibly new bucket (nil for case 1, or ab (non-nil) for case 2),
			// and retry until the bucket becomes nil, which indicates that t has finally been removed.
		}
		return stopped
	}
	return false
}

type bucket struct {
	// 64-bit atomic operations require 64-bit alignment, but 32-bit
	// compilers do not ensure it. So we must keep the 64-bit field
	// as the first field of the struct.
	//
	// For more explanations, see https://golang.org/pkg/sync/atomic/#pkg-note-BUG
	// and https://go101.org/article/memory-layout.html.
	expiration int64

	mu     sync.Mutex
	timers *list.List
}

func newBucket() *bucket {
	return &bucket{
		timers:     list.New(),
		expiration: -1,
	}
}

func (b *bucket) Expiration() int64 {
	if b != nil {
		return atomic.LoadInt64(&b.expiration)
	}
	return -1
}

// Add add timer to bucket
func (b *bucket) Add(t *Timer, expiration int64) bool {
	if b != nil {
		b.mu.Lock()
		defer b.mu.Unlock()
		e := b.timers.PushBack(t)
		t.setBucket(b)
		t.element = e
		return atomic.SwapInt64(&b.expiration, expiration) != expiration
	}
	return false
}

func (b *bucket) remove(t *Timer) bool {
	if b != nil {
		if t.getBucket() != b {
			// If remove is called from t.Stop, and this happens just after the timing wheel's goroutine has:
			//     1. removed t from b (through b.Flush -> b.remove)
			//     2. moved t from b to another bucket c (through b.Flush -> b.remove and c.Add)
			// then t.getBucket will return nil for case 1, or c (non-nil) for case 2.
			// In either case, the returned value does not equal to b.
			return false
		}
		b.timers.Remove(t.element)
		t.setBucket(nil)
		t.element = nil
		return true
	}
	return false
}

// Remove remove timer from bucket
func (b *bucket) Remove(t *Timer) bool {
	if b != nil {
		b.mu.Lock()
		defer b.mu.Unlock()
		return b.remove(t)
	}
	return false
}

// Flush flush timers from bucket
func (b *bucket) Flush(reinsert func(*Timer)) {
	var ts []*Timer

	if b != nil {
		b.mu.Lock()
		for e := b.timers.Front(); e != nil; {
			next := e.Next()

			t := e.Value.(*Timer)
			b.remove(t)
			ts = append(ts, t)

			e = next
		}
		atomic.SwapInt64(&b.expiration, -1)
		b.mu.Unlock()
	}

	for _, t := range ts {
		// addOrRun()
		//     1. already expired
		//     2. add to another bucket c (high timingwheel to low timingwheel)
		reinsert(t)
	}
}
