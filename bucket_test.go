package timingwheel

import (
	"testing"
	"time"
)

func TestBucket_Flush(t *testing.T) {
	b := newBucket()

	b.Add(&Timer{}, timeToMs(time.Now().UTC()))
	b.Add(&Timer{}, timeToMs(time.Now().UTC()))
	l1 := b.timers.Len()
	if l1 != 2 {
		t.Fatalf("Got (%+v) != Want (%+v)", l1, 2)
	}

	b.Flush(func(*Timer) {})
	l2 := b.timers.Len()
	if l2 != 0 {
		t.Fatalf("Got (%+v) != Want (%+v)", l2, 0)
	}
}
