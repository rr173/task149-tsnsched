package scheduler

import (
	"testing"
	"example.com/task149/tsnsched/internal/model"
	"example.com/task149/tsnsched/internal/topology"
)

func TestBug08DeadlineUsesUnwrappedLatency(t *testing.T) { g:=topology.New();_ = g.AddNode(model.Node{ID:"a",Name:"a"});_ = g.AddNode(model.Node{ID:"b",Name:"b"});_ = g.AddPort(model.Port{ID:"pa",NodeID:"a",RateBits:1_000_000_000});_ = g.AddPort(model.Port{ID:"pb",NodeID:"b",RateBits:1_000_000_000});_ = g.AddLink(model.Link{ID:"l",FromPort:"pa",ToPort:"pb",PropagationNS:0,PeriodNS:100,Enabled:true});stream:=model.Stream{ID:"s",PeriodNS:100,FrameBits:10,ReleaseNS:99,DeadlineNS:5,MaxJitterNS:1000,PathPorts:[]string{"pa","pb"}};trace,e:=BuildTrace(stream,[]model.Link{{ID:"l",PropagationNS:0}},[]int64{1_000_000_000},100);if e!=nil||trace.CompletionNS!=109||TraceWithinDeadline(trace,stream.DeadlineNS){t.Errorf("trace lost unwrapped latency: %+v %v",trace,e)};p:=New(g,Config{NetworkPeriodNS:100});_,v:=p.Plan([]model.Stream{stream},"d");if len(v)==0{t.Errorf("cross-cycle latency exceeded deadline but was accepted")} }
