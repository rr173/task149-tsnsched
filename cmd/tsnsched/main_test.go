package main

import "testing"

func TestFixtureBuilders(t *testing.T) {
	if modelNode("a", "b").ID != "a" || modelPort("p", "a").RateBits <= 0 || modelLink("l", "a", "b").ID != "l" {
		t.Fatal("fixture builder failed")
	}
}
