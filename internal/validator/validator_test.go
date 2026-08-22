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

// TestReleaseOnPeriodBoundaryRejected guards the cycle-edge contract. A release
// landing exactly on the period boundary (release == period) is ambiguous on the
// ring: cycle.Normalize(period, period) == 0, so it coincides with the next
// cycle's start and must be rejected rather than silently wrapped. The previous
// `ReleaseNS > PeriodNS` test let this case through as a false pass.
func TestReleaseOnPeriodBoundaryRejected(t *testing.T) {
	r := New(100).Check(nil, []model.Stream{{ID: "s", PeriodNS: 100, ReleaseNS: 100, DeadlineNS: 1000}}, nil)
	if r.Valid {
		t.Fatalf("boundary release accepted: %+v", r)
	}
	if !containsKind(r.Violations, "release-period-s") {
		t.Fatalf("missing release-period violation: %+v", r.Violations)
	}
}

// TestReleaseJustInsidePeriodAccepted confirms the boundary check rejects only
// the edge and beyond, not a release safely inside the window.
func TestReleaseJustInsidePeriodAccepted(t *testing.T) {
	r := New(100).Check(nil, []model.Stream{{ID: "s", PeriodNS: 100, ReleaseNS: 99, DeadlineNS: 1000}}, nil)
	if !r.Valid {
		t.Fatalf("valid in-period release rejected: %+v", r.Violations)
	}
}

func containsKind(vs []model.Violation, id string) bool {
	for _, v := range vs {
		if v.ID == id {
			return true
		}
	}
	return false
}

