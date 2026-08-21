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

// TestCreateDraftBindsValidationToVersion locks the version-binding contract:
// a freshly generated draft must carry validation evidence bound to its own id
// with a non-empty checked_at, queryable by that id, not parked on a shared
// "unbound" sentinel that a later commit cannot read.
func TestCreateDraftBindsValidationToVersion(t *testing.T) {
	s, e := store.Open(filepath.Join(t.TempDir(), "v.db"))
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

	// Generate two draft versions; each must own its own evidence.
	d1, e := x.CreateDraft(ctx, "n", 1)
	if e != nil {
		t.Fatalf("create v1: %v", e)
	}
	d2, e := x.CreateDraft(ctx, "n", 2)
	if e != nil {
		t.Fatalf("create v2: %v", e)
	}

	for _, d := range []struct{ id string }{{d1.Draft.ID}, {d2.Draft.ID}} {
		sum, e := x.Summary(ctx, d.id)
		if e != nil {
			t.Fatalf("summary %s: %v", d.id, e)
		}
		v := sum.Validation
		if v.DraftID != d.id {
			t.Fatalf("DraftID=%q want %q (evidence bound to wrong version)", v.DraftID, d.id)
		}
		if v.CheckedAt.IsZero() {
			t.Fatalf("CheckedAt empty for %s — evidence not persisted", d.id)
		}
		if !v.Valid {
			t.Fatalf("Valid=false with %d violations for %s — evidence dropped", len(v.Violations), d.id)
		}
	}

	// Re-validating v1 must refresh only v1's evidence and leave v2 intact.
	r1, e := x.Validate(ctx, d1.Draft.ID)
	if e != nil {
		t.Fatalf("revalidate v1: %v", e)
	}
	if r1.DraftID != d1.Draft.ID || r1.CheckedAt.IsZero() {
		t.Fatalf("revalidate lost binding: %+v", r1)
	}
	v2, e := x.Summary(ctx, d2.Draft.ID)
	if e != nil || !v2.Validation.Valid || v2.Validation.DraftID != d2.Draft.ID {
		t.Fatalf("v2 evidence clobbered by revalidate: %+v", v2.Validation)
	}

	// Committing v2 (the newer version) reads only its own evidence and succeeds.
	if e = x.Commit(ctx, "n", d2.Draft.ID); e != nil {
		t.Fatalf("commit v2: %v", e)
	}
	// v1 must not be committable now: its own evidence says valid, but the active
	// version moved on — this is the version-isolation guarantee, not a gate leak.
	if e = x.Commit(ctx, "n", d1.Draft.ID); e == nil {
		t.Fatal("committing superseded v1 should fail")
	}
}
