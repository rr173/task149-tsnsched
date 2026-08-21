package topology

import (
	"example.com/task149/tsnsched/internal/model"
	"fmt"
)

type Graph struct {
	Nodes map[string]model.Node
	Ports map[string]model.Port
	Links map[string]model.Link
	Out   map[string][]model.Link
}

func New() *Graph {
	return &Graph{Nodes: map[string]model.Node{}, Ports: map[string]model.Port{}, Links: map[string]model.Link{}, Out: map[string][]model.Link{}}
}
func (g *Graph) AddNode(n model.Node) error {
	if err := n.Validate(); err != nil {
		return err
	}
	g.Nodes[n.ID] = n
	return nil
}
func (g *Graph) AddPort(p model.Port) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if _, ok := g.Nodes[p.NodeID]; !ok {
		return fmt.Errorf("node %s missing", p.NodeID)
	}
	g.Ports[p.ID] = p
	return nil
}
func (g *Graph) AddLink(l model.Link) error {
	if err := l.Validate(); err != nil {
		return err
	}
	if _, exists := g.Links[l.ID]; exists { return nil }
	a, aok := g.Ports[l.FromPort]
	b, bok := g.Ports[l.ToPort]
	if !aok || !bok || a.NodeID == b.NodeID {
		return fmt.Errorf("link endpoints invalid")
	}
	g.Links[l.ID] = l
	g.Out[l.FromPort] = append(g.Out[l.FromPort], l)
	return nil
}
func (g *Graph) LinkBetween(from, to string) (model.Link, bool) {
	for _, l := range g.Out[from] {
		if l.ToPort == to && l.Enabled {
			return l, true
		}
	}
	return model.Link{}, false
}
func (g *Graph) ValidatePath(path []string) error {
	if len(path) < 2 {
		return fmt.Errorf("path too short")
	}
	seen := map[string]bool{}
	for _, p := range path {
		if seen[p] {
			return fmt.Errorf("repeated port %s", p)
		}
		seen[p] = true
		if _, ok := g.Ports[p]; !ok {
			return fmt.Errorf("port %s missing", p)
		}
	}
	for i := 0; i < len(path)-1; i++ {
		if _, ok := g.LinkBetween(path[i], path[i+1]); !ok {
			return fmt.Errorf("missing directed link %s->%s", path[i], path[i+1])
		}
	}
	return nil
}
func (g *Graph) LinksForPath(path []string) []model.Link {
	out := make([]model.Link, 0)
	for i := 0; i < len(path)-1; i++ {
		if l, ok := g.LinkBetween(path[i], path[i+1]); ok {
			out = append(out, l)
		}
	}
	return out
}
func (g *Graph) PortRate(id string) int64 {
	if p, ok := g.Ports[id]; ok {
		return p.RateBits
	}
	return 0
}
