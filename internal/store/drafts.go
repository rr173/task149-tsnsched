package store

import (
	"context"
	"database/sql"
	"example.com/task149/tsnsched/internal/model"
	"time"
)

func (s *Store) CreateDraft(ctx context.Context, d model.ScheduleDraft, alloc []model.Allocation) error {
	return s.Tx(ctx, func(tx *sql.Tx) error {
		now := time.Now().UTC()
		d.CreatedAt = now
		d.UpdatedAt = now
		if d.Status == "" {
			d.Status = model.Draft
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO drafts VALUES(?,?,?,?,?,?,?)`, d.ID, d.NetworkID, d.Version, d.Status, d.PeriodNS, text(now), text(now)); err != nil {
			return err
		}
		for _, a := range alloc {
			if _, err := tx.ExecContext(ctx, `INSERT INTO allocations VALUES(?,?,?,?,?,?,?,?,?,?,?)`, a.ID, a.DraftID, a.StreamID, a.LinkID, a.PortID, a.StartNS, a.EndNS, a.GuardBeforeNS, a.GuardAfterNS, a.ArrivalNS, a.DepartureNS); err != nil {
				return err
			}
		}
		return nil
	})
}
func (s *Store) GetDraft(ctx context.Context, id string) (model.ScheduleDraft, error) {
	d := model.ScheduleDraft{}
	var st, at, up string
	err := s.db.QueryRowContext(ctx, `SELECT id,network_id,version,status,period_ns,created_at,updated_at FROM drafts WHERE id=?`, id).Scan(&d.ID, &d.NetworkID, &d.Version, &st, &d.PeriodNS, &at, &up)
	if err != nil {
		return d, requireRows(err)
	}
	d.Status = model.Status(st)
	d.CreatedAt, _ = parse(at)
	d.UpdatedAt, _ = parse(up)
	return d, nil
}
func (s *Store) Allocations(ctx context.Context, draftID string) ([]model.Allocation, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,draft_id,stream_id,link_id,port_id,start_ns,end_ns,guard_before_ns,guard_after_ns,arrival_ns,departure_ns FROM allocations WHERE draft_id=? ORDER BY start_ns,id`, draftID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Allocation{}
	for rows.Next() {
		var a model.Allocation
		if err := rows.Scan(&a.ID, &a.DraftID, &a.StreamID, &a.LinkID, &a.PortID, &a.StartNS, &a.EndNS, &a.GuardBeforeNS, &a.GuardAfterNS, &a.ArrivalNS, &a.DepartureNS); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
func (s *Store) SetDraftStatus(ctx context.Context, id string, status model.Status) error {
	_, err := s.db.ExecContext(ctx, `UPDATE drafts SET status=?,updated_at=? WHERE id=?`, status, text(time.Now().UTC()), id)
	return err
}
func (s *Store) Drafts(ctx context.Context) ([]model.ScheduleDraft, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,network_id,version,status,period_ns,created_at,updated_at FROM drafts WHERE status IN ('draft','validated') ORDER BY version DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.ScheduleDraft{}
	for rows.Next() {
		d, e := scanDraft(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, d)
	}
	return out, rows.Err()
}
