package scheduler

import (
	"example.com/task149/tsnsched/internal/cycle"
	"example.com/task149/tsnsched/internal/model"
	"sort"
)

func SortAllocations(items []model.Allocation) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].PortID == items[j].PortID {
			return items[i].StartNS < items[j].StartNS
		}
		return items[i].PortID < items[j].PortID
	})
}

func InWindow(a model.Allocation, from, to, period int64) bool {
	if to <= from {
		return false
	}
	if to-from >= period {
		return true
	}
	window := cycle.Split(from, to-from, period)
	occupied := cycle.Reserved(a.StartNS, a.EndNS-a.StartNS, period, cycle.Guard{Before: a.GuardBeforeNS, After: a.GuardAfterNS})
	return cycle.Overlap(window, occupied)
}

func GroupByPort(items []model.Allocation) map[string][]model.Allocation {
	groups := map[string][]model.Allocation{}
	for _, a := range items {
		groups[a.PortID] = append(groups[a.PortID], a)
	}
	for k := range groups {
		SortAllocations(groups[k])
	}
	return groups
}
