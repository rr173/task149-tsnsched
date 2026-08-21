package service

import (
	"context"
	"example.com/task149/tsnsched/internal/model"
	"example.com/task149/tsnsched/internal/scheduler"
	"example.com/task149/tsnsched/internal/store"
)

func (s *Service) Summary(ctx context.Context, id string) (model.ScheduleSummary, error) {
	return s.Data.Summary(ctx, id)
}
func (s *Service) Drafts(ctx context.Context) ([]model.ScheduleDraft, error) {
	return s.Data.Drafts(ctx)
}
func (s *Service) Window(ctx context.Context, id string, from, to int64) (model.Window, error) {
	a, e := s.Data.Allocations(ctx, id)
	if e != nil {
		return model.Window{}, e
	}
	out := []model.Allocation{}
	for _, v := range a {
		if scheduler.InWindow(v, from, to, s.PeriodNS) {
			out = append(out, v)
		}
	}
	return model.Window{FromNS: from, ToNS: to, Allocations: out}, nil
}
func (s *Service) Topology(ctx context.Context) (*store.Store, error) { return s.Data, nil }
func (s *Service) Capacity(ctx context.Context, id string) (scheduler.CapacityReport, error) {
	a, err := s.Data.Allocations(ctx, id)
	if err != nil {
		return scheduler.CapacityReport{}, err
	}
	return scheduler.Capacity(a, s.PeriodNS), nil
}
func (s *Service) Audit(ctx context.Context, limit int) ([]store.AuditEvent, error) {
	return s.Data.Audit(ctx, limit)
}
