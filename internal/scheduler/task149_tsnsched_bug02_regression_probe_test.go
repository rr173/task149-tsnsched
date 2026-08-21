package scheduler_test

import (
	"testing"
	"example.com/task149/tsnsched/internal/model"
	"example.com/task149/tsnsched/internal/scheduler"
	"example.com/task149/tsnsched/internal/validator"
)

func TestBug02CeilingFrameDuration(t *testing.T) { got,err:=scheduler.FrameDuration(3,2_000_000_000);if err!=nil||got!=2{t.Errorf("duration=%d err=%v, want ceil(1.5)=2",got,err)};if issues:=validator.CheckStreamSet([]model.Stream{{ID:"boundary",PeriodNS:100,FrameBits:3,ReleaseNS:100,DeadlineNS:120,MaxJitterNS:20,PathPorts:[]string{"a","b"}}});len(issues)==0{t.Errorf("release at the cycle boundary must be rejected")} }
