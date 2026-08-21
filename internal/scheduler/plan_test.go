package scheduler

import (
	"example.com/task149/tsnsched/internal/model"
	"example.com/task149/tsnsched/internal/topology"
	"testing"
)

func TestPlanKeepsCrossCycleEnd(t *testing.T) {
	g := topology.New()
	_ = g.AddNode(model.Node{ID: "a", Name: "a"})
	_ = g.AddNode(model.Node{ID: "b", Name: "b"})
	_ = g.AddPort(model.Port{ID: "pa", NodeID: "a", RateBits: 1_000_000_000})
	_ = g.AddPort(model.Port{ID: "pb", NodeID: "b", RateBits: 1_000_000_000})
	_ = g.AddLink(model.Link{ID: "l", FromPort: "pa", ToPort: "pb", PeriodNS: 100, Enabled: true})
	p := New(g, Config{NetworkPeriodNS: 100})
	a, _ := p.Plan([]model.Stream{{ID: "s", PeriodNS: 100, FrameBits: 10, ReleaseNS: 99, DeadlineNS: 1000, MaxJitterNS: 1000, PathPorts: []string{"pa", "pb"}}}, "d")
	if len(a) != 1 || a[0].EndNS <= 100 {
		t.Fatalf("cross cycle end lost: %#v", a)
	}
}
