package service

import (
	"context"
	"example.com/task149/tsnsched/internal/model"
	"example.com/task149/tsnsched/internal/store"
	"path/filepath"
	"testing"
)

// seedTopology builds a minimal valid TSN topology (two nodes, two ports, one
// directed link) and a single stream over the link, returning the in-memory
// graph node/port/link counts expected after a rebuild.
func seedTopology(t *testing.T, x *Service, ctx context.Context) {
	t.Helper()
	if e := x.AddNode(ctx, model.Node{ID: "a", Name: "a", Enabled: true}); e != nil {
		t.Fatal(e)
	}
	if e := x.AddNode(ctx, model.Node{ID: "b", Name: "b", Enabled: true}); e != nil {
		t.Fatal(e)
	}
	if e := x.AddPort(ctx, model.Port{ID: "pa", NodeID: "a", RateBits: 1_000_000_000}); e != nil {
		t.Fatal(e)
	}
	if e := x.AddPort(ctx, model.Port{ID: "pb", NodeID: "b", RateBits: 1_000_000_000}); e != nil {
		t.Fatal(e)
	}
	if e := x.AddLink(ctx, model.Link{ID: "l", FromPort: "pa", ToPort: "pb", PeriodNS: 1_000_000, Enabled: true}); e != nil {
		t.Fatal(e)
	}
	if e := x.AddStream(ctx, model.Stream{ID: "s", Name: "s", PeriodNS: 1_000_000, FrameBits: 10, DeadlineNS: 1000, MaxJitterNS: 10000, PathPorts: []string{"pa", "pb"}}); e != nil {
		t.Fatal(e)
	}
}

// TestReloadPreservesCommittedActiveAndRebuildsTopology guards the TSN
// state-reload boundary end to end: after Reload, the already-committed active
// version must remain queryable and the persisted node/link topology must be
// restored into the in-memory index, while the committed draft is neither
// promoted beyond committed nor deleted.
func TestReloadPreservesCommittedActiveAndRebuildsTopology(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.db")
	s, e := store.Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	x := New(s)
	ctx := context.Background()
	seedTopology(t, x, ctx)

	// Commit a validated draft: this writes the active version row that a
	// reload must never destroy, and a committed draft row that must survive.
	committed, e := x.CreateDraft(ctx, "net", 1)
	if e != nil || !committed.Validation.Valid {
		t.Fatalf("draft=%+v err=%v", committed, e)
	}
	if e = x.Commit(ctx, "net", committed.Draft.ID); e != nil {
		t.Fatal(e)
	}
	before, e := x.Active(ctx, "net")
	if e != nil {
		t.Fatalf("active before reload: %v", e)
	}
	if before.DraftID != committed.Draft.ID || before.Version != 1 {
		t.Fatalf("unexpected active before reload: %+v", before)
	}

	// Reload — the boundary under test.
	if e = x.Reload(ctx); e != nil {
		t.Fatalf("reload: %v", e)
	}

	// (1) Already-committed active version must remain queryable and identical.
	after, e := x.Active(ctx, "net")
	if e != nil {
		t.Fatalf("active after reload: %v", e)
	}
	if after.DraftID != before.DraftID || after.Version != before.Version {
		t.Fatalf("active version changed across reload: before=%+v after=%+v", before, after)
	}

	// (2) Persisted node/port/link topology must be restored into the
	// in-memory index. Before the fix, Graph() discarded the rebuilt graph and
	// returned an empty one, so ValidatePath failed with "port missing".
	g := x.GraphSnapshot()
	if g == nil {
		t.Fatal("graph nil after reload")
	}
	if len(g.Nodes) != 2 || len(g.Ports) != 2 || len(g.Links) != 1 {
		t.Fatalf("topology not rebuilt into memory index: nodes=%d ports=%d links=%d", len(g.Nodes), len(g.Ports), len(g.Links))
	}
	if err := g.ValidatePath([]string{"pa", "pb"}); err != nil {
		t.Fatalf("rebuilt graph missing directed link: %v", err)
	}

	// (3) The committed draft must survive reload with its committed status
	// intact — Reload must not promote, demote, or delete persisted drafts.
	cd, e := x.Data.GetDraft(ctx, committed.Draft.ID)
	if e != nil {
		t.Fatalf("committed draft disappeared after reload: %v", e)
	}
	if cd.Status != model.Committed {
		t.Fatalf("committed draft status changed across reload: got %q", cd.Status)
	}
	// And its allocations must still be queryable through the recovered store.
	allocs, e := x.Data.Allocations(ctx, committed.Draft.ID)
	if e != nil {
		t.Fatalf("committed allocations unreadable after reload: %v", e)
	}
	if len(allocs) != len(committed.Allocations) {
		t.Fatalf("allocation count changed across reload: before=%d after=%d", len(committed.Allocations), len(allocs))
	}
}

// TestRecoverRehydratesPersistedTopologyAfterProcessRestart simulates a process
// restart by reopening the same SQLite file and calling Recover: the already
// persisted nodes and links must be rehydrated into the fresh in-memory index
// (not an empty graph), so the service can continue serving the restored
// topology.
func TestRecoverRehydratesPersistedTopologyAfterProcessRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.db")
	s, e := store.Open(path)
	if e != nil {
		t.Fatal(e)
	}
	x := New(s)
	ctx := context.Background()
	seedTopology(t, x, ctx)
	if e = s.Close(); e != nil {
		t.Fatal(e)
	}

	// Reopen the same database — the process-restart case.
	s, e = store.Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	x = New(s)
	if e = x.Recover(ctx); e != nil {
		t.Fatalf("recover after reopen: %v", e)
	}
	g := x.GraphSnapshot()
	if g == nil || len(g.Nodes) != 2 || len(g.Ports) != 2 || len(g.Links) != 1 {
		t.Fatalf("persisted topology not rehydrated into memory index: %+v", g)
	}
	if err := g.ValidatePath([]string{"pa", "pb"}); err != nil {
		t.Fatalf("rehydrated graph missing directed link: %v", err)
	}
}
