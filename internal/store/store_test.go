package store

import (
	"context"
	"errors"
	"example.com/task149/tsnsched/internal/model"
	"path/filepath"
	"testing"
	"time"
)

func TestDraftPersistsAcrossOpen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.db")
	s, e := Open(path)
	if e != nil {
		t.Fatal(e)
	}
	ctx := context.Background()
	d := model.ScheduleDraft{ID: "d", NetworkID: "n", Version: 1, PeriodNS: 100}
	if e = s.CreateDraft(ctx, d, nil); e != nil {
		t.Fatal(e)
	}
	_ = s.Close()
	s, e = Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	got, e := s.GetDraft(ctx, "d")
	if e != nil || got.ID != "d" {
		t.Fatalf("%+v %v", got, e)
	}
}

// seedValidatedDraft inserts a draft with the given status plus a passing
// validation result, so the commit gate can be exercised at the store layer.
func seedValidatedDraft(t *testing.T, s *Store, ctx context.Context, id, network string, version int64, status model.Status) {
	t.Helper()
	d := model.ScheduleDraft{ID: id, NetworkID: network, Version: version, Status: status, PeriodNS: 100}
	if e := s.CreateDraft(ctx, d, nil); e != nil {
		t.Fatal(e)
	}
	if _, e := s.db.ExecContext(ctx, `INSERT INTO validation_results(draft_id,valid,checked_at) VALUES(?,1,?)`, id, text(time.Now().UTC())); e != nil {
		t.Fatal(e)
	}
}

// TestCommitOlderVersionRejected guards the version-monotonicity constraint at
// the active-pointer layer: once a newer version is active, committing an
// older version of the same network must fail with ErrConflict and leave the
// active pointer untouched.
func TestCommitOlderVersionRejected(t *testing.T) {
	s, e := Open(filepath.Join(t.TempDir(), "s.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	ctx := context.Background()
	seedValidatedDraft(t, s, ctx, "d1", "n", 1, model.Validated)
	seedValidatedDraft(t, s, ctx, "d2", "n", 2, model.Validated)
	if e := s.Commit(ctx, "n", "d1"); e != nil {
		t.Fatalf("commit v1: %v", e)
	}
	if e := s.Commit(ctx, "n", "d2"); e != nil {
		t.Fatalf("commit v2: %v", e)
	}
	if e := s.Commit(ctx, "n", "d1"); !errors.Is(e, model.ErrConflict) {
		t.Fatalf("older recommit: want ErrConflict, got %v", e)
	}
	a, e := s.Active(ctx, "n")
	if e != nil || a.Version != 2 || a.DraftID != "d2" {
		t.Fatalf("active=%+v want v2/d2 err=%v", a, e)
	}
}

// TestCommitFinishedAfterRollbackRejected guards the draft lifecycle gate: a
// committed draft whose active pointer was removed by Rollback is finished and
// must be rejected from re-entering the commit gate.
func TestCommitFinishedAfterRollbackRejected(t *testing.T) {
	s, e := Open(filepath.Join(t.TempDir(), "s.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	ctx := context.Background()
	seedValidatedDraft(t, s, ctx, "d1", "n", 1, model.Validated)
	if e := s.Commit(ctx, "n", "d1"); e != nil {
		t.Fatalf("commit: %v", e)
	}
	if e := s.Rollback(ctx, "n"); e != nil {
		t.Fatalf("rollback: %v", e)
	}
	if e := s.Commit(ctx, "n", "d1"); !errors.Is(e, model.ErrConflict) {
		t.Fatalf("recommit after rollback: want ErrConflict, got %v", e)
	}
}

// TestCommitCurrentVersionIdempotent confirms the idempotent short-circuit:
// re-committing the active draft is a no-op success and does not drift the
// active pointer or its version.
func TestCommitCurrentVersionIdempotent(t *testing.T) {
	s, e := Open(filepath.Join(t.TempDir(), "s.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	ctx := context.Background()
	seedValidatedDraft(t, s, ctx, "d1", "n", 1, model.Validated)
	if e := s.Commit(ctx, "n", "d1"); e != nil {
		t.Fatalf("commit: %v", e)
	}
	for i := 0; i < 3; i++ {
		if e := s.Commit(ctx, "n", "d1"); e != nil {
			t.Fatalf("idempotent recommit %d: %v", i, e)
		}
	}
	a, e := s.Active(ctx, "n")
	if e != nil || a.Version != 1 || a.DraftID != "d1" {
		t.Fatalf("active drift: %+v err=%v", a, e)
	}
}

// TestCommitRejectedDraftBlocked confirms the lifecycle gate rejects drafts in
// terminal failure states (rejected) even before reaching the active pointer.
func TestCommitRejectedDraftBlocked(t *testing.T) {
	s, e := Open(filepath.Join(t.TempDir(), "s.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	ctx := context.Background()
	seedValidatedDraft(t, s, ctx, "d1", "n", 1, model.Rejected)
	if e := s.Commit(ctx, "n", "d1"); !errors.Is(e, model.ErrNotReady) {
		t.Fatalf("rejected commit: want ErrNotReady, got %v", e)
	}
}
