package scheduler

import (
	"fmt"
	"time"

	"example.com/task149/tsnsched/internal/cycle"
	"example.com/task149/tsnsched/internal/model"
	"example.com/task149/tsnsched/internal/topology"
)

type Config struct {
	GuardBeforeNS   int64
	GuardAfterNS    int64
	NetworkPeriodNS int64
}

type Planner struct {
	Graph  *topology.Graph
	Config Config
}

func New(g *topology.Graph, c Config) *Planner { return &Planner{Graph: g, Config: c} }

func (p *Planner) Plan(streams []model.Stream, draft string) ([]model.Allocation, []model.Violation) {
	alloc := []model.Allocation{}
	viol := []model.Violation{}
	for _, s := range streams {
		if err := p.Graph.ValidatePath(s.PathPorts); err != nil {
			viol = append(viol, model.Violation{ID: "path-" + s.ID, DraftID: draft, Kind: model.PathError, StreamA: s.ID, Detail: err.Error(), CreatedAt: time.Now().UTC()})
			continue
		}
		links := p.Graph.LinksForPath(s.PathPorts)
		cursor := s.ReleaseNS
		var first, final int64
		for i, l := range links {
			rate := p.Graph.PortRate(s.PathPorts[i])
			if rate <= 0 {
				viol = append(viol, model.Violation{ID: "rate-" + s.ID, DraftID: draft, Kind: model.PathError, StreamA: s.ID, PortID: s.PathPorts[i], Detail: "invalid port rate", CreatedAt: time.Now().UTC()})
				continue
			}
			duration, err := FrameDuration(s.FrameBits, rate)
			if err != nil {
				viol = append(viol, model.Violation{ID: "duration-" + s.ID, DraftID: draft, Kind: model.PathError, StreamA: s.ID, Detail: err.Error(), CreatedAt: time.Now().UTC()})
				continue
			}
			start := cycle.Normalize(cursor, p.Config.NetworkPeriodNS)
			if i == 0 {
				first = cursor
			}
			end := start + duration
			a := model.Allocation{ID: fmt.Sprintf("%s-%s-%d", s.ID, draft, i), DraftID: draft, StreamID: s.ID, LinkID: l.ID, PortID: s.PathPorts[i], StartNS: start, EndNS: end, GuardBeforeNS: p.Config.GuardBeforeNS, GuardAfterNS: p.Config.GuardAfterNS, ArrivalNS: cursor, DepartureNS: cursor + duration}
			alloc = append(alloc, a)
			cursor += duration + l.PropagationNS
			if i == len(links)-1 {
				final = cursor
			}
		}
		if final-first > s.DeadlineNS {
			viol = append(viol, model.Violation{ID: "deadline-" + s.ID, DraftID: draft, Kind: model.Deadline, StreamA: s.ID, Detail: fmt.Sprintf("path latency %d exceeds %d", final-first, s.DeadlineNS), CreatedAt: time.Now().UTC()})
		}
		if cycle.Distance(first, final, p.Config.NetworkPeriodNS) > s.MaxJitterNS && s.MaxJitterNS > 0 {
			viol = append(viol, model.Violation{ID: "jitter-path-" + s.ID, DraftID: draft, Kind: model.Jitter, StreamA: s.ID, Detail: fmt.Sprintf("path jitter %d exceeds %d", cycle.Distance(first, final, p.Config.NetworkPeriodNS), s.MaxJitterNS), CreatedAt: time.Now().UTC()})
		}
	}
	return alloc, viol
}
