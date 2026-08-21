package service

import (
	"context"
	"example.com/task149/tsnsched/internal/model"
	"example.com/task149/tsnsched/internal/store"
	"path/filepath"
	"testing"
)

func setupService(t *testing.T) (*Service, context.Context) {
	t.Helper()
	s, e := store.Open(filepath.Join(t.TempDir(), "s.db"))
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { s.Close() })
	x := New(s)
	x.PeriodNS = 100
	x.Guard.NetworkPeriodNS = 100
	x.Guard.GuardBeforeNS = 0
	x.Guard.GuardAfterNS = 0
	ctx := context.Background()
	_ = x.AddNode(ctx, model.Node{ID: "a", Name: "a", Enabled: true})
	_ = x.AddNode(ctx, model.Node{ID: "b", Name: "b", Enabled: true})
	_ = x.AddPort(ctx, model.Port{ID: "pa", NodeID: "a", RateBits: 1_000_000_000})
	_ = x.AddPort(ctx, model.Port{ID: "pb", NodeID: "b", RateBits: 1_000_000_000})
	_ = x.AddLink(ctx, model.Link{ID: "l", FromPort: "pa", ToPort: "pb", PeriodNS: 100, Enabled: true})
	return x, ctx
}

func TestCommitIsIdempotent(t *testing.T) {
	x, ctx := setupService(t)
	_ = x.AddStream(ctx, model.Stream{ID: "s", Name: "s", PeriodNS: 1_000_000, FrameBits: 10, DeadlineNS: 1000, MaxJitterNS: 10000, PathPorts: []string{"pa", "pb"}})
	d, e := x.CreateDraft(ctx, "n", 1)
	if e != nil || !d.Validation.Valid {
		t.Fatalf("draft=%+v err=%v", d, e)
	}
	if e = x.Commit(ctx, "n", d.Draft.ID); e != nil {
		t.Fatal(e)
	}
	if e = x.Commit(ctx, "n", d.Draft.ID); e != nil {
		t.Fatal(e)
	}
}

func TestWindowIncludesWrapAroundOccupancy(t *testing.T) {
	x, ctx := setupService(t)
	// Release near the cycle end so the planned slot wraps from ~95 to ~5.
	_ = x.AddStream(ctx, model.Stream{ID: "s", Name: "s", PeriodNS: 100, FrameBits: 100, ReleaseNS: 95, DeadlineNS: 1000, MaxJitterNS: 10000, PathPorts: []string{"pa", "pb"}})
	d, e := x.CreateDraft(ctx, "n", 1)
	if e != nil || !d.Validation.Valid {
		t.Fatalf("draft=%+v err=%v", d, e)
	}
	w, e := x.Window(ctx, d.Draft.ID, 0, 10)
	if e != nil {
		t.Fatal(e)
	}
	if len(w.Allocations) != 1 {
		t.Fatalf("window at cycle start must include the wrap-around allocation, got %d", len(w.Allocations))
	}
}

func TestWindowFullPeriod(t *testing.T) {
	x, ctx := setupService(t)
	_ = x.AddStream(ctx, model.Stream{ID: "s", Name: "s", PeriodNS: 100, FrameBits: 10, ReleaseNS: 10, DeadlineNS: 1000, MaxJitterNS: 10000, PathPorts: []string{"pa", "pb"}})
	d, e := x.CreateDraft(ctx, "n", 1)
	if e != nil || !d.Validation.Valid {
		t.Fatalf("draft=%+v err=%v", d, e)
	}
	w, e := x.Window(ctx, d.Draft.ID, 0, 100)
	if e != nil {
		t.Fatal(e)
	}
	if len(w.Allocations) != 1 {
		t.Fatalf("full-period window must include the allocation, got %d", len(w.Allocations))
	}
}

func TestWindowExcludesNonOverlapping(t *testing.T) {
	x, ctx := setupService(t)
	_ = x.AddStream(ctx, model.Stream{ID: "s", Name: "s", PeriodNS: 100, FrameBits: 10, ReleaseNS: 10, DeadlineNS: 1000, MaxJitterNS: 10000, PathPorts: []string{"pa", "pb"}})
	d, e := x.CreateDraft(ctx, "n", 1)
	if e != nil || !d.Validation.Valid {
		t.Fatalf("draft=%+v err=%v", d, e)
	}
	w, e := x.Window(ctx, d.Draft.ID, 80, 95)
	if e != nil {
		t.Fatal(e)
	}
	if len(w.Allocations) != 0 {
		t.Fatalf("non-overlapping window must be empty, got %d", len(w.Allocations))
	}
}
