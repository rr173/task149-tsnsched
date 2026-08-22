package service

import (
	"context"
	"example.com/task149/tsnsched/internal/model"
)

func (s *Service) CanCommit(ctx context.Context, id string) bool {
	d, e := s.Data.GetDraft(ctx, id)
	if e != nil {
		return false
	}
	// Only Validated drafts may enter the commit gate. A finished draft
	// (Committed/RolledBack/Rejected) must not re-enter; the idempotent
	// re-commit of the *current* active version is the sole exception and is
	// enforced atomically inside Store.Commit against the active pointer, so
	// this gate stays intentionally narrow.
	return d.Status == model.Validated
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
