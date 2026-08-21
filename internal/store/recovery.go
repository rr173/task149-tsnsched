package store

import (
	"context"
	"example.com/task149/tsnsched/internal/model"
	"example.com/task149/tsnsched/internal/topology"
)

// Graph reconstructs the in-memory topology index from the persisted nodes,
// ports and links. It is the database recovery read used by state rebuild
// (service.Recover) and the reload entrypoint (service.Reload). It must return
// the freshly built graph so persisted topology is rehydrated into memory;
// callers that mutate committed/active state are responsible for leaving the
// active_versions table untouched.
func (s *Store) Graph(ctx context.Context) (*topology.Graph, error) {
	n, e := s.Nodes(ctx)
	if e != nil {
		return nil, e
	}
	p, e := s.Ports(ctx)
	if e != nil {
		return nil, e
	}
	l, e := s.Links(ctx)
	if e != nil {
		return nil, e
	}
	return topology.Build(n, p, l), nil
}
func (s *Store) Summary(ctx context.Context, id string) (model.ScheduleSummary, error) {
	d, e := s.GetDraft(ctx, id)
	if e != nil {
		return model.ScheduleSummary{}, e
	}
	a, e := s.Allocations(ctx, id)
	if e != nil {
		return model.ScheduleSummary{}, e
	}
	v, e := s.GetValidation(ctx, id)
	return model.ScheduleSummary{Draft: d, Allocations: a, Validation: v}, e
}
func (s *Store) Stats(ctx context.Context) map[string]int {
	out := map[string]int{}
	for _, q := range []struct{ k, q string }{{"nodes", "SELECT COUNT(*) FROM nodes"}, {"ports", "SELECT COUNT(*) FROM ports"}, {"links", "SELECT COUNT(*) FROM links"}, {"streams", "SELECT COUNT(*) FROM streams"}, {"drafts", "SELECT COUNT(*) FROM drafts"}, {"violations", "SELECT COUNT(*) FROM violations"}} {
		var n int
		if s.db.QueryRowContext(ctx, q.q).Scan(&n) == nil {
			out[q.k] = n
		}
	}
	return out
}
