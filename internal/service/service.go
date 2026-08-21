package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"example.com/task149/tsnsched/internal/model"
	"example.com/task149/tsnsched/internal/scheduler"
	"example.com/task149/tsnsched/internal/store"
	"example.com/task149/tsnsched/internal/topology"
	"example.com/task149/tsnsched/internal/validator"
	"fmt"
	"sync"
	"time"
)

type Service struct {
	Data     *store.Store
	Graph    *topology.Graph
	mu       sync.Mutex
	PeriodNS int64
	Guard    scheduler.Config
}

func New(d *store.Store) *Service {
	return &Service{Data: d, PeriodNS: 1_000_000, Guard: scheduler.Config{NetworkPeriodNS: 1_000_000, GuardBeforeNS: 100, GuardAfterNS: 100}}
}
func (s *Service) Recover(ctx context.Context) error {
	g, e := s.Data.Graph(ctx)
	if e != nil {
		return e
	}
	s.mu.Lock()
	s.Graph = g
	s.mu.Unlock()
	return nil
}
func (s *Service) AddNode(ctx context.Context, n model.Node) error {
	if err := s.Data.PutNode(ctx, n); err != nil {
		return err
	}
	return s.Recover(ctx)
}
func (s *Service) AddPort(ctx context.Context, p model.Port) error {
	if err := s.Data.PutPort(ctx, p); err != nil {
		return err
	}
	return s.Recover(ctx)
}
func (s *Service) AddLink(ctx context.Context, l model.Link) error {
	if err := s.Data.PutLink(ctx, l); err != nil {
		return err
	}
	return s.Recover(ctx)
}
func (s *Service) AddStream(ctx context.Context, v model.Stream) error {
	return s.Data.PutStream(ctx, v)
}
func (s *Service) CreateDraft(ctx context.Context, network string, version int64) (model.ScheduleSummary, error) {
	s.mu.Lock()
	g := s.Graph
	s.mu.Unlock()
	if g == nil {
		if e := s.Recover(ctx); e != nil {
			return model.ScheduleSummary{}, e
		}
		g = s.Graph
	}
	streams, e := s.Data.Streams(ctx)
	if e != nil {
		return model.ScheduleSummary{}, e
	}
	id := newID("draft")
	p := scheduler.New(g, scheduler.Config{NetworkPeriodNS: s.PeriodNS, GuardBeforeNS: s.Guard.GuardBeforeNS, GuardAfterNS: s.Guard.GuardAfterNS})
	alloc, seed := p.Plan(streams, id)
	d := model.ScheduleDraft{ID: id, NetworkID: network, Version: version, Status: model.Draft, PeriodNS: s.PeriodNS}
	if e = s.Data.CreateDraft(ctx, d, alloc); e != nil {
		return model.ScheduleSummary{}, e
	}
	r := validator.New(s.PeriodNS).Check(alloc, streams, seed)
	// Bind the validation evidence to this exact draft version so it is
	// queryable by draft id and cannot be overwritten onto a shared sentinel.
	r.DraftID = id
	r.CheckedAt = time.Now().UTC()
	_ = s.Data.RecordAudit(ctx, "draft_validated", id, fmt.Sprintf("valid=%t violations=%d", r.Valid, len(r.Violations)))
	if e = s.Data.SaveValidation(ctx, r); e != nil {
		return model.ScheduleSummary{}, e
	}
	if r.Valid {
		_ = s.Data.SetDraftStatus(ctx, id, model.Validated)
	} else {
		_ = s.Data.SetDraftStatus(ctx, id, model.Rejected)
	}
	return s.Data.Summary(ctx, id)
}
func (s *Service) Validate(ctx context.Context, id string) (model.ValidationResult, error) {
	d, e := s.Data.GetDraft(ctx, id)
	if e != nil {
		return model.ValidationResult{}, e
	}
	a, e := s.Data.Allocations(ctx, id)
	if e != nil {
		return model.ValidationResult{}, e
	}
	streams, e := s.Data.Streams(ctx)
	if e != nil {
		return model.ValidationResult{}, e
	}
	r := validator.New(d.PeriodNS).Check(a, streams, nil)
	r.DraftID = id
	r.CheckedAt = time.Now().UTC()
	if e = s.Data.SaveValidation(ctx, r); e != nil {
		return r, e
	}
	if r.Valid {
		e = s.Data.SetDraftStatus(ctx, id, model.Validated)
	} else {
		e = s.Data.SetDraftStatus(ctx, id, model.Rejected)
	}
	return r, e
}
func (s *Service) Commit(ctx context.Context, network, id string) error {
	return s.Data.Commit(ctx, network, id)
}
func (s *Service) Rollback(ctx context.Context, network string) error {
	return s.Data.Rollback(ctx, network)
}
func (s *Service) Active(ctx context.Context, network string) (model.ActiveVersion, error) {
	return s.Data.Active(ctx, network)
}
func (s *Service) Stats(ctx context.Context) map[string]int { return s.Data.Stats(ctx) }
func (s *Service) GraphSnapshot() *topology.Graph           { s.mu.Lock(); defer s.mu.Unlock(); return s.Graph }
func newID(prefix string) string {
	b := make([]byte, 8)
	if _, e := rand.Read(b); e != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return prefix + "-" + hex.EncodeToString(b)
}
