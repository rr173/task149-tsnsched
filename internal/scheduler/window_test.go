package scheduler

import (
	"example.com/task149/tsnsched/internal/model"
	"testing"
)

func TestInWindowCatchesWrapAroundOccupancy(t *testing.T) {
	// Slot occupies [95,5) across the boundary (no guard band).
	a := model.Allocation{StartNS: 95, EndNS: 105}
	if !InWindow(a, 0, 10, 100) {
		t.Fatal("window at cycle start must include wrap-around occupancy")
	}
}

func TestInWindowFullPeriodWindow(t *testing.T) {
	// A window spanning the whole period contains every occupancy.
	a := model.Allocation{StartNS: 95, EndNS: 105}
	if !InWindow(a, 0, 100, 100) {
		t.Fatal("full-period window must include any allocation")
	}
}

func TestInWindowNormalNonWrap(t *testing.T) {
	a := model.Allocation{StartNS: 10, EndNS: 20}
	if !InWindow(a, 15, 30, 100) {
		t.Fatal("ordinary non-wrap overlap must be detected")
	}
	if InWindow(a, 0, 5, 100) {
		t.Fatal("non-overlapping non-wrap allocation must be excluded")
	}
}

func TestInWindowGuardBandWrap(t *testing.T) {
	// Slot body [2,4) with a 3ns before-guard reaches back to 99, wrapping to [0,4).
	a := model.Allocation{StartNS: 2, EndNS: 6, GuardBeforeNS: 3}
	if !InWindow(a, 0, 10, 100) {
		t.Fatal("guard band wrapping into the cycle start must be included")
	}
}

func TestInWindowRejectsInvertedWindow(t *testing.T) {
	a := model.Allocation{StartNS: 10, EndNS: 20}
	if InWindow(a, 30, 10, 100) {
		t.Fatal("inverted window must match nothing")
	}
}
