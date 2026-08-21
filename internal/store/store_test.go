package store

import (
	"context"
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

// TestValidationBoundToDraftVersion locks the version-binding contract: every
// generated or re-validated result is persisted under its own draft id, is
// queryable with a non-empty checked_at, and a sibling draft's evidence cannot
// overwrite it onto a shared sentinel.
func TestValidationBoundToDraftVersion(t *testing.T) {
	s, e := Open(filepath.Join(t.TempDir(), "v.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	v1 := model.ValidationResult{DraftID: "draft-v1", Valid: true, CheckedAt: now}
	if e = s.SaveValidation(ctx, v1); e != nil {
		t.Fatalf("save v1: %v", e)
	}
	v2 := model.ValidationResult{DraftID: "draft-v2", Valid: false, Violations: []model.Violation{{Kind: model.Conflict, StreamA: "x", StreamB: "y", Detail: "overlap"}}, CheckedAt: now.Add(time.Second)}
	if e = s.SaveValidation(ctx, v2); e != nil {
		t.Fatalf("save v2: %v", e)
	}

	// v1 reads back its own evidence with a non-empty checked_at.
	got1, e := s.GetValidation(ctx, "draft-v1")
	if e != nil || !got1.Valid || got1.CheckedAt.IsZero() {
		t.Fatalf("v1 evidence lost: %+v err=%v", got1, e)
	}
	if got1.DraftID != "draft-v1" {
		t.Fatalf("v1 DraftID unbound: %q", got1.DraftID)
	}

	// v2 reads back its own (invalid) evidence independently.
	got2, e := s.GetValidation(ctx, "draft-v2")
	if e != nil || got2.Valid || len(got2.Violations) != 1 {
		t.Fatalf("v2 evidence lost: %+v err=%v", got2, e)
	}

	// No sentinel "unbound" row may exist; each version owns exactly one row.
	var n, unbound int
	if e = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM validation_results WHERE draft_id='unbound'`).Scan(&unbound); e != nil {
		t.Fatal(e)
	}
	if e = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM validation_results WHERE draft_id LIKE 'draft-v%'`).Scan(&n); e != nil {
		t.Fatal(e)
	}
	if unbound != 0 || n != 2 {
		t.Fatalf("sentinel=%d bound=%d, want sentinel=0 bound=2", unbound, n)
	}
}

// TestSaveValidationRejectsUnbound ensures a result with no draft id is rejected
// rather than being parked on a shared sentinel that later commits cannot read.
func TestSaveValidationRejectsUnbound(t *testing.T) {
	s, e := Open(filepath.Join(t.TempDir(), "u.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	ctx := context.Background()
	err := s.SaveValidation(ctx, model.ValidationResult{DraftID: "", Valid: true, CheckedAt: time.Now().UTC()})
	if err == nil {
		t.Fatal("expected error saving unbound validation")
	}
	var n int
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM validation_results WHERE draft_id='unbound'`).Scan(&n)
	if n != 0 {
		t.Fatalf("unbound sentinel leaked %d row(s)", n)
	}
}

// TestCommitRequiresOwnValidationEvidence locks the commit gate: a draft with no
// validation evidence bound to its own id must be rejected, even if another draft
// version in the same network was validated.
func TestCommitRequiresOwnValidationEvidence(t *testing.T) {
	s, e := Open(filepath.Join(t.TempDir(), "c.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	ctx := context.Background()

	// Two drafts for the same network; only v1 carries validation evidence.
	d1 := model.ScheduleDraft{ID: "d1", NetworkID: "net", Version: 1, PeriodNS: 100}
	if e = s.CreateDraft(ctx, d1, nil); e != nil {
		t.Fatal(e)
	}
	d2 := model.ScheduleDraft{ID: "d2", NetworkID: "net", Version: 2, PeriodNS: 100}
	if e = s.CreateDraft(ctx, d2, nil); e != nil {
		t.Fatal(e)
	}
	if e = s.SaveValidation(ctx, model.ValidationResult{DraftID: "d1", Valid: true, CheckedAt: time.Now().UTC()}); e != nil {
		t.Fatal(e)
	}

	// Committing the validated draft succeeds.
	if e = s.Commit(ctx, "net", "d1"); e != nil {
		t.Fatalf("commit d1: %v", e)
	}
	// Committing the unvalidated sibling must fail with ErrNotReady — its evidence
	// is not bound to d2 and must not be inherited from d1's sentinel-free record.
	err := s.Commit(ctx, "net", "d2")
	if err != model.ErrNotReady {
		t.Fatalf("commit d2 = %v, want ErrNotReady", err)
	}
}
