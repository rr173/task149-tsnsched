package main

import "example.com/task149/tsnsched/internal/model"

func modelNode(id, name string) model.Node {
	return model.Node{ID: id, Name: name, ClockDomain: "g1", Enabled: true}
}
func modelPort(id, node string) model.Port {
	return model.Port{ID: id, NodeID: node, Name: id, Direction: model.Egress, RateBits: 1_000_000_000}
}
func modelLink(id, from, to string) model.Link {
	return model.Link{ID: id, FromPort: from, ToPort: to, PropagationNS: 10, PeriodNS: 1_000_000, Enabled: true}
}
func modelStream(id string, path []string) model.Stream {
	return model.Stream{ID: id, Name: id, PeriodNS: 1_000_000, FrameBits: 100, ReleaseNS: 10, DeadlineNS: 10_000, MaxJitterNS: 20_000, PathPorts: path}
}
