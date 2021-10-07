package timingwheel

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/GiterLab/timingwheel/utils"
	"github.com/GiterLab/timingwheel/utils/delayqueue"
	"github.com/tobyzxj/uuid"
)

// TimingWheel is an implementation of Hierarchical Timing Wheels.
type TimingWheel struct {
	tick      int64 // in milliseconds
	wheelSize int64

	interval    int64 // in milliseconds, interval = tick * wheelSize
	currentTime int64 // in milliseconds
	buckets     []*bucket
	queue       *delayqueue.DelayQueue

	// The higher-level overflow wheel.
	//
	// NOTE: This field may be updated and read concurrently, through Add().
	overflowWheel unsafe.Pointer // type: *TimingWheel

	exitC     chan struct{}
	waitGroup utils.WaitGroupWrapper

	// Locks used to protect data structures while ticking
	readWriteLock *sync.RWMutex
}

// NewTimingWheel creates an instance of TimingWheel with the given tick and wheelSize.
func NewTimingWheel(tick time.Duration, wheelSize int64) *TimingWheel {
	tickMs := int64(tick / time.Millisecond)
	if tickMs <= 0 {
		panic(errors.New("tick must be greater than or equal to 1ms"))
	}
	startMs := utils.TimeToMs(time.Now().UTC())
	return newTimingWheel(
		tickMs,
		wheelSize,
		startMs,
		delayqueue.New(int(wheelSize)),
		new(sync.RWMutex),
	)
}

// newTimingWheel is an internal helper function that really creates an instance of TimingWheel.
func newTimingWheel(tickMs int64, wheelSize int64, startMs int64, queue *delayqueue.DelayQueue, readWriteLock *sync.RWMutex) *TimingWheel {
	buckets := make([]*bucket, wheelSize)
	for i := range buckets {
		buckets[i] = newBucket()
	}
	return &TimingWheel{
		tick:          tickMs,
		wheelSize:     wheelSize,
		currentTime:   utils.Truncate(startMs, tickMs),
		interval:      tickMs * wheelSize,
		buckets:       buckets,
		queue:         queue,
		exitC:         make(chan struct{}),
		readWriteLock: readWriteLock,
	}
}

// add inserts the timer t into the current timing wheel.
func (tw *TimingWheel) add(t *Timer) (bool, error) {
	if tw == nil {
		return false, errors.New("tw is nil")
	}

	currentTime := atomic.LoadInt64(&tw.currentTime)
	if t.expiration < currentTime+tw.tick {
		// Already expired
		return false, nil
	} else if t.expiration < currentTime+tw.interval {
		// Put it into its own bucket
		virtualID := t.expiration / tw.tick
		b := tw.buckets[virtualID%tw.wheelSize]

		// Set the bucket expiration time
		if b.Add(t, virtualID*tw.tick) {
			// The bucket needs to be enqueued since it was an expired bucket.
			// We only need to enqueue the bucket when its expiration time has changed,
			// i.e. the wheel has advanced and this bucket get reused with a new expiration.
			// Any further calls to set the expiration within the same wheel cycle will
			// pass in the same value and hence return false, thus the bucket with the
			// same expiration will not be enqueued multiple times.
			tw.queue.Offer(b, b.Expiration())
		}
		return true, nil
	} else {
		// Out of the interval. Put it into the overflow wheel
		overflowWheel := atomic.LoadPointer(&tw.overflowWheel)
		if overflowWheel == nil {
			atomic.CompareAndSwapPointer(
				&tw.overflowWheel,
				nil,
				unsafe.Pointer(newTimingWheel(
					tw.interval,
					tw.wheelSize,
					currentTime,
					tw.queue,
					tw.readWriteLock,
				)),
			)
			overflowWheel = atomic.LoadPointer(&tw.overflowWheel)
		}
		return (*TimingWheel)(overflowWheel).add(t)
	}
}

// addOrRun inserts the timer t into the current timing wheel, or run the
// timer's task if it has already expired.
func (tw *TimingWheel) addOrRun(t *Timer) {
	if tw != nil {
		tw.readWriteLock.RLocker().Lock()
		isExpired, err := tw.add(t)
		if err != nil {
			TraceError("addOrRun error: %v", err)
		}
		tw.readWriteLock.RLocker().Unlock()
		if !isExpired {
			// Already expired

			// Like the standard time.AfterFunc (https://golang.org/pkg/time/#AfterFunc),
			// always execute the timer's task in its own goroutine.
			go t.task(t.taskID, t.taskArgs)
		}
	} else {
		TraceError("addOrRun error, tw is nil")
	}
}

func (tw *TimingWheel) advanceClock(expiration int64) {
	if tw != nil {
		currentTime := atomic.LoadInt64(&tw.currentTime)
		if expiration >= currentTime+tw.tick {
			currentTime = utils.Truncate(expiration, tw.tick)
			atomic.StoreInt64(&tw.currentTime, currentTime)

			// Try to advance the clock of the overflow wheel if present
			overflowWheel := atomic.LoadPointer(&tw.overflowWheel)
			if overflowWheel != nil {
				(*TimingWheel)(overflowWheel).advanceClock(currentTime)
			}
		}
	} else {
		TraceError("advanceClock error, tw is nil")
	}
}

// Start starts the current timing wheel.
func (tw *TimingWheel) Start() error {
	if tw == nil || tw.readWriteLock == nil {
		return errors.New("tw is nil or tw.readWriteLock == nil")
	}

	tw.waitGroup.Wrap(func() {
		tw.queue.Poll(tw.exitC, func() int64 {
			return utils.TimeToMs(time.Now().UTC())
		})
	})

	tw.waitGroup.Wrap(func() {
		for {
			select {
			case elem := <-tw.queue.C:
				b := elem.(*bucket)
				if b != nil {
					tw.readWriteLock.Lock()
					tw.advanceClock(b.Expiration())
					tw.readWriteLock.Unlock()
					b.Flush(tw.addOrRun)
				}
			case <-tw.exitC:
				return
			}
		}
	})
	return nil
}

// Stop stops the current timing wheel.
//
// If there is any timer's task being running in its own goroutine, Stop does
// not wait for the task to complete before returning. If the caller needs to
// know whether the task is completed, it must coordinate with the task explicitly.
func (tw *TimingWheel) Stop() {
	if tw != nil {
		close(tw.exitC)
		tw.waitGroup.Wait()
	}
}

// AfterFunc waits for the duration to elapse and then calls f in its own goroutine.
// It returns a Timer that can be used to cancel the call using its Stop method.
func (tw *TimingWheel) AfterFunc(d time.Duration, f func()) *Timer {
	if tw != nil {
		t := &Timer{
			expiration: utils.TimeToMs(time.Now().UTC().Add(d)),
			task: func(id string, args interface{}) {
				TraceInfo("task-id: %v", id)
				f()
			},
			taskID: fmt.Sprintf("auto_after_func_%v_%v", uuid.New(), utils.TimeToMs(time.Now().UTC())),
		}
		tw.addOrRun(t)
		return t
	}
	return nil
}

// AfterFuncWithArgs the same as AfterFunc, but more user arguments
func (tw *TimingWheel) AfterFuncWithArgs(d time.Duration, f func(string, interface{}), id string, args interface{}) *Timer {
	if tw != nil {
		if id == "" {
			id = fmt.Sprintf("auto_after_func_with_args_%v_%v", uuid.New(), utils.TimeToMs(time.Now().UTC()))
		}
		t := &Timer{
			expiration: utils.TimeToMs(time.Now().UTC().Add(d)),
			task: func(id string, args interface{}) {
				TraceInfo("task-id: %v, task-args: %v", id, args)
				f(id, args)
			},
			taskID:   id,
			taskArgs: args,
		}
		tw.addOrRun(t)
		return t
	}
	return nil
}

// Scheduler determines the execution plan of a task.
type Scheduler interface {
	// Next returns the next execution time after the given (previous) time.
	// It will return a zero time if no next time is scheduled.
	//
	// All times must be UTC.
	Next(time.Time) time.Time
}

// ScheduleFunc calls f (in its own goroutine) according to the execution
// plan scheduled by s. It returns a Timer that can be used to cancel the
// call using its Stop method.
//
// If the caller want to terminate the execution plan halfway, it must
// stop the timer and ensure that the timer is stopped actually, since in
// the current implementation, there is a gap between the expiring and the
// restarting of the timer. The wait time for ensuring is short since the
// gap is very small.
//
// Internally, ScheduleFunc will ask the first execution time (by calling
// s.Next()) initially, and create a timer if the execution time is non-zero.
// Afterwards, it will ask the next execution time each time f is about to
// be executed, and f will be called at the next execution time if the time
// is non-zero.
func (tw *TimingWheel) ScheduleFunc(s Scheduler, f func()) (t *Timer) {
	expiration := s.Next(time.Now().UTC())
	if expiration.IsZero() {
		// No time is scheduled, return nil.
		return nil
	}
	if tw != nil {
		t = &Timer{
			expiration: utils.TimeToMs(expiration),
			task: func(id string, args interface{}) {
				// Schedule the task to execute at the next time if possible.
				expiration := s.Next(utils.MsToTime(t.expiration))
				if !expiration.IsZero() {
					t.expiration = utils.TimeToMs(expiration)
					tw.addOrRun(t)
				}

				// Actually execute the task.
				TraceInfo("task-id: %v", id)
				f()
			},
			taskID: fmt.Sprintf("auto_scheduler_%v_%v", uuid.New(), utils.TimeToMs(time.Now().UTC())),
		}
		tw.addOrRun(t)
		return t
	}
	return nil
}

// ScheduleFuncWithArgs the same as ScheduleFunc, but more user arguments
func (tw *TimingWheel) ScheduleFuncWithArgs(s Scheduler, f func(string, interface{}), id string, args interface{}) (t *Timer) {
	expiration := s.Next(time.Now().UTC())
	if expiration.IsZero() {
		// No time is scheduled, return nil.
		return nil
	}
	if tw != nil {
		if id == "" {
			id = fmt.Sprintf("auto_scheduler_with_args_%v_%v", uuid.New(), utils.TimeToMs(time.Now().UTC()))
		}
		t = &Timer{
			expiration: utils.TimeToMs(expiration),
			task: func(id string, args interface{}) {
				// Schedule the task to execute at the next time if possible.
				expiration := s.Next(utils.MsToTime(t.expiration))
				if !expiration.IsZero() {
					t.expiration = utils.TimeToMs(expiration)
					tw.addOrRun(t)
				}

				// Actually execute the task.
				TraceInfo("task-id: %v, task-args: %v", id, args)
				f(id, args)
			},
			taskID:   id,
			taskArgs: args,
		}
		tw.addOrRun(t)
		return t
	}
	return nil
}
