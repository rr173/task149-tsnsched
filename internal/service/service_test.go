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

// TestRepointedLinkIsPickedUpByNewDraft guards BUG03: after an engineer
// re-points an existing directed link's endpoints, the service's in-process
// topology (the graph CreateDraft reads from) must reflect the new direction
// immediately. It must not keep using the stale topology, nor keep the stale
// edge just because the link id is unchanged. Unmodified links keep their
// path-validation behavior.
func TestRepointedLinkIsPickedUpByNewDraft(t *testing.T) {
	s, e := store.Open(filepath.Join(t.TempDir(), "s.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	x := New(s)
	ctx := context.Background()

	_ = x.AddNode(ctx, model.Node{ID: "a", Name: "a", Enabled: true})
	_ = x.AddNode(ctx, model.Node{ID: "b", Name: "b", Enabled: true})
	_ = x.AddNode(ctx, model.Node{ID: "c", Name: "c", Enabled: true})
	_ = x.AddPort(ctx, model.Port{ID: "pa", NodeID: "a", RateBits: 1_000_000_000})
	_ = x.AddPort(ctx, model.Port{ID: "pb", NodeID: "b", RateBits: 1_000_000_000})
	_ = x.AddPort(ctx, model.Port{ID: "pc", NodeID: "c", RateBits: 1_000_000_000})

	// Original direction: l points pa -> pb.
	_ = x.AddLink(ctx, model.Link{ID: "l", FromPort: "pa", ToPort: "pb", PeriodNS: 1_000_000, Enabled: true})
	g := x.GraphSnapshot()
	if err := g.ValidatePath([]string{"pa", "pb"}); err != nil {
		t.Fatalf("original direction pa->pb rejected: %v", err)
	}

	// Engineer re-points the same link id l to pb -> pc.
	_ = x.AddLink(ctx, model.Link{ID: "l", FromPort: "pb", ToPort: "pc", PeriodNS: 1_000_000, Enabled: true})

	// The persisted row in the database must reflect the new endpoints.
	got, e := x.Data.GetLink(ctx, "l")
	if e != nil {
		t.Fatal(e)
	}
	if got.FromPort != "pb" || got.ToPort != "pc" {
		t.Fatalf("persisted link not updated: %+v", got)
	}

	// The scheduler reads s.Graph; it must already hold the new direction and
	// must NOT retain the stale edge. A draft created now validates against
	// this graph, so this is the consistency point between persistence, the
	// topology index, and the scheduling read path.
	g = x.GraphSnapshot()
	if err := g.ValidatePath([]string{"pb", "pc"}); err != nil {
		t.Fatalf("new direction pb->pc not visible to scheduler: %v", err)
	}
	if err := g.ValidatePath([]string{"pa", "pb"}); err == nil {
		t.Fatal("stale directed edge pa->pb still in scheduler's graph")
	}
	if link, ok := g.Links["l"]; !ok || link.FromPort != "pb" || link.ToPort != "pc" {
		t.Fatalf("graph link l not replaced: %+v", link)
	}

	// Unmodified-link behavior: a link that was never re-pointed still validates
	// its path normally alongside the re-pointed one.
	_ = x.AddNode(ctx, model.Node{ID: "d", Name: "d", Enabled: true})
	_ = x.AddPort(ctx, model.Port{ID: "pd", NodeID: "d", RateBits: 1_000_000_000})
	_ = x.AddPort(ctx, model.Port{ID: "pe", NodeID: "a", RateBits: 1_000_000_000})
	_ = x.AddLink(ctx, model.Link{ID: "l2", FromPort: "pc", ToPort: "pd", PeriodNS: 1_000_000, Enabled: true})
	_ = x.AddLink(ctx, model.Link{ID: "l3", FromPort: "pd", ToPort: "pe", PeriodNS: 1_000_000, Enabled: true})
	g = x.GraphSnapshot()
	if err := g.ValidatePath([]string{"pc", "pd", "pe"}); err != nil {
		t.Fatalf("unmodified link path rejected: %v", err)
	}
}

// TestRepointedLinkAcrossReopen guards the persistence layer of BUG03: after a
// re-point and a process restart (store reopen), the link must load with the
// new endpoints, not the originals.
func TestRepointedLinkAcrossReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "s.db")
	s, e := store.Open(path)
	if e != nil {
		t.Fatal(e)
	}
	x := New(s)
	ctx := context.Background()
	_ = x.AddNode(ctx, model.Node{ID: "a", Name: "a", Enabled: true})
	_ = x.AddNode(ctx, model.Node{ID: "b", Name: "b", Enabled: true})
	_ = x.AddNode(ctx, model.Node{ID: "c", Name: "c", Enabled: true})
	_ = x.AddPort(ctx, model.Port{ID: "pa", NodeID: "a", RateBits: 1})
	_ = x.AddPort(ctx, model.Port{ID: "pb", NodeID: "b", RateBits: 1})
	_ = x.AddPort(ctx, model.Port{ID: "pc", NodeID: "c", RateBits: 1})
	_ = x.AddLink(ctx, model.Link{ID: "l", FromPort: "pa", ToPort: "pb", PeriodNS: 1, Enabled: true})
	_ = x.AddLink(ctx, model.Link{ID: "l", FromPort: "pb", ToPort: "pc", PeriodNS: 1, Enabled: true})
	_ = s.Close()

	s, e = store.Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	got, e := s.GetLink(ctx, "l")
	if e != nil {
		t.Fatal(e)
	}
	if got.FromPort != "pb" || got.ToPort != "pc" {
		t.Fatalf("re-pointed link did not survive reopen: %+v", got)
	}
}
