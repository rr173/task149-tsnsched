package store

import (
	"context"
	"example.com/task149/tsnsched/internal/model"
	"path/filepath"
	"testing"
)

func TestDraftPersistsAcrossOpen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.db")
	s, e := Open(path)
	if e != nil {
		t.Fatal(e)
	}
	ctx := context.Background()
	d := model.ScheduleDraft{ID: "d", NetworkID: "n", Version: 1, PeriodNS: 100}
	if e = s.CreateDraft(ctx, d, nil); e != nil {
		t.Fatal(e)
	}
	_ = s.Close()
	s, e = Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	got, e := s.GetDraft(ctx, "d")
	if e != nil || got.ID != "d" {
		t.Fatalf("%+v %v", got, e)
	}
}
