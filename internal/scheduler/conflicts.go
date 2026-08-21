package scheduler

import (
	"example.com/task149/tsnsched/internal/cycle"
	"example.com/task149/tsnsched/internal/model"
	"fmt"
	"time"
)

func DetectConflicts(alloc []model.Allocation, period int64) []model.Violation {
	out := []model.Violation{}
	for i := range alloc {
		for j := i + 1; j < len(alloc); j++ {
			a, b := alloc[i], alloc[j]
			if a.PortID != b.PortID {
				continue
			}
			// Compare the reserved ranges (slot + guard bands) rather than the
			// raw transmission intervals. A trailing guard band of one
			// transmission crossing the start of another's leading guard band
			// counts as a conflict even when the two raw intervals are disjoint,
			// which is exactly the case the capacity/window code already
			// reserves for via cycle.Expand.
			ra := cycle.Reserved(a.StartNS, a.EndNS-a.StartNS, period, cycle.Guard{Before: a.GuardBeforeNS, After: a.GuardAfterNS})
			rb := cycle.Reserved(b.StartNS, b.EndNS-b.StartNS, period, cycle.Guard{Before: b.GuardBeforeNS, After: b.GuardAfterNS})
			if cycle.Overlap(ra, rb) {
				out = append(out, model.Violation{ID: fmt.Sprintf("conflict-%s-%s", a.ID, b.ID), Kind: model.Conflict, StreamA: a.StreamID, StreamB: b.StreamID, PortID: a.PortID, Detail: fmt.Sprintf("overlapping ring slots %d/%d", a.StartNS, b.StartNS), CreatedAt: time.Now().UTC()})
			}
		}
	}
	return out
}
func MaxJitter(alloc []model.Allocation) int64 {
	var max int64
	for _, a := range alloc {
		if d := a.DepartureNS - a.ArrivalNS; d > max {
			max = d
		}
	}
	return max
}
