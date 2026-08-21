package store

import (
	"context"
	"example.com/task149/tsnsched/internal/model"
	"time"
)

func (s *Store) PutNode(ctx context.Context, n model.Node) error {
	if err := n.Validate(); err != nil {
		return err
	}
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now().UTC()
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO nodes VALUES(?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET name=excluded.name,clock_domain=excluded.clock_domain,enabled=excluded.enabled`, n.ID, n.Name, n.ClockDomain, yes(n.Enabled), text(n.CreatedAt))
	return fmtErr("put node", err)
}
func (s *Store) PutPort(ctx context.Context, p model.Port) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now().UTC()
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO ports VALUES(?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET node_id=excluded.node_id,name=excluded.name,direction=excluded.direction,rate_bits=excluded.rate_bits`, p.ID, p.NodeID, p.Name, p.Direction, p.RateBits, text(p.CreatedAt))
	return fmtErr("put port", err)
}
func (s *Store) PutLink(ctx context.Context, l model.Link) error {
	if err := l.Validate(); err != nil {
		return err
	}
	if l.CreatedAt.IsZero() {
		l.CreatedAt = time.Now().UTC()
	}
	// Upsert so re-pointing an existing link id (changing from_port/to_port
	// and the other fields) actually persists. INSERT OR IGNORE would silently
	// keep the old endpoints once the id already exists, leaving the persisted
	// topology stale. created_at is intentionally not overwritten, mirroring
	// PutNode/PutPort.
	_, err := s.db.ExecContext(ctx, `INSERT INTO links VALUES(?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET from_port=excluded.from_port,to_port=excluded.to_port,propagation_ns=excluded.propagation_ns,period_ns=excluded.period_ns,enabled=excluded.enabled`, l.ID, l.FromPort, l.ToPort, l.PropagationNS, l.PeriodNS, yes(l.Enabled), text(l.CreatedAt))
	return fmtErr("put link", err)
}
func (s *Store) PutStream(ctx context.Context, v model.Stream) error {
	if err := v.Validate(); err != nil {
		return err
	}
	if v.CreatedAt.IsZero() {
		v.CreatedAt = time.Now().UTC()
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO streams VALUES(?,?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET name=excluded.name,period_ns=excluded.period_ns,frame_bits=excluded.frame_bits,release_ns=excluded.release_ns,deadline_ns=excluded.deadline_ns,max_jitter_ns=excluded.max_jitter_ns,path_json=excluded.path_json`, v.ID, v.Name, v.PeriodNS, v.FrameBits, v.ReleaseNS, v.DeadlineNS, v.MaxJitterNS, pathJSON(v.PathPorts), text(v.CreatedAt))
	return fmtErr("put stream", err)
}
func (s *Store) Nodes(ctx context.Context) ([]model.Node, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,name,clock_domain,enabled,created_at FROM nodes ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Node{}
	for rows.Next() {
		v, e := scanNode(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) Ports(ctx context.Context) ([]model.Port, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,node_id,name,direction,rate_bits,created_at FROM ports ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Port{}
	for rows.Next() {
		v, e := scanPort(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) Links(ctx context.Context) ([]model.Link, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,from_port,to_port,propagation_ns,period_ns,enabled,created_at FROM links ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Link{}
	for rows.Next() {
		v, e := scanLink(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) Streams(ctx context.Context) ([]model.Stream, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,name,period_ns,frame_bits,release_ns,deadline_ns,max_jitter_ns,path_json,created_at FROM streams ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Stream{}
	for rows.Next() {
		v, e := scanStream(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
