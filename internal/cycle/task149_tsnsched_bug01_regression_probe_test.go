package cycle_test

import (
	"testing"
	"example.com/task149/tsnsched/internal/cycle"
)

func TestBug01CrossCycleGuardCollision(t *testing.T) { a:=cycle.Expand(98,3,100,cycle.Guard{After:2});b:=cycle.Split(0,3,100);if !cycle.Overlap(a,b){t.Fatalf("wrapped guard interval %#v must collide with %#v",a,b)} }
