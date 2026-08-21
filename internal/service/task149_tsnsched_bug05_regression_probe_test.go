package service

import (
	"context"
	"path/filepath"
	"testing"
	"example.com/task149/tsnsched/internal/model"
	"example.com/task149/tsnsched/internal/store"
)

func TestBug05CommitSameDraftIsIdempotent(t *testing.T) { s,e:=store.Open(filepath.Join(t.TempDir(),"state.db"));if e!=nil{t.Fatal(e)};defer s.Close();x:=New(s);ctx:=context.Background();d:=model.ScheduleDraft{ID:"d",NetworkID:"n",Version:1,Status:model.Validated,PeriodNS:100};if e=s.CreateDraft(ctx,d,nil);e!=nil{t.Fatal(e)};if e=s.SaveValidation(ctx,model.ValidationResult{DraftID:"d",Valid:true});e!=nil{t.Fatal(e)};if e=x.Commit(ctx,"n","d");e!=nil{t.Fatal(e)};if x.CanCommit(ctx,"d"){t.Errorf("a committed draft must not re-enter the validation gate")};if e=x.Commit(ctx,"n","d");e!=nil{t.Errorf("second identical commit must be harmless: %v",e)} }
