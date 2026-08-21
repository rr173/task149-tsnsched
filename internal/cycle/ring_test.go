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
