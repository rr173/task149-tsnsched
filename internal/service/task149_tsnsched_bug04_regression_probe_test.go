package service

import (
	"context"
	"path/filepath"
	"testing"
	"example.com/task149/tsnsched/internal/store"
)

func TestBug04ValidationEvidenceBelongsToDraft(t *testing.T) { s,e:=store.Open(filepath.Join(t.TempDir(),"state.db"));if e!=nil{t.Fatal(e)};defer s.Close();x:=New(s);ctx:=context.Background();d,e:=x.CreateDraft(ctx,"network",1);if e!=nil{t.Fatal(e)};if d.Validation.DraftID!=d.Draft.ID||d.Validation.CheckedAt.IsZero(){t.Errorf("validation evidence was not persisted for draft %q: %+v",d.Draft.ID,d.Validation)};if _,e=x.CommitIfValid(ctx,"network",d.Draft.ID);e!=nil{t.Errorf("draft with its own validation evidence must be committable: %v",e)} }
