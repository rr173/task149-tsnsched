package validator

import (
	"example.com/task149/tsnsched/internal/model"
	"example.com/task149/tsnsched/internal/scheduler"
	"fmt"
)

type Validator struct{ PeriodNS int64 }

func New(period int64) Validator { return Validator{PeriodNS: period} }
func (v Validator) Check(alloc []model.Allocation, streams []model.Stream, seed []model.Violation) model.ValidationResult {
	r := model.ValidationResult{Valid: true, Violations: append([]model.Violation{}, seed...)}
	r.Violations = append(r.Violations, scheduler.DetectConflicts(alloc, v.PeriodNS)...)
	by := map[string]model.Stream{}
	for _, s := range streams {
		by[s.ID] = s
	}
	for _, a := range alloc {
		if s, ok := by[a.StreamID]; ok && a.DepartureNS-a.ArrivalNS > s.MaxJitterNS {
			r.Violations = append(r.Violations, model.Violation{ID: "jitter-" + a.ID, DraftID: a.DraftID, Kind: model.Jitter, StreamA: a.StreamID, PortID: a.PortID, Detail: fmt.Sprintf("jitter %d exceeds %d", a.DepartureNS-a.ArrivalNS, s.MaxJitterNS)})
		}
	}
	r.Valid = len(r.Violations) == 0
	return r
}
func Explain(r model.ValidationResult) []string {
	out := make([]string, 0, len(r.Violations))
	for _, v := range r.Violations {
		out = append(out, fmt.Sprintf("%s: %s", v.Kind, v.Detail))
	}
	return out
}
