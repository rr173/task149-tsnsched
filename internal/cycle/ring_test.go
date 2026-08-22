package cycle

import "testing"

func TestSplitWrapsAtPeriod(t *testing.T) {
	got := Split(95, 10, 100)
	if len(got) != 2 || got[0].Start != 95 || got[1].End != 5 {
		t.Fatalf("unexpected %#v", got)
	}
}
func TestGuardOverlap(t *testing.T) {
	if !Overlap(Expand(2, 4, 100, Guard{Before: 3}), Split(98, 4, 100)) {
		t.Fatal("guard should overlap across boundary")
	}
}

func TestSplitFullPeriod(t *testing.T) {
	for _, start := range []int64{0, 50} {
		got := Split(start, 100, 100)
		if len(got) != 1 || got[0] != (Segment{0, 100}) {
			t.Fatalf("start=%d full period: unexpected %#v", start, got)
		}
	}
}
func TestSplitNoWrap(t *testing.T) {
	got := Split(10, 5, 100)
	if len(got) != 1 || got[0] != (Segment{10, 15}) {
		t.Fatalf("non-wrap segment changed: %#v", got)
	}
}
func TestExpandAfterGuardBand(t *testing.T) {
	got := Expand(95, 5, 100, Guard{After: 3})
	if len(got) != 2 || got[0] != (Segment{95, 100}) || got[1] != (Segment{0, 3}) {
		t.Fatalf("after guard band not wrapped: %#v", got)
	}
}
func TestExpandBeforeAfterWrap(t *testing.T) {
	// duration 4 + Before 3 + After 1 = 8, starting before 0 -> wraps to [99,100) + [0,7)
	got := Expand(2, 4, 100, Guard{Before: 3, After: 1})
	if len(got) != 2 || got[0] != (Segment{99, 100}) || got[1] != (Segment{0, 7}) {
		t.Fatalf("before+after guard band not wrapped: %#v", got)
	}
}
func TestWidthWrapAround(t *testing.T) {
	if w := Width(Split(95, 10, 100)); w != 10 {
		t.Fatalf("wrap width = %d, want 10", w)
	}
}
func TestOverlapWrapAroundBoundary(t *testing.T) {
	// occupied {95,10} wraps to [95,100)+[0,10); a window at the cycle start overlaps.
	if !Overlap(Split(0, 10, 100), Expand(95, 10, 100, Guard{})) {
		t.Fatal("wrapped occupancy at cycle start should overlap")
	}
}
