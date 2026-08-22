package store

import (
	"context"
	"database/sql"
	"example.com/task149/tsnsched/internal/model"
	"time"
)

func (s *Store) SaveValidation(ctx context.Context, r model.ValidationResult) error {
	return s.Tx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM violations WHERE draft_id=?`, r.DraftID); err != nil {
			return err
		}
		for _, v := range r.Violations {
			if _, err := tx.ExecContext(ctx, `INSERT INTO violations VALUES(?,?,?,?,?,?,?,?,?)`, v.ID, v.DraftID, v.Kind, v.StreamA, v.StreamB, v.PortID, v.Detail, yes(v.Resolved), text(v.CreatedAt)); err != nil {
				return err
			}
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO validation_results VALUES(?,?,?) ON CONFLICT(draft_id) DO UPDATE SET valid=excluded.valid,checked_at=excluded.checked_at`, r.DraftID, yes(r.Valid), text(r.CheckedAt))
		return err
	})
}
func (s *Store) GetValidation(ctx context.Context, id string) (model.ValidationResult, error) {
	r := model.ValidationResult{DraftID: id}
	var valid int
	var at string
	err := s.db.QueryRowContext(ctx, `SELECT valid,checked_at FROM validation_results WHERE draft_id=?`, id).Scan(&valid, &at)
	if err != nil {
		if err == sql.ErrNoRows {
			return r, nil
		}
		return r, err
	}
	r.Valid = valid != 0
	r.CheckedAt, _ = parse(at)
	r.Violations, _ = s.Violations(ctx, id)
	return r, nil
}
func (s *Store) Violations(ctx context.Context, id string) ([]model.Violation, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,draft_id,kind,stream_a,stream_b,port_id,detail,resolved,created_at FROM violations WHERE draft_id=? ORDER BY id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Violation{}
	for rows.Next() {
		var v model.Violation
		var res int
		var at string
		if err := rows.Scan(&v.ID, &v.DraftID, &v.Kind, &v.StreamA, &v.StreamB, &v.PortID, &v.Detail, &res, &at); err != nil {
			return nil, err
		}
		v.Resolved = res != 0
		v.CreatedAt, _ = parse(at)
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) Commit(ctx context.Context, network, id string) error {
	return s.Tx(ctx, func(tx *sql.Tx) error {
		var d model.ScheduleDraft
		var st, created, updated string
		if err := tx.QueryRowContext(ctx, `SELECT id,network_id,version,status,period_ns,created_at,updated_at FROM drafts WHERE id=?`, id).Scan(&d.ID, &d.NetworkID, &d.Version, &st, &d.PeriodNS, &created, &updated); err != nil {
			return requireRows(err)
		}
		d.Status = model.Status(st)
		d.CreatedAt, _ = parse(created)
		d.UpdatedAt, _ = parse(updated)
		if d.NetworkID != network || st == string(model.Rejected) {
			return model.ErrConflict
		}
		var valid int
		if err := tx.QueryRowContext(ctx, `SELECT valid FROM validation_results WHERE draft_id=?`, id).Scan(&valid); err != nil || valid == 0 {
			return model.ErrNotReady
		}
		// Inspect the active pointer inside the same transaction so the gate is
		// authoritative: a rejected draft, a stale version, an idempotent
		// re-commit, and a version swap are all decided here rather than split
		// across a separate read in the service layer.
		var current string
		var currentVersion int64
		err := tx.QueryRowContext(ctx, `SELECT draft_id,version FROM active_versions WHERE network_id=?`, network).Scan(&current, &currentVersion)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		now := text(time.Now().UTC())
		switch {
		case err != sql.ErrNoRows && current == id:
			// Idempotent re-commit of the version that is already active.
			// Reuse the existing active pointer; never INSERT again (that
			// would trip the network_id PRIMARY KEY unique constraint). A
			// committed draft stays committed.
			if _, e := tx.ExecContext(ctx, `UPDATE active_versions SET updated_at=? WHERE network_id=?`, now, network); e != nil {
				return e
			}
			if st != string(model.Committed) {
				if _, e := tx.ExecContext(ctx, `UPDATE drafts SET status='committed',updated_at=? WHERE id=?`, now, id); e != nil {
					return e
				}
			}
			return nil
		case err != sql.ErrNoRows && currentVersion >= d.Version:
			// A different draft is active and its version is not older than
			// this one: the active-version swap rule forbids regressing, so
			// reject the commit rather than overwriting the pointer.
			return model.ErrConflict
		default:
			// Either no active version exists yet, or the active version is
			// strictly older and must be replaced. UPSERT the pointer onto the
			// network row instead of an unconditional INSERT+UPDATE, so the
			// network_id PRIMARY KEY is safely reused in both cases.
			if _, e := tx.ExecContext(ctx, `INSERT INTO active_versions(network_id,draft_id,version,updated_at) VALUES(?,?,?,?) ON CONFLICT(network_id) DO UPDATE SET draft_id=excluded.draft_id,version=excluded.version,updated_at=excluded.updated_at`, network, id, d.Version, now); e != nil {
				return e
			}
			if st != string(model.Committed) {
				if _, e := tx.ExecContext(ctx, `UPDATE drafts SET status='committed',updated_at=? WHERE id=?`, now, id); e != nil {
					return e
				}
			}
			return nil
		}
	})
}
func (s *Store) Active(ctx context.Context, network string) (model.ActiveVersion, error) {
	var a model.ActiveVersion
	var at string
	err := s.db.QueryRowContext(ctx, `SELECT network_id,draft_id,version,updated_at FROM active_versions WHERE network_id=?`, network).Scan(&a.NetworkID, &a.DraftID, &a.Version, &at)
	if err != nil {
		return a, requireRows(err)
	}
	a.UpdatedAt, _ = parse(at)
	return a, nil
}
func (s *Store) Rollback(ctx context.Context, network string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM active_versions WHERE network_id=?`, network)
	return err
}
