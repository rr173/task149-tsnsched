package cycle

import "testing"

func TestSplitWrapsAtPeriod(t *testing.T) {
	got := Split(95, 10, 100)
	if len(got) != 2 || got[0].Start != 95 || got[0].End != 100 || got[1].Start != 0 || got[1].End != 5 {
		t.Fatalf("unexpected %#v", got)
	}
}
func TestSplitNoWrap(t *testing.T) {
	got := Split(10, 5, 100)
	if len(got) != 1 || got[0].Start != 10 || got[0].End != 15 {
		t.Fatalf("unexpected %#v", got)
	}
}
func TestSplitFullPeriod(t *testing.T) {
	for _, start := range []int64{0, 50, 95} {
		got := Split(start, 100, 100)
		if len(got) != 1 || got[0].Start != 0 || got[0].End != 100 {
			t.Fatalf("start=%d unexpected %#v", start, got)
		}
	}
}
func TestOverlapWrapAroundBoundary(t *testing.T) {
	// Occupancy [95,100) + [0,5) wraps; a window at the cycle start must hit it.
	occupied := Expand(95, 10, 100, Guard{})
	if len(occupied) != 2 || !Overlap([]Segment{{Start: 0, End: 10}}, occupied) {
		t.Fatalf("wrap-around occupancy missed: %#v", occupied)
	}
}
func TestGuardOverlap(t *testing.T) {
	if !Overlap(Expand(2, 4, 100, Guard{Before: 3}), Split(98, 4, 100)) {
		t.Fatal("guard should overlap across boundary")
	}
}
