package scheduler

import (
	"example.com/task149/tsnsched/internal/cycle"
	"example.com/task149/tsnsched/internal/model"
	"testing"
)

func TestCapacityWrapAroundReserved(t *testing.T) {
	// a=[95,100)+[0,5) -> 10 ns reserved on the ring, not 5.
	alloc := []model.Allocation{{ID: "a", PortID: "p", StartNS: 95, EndNS: 105}}
	rep := Capacity(alloc, 100)
	if len(rep.Ports) != 1 || rep.Ports[0].ReservedNS != 10 {
		t.Fatalf("wrap reserved = %d, want 10 (report=%+v)", rep.Ports[0].ReservedNS, rep)
	}
}

func TestCapacityGuardBandReserved(t *testing.T) {
	// a=[95,100)+[0,3): duration 5 + After 3 = 8 ns reserved — same range the
	// conflict check uses, so capacity and conflict evidence agree.
	alloc := []model.Allocation{{ID: "a", PortID: "p", StartNS: 95, EndNS: 100, GuardAfterNS: 3}}
	rep := Capacity(alloc, 100)
	if rep.Ports[0].ReservedNS != 8 {
		t.Fatalf("guard reserved = %d, want 8", rep.Ports[0].ReservedNS)
	}
}

func TestCapacityOverCapacityWithGuard(t *testing.T) {
	// Guard bands push a single slot past the whole cycle -> over capacity.
	alloc := []model.Allocation{{ID: "a", PortID: "p", StartNS: 0, EndNS: 100, GuardBeforeNS: 10, GuardAfterNS: 10}}
	rep := Capacity(alloc, 100)
	if !rep.Ports[0].OverCapacity {
		t.Fatalf("expected over capacity, got %+v", rep.Ports[0])
	}
	if Feasible(rep) {
		t.Fatal("feasible despite over capacity")
	}
}

func TestSlotGapsWrapAround(t *testing.T) {
	// a=[95,100)+[0,5) occupies both ends; the free gap is the middle [5,95).
	alloc := []model.Allocation{{ID: "a", PortID: "p", StartNS: 95, EndNS: 105}}
	gaps := SlotGaps(alloc, 100)
	if len(gaps) != 1 || gaps[0] != (cycle.Segment{Start: 5, End: 95}) {
		t.Fatalf("wrap gaps = %#v, want [{5 95}]", gaps)
	}
}

func TestSlotGapsMergeWrapAndNonWrap(t *testing.T) {
	// a wraps [98,100)+[0,4); b=[10,20). Free gaps: [4,10) and [20,98).
	alloc := []model.Allocation{
		{ID: "a", PortID: "p", StartNS: 98, EndNS: 104},
		{ID: "b", PortID: "p", StartNS: 10, EndNS: 20},
	}
	gaps := SlotGaps(alloc, 100)
	want := []cycle.Segment{{Start: 4, End: 10}, {Start: 20, End: 98}}
	if len(gaps) != len(want) {
		t.Fatalf("gaps = %#v, want %#v", gaps, want)
	}
	for i := range want {
		if gaps[i] != want[i] {
			t.Fatalf("gap[%d] = %#v, want %#v", i, gaps[i], want[i])
		}
	}
}
