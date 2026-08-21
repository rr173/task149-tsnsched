package validator

import (
	"example.com/task149/tsnsched/internal/model"
	"fmt"
)

func CheckStreamSet(streams []model.Stream) []model.Violation {
	seen := map[string]bool{}
	out := []model.Violation{}
	for _, s := range streams {
		if seen[s.ID] {
			out = append(out, model.Violation{ID: "duplicate-stream-" + s.ID, Kind: model.PathError, StreamA: s.ID, Detail: "stream id is repeated"})
		}
		seen[s.ID] = true
		if s.ReleaseNS > s.PeriodNS {
			out = append(out, model.Violation{ID: "release-period-" + s.ID, Kind: model.Deadline, StreamA: s.ID, Detail: fmt.Sprintf("release %d is outside period %d", s.ReleaseNS, s.PeriodNS)})
		}
		if s.DeadlineNS < s.ReleaseNS {
			out = append(out, model.Violation{ID: "deadline-release-" + s.ID, Kind: model.Deadline, StreamA: s.ID, Detail: "deadline precedes release"})
		}
	}
	return out
}

func HasBlockingViolations(r model.ValidationResult) bool {
	for _, v := range r.Violations {
		if !v.Resolved {
			return true
		}
	}
	return false
}
