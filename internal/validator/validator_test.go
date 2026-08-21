package validator

import (
	"example.com/task149/tsnsched/internal/model"
	"testing"
)

func TestConflictsAreInvalid(t *testing.T) {
	r := New(100).Check([]model.Allocation{{ID: "a", DraftID: "d", StreamID: "a", PortID: "p", StartNS: 98, EndNS: 104, GuardAfterNS: 2}, {ID: "b", DraftID: "d", StreamID: "b", PortID: "p", StartNS: 1, EndNS: 5}}, nil, nil)
	if r.Valid {
		t.Fatal("overlap accepted")
	}
}
