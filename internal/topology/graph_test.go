package topology

import (
	"example.com/task149/tsnsched/internal/model"
	"testing"
)

func TestValidatePathRequiresDirection(t *testing.T) {
	g := New()
	_ = g.AddNode(model.Node{ID: "a", Name: "a"})
	_ = g.AddNode(model.Node{ID: "b", Name: "b"})
	_ = g.AddPort(model.Port{ID: "pa", NodeID: "a", RateBits: 1})
	_ = g.AddPort(model.Port{ID: "pb", NodeID: "b", RateBits: 1})
	if err := g.ValidatePath([]string{"pa", "pb"}); err == nil {
		t.Fatal("missing link accepted")
	}
}

// TestRepointingLinkReplacesStaleDirection guards the topology index layer of
// the BUG03 fix: when an existing link id is re-added with a new direction, the
// old directed edge must not linger in the Out index, and the new direction must
// be usable immediately. The same-link-id guard previously made AddLink a no-op,
// so the stale edge survived.
func TestRepointingLinkReplacesStaleDirection(t *testing.T) {
	g := New()
	_ = g.AddNode(model.Node{ID: "a", Name: "a"})
	_ = g.AddNode(model.Node{ID: "b", Name: "b"})
	_ = g.AddNode(model.Node{ID: "c", Name: "c"})
	_ = g.AddPort(model.Port{ID: "pa", NodeID: "a", RateBits: 1})
	_ = g.AddPort(model.Port{ID: "pb", NodeID: "b", RateBits: 1})
	_ = g.AddPort(model.Port{ID: "pc", NodeID: "c", RateBits: 1})

	_ = g.AddLink(model.Link{ID: "l", FromPort: "pa", ToPort: "pb", PeriodNS: 1, Enabled: true})
	if err := g.ValidatePath([]string{"pa", "pb"}); err != nil {
		t.Fatalf("original direction rejected: %v", err)
	}

	// Engineer re-points link l from pa->pb to pb->pc, reusing the same id.
	_ = g.AddLink(model.Link{ID: "l", FromPort: "pb", ToPort: "pc", PeriodNS: 1, Enabled: true})

	if err := g.ValidatePath([]string{"pa", "pb"}); err == nil {
		t.Fatal("stale direction pa->pb still accepted after re-point")
	}
	if err := g.ValidatePath([]string{"pb", "pc"}); err != nil {
		t.Fatalf("new direction pb->pc rejected: %v", err)
	}
	if got := g.Links["l"]; got.FromPort != "pb" || got.ToPort != "pc" {
		t.Fatalf("link l not updated: %+v", got)
	}
	// Old source must not retain a dangling entry for the re-pointed link id.
	for _, l := range g.Out["pa"] {
		if l.ID == "l" {
			t.Fatalf("stale edge lingered under pa: %+v", g.Out["pa"])
		}
	}
}

// TestRepointingLinkSameDirectionKeepsPath verifies the "keep unmodified link
// behavior" half of the fix: re-adding a link with the same id and same
// direction must leave path validation working, not break the graph.
func TestRepointingLinkSameDirectionKeepsPath(t *testing.T) {
	g := New()
	_ = g.AddNode(model.Node{ID: "a", Name: "a"})
	_ = g.AddNode(model.Node{ID: "b", Name: "b"})
	_ = g.AddPort(model.Port{ID: "pa", NodeID: "a", RateBits: 1})
	_ = g.AddPort(model.Port{ID: "pb", NodeID: "b", RateBits: 1})
	_ = g.AddLink(model.Link{ID: "l", FromPort: "pa", ToPort: "pb", PeriodNS: 1, Enabled: true})
	// Re-assert the same direction (e.g. an idempotent re-post).
	_ = g.AddLink(model.Link{ID: "l", FromPort: "pa", ToPort: "pb", PeriodNS: 1, Enabled: true})
	if err := g.ValidatePath([]string{"pa", "pb"}); err != nil {
		t.Fatalf("unmodified link path rejected: %v", err)
	}
	if n := len(g.Out["pa"]); n != 1 {
		t.Fatalf("duplicate edge entries for unmodified link: %d", n)
	}
}
