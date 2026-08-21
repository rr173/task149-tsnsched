package service

import (
	"context"
	"path/filepath"
	"testing"
	"example.com/task149/tsnsched/internal/model"
	"example.com/task149/tsnsched/internal/store"
)

func TestBug03PersistedLinkUpdateRefreshesPlanner(t *testing.T) {s,e:=store.Open(filepath.Join(t.TempDir(),"state.db"));if e!=nil{t.Fatal(e)};defer s.Close();x:=New(s);ctx:=context.Background();for _,n:=range []model.Node{{ID:"a",Name:"a"},{ID:"b",Name:"b"},{ID:"c",Name:"c"}}{if e=x.AddNode(ctx,n);e!=nil{t.Fatal(e)}};for _,p:=range []model.Port{{ID:"pa",NodeID:"a",RateBits:1_000_000_000},{ID:"pb",NodeID:"b",RateBits:1_000_000_000},{ID:"pc",NodeID:"c",RateBits:1_000_000_000}}{if e=x.AddPort(ctx,p);e!=nil{t.Fatal(e)}};if e=x.AddLink(ctx,model.Link{ID:"l",FromPort:"pa",ToPort:"pb",PeriodNS:1_000_000,Enabled:true});e!=nil{t.Fatal(e)};if e=x.AddStream(ctx,model.Stream{ID:"s",Name:"s",PeriodNS:1_000_000,FrameBits:10,DeadlineNS:1000,MaxJitterNS:10000,PathPorts:[]string{"pa","pc"}});e!=nil{t.Fatal(e)};if e=x.AddLink(ctx,model.Link{ID:"l",FromPort:"pa",ToPort:"pc",PeriodNS:1_000_000,Enabled:true});e!=nil{t.Fatal(e)};draft,e:=x.CreateDraft(ctx,"n",1);if e!=nil{t.Fatal(e)};if !draft.Validation.Valid{t.Fatalf("updated link was not used: %#v",draft.Validation.Violations)} }
