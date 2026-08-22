package scheduler

import (
	"example.com/task149/tsnsched/internal/model"
	"testing"
)

func TestDetectConflictsWrapAround(t *testing.T) {
	// a occupies [95,100)+[0,5); b occupies [1,5). On the ring they collide
	// only because the wrap-around segment is no longer dropped.
	alloc := []model.Allocation{
		{ID: "a", PortID: "p", StartNS: 95, EndNS: 105},
		{ID: "b", PortID: "p", StartNS: 1, EndNS: 5},
	}
	if got := DetectConflicts(alloc, 100); len(got) != 1 {
		t.Fatalf("wrap-around conflict missed: %#v", got)
	}
}

func TestDetectConflictsGuardAfterOnly(t *testing.T) {
	// Original intervals [90,95) and [98,102) (wraps to [98,100)+[0,2)) do not
	// overlap, but a's after guard band extends to [0,1) and crosses b.
	alloc := []model.Allocation{
		{ID: "a", PortID: "p", StartNS: 90, EndNS: 95, GuardAfterNS: 6},
		{ID: "b", PortID: "p", StartNS: 98, EndNS: 102},
	}
	if got := DetectConflicts(alloc, 100); len(got) != 1 {
		t.Fatalf("after-guard conflict missed: %#v", got)
	}
}

func TestDetectConflictsNoFalsePositiveDifferentPorts(t *testing.T) {
	alloc := []model.Allocation{
		{ID: "a", PortID: "p1", StartNS: 95, EndNS: 105},
		{ID: "b", PortID: "p2", StartNS: 1, EndNS: 5},
	}
	if got := DetectConflicts(alloc, 100); len(got) != 0 {
		t.Fatalf("different ports flagged: %#v", got)
	}
}

func TestDetectConflictsNonOverlappingSamePort(t *testing.T) {
	alloc := []model.Allocation{
		{ID: "a", PortID: "p", StartNS: 0, EndNS: 10},
		{ID: "b", PortID: "p", StartNS: 20, EndNS: 30},
	}
	if got := DetectConflicts(alloc, 100); len(got) != 0 {
		t.Fatalf("non-overlapping same port flagged: %#v", got)
	}
}

func TestDetectConflictsGuardBeforeWrap(t *testing.T) {
	// b starts near the cycle end; a's before guard reaches across the boundary.
	// a=[1,4) Before=3 -> reserved [-2,4) -> wraps [98,100)+[0,4). b=[95,99).
	alloc := []model.Allocation{
		{ID: "a", PortID: "p", StartNS: 1, EndNS: 4, GuardBeforeNS: 3},
		{ID: "b", PortID: "p", StartNS: 95, EndNS: 99},
	}
	if got := DetectConflicts(alloc, 100); len(got) != 1 {
		t.Fatalf("before-guard wrap conflict missed: %#v", got)
	}
}
