package scheduler

import (
	"example.com/task149/tsnsched/internal/cycle"
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

// TestPlanSlotMatchesFrameDuration shares the integer duration contract between
// the planner and FrameDuration/BuildTrace. A non-divisible rate used to let the
// planner trim the slot (duration-- on top of truncated division), diverging from
// the trace the validator reasons about. The slot must now span exactly the
// ceiling duration, so the reserved window carries every bit of the frame.
func TestPlanSlotMatchesFrameDuration(t *testing.T) {
	const frame, rate = int64(100), int64(7)
	g := topology.New()
	_ = g.AddNode(model.Node{ID: "a", Name: "a"})
	_ = g.AddNode(model.Node{ID: "b", Name: "b"})
	_ = g.AddPort(model.Port{ID: "pa", NodeID: "a", RateBits: rate})
	_ = g.AddPort(model.Port{ID: "pb", NodeID: "b", RateBits: rate})
	_ = g.AddLink(model.Link{ID: "l", FromPort: "pa", ToPort: "pb", PeriodNS: 1_000_000, Enabled: true})
	p := New(g, Config{NetworkPeriodNS: 1_000_000})
	const release = int64(10)
	a, viol := p.Plan([]model.Stream{{ID: "s", PeriodNS: 1_000_000, FrameBits: frame, ReleaseNS: release, DeadlineNS: 1 << 62, MaxJitterNS: 1 << 62, PathPorts: []string{"pa", "pb"}}}, "d")
	if len(viol) != 0 {
		t.Fatalf("unexpected violations: %+v", viol)
	}
	if len(a) != 1 {
		t.Fatalf("want 1 alloc, got %d", len(a))
	}
	want, _ := FrameDuration(frame, rate)
	got := a[0].EndNS - a[0].StartNS
	if got != want {
		t.Fatalf("slot duration=%d want ceiling FrameDuration=%d", got, want)
	}
	// Departure time must be sized by the same duration, not duration-1.
	if a[0].DepartureNS-a[0].ArrivalNS != want {
		t.Fatalf("departure span=%d want %d", a[0].DepartureNS-a[0].ArrivalNS, want)
	}
	// Sanity: the slot product with the rate covers the whole frame.
	if got*rate < frame*1_000_000_000 {
		t.Fatalf("slot under-covers frame: %d*%d < %d", got, rate, frame*1_000_000_000)
	}
}

// TestPlanReleasesAcrossCycleBoundaryUsesSharedNormalize confirms the planner's
// cycle.Normalize is the same integer time semantic the validator's boundary
// check reasons about: a release landing exactly on a period edge normalises to 0
// and is therefore a boundary release the stream-set check must reject (see
// validator rules). Here we only assert the planner itself does not silently
// relocate such a release.
func TestPlanReleasesAcrossCycleBoundaryUsesSharedNormalize(t *testing.T) {
	const period = int64(100)
	g := topology.New()
	_ = g.AddNode(model.Node{ID: "a", Name: "a"})
	_ = g.AddNode(model.Node{ID: "b", Name: "b"})
	_ = g.AddPort(model.Port{ID: "pa", NodeID: "a", RateBits: 1_000_000_000})
	_ = g.AddPort(model.Port{ID: "pb", NodeID: "b", RateBits: 1_000_000_000})
	_ = g.AddLink(model.Link{ID: "l", FromPort: "pa", ToPort: "pb", PeriodNS: period, Enabled: true})
	p := New(g, Config{NetworkPeriodNS: period})
	// Release exactly on the period boundary: normalises to 0.
	a, _ := p.Plan([]model.Stream{{ID: "s", PeriodNS: period, FrameBits: 1, ReleaseNS: period, DeadlineNS: 1 << 62, MaxJitterNS: 1 << 62, PathPorts: []string{"pa", "pb"}}}, "d")
	if len(a) != 1 {
		t.Fatalf("want 1 alloc, got %d", len(a))
	}
	if got := cycle.Normalize(period, period); got != 0 {
		t.Fatalf("boundary release %d normalises to %d, not 0", period, got)
	}
	if a[0].StartNS != 0 {
		t.Fatalf("planner start=%d, want 0 for boundary release", a[0].StartNS)
	}
}

