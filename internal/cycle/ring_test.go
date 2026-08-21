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

// TestGuardCrossesBoundary is the bug scenario: a transmission whose trailing
// guard band wraps the cycle and lands on the start of the next transmission.
// The raw intervals [98,100) and [0,2) are disjoint, yet the guard band of
// the first reaches [0,1) and therefore the reserved ranges must overlap.
func TestGuardCrossesBoundary(t *testing.T) {
	period := int64(100)
	a := Reserved(98, 2, period, Guard{After: 1}) // slot [98,100) + after-guard [0,1)
	b := Reserved(0, 2, period, Guard{})          // slot [0,2)
	if Overlap(Split(98, 2, period), Split(0, 2, period)) {
		t.Fatal("raw intervals must be disjoint for this scenario")
	}
	if !Overlap(a, b) {
		t.Fatal("trailing guard band crossing next start must overlap")
	}
}

// TestGuardLeadingCollision mirrors the scenario on the other side: the next
// transmission's leading guard band reaches back into the previous slot.
func TestGuardLeadingCollision(t *testing.T) {
	period := int64(100)
	a := Reserved(96, 2, period, Guard{})          // slot [96,98)
	b := Reserved(98, 2, period, Guard{Before: 2}) // before-guard [96,98)
	if Overlap(Split(96, 2, period), Split(98, 2, period)) {
		t.Fatal("raw intervals must be disjoint for this scenario")
	}
	if !Overlap(a, b) {
		t.Fatal("leading guard band reaching prior slot must overlap")
	}
}
