package topology

import "example.com/task149/tsnsched/internal/model"

func Build(nodes []model.Node, ports []model.Port, links []model.Link) *Graph {
	g := New()
	for _, n := range nodes {
		_ = g.AddNode(n)
	}
	for _, p := range ports {
		_ = g.AddPort(p)
	}
	for _, l := range links {
		_ = g.AddLink(l)
	}
	return g
}
func SameClock(g *Graph, path []string) bool {
	if len(path) == 0 {
		return true
	}
	first := g.Nodes[g.Ports[path[0]].NodeID].ClockDomain
	for _, id := range path {
		if g.Nodes[g.Ports[id].NodeID].ClockDomain != first {
			return false
		}
	}
	return true
}
func PathNodes(g *Graph, path []string) []string {
	out := make([]string, 0, len(path))
	for _, p := range path {
		out = append(out, g.Ports[p].NodeID)
	}
	return out
}
