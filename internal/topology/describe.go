package topology

import "example.com/task149/tsnsched/internal/model"

type Snapshot struct {
	Nodes   []model.Node   `json:"nodes"`
	Ports   []model.Port   `json:"ports"`
	Links   []model.Link   `json:"links"`
	Streams []model.Stream `json:"streams"`
}

func (g *Graph) Snapshot(streams []model.Stream) Snapshot {
	out := Snapshot{Nodes: make([]model.Node, 0, len(g.Nodes)), Ports: make([]model.Port, 0, len(g.Ports)), Links: make([]model.Link, 0, len(g.Links)), Streams: streams}
	for _, n := range g.Nodes {
		out.Nodes = append(out.Nodes, n)
	}
	for _, p := range g.Ports {
		out.Ports = append(out.Ports, p)
	}
	for _, l := range g.Links {
		out.Links = append(out.Links, l)
	}
	return out
}

func (g *Graph) ValidateNetwork() []string {
	issues := []string{}
	for id, l := range g.Links {
		if _, ok := g.Ports[l.FromPort]; !ok {
			issues = append(issues, "link "+id+" source missing")
		}
		if _, ok := g.Ports[l.ToPort]; !ok {
			issues = append(issues, "link "+id+" target missing")
		}
	}
	return issues
}
