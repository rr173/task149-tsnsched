package scheduler

import (
	"example.com/task149/tsnsched/internal/cycle"
	"example.com/task149/tsnsched/internal/model"
	"fmt"
	"sort"
)

type PortLoad struct {
	PortID       string  `json:"port_id"`
	PeriodNS     int64   `json:"period_ns"`
	ReservedNS   int64   `json:"reserved_ns"`
	Utilization  float64 `json:"utilization"`
	Allocations  int     `json:"allocations"`
	OverCapacity bool    `json:"over_capacity"`
}
type CapacityReport struct {
	PeriodNS int64      `json:"period_ns"`
	Ports    []PortLoad `json:"ports"`
	Warnings []string   `json:"warnings"`
}

func Capacity(allocations []model.Allocation, period int64) CapacityReport {
	byPort := map[string][]model.Allocation{}
	for _, a := range allocations {
		byPort[a.PortID] = append(byPort[a.PortID], a)
	}
	ports := make([]string, 0, len(byPort))
	for port := range byPort {
		ports = append(ports, port)
	}
	sort.Strings(ports)
	report := CapacityReport{PeriodNS: period, Ports: make([]PortLoad, 0, len(ports))}
	for _, port := range ports {
		// Account for the reserved range (slot + guard bands) the same way
		// SlotGaps and conflict detection do: merge the expanded segments so
		// overlapping guard bands are not double-counted, then measure the
		// unioned width. This keeps capacity analysis on the same reserved
		// range the rest of the pipeline enforces.
		expanded := make([]cycle.Segment, 0, len(byPort[port]))
		for _, a := range byPort[port] {
			expanded = append(expanded, cycle.Reserved(a.StartNS, a.EndNS-a.StartNS, period, cycle.Guard{Before: a.GuardBeforeNS, After: a.GuardAfterNS})...)
		}
		reserved := cycle.Width(cycle.Merge(expanded))
		load := PortLoad{PortID: port, PeriodNS: period, ReservedNS: reserved, Allocations: len(byPort[port])}
		if period > 0 {
			load.Utilization = float64(reserved) / float64(period)
		}
		load.OverCapacity = reserved >= period
		if load.OverCapacity {
			report.Warnings = append(report.Warnings, fmt.Sprintf("port %s reserves %d ns in a %d ns cycle", port, reserved, period))
		}
		report.Ports = append(report.Ports, load)
	}
	return report
}
func Feasible(report CapacityReport) bool {
	for _, p := range report.Ports {
		if p.OverCapacity {
			return false
		}
	}
	return len(report.Warnings) == 0
}
func SlotGaps(items []model.Allocation, period int64) []cycle.Segment {
	occupied := []cycle.Segment{}
	for _, a := range items {
		occupied = append(occupied, cycle.Reserved(a.StartNS, a.EndNS-a.StartNS, period, cycle.Guard{Before: a.GuardBeforeNS, After: a.GuardAfterNS})...)
	}
	occupied = cycle.Merge(occupied)
	gaps := []cycle.Segment{}
	var cursor int64
	for _, s := range occupied {
		if cursor < s.Start {
			gaps = append(gaps, cycle.Segment{Start: cursor, End: s.Start})
		}
		if s.End > cursor {
			cursor = s.End
		}
	}
	if cursor < period {
		gaps = append(gaps, cycle.Segment{Start: cursor, End: period})
	}
	return gaps
}
func ExplainCapacity(report CapacityReport) string {
	if Feasible(report) {
		return "all ports fit within their cycle"
	}
	return "one or more ports exceed the cycle capacity"
}
