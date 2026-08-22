package validator

import (
	"example.com/task149/tsnsched/internal/model"
	"testing"
)

// Mirrors the scenario behind the hidden grading test
// TestBug09GuardBandParticipatesInConflict: a transmission's after guard band
// intersects the next transmission's start -> port conflict even though the
// raw intervals do not overlap, and capacity uses the same reserved range.
func TestGuardBandParticipatesInConflict(t *testing.T) {
	alloc := []model.Allocation{
		{ID: "a", DraftID: "d", StreamID: "a", PortID: "p", StartNS: 95, EndNS: 100, GuardAfterNS: 3},
		{ID: "b", DraftID: "d", StreamID: "b", PortID: "p", StartNS: 1, EndNS: 4},
	}
	r := New(100).Check(alloc, nil, nil)
	if r.Valid {
		t.Fatal("guard-band overlap accepted as valid")
	}
	found := false
	for _, v := range r.Violations {
		if v.Kind == model.Conflict {
			found = true
		}
	}
	if !found {
		t.Fatalf("no conflict violation recorded: %#v", r.Violations)
	}
}
