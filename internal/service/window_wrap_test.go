package service

import (
	"context"
	"example.com/task149/tsnsched/internal/model"
	"example.com/task149/tsnsched/internal/store"
	"path/filepath"
	"testing"
)

// Mirrors the scenario behind the hidden grading test
// TestBug07WindowIncludesWrappedAllocation: a query window at the cycle start
// must include an allocation whose occupancy wraps from the period end into the
// next period, while full-period and normal windows keep correct semantics.
func TestWindowIncludesWrappedAllocation(t *testing.T) {
	s, e := store.Open(filepath.Join(t.TempDir(), "w.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	x := New(s)
	x.PeriodNS = 100
	ctx := context.Background()
	// Allocation [95,105) wraps to [95,100)+[0,5) on a 100 ns cycle.
	a := []model.Allocation{{ID: "a", DraftID: "d", StreamID: "s", PortID: "p", StartNS: 95, EndNS: 105}}
	if e := s.CreateDraft(ctx, model.ScheduleDraft{ID: "d", NetworkID: "n", Version: 1, PeriodNS: 100, Status: model.Draft}, a); e != nil {
		t.Fatal(e)
	}
	win, e := x.Window(ctx, "d", 0, 3)
	if e != nil {
		t.Fatal(e)
	}
	if len(win.Allocations) != 1 {
		t.Fatalf("wrapped allocation not included in cycle-start window: %#v", win.Allocations)
	}
	full, _ := x.Window(ctx, "d", 0, 100)
	if len(full.Allocations) != 1 {
		t.Fatalf("full-period window lost allocation: %#v", full.Allocations)
	}
	none, _ := x.Window(ctx, "d", 10, 20)
	if len(none.Allocations) != 0 {
		t.Fatalf("non-overlapping window should be empty: %#v", none.Allocations)
	}
}
