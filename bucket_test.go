package timingwheel

import (
	"testing"
	"time"

	"github.com/GiterLab/timingwheel/utils"
)

func TestBucket_Flush(t *testing.T) {
	b := newBucket()

	b.Add(&Timer{}, utils.TimeToMs(time.Now().UTC()))
	b.Add(&Timer{}, utils.TimeToMs(time.Now().UTC()))
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
