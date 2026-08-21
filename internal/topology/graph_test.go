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
