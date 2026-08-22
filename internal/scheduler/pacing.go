package scheduler

import (
	"example.com/task149/tsnsched/internal/cycle"
	"example.com/task149/tsnsched/internal/model"
	"fmt"
)

type HopTiming struct {
	Hop           int    `json:"hop"`
	PortID        string `json:"port_id"`
	LinkID        string `json:"link_id"`
	ArrivalNS     int64  `json:"arrival_ns"`
	StartNS       int64  `json:"start_ns"`
	FinishNS      int64  `json:"finish_ns"`
	PropagationNS int64  `json:"propagation_ns"`
}
type TimingTrace struct {
	StreamID     string      `json:"stream_id"`
	Hops         []HopTiming `json:"hops"`
	ReleaseNS    int64       `json:"release_ns"`
	CompletionNS int64       `json:"completion_ns"`
	CrossedCycle bool        `json:"crossed_cycle"`
}

func FrameDuration(frameBits, rateBits int64) (int64, error) {
	if frameBits <= 0 || rateBits <= 0 {
		return 0, fmt.Errorf("frame and rate must be positive")
	}
	// Duration is measured in whole nanoseconds and must hold every bit of the
	// frame. When the port rate does not evenly divide the frame bit count, the
	// exact wire time is fractional; truncating to the quotient would reserve a
	// slot too short to carry the last bit. Round up to the smallest integral
	// number of nanoseconds whose product with the rate covers the full frame.
	num := frameBits * 1_000_000_000
	q := num / rateBits
	if num%rateBits != 0 {
		q++
	}
	return q, nil
}

func BuildTrace(stream model.Stream, links []model.Link, rates []int64, period int64) (TimingTrace, error) {
	if len(links) != len(rates) {
		return TimingTrace{}, fmt.Errorf("links and rates have different lengths")
	}
	trace := TimingTrace{StreamID: stream.ID, ReleaseNS: stream.ReleaseNS, Hops: make([]HopTiming, 0, len(links))}
	cursor := stream.ReleaseNS
	for i, link := range links {
		duration, err := FrameDuration(stream.FrameBits, rates[i])
		if err != nil {
			return TimingTrace{}, err
		}
		start := cycle.Normalize(cursor, period)
		finish := cursor + duration
		trace.Hops = append(trace.Hops, HopTiming{Hop: i, PortID: stream.PathPorts[i], LinkID: link.ID, ArrivalNS: cursor, StartNS: start, FinishNS: finish, PropagationNS: link.PropagationNS})
		cursor = finish + link.PropagationNS
	}
	trace.CompletionNS = cursor
	trace.CrossedCycle = period > 0 && cursor/period > stream.ReleaseNS/period
	return trace, nil
}

func TraceWithinDeadline(t TimingTrace, deadline int64) bool {
	return t.CompletionNS-t.ReleaseNS <= deadline
}
func TraceJitter(t TimingTrace) int64 {
	if len(t.Hops) == 0 {
		return 0
	}
	return t.CompletionNS - t.ReleaseNS
}
func TraceSummary(t TimingTrace) string {
	return fmt.Sprintf("stream=%s hops=%d completion=%d crossed_cycle=%t", t.StreamID, len(t.Hops), t.CompletionNS, t.CrossedCycle)
}

func NormalizeTrace(t TimingTrace, period int64) TimingTrace {
	for i := range t.Hops {
		t.Hops[i].StartNS = cycle.Normalize(t.Hops[i].StartNS, period)
		t.Hops[i].FinishNS = cycle.Normalize(t.Hops[i].FinishNS, period)
	}
	t.CompletionNS = cycle.Normalize(t.CompletionNS, period)
	return t
}
