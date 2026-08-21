package store

import (
	"context"
	"database/sql"
	"example.com/task149/tsnsched/internal/model"
)

func (s *Store) GetNode(ctx context.Context, id string) (model.Node, error) {
	var n model.Node
	var enabled int
	var created string
	err := s.db.QueryRowContext(ctx, `SELECT id,name,clock_domain,enabled,created_at FROM nodes WHERE id=?`, id).Scan(&n.ID, &n.Name, &n.ClockDomain, &enabled, &created)
	if err != nil {
		return n, requireRows(err)
	}
	n.Enabled = enabled != 0
	n.CreatedAt, _ = parse(created)
	return n, nil
}
func (s *Store) GetPort(ctx context.Context, id string) (model.Port, error) {
	var p model.Port
	var created string
	err := s.db.QueryRowContext(ctx, `SELECT id,node_id,name,direction,rate_bits,created_at FROM ports WHERE id=?`, id).Scan(&p.ID, &p.NodeID, &p.Name, &p.Direction, &p.RateBits, &created)
	if err != nil {
		return p, requireRows(err)
	}
	p.CreatedAt, _ = parse(created)
	return p, nil
}
func (s *Store) GetLink(ctx context.Context, id string) (model.Link, error) {
	var l model.Link
	var enabled int
	var created string
	err := s.db.QueryRowContext(ctx, `SELECT id,from_port,to_port,propagation_ns,period_ns,enabled,created_at FROM links WHERE id=?`, id).Scan(&l.ID, &l.FromPort, &l.ToPort, &l.PropagationNS, &l.PeriodNS, &enabled, &created)
	if err != nil {
		return l, requireRows(err)
	}
	l.Enabled = enabled != 0
	l.CreatedAt, _ = parse(created)
	return l, nil
}
func (s *Store) GetStream(ctx context.Context, id string) (model.Stream, error) {
	var v model.Stream
	var path, created string
	err := s.db.QueryRowContext(ctx, `SELECT id,name,period_ns,frame_bits,release_ns,deadline_ns,max_jitter_ns,path_json,created_at FROM streams WHERE id=?`, id).Scan(&v.ID, &v.Name, &v.PeriodNS, &v.FrameBits, &v.ReleaseNS, &v.DeadlineNS, &v.MaxJitterNS, &path, &created)
	if err != nil {
		return v, requireRows(err)
	}
	v.PathPorts, _ = readPath(path)
	v.CreatedAt, _ = parse(created)
	return v, nil
}
func (s *Store) Count(ctx context.Context, table string) (int, error) {
	allowed := map[string]bool{"nodes": true, "ports": true, "links": true, "streams": true, "drafts": true, "violations": true}
	if !allowed[table] {
		return 0, model.ErrInvalid
	}
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM `+table).Scan(&n)
	return n, err
}

func nullableString(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}
