package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"example.com/task149/tsnsched/internal/model"
	"fmt"
	_ "modernc.org/sqlite"
	"time"
)

type Store struct{ db *sql.DB }

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if _, err = db.Exec("PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;" + schema); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}
func (s *Store) DB() *sql.DB                    { return s.db }
func (s *Store) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }
func (s *Store) Tx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	if err = fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
func text(t time.Time) string           { return t.UTC().Format(time.RFC3339Nano) }
func parse(v string) (time.Time, error) { return time.Parse(time.RFC3339Nano, v) }
func yes(v bool) int {
	if v {
		return 1
	}
	return 0
}
func pathJSON(p []string) string { b, _ := json.Marshal(p); return string(b) }
func readPath(v string) ([]string, error) {
	var p []string
	err := json.Unmarshal([]byte(v), &p)
	return p, err
}
func scanNode(s interface{ Scan(...any) error }) (model.Node, error) {
	var n model.Node
	var at string
	var en int
	err := s.Scan(&n.ID, &n.Name, &n.ClockDomain, &en, &at)
	n.Enabled = en != 0
	n.CreatedAt, _ = parse(at)
	return n, err
}
func scanPort(s interface{ Scan(...any) error }) (model.Port, error) {
	var p model.Port
	var at string
	err := s.Scan(&p.ID, &p.NodeID, &p.Name, &p.Direction, &p.RateBits, &at)
	p.CreatedAt, _ = parse(at)
	return p, err
}
func scanLink(s interface{ Scan(...any) error }) (model.Link, error) {
	var l model.Link
	var at string
	var en int
	err := s.Scan(&l.ID, &l.FromPort, &l.ToPort, &l.PropagationNS, &l.PeriodNS, &en, &at)
	l.Enabled = en != 0
	l.CreatedAt, _ = parse(at)
	return l, err
}
func scanStream(s interface{ Scan(...any) error }) (model.Stream, error) {
	var v model.Stream
	var p, at string
	err := s.Scan(&v.ID, &v.Name, &v.PeriodNS, &v.FrameBits, &v.ReleaseNS, &v.DeadlineNS, &v.MaxJitterNS, &p, &at)
	v.PathPorts, _ = readPath(p)
	v.CreatedAt, _ = parse(at)
	return v, err
}
func scanDraft(s interface{ Scan(...any) error }) (model.ScheduleDraft, error) {
	var d model.ScheduleDraft
	var st, at, up string
	err := s.Scan(&d.ID, &d.NetworkID, &d.Version, &st, &d.PeriodNS, &at, &up)
	d.Status = model.Status(st)
	d.CreatedAt, _ = parse(at)
	d.UpdatedAt, _ = parse(up)
	return d, err
}
func requireRows(err error) error {
	if err == sql.ErrNoRows {
		return model.ErrNotFound
	}
	return err
}
func fmtErr(op string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", op, err)
}
