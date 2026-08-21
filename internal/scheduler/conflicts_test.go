package scheduler

import (
	"example.com/task149/tsnsched/internal/model"
	"testing"
)

// TestDetectConflictGuardBandAcrossBoundary is the bug report: a trailing
// guard band of one transmission crosses the cycle boundary and lands on the
// start of the next transmission. The raw intervals [98,100) and [1,5) do
// NOT overlap, but the reserved ranges must, so a conflict is reported.
func TestDetectConflictGuardBandAcrossBoundary(t *testing.T) {
	alloc := []model.Allocation{
		{ID: "a", StreamID: "a", PortID: "p", StartNS: 98, EndNS: 100, GuardAfterNS: 2},
		{ID: "b", StreamID: "b", PortID: "p", StartNS: 1, EndNS: 5},
	}
	v := DetectConflicts(alloc, 100)
	if len(v) == 0 {
		t.Fatalf("expected guard-band conflict across boundary, got none")
	}
}

// TestDetectConflictGuardBandLeading mirrors the scenario when the next
// transmission's leading guard band reaches back into the prior slot.
func TestDetectConflictGuardBandLeading(t *testing.T) {
	alloc := []model.Allocation{
		{ID: "a", StreamID: "a", PortID: "p", StartNS: 96, EndNS: 98},
		{ID: "b", StreamID: "b", PortID: "p", StartNS: 98, EndNS: 100, GuardBeforeNS: 2},
	}
	v := DetectConflicts(alloc, 100)
	if len(v) == 0 {
		t.Fatalf("expected leading guard-band conflict, got none")
	}
}

// TestDetectConflictNoFalsePositive proves the fix does not over-report:
// disjoint reserved ranges leave no conflict.
func TestDetectConflictNoFalsePositive(t *testing.T) {
	alloc := []model.Allocation{
		{ID: "a", StreamID: "a", PortID: "p", StartNS: 10, EndNS: 20, GuardBeforeNS: 2, GuardAfterNS: 2},
		{ID: "b", StreamID: "b", PortID: "p", StartNS: 40, EndNS: 50, GuardBeforeNS: 2, GuardAfterNS: 2},
	}
	if v := DetectConflicts(alloc, 100); len(v) != 0 {
		t.Fatalf("expected no conflict, got %+v", v)
	}
}

// TestDetectConflictIgnoresOtherPorts ensures conflicts are still port-scoped.
func TestDetectConflictIgnoresOtherPorts(t *testing.T) {
	alloc := []model.Allocation{
		{ID: "a", StreamID: "a", PortID: "p1", StartNS: 10, EndNS: 20, GuardAfterNS: 5},
		{ID: "b", StreamID: "b", PortID: "p2", StartNS: 12, EndNS: 18, GuardBeforeNS: 5},
	}
	if v := DetectConflicts(alloc, 100); len(v) != 0 {
		t.Fatalf("expected no conflict across different ports, got %+v", v)
	}
}

// TestCapacityAccountsForGuardBands ensures the reserved accounting folds the
// guard bands in and merges overlapping bands instead of summing raw widths.
func TestCapacityAccountsForGuardBands(t *testing.T) {
	alloc := []model.Allocation{
		{ID: "a", PortID: "p", StartNS: 10, EndNS: 20, GuardBeforeNS: 2, GuardAfterNS: 2}, // reserved [8,22)
	}
	rep := Capacity(alloc, 100)
	if len(rep.Ports) != 1 || rep.Ports[0].PortID != "p" {
		t.Fatalf("unexpected report: %+v", rep)
	}
	// raw width would be 10; with guards the reserved width is 14.
	if rep.Ports[0].ReservedNS != 14 {
		t.Fatalf("expected reserved=14 (slot 10 + 2 guards), got %d", rep.Ports[0].ReservedNS)
	}
}

// TestCapacityMergesOverlappingGuardBands ensures overlapping guard bands are
// not double-counted: two slots whose guards touch must be unioned, not added.
func TestCapacityMergesOverlappingGuardBands(t *testing.T) {
	alloc := []model.Allocation{
		{ID: "a", PortID: "p", StartNS: 10, EndNS: 20, GuardBeforeNS: 0, GuardAfterNS: 5},  // [10,25)
		{ID: "b", PortID: "p", StartNS: 22, EndNS: 30, GuardBeforeNS: 5, GuardAfterNS: 0}, // [17,30)
	}
	rep := Capacity(alloc, 100)
	if rep.Ports[0].ReservedNS != 20 { // union [10,30) == 20, not 25
		t.Fatalf("expected merged reserved=20, got %d", rep.Ports[0].ReservedNS)
	}
}
