package store

import (
	"context"
	"time"
)

type AuditEvent struct {
	ID        int64     `json:"id"`
	Kind      string    `json:"kind"`
	EntityID  string    `json:"entity_id"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Store) EnsureAudit(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS audit_events(id INTEGER PRIMARY KEY AUTOINCREMENT,kind TEXT NOT NULL,entity_id TEXT NOT NULL,detail TEXT NOT NULL,created_at TEXT NOT NULL)`)
	return err
}
func (s *Store) RecordAudit(ctx context.Context, kind, entityID, detail string) error {
	if err := s.EnsureAudit(ctx); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO audit_events(kind,entity_id,detail,created_at) VALUES(?,?,?,?)`, kind, entityID, detail, text(time.Now().UTC()))
	return err
}
func (s *Store) Audit(ctx context.Context, limit int) ([]AuditEvent, error) {
	if err := s.EnsureAudit(ctx); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,kind,entity_id,detail,created_at FROM audit_events ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []AuditEvent{}
	for rows.Next() {
		var e AuditEvent
		var at string
		if err := rows.Scan(&e.ID, &e.Kind, &e.EntityID, &e.Detail, &at); err != nil {
			return nil, err
		}
		e.CreatedAt, _ = parse(at)
		out = append(out, e)
	}
	return out, rows.Err()
}
func (s *Store) DraftHistory(ctx context.Context, network string) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,version,status,updated_at FROM drafts WHERE network_id=? ORDER BY version`, network)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, st, at string
		var version int64
		if err := rows.Scan(&id, &version, &st, &at); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"id": id, "version": version, "status": st, "updated_at": at})
	}
	return out, rows.Err()
}
