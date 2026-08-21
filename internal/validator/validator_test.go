package validator

import (
	"example.com/task149/tsnsched/internal/model"
	"testing"
)

func TestConflictsAreInvalid(t *testing.T) {
	r := New(100).Check([]model.Allocation{{ID: "a", DraftID: "d", StreamID: "a", PortID: "p", StartNS: 98, EndNS: 104, GuardAfterNS: 2}, {ID: "b", DraftID: "d", StreamID: "b", PortID: "p", StartNS: 1, EndNS: 5}}, nil, nil)
	if r.Valid {
		t.Fatal("overlap accepted")
	}
}

// TestGuardBandConflictAcrossBoundary is the bug report: a transmission's
// trailing guard band crosses the cycle boundary onto the next transmission's
// start even though the two raw intervals are disjoint. The validator must
// flag a port conflict here, not treat it as safe.
func TestGuardBandConflictAcrossBoundary(t *testing.T) {
	alloc := []model.Allocation{
		{ID: "a", DraftID: "d", StreamID: "a", PortID: "p", StartNS: 98, EndNS: 100, GuardAfterNS: 2},
		{ID: "b", DraftID: "d", StreamID: "b", PortID: "p", StartNS: 1, EndNS: 4, GuardBeforeNS: 1},
	}
	r := New(100).Check(alloc, nil, nil)
	if r.Valid {
		t.Fatalf("expected a guard-band conflict, got valid result: %+v", r)
	}
	var got bool
	for _, v := range r.Violations {
		if v.Kind == model.Conflict && v.PortID == "p" {
			got = true
		}
	}
	if !got {
		t.Fatalf("expected a conflict violation, got %+v", r.Violations)
	}
}

// TestNoFalseConflictWhenGuardLeavesGap proves the fix is precise: when the
// guard bands genuinely leave a gap, no conflict is reported.
func TestNoFalseConflictWhenGuardLeavesGap(t *testing.T) {
	alloc := []model.Allocation{
		{ID: "a", DraftID: "d", StreamID: "a", PortID: "p", StartNS: 10, EndNS: 20, GuardBeforeNS: 2, GuardAfterNS: 2},
		{ID: "b", DraftID: "d", StreamID: "b", PortID: "p", StartNS: 40, EndNS: 50, GuardBeforeNS: 2, GuardAfterNS: 2},
	}
	r := New(100).Check(alloc, nil, nil)
	if !r.Valid {
		t.Fatalf("expected valid, got violations: %+v", r.Violations)
	}
}
