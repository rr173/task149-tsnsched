package store_test

import (
	"context"
	"path/filepath"
	"testing"
	"example.com/task149/tsnsched/internal/model"
	"example.com/task149/tsnsched/internal/service"
	"example.com/task149/tsnsched/internal/store"
)

func TestBug10ActiveVersionCannotMoveBackward(t *testing.T) { s,e:=store.Open(filepath.Join(t.TempDir(),"state.db"));if e!=nil{t.Fatal(e)};defer s.Close();ctx:=context.Background();if (model.ScheduleDraft{Status:model.Committed}).CanCommit(){t.Errorf("committed draft must not pass the lifecycle gate")};for _,d:=range []model.ScheduleDraft{{ID:"v1",NetworkID:"n",Version:1,Status:model.Validated,PeriodNS:100},{ID:"v2",NetworkID:"n",Version:2,Status:model.Validated,PeriodNS:100}}{if e=s.CreateDraft(ctx,d,nil);e!=nil{t.Fatal(e)};if e=s.SaveValidation(ctx,model.ValidationResult{DraftID:d.ID,Valid:true});e!=nil{t.Fatal(e)}};x:=service.New(s);if _,e=x.CommitIfValid(ctx,"n","v2");e!=nil{t.Errorf("newer version should commit: %v",e)};if _,e=x.CommitIfValid(ctx,"n","v1");e==nil{t.Errorf("older version became active")} }
