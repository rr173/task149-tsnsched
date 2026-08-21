package service

import (
	"context"
	"example.com/task149/tsnsched/internal/model"
)

func (s *Service) Health(ctx context.Context) map[string]any {
	result := map[string]any{"status": "ok", "period_ns": s.PeriodNS}
	if err := s.Data.Ping(ctx); err != nil {
		result["status"] = "degraded"
		result["error"] = err.Error()
	}
	return result
}

func (s *Service) Reload(ctx context.Context) error { return s.Recover(ctx) }

func (s *Service) Compare(ctx context.Context, left, right string) (map[string]any, error) {
	a, err := s.Data.Allocations(ctx, left)
	if err != nil {
		return nil, err
	}
	b, err := s.Data.Allocations(ctx, right)
	if err != nil {
		return nil, err
	}
	return map[string]any{"left": left, "right": right, "left_count": len(a), "right_count": len(b), "same_count": len(a) == len(b)}, nil
}

func (s *Service) Streams(ctx context.Context) ([]model.Stream, error) { return s.Data.Streams(ctx) }
