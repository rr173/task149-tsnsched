package scheduler

import (
	"example.com/task149/tsnsched/internal/cycle"
	"example.com/task149/tsnsched/internal/model"
	"testing"
)

func TestInWindowCatchesWrapAroundOccupancy(t *testing.T) {
	// Occupancy [95,100)+[0,5) (wraps); a query window at the cycle start hits it.
	a := model.Allocation{ID: "a", PortID: "p", StartNS: 95, EndNS: 105}
	if !InWindow(a, 0, 10, 100) {
		t.Fatal("window at cycle start must hit wrapped occupancy")
	}
}

func TestInWindowFullPeriodWindow(t *testing.T) {
	a := model.Allocation{ID: "a", PortID: "p", StartNS: 50, EndNS: 60}
	if !InWindow(a, 0, 100, 100) {
		t.Fatal("full-period window must contain any occupancy")
	}
}

func TestInWindowNormalNonWrap(t *testing.T) {
	a := model.Allocation{ID: "a", PortID: "p", StartNS: 10, EndNS: 20}
	if !InWindow(a, 15, 30, 100) {
		t.Fatal("non-wrap occupancy in window must match")
	}
	if InWindow(a, 0, 5, 100) {
		t.Fatal("non-wrap occupancy outside window must not match")
	}
}

func TestInWindowGuardBandWrap(t *testing.T) {
	// a=[2,4) Before=3 -> reserved [99,100)+[0,6); window [0,10) hits the wrap.
	a := model.Allocation{ID: "a", PortID: "p", StartNS: 2, EndNS: 4, GuardBeforeNS: 3}
	if !InWindow(a, 0, 10, 100) {
		t.Fatal("guard-band wrap occupancy must hit cycle-start window")
	}
}

func TestInWindowEmptyWindow(t *testing.T) {
	a := model.Allocation{ID: "a", PortID: "p", StartNS: 0, EndNS: 10}
	if InWindow(a, 5, 5, 100) {
		t.Fatal("empty window must not match")
	}
	if InWindow(a, 10, 5, 100) {
		t.Fatal("reversed window must not match")
	}
}

// TestConflictAndCapacityShareGuardRange pins the chain consistency the fix
// targets: the conflict evidence and the capacity reservation must use the
// same guard-expanded ring occupancy, so a guard band that triggers a conflict
// is also reflected in the reserved ns.
func TestConflictAndCapacityShareGuardRange(t *testing.T) {
	alloc := []model.Allocation{
		{ID: "a", PortID: "p", StartNS: 90, EndNS: 95, GuardAfterNS: 6},
		{ID: "b", PortID: "p", StartNS: 98, EndNS: 102},
	}
	viol := DetectConflicts(alloc, 100)
	if len(viol) != 1 {
		t.Fatalf("expected 1 conflict, got %#v", viol)
	}
	rep := Capacity(alloc, 100)
	// a: [90,95) + After 6 -> Split(90,11,100) = [90,100)+[0,1) = 11 ns (wraps).
	// b: [98,102) -> [98,100)+[0,2) = 4 ns. The capacity reservation sums the
	// same guard-expanded ring occupancy the conflict check used to flag them.
	aRange := cycle.Expand(90, 5, 100, cycle.Guard{After: 6})
	bRange := cycle.Expand(98, 4, 100, cycle.Guard{})
	want := cycle.Width(aRange) + cycle.Width(bRange)
	if rep.Ports[0].ReservedNS != want {
		t.Fatalf("reserved = %d, want %d (guard range must match conflict check)", rep.Ports[0].ReservedNS, want)
	}
	if !cycle.Overlap(aRange, bRange) {
		t.Fatal("guard-expanded ranges must overlap (conflict evidence != capacity range)")
	}
}
