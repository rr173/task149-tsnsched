package service

import (
	"context"
	"path/filepath"
	"testing"
	"example.com/task149/tsnsched/internal/model"
	"example.com/task149/tsnsched/internal/store"
)

func TestBug07WindowIncludesWrappedAllocation(t *testing.T) {s,e:=store.Open(filepath.Join(t.TempDir(),"state.db"));if e!=nil{t.Fatal(e)};defer s.Close();x:=New(s);x.PeriodNS=100;d:=model.ScheduleDraft{ID:"d",NetworkID:"n",Version:1,PeriodNS:100};a:=model.Allocation{ID:"a",DraftID:"d",StreamID:"s",PortID:"p",StartNS:98,EndNS:104};if e=s.CreateDraft(context.Background(),d,[]model.Allocation{a});e!=nil{t.Fatal(e)};w,e:=x.Window(context.Background(),"d",0,5);if e!=nil{t.Fatal(e)};if len(w.Allocations)!=1{t.Fatalf("wrapped allocation missing from window: %#v",w)} }
