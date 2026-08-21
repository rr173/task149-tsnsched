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
			ag := cycle.Guard{Before: a.GuardBeforeNS, After: a.GuardAfterNS}
			bg := cycle.Guard{Before: b.GuardBeforeNS, After: b.GuardAfterNS}
			if cycle.Overlap(cycle.Split(a.StartNS, a.EndNS-a.StartNS, period), cycle.Split(b.StartNS, b.EndNS-b.StartNS, period)) {
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
