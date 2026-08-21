package validator

import (
	"testing"
	"example.com/task149/tsnsched/internal/model"
)

func TestBug09GuardBandParticipatesInConflict(t *testing.T) {r:=New(100).Check([]model.Allocation{{ID:"a",DraftID:"d",StreamID:"a",PortID:"p",StartNS:10,EndNS:20,GuardAfterNS:5},{ID:"b",DraftID:"d",StreamID:"b",PortID:"p",StartNS:24,EndNS:30}},nil,nil);if r.Valid{t.Fatal("guard band collision was accepted")} }
