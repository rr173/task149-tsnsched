package store

import (
	"context"
	"encoding/json"
	"example.com/task149/tsnsched/internal/model"
)

func (s *Store) Export(ctx context.Context, draftID string) ([]byte, error) {
	summary, err := s.Summary(ctx, draftID)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(summary, "", "  ")
}

func (s *Store) ActiveOrEmpty(ctx context.Context, network string) (model.ActiveVersion, error) {
	a, err := s.Active(ctx, network)
	if err == model.ErrNotFound {
		return model.ActiveVersion{NetworkID: network}, nil
	}
	return a, err
}
