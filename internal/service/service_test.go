package service

import (
	"context"
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

// seedPlanableTopology builds the smallest topology that schedules cleanly, so
// the remaining commit tests can create validated drafts at arbitrary versions.
func seedPlanableTopology(t *testing.T, x *Service, ctx context.Context) {
	t.Helper()
	_ = x.AddNode(ctx, model.Node{ID: "a", Name: "a", Enabled: true})
	_ = x.AddNode(ctx, model.Node{ID: "b", Name: "b", Enabled: true})
	_ = x.AddPort(ctx, model.Port{ID: "pa", NodeID: "a", RateBits: 1_000_000_000})
	_ = x.AddPort(ctx, model.Port{ID: "pb", NodeID: "b", RateBits: 1_000_000_000})
	_ = x.AddLink(ctx, model.Link{ID: "l", FromPort: "pa", ToPort: "pb", PeriodNS: 1_000_000, Enabled: true})
	_ = x.AddStream(ctx, model.Stream{ID: "s", Name: "s", PeriodNS: 1_000_000, FrameBits: 10, DeadlineNS: 1000, MaxJitterNS: 10000, PathPorts: []string{"pa", "pb"}})
}

// TestCommitIdempotentReuse verifies the core BUG05 fix: re-committing the draft
// that is already the active version must take the idempotent path. The service
// gate must not reject it as a conflict, and the store must reuse the
// active-pointer row instead of inserting again (which would trip the
// network_id PRIMARY KEY unique constraint).
func TestCommitIdempotentReuse(t *testing.T) {
	s, e := store.Open(filepath.Join(t.TempDir(), "s.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	x := New(s)
	ctx := context.Background()
	seedPlanableTopology(t, x, ctx)
	d, e := x.CreateDraft(ctx, "n", 1)
	if e != nil || !d.Validation.Valid {
		t.Fatalf("draft=%+v err=%v", d, e)
	}
	if e := x.Commit(ctx, "n", d.Draft.ID); e != nil {
		t.Fatalf("initial commit: %v", e)
	}
	before, e := x.Active(ctx, "n")
	if e != nil {
		t.Fatal(e)
	}
	// Re-commit the same draft repeatedly; each call must succeed and keep the
	// pointer pointing at the same draft/version.
	for i := 0; i < 3; i++ {
		if e := x.Commit(ctx, "n", d.Draft.ID); e != nil {
			t.Fatalf("recommit %d: %v", i, e)
		}
	}
	after, e := x.Active(ctx, "n")
	if e != nil {
		t.Fatal(e)
	}
	if after.DraftID != d.Draft.ID || after.Version != 1 {
		t.Fatalf("active pointer moved: %+v", after)
	}
	if after.UpdatedAt.Before(before.UpdatedAt) {
		t.Fatalf("updated_at went backwards: before=%v after=%v", before.UpdatedAt, after.UpdatedAt)
	}
	// Exactly one active row survives: the re-commit never created a duplicate.
	n := 0
	if e := s.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM active_versions WHERE network_id=?`, "n").Scan(&n); e != nil {
		t.Fatal(e)
	}
	if n != 1 {
		t.Fatalf("expected exactly 1 active row, got %d", n)
	}
}

// TestCommitReplacesDifferentVersion verifies that the active-version swap rule
// still holds: a newer version supersedes the current active one, but a draft
// whose version is not newer is rejected as a conflict. The second draft is
// built through the store layer (with unique allocation ids) because the
// planner derives allocation ids from the stream id, so two planner-created
// drafts over the same stream collide on the allocations primary key — a
// pre-existing constraint unrelated to the idempotency fix.
func TestCommitReplacesDifferentVersion(t *testing.T) {
	s, e := store.Open(filepath.Join(t.TempDir(), "s.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	x := New(s)
	ctx := context.Background()
	seedPlanableTopology(t, x, ctx)

	v1, e := x.CreateDraft(ctx, "n", 1)
	if e != nil || !v1.Validation.Valid {
		t.Fatalf("v1 draft=%+v err=%v", v1, e)
	}
	if e := x.Commit(ctx, "n", v1.Draft.ID); e != nil {
		t.Fatalf("commit v1: %v", e)
	}

	// A strictly newer version supersedes the active pointer. Built directly
	// through the store with unique allocation ids so the planner's stream-based
	// allocation ids don't collide with v1's rows.
	if e := s.CreateDraft(ctx, model.ScheduleDraft{ID: "d2", NetworkID: "n", Version: 2, PeriodNS: x.PeriodNS}, []model.Allocation{{
		ID: "v2-s-0", DraftID: "d2", StreamID: "s", LinkID: "l", PortID: "pa",
		StartNS: 0, EndNS: 1, GuardBeforeNS: 1, GuardAfterNS: 1, ArrivalNS: 0, DepartureNS: 1,
	}}); e != nil {
		t.Fatalf("store-create v2: %v", e)
	}
	if e := s.SaveValidation(ctx, model.ValidationResult{DraftID: "d2", Valid: true, CheckedAt: v1.Draft.CreatedAt}); e != nil {
		t.Fatalf("save-validation v2: %v", e)
	}
	if e := x.Commit(ctx, "n", "d2"); e != nil {
		t.Fatalf("commit v2 should replace v1: %v", e)
	}
	active, e := x.Active(ctx, "n")
	if e != nil {
		t.Fatal(e)
	}
	if active.DraftID != "d2" || active.Version != 2 {
		t.Fatalf("active not advanced to v2: %+v", active)
	}

	// Re-committing the now-superseded v1 must NOT roll the pointer back: its
	// version is not newer than the active one, so the swap rule rejects it.
	if e := x.Commit(ctx, "n", v1.Draft.ID); e != model.ErrConflict {
		t.Fatalf("committing stale v1 want ErrConflict, got %v", e)
	}
	active, e = x.Active(ctx, "n")
	if e != nil {
		t.Fatal(e)
	}
	if active.DraftID != "d2" || active.Version != 2 {
		t.Fatalf("active regressed after stale commit: %+v", active)
	}
	// Still exactly one active row: the swap reused the network_id primary key.
	n := 0
	if e := s.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM active_versions WHERE network_id=?`, "n").Scan(&n); e != nil {
		t.Fatal(e)
	}
	if n != 1 {
		t.Fatalf("expected exactly 1 active row, got %d", n)
	}
}
