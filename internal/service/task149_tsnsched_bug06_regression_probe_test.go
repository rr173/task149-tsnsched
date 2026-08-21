package service

import (
	"context"
	"path/filepath"
	"testing"
	"example.com/task149/tsnsched/internal/model"
	"example.com/task149/tsnsched/internal/store"
)

func TestBug06ReloadPreservesActiveVersion(t *testing.T) { path:=filepath.Join(t.TempDir(),"state.db");s,e:=store.Open(path);if e!=nil{t.Fatal(e)};defer s.Close();x:=New(s);ctx:=context.Background();if e=x.AddNode(ctx,model.Node{ID:"node",Name:"node"});e!=nil{t.Fatal(e)};d:=model.ScheduleDraft{ID:"active",NetworkID:"n",Version:1,Status:model.Validated,PeriodNS:100};if e=s.CreateDraft(ctx,d,nil);e!=nil{t.Fatal(e)};if e=s.SaveValidation(ctx,model.ValidationResult{DraftID:"active",Valid:true});e!=nil{t.Fatal(e)};if e=x.Commit(ctx,"n","active");e!=nil{t.Fatal(e)};if e=x.Reload(ctx);e!=nil{t.Fatal(e)};a,e:=x.Active(ctx,"n");if e!=nil||a.DraftID!="active"{t.Errorf("reload lost active version: %+v %v",a,e)};if x.GraphSnapshot()==nil||len(x.GraphSnapshot().Nodes)!=1{t.Errorf("reload lost persisted topology: %#v",x.GraphSnapshot())} }
