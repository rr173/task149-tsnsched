package service

import (
	"context"
	"errors"
	"example.com/task149/tsnsched/internal/model"
	"example.com/task149/tsnsched/internal/store"
	"path/filepath"
	"testing"
)

func TestCommitIsIdempotent(t *testing.T) {
	s, e := store.Open(filepath.Join(t.TempDir(), "s.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	x := New(s)
	ctx := context.Background()
	_ = x.AddNode(ctx, model.Node{ID: "a", Name: "a", Enabled: true})
	_ = x.AddNode(ctx, model.Node{ID: "b", Name: "b", Enabled: true})
	_ = x.AddPort(ctx, model.Port{ID: "pa", NodeID: "a", RateBits: 1_000_000_000})
	_ = x.AddPort(ctx, model.Port{ID: "pb", NodeID: "b", RateBits: 1_000_000_000})
	_ = x.AddLink(ctx, model.Link{ID: "l", FromPort: "pa", ToPort: "pb", PeriodNS: 1_000_000, Enabled: true})
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

// fixtureService builds a service wired to a single-link, single-stream topology
// whose drafts validate, so version-ordering behavior can be exercised.
func fixtureService(t *testing.T) *Service {
	s, e := store.Open(filepath.Join(t.TempDir(), "s.db"))
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { s.Close() })
	x := New(s)
	ctx := context.Background()
	_ = x.AddNode(ctx, model.Node{ID: "a", Name: "a", Enabled: true})
	_ = x.AddNode(ctx, model.Node{ID: "b", Name: "b", Enabled: true})
	_ = x.AddPort(ctx, model.Port{ID: "pa", NodeID: "a", RateBits: 1_000_000_000})
	_ = x.AddPort(ctx, model.Port{ID: "pb", NodeID: "b", RateBits: 1_000_000_000})
	_ = x.AddLink(ctx, model.Link{ID: "l", FromPort: "pa", ToPort: "pb", PeriodNS: 1_000_000, Enabled: true})
	_ = x.AddStream(ctx, model.Stream{ID: "s", Name: "s", PeriodNS: 1_000_000, FrameBits: 10, DeadlineNS: 1000, MaxJitterNS: 10000, PathPorts: []string{"pa", "pb"}})
	return x
}

// TestOlderVersionCannotRecommit guards the version-monotonicity constraint at
// the service layer: once a newer version is the active version for a network,
// an older committed draft must not become the active version again. Two
// independent services (separate databases) are used so each can build a real
// validated draft without their allocations colliding (the planner derives
// allocation IDs from stream IDs).
func TestOlderVersionCannotRecommit(t *testing.T) {
	ctx := context.Background()

	x1 := fixtureService(t)
	d1, e := x1.CreateDraft(ctx, "n", 1)
	if e != nil || !d1.Validation.Valid {
		t.Fatalf("draft v1=%+v err=%v", d1, e)
	}

	// Second service on its own DB builds version 2 for the *same* network id,
	// then we hand both drafts to one store so the ordering gate is exercised
	// against a single active_versions row.
	x2 := fixtureService(t)
	d2, e := x2.CreateDraft(ctx, "n", 2)
	if e != nil || !d2.Validation.Valid {
		t.Fatalf("draft v2=%+v err=%v", d2, e)
	}
	// Copy d2 into x1's store so both drafts live in one database. The drafts
	// table's UNIQUE(network_id,version) holds because the versions differ.
	if e := x1.Data.CreateDraft(ctx, d2.Draft, nil); e != nil {
		t.Fatalf("copy d2: %v", e)
	}
	if _, e := x1.Data.DB().ExecContext(ctx, `INSERT INTO validation_results(draft_id,valid,checked_at) VALUES(?,1,'2006-01-02T15:04:05Z')`, d2.Draft.ID); e != nil {
		t.Fatalf("seed d2 validation: %v", e)
	}

	if e = x1.Commit(ctx, "n", d1.Draft.ID); e != nil {
		t.Fatalf("commit v1: %v", e)
	}
	if e = x1.Commit(ctx, "n", d2.Draft.ID); e != nil {
		t.Fatalf("commit v2: %v", e)
	}
	// Active pointer is now the newer version 2.
	if a, _ := x1.Active(ctx, "n"); a.Version != 2 || a.DraftID != d2.Draft.ID {
		t.Fatalf("active=%+v want v2", a)
	}
	// Re-committing the older v1 must be rejected and must not regress the pointer.
	if e = x1.Commit(ctx, "n", d1.Draft.ID); !errors.Is(e, model.ErrConflict) {
		t.Fatalf("older recommit: want ErrConflict, got %v", e)
	}
	if a, _ := x1.Active(ctx, "n"); a.Version != 2 || a.DraftID != d2.Draft.ID {
		t.Fatalf("active regressed after rejected recommit: %+v", a)
	}
}

// TestFinishedDraftCannotRecommit guards the draft lifecycle gate: a committed
// draft that has been rolled back (no longer the active pointer) is finished and
// must be rejected from the commit gate, even though its status row still reads
// "committed".
func TestFinishedDraftCannotRecommit(t *testing.T) {
	x := fixtureService(t)
	ctx := context.Background()
	d, e := x.CreateDraft(ctx, "n", 1)
	if e != nil || !d.Validation.Valid {
		t.Fatalf("draft=%+v err=%v", d, e)
	}
	if e = x.Commit(ctx, "n", d.Draft.ID); e != nil {
		t.Fatalf("commit: %v", e)
	}
	if e = x.Rollback(ctx, "n"); e != nil {
		t.Fatalf("rollback: %v", e)
	}
	// The draft row is still "committed" but the active pointer is gone; the
	// finished draft must not re-enter the commit gate.
	if e = x.Commit(ctx, "n", d.Draft.ID); !errors.Is(e, model.ErrConflict) {
		t.Fatalf("recommit after rollback: want ErrConflict, got %v", e)
	}
	if a, e := x.Active(ctx, "n"); e == nil {
		t.Fatalf("active pointer should be absent after rejected recommit, got %+v", a)
	}
}

// TestRecommitCurrentVersionIsIdempotentAtService confirms the idempotent
// short-circuit survives a narrow service gate: re-submitting the active draft
// repeatedly is a no-op success and does not drift the active pointer.
func TestRecommitCurrentVersionIsIdempotentAtService(t *testing.T) {
	x := fixtureService(t)
	ctx := context.Background()
	d, e := x.CreateDraft(ctx, "n", 1)
	if e != nil || !d.Validation.Valid {
		t.Fatalf("draft=%+v err=%v", d, e)
	}
	if e = x.Commit(ctx, "n", d.Draft.ID); e != nil {
		t.Fatalf("commit: %v", e)
	}
	for i := 0; i < 3; i++ {
		if e = x.Commit(ctx, "n", d.Draft.ID); e != nil {
			t.Fatalf("idempotent recommit %d: %v", i, e)
		}
	}
	a, e := x.Active(ctx, "n")
	if e != nil || a.DraftID != d.Draft.ID || a.Version != 1 {
		t.Fatalf("active drift: %+v err=%v", a, e)
	}
}

