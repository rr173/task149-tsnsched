package service

import (
	"context"
	"example.com/task149/tsnsched/internal/model"
)

// CanCommit reports whether a draft is backed by its own persisted validation
// evidence for this exact version. The validation result must be bound to the
// draft id (never an empty or sentinel id) and carry a non-empty checked_at, so
// that a later commit reads only this version's evidence and is not misled by a
// shared "unbound" record left over from generation.
func (s *Service) CanCommit(ctx context.Context, id string) bool {
	if id == "" {
		return false
	}
	v, e := s.Data.GetValidation(ctx, id)
	if e != nil {
		return false
	}
	return v.DraftID == id && v.Valid && !v.CheckedAt.IsZero()
}
func (s *Service) CommitIfValid(ctx context.Context, network, id string) (model.ActiveVersion, error) {
	if !s.CanCommit(ctx, id) {
		return model.ActiveVersion{}, model.ErrNotReady
	}
	if e := s.Commit(ctx, network, id); e != nil {
		return model.ActiveVersion{}, e
	}
	return s.Active(ctx, network)
}
