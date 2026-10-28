package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (r *channelRepository) ReadUpstreamURLRemark(ctx context.Context, upstream string) (*service.UpstreamURLRemark, error) {
	if err := service.ValidateUpstreamURLRemarkIdentity(upstream); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	out := &service.UpstreamURLRemark{Upstream: upstream}
	err := r.db.QueryRowContext(ctx, `SELECT remark,revision FROM upstream_url_channel_remarks WHERE upstream=$1`, upstream).Scan(&out.Remark, &out.Revision)
	if err == sql.ErrNoRows {
		return out, nil
	}
	return out, err
}
func (r *channelRepository) SaveUpstreamURLRemark(ctx context.Context, m service.UpstreamURLRemarkMutation) (*service.UpstreamURLRemark, error) {
	if err := service.ValidateUpstreamURLRemarkMutation(&m); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	out := &service.UpstreamURLRemark{Upstream: m.Upstream}
	err := r.runInTx(ctx, func(tx *sql.Tx) error {
		now := time.Now().UTC()
		raw, err := operationsAccounts(ctx, tx, service.UpstreamOperationsFilter{From: now.Add(-31 * 24 * time.Hour), Until: now})
		if err != nil {
			return err
		}
		var accounts []operationsAccount
		if err = json.Unmarshal(raw, &accounts); err != nil {
			return err
		}
		found := false
		for _, a := range accounts {
			if a.Key == m.Upstream && a.Attribution == "current_config" {
				found = true
				break
			}
		}
		if !found {
			return service.ErrUpstreamURLRemarkValidation
		}
		if *m.Revision == 0 {
			err = tx.QueryRowContext(ctx, `INSERT INTO upstream_url_channel_remarks(upstream,remark,revision) VALUES($1,$2,1) ON CONFLICT(upstream) DO NOTHING RETURNING remark,revision`, m.Upstream, m.Remark).Scan(&out.Remark, &out.Revision)
		} else {
			err = tx.QueryRowContext(ctx, `UPDATE upstream_url_channel_remarks SET remark=$2,revision=revision+1,updated_at=CURRENT_TIMESTAMP WHERE upstream=$1 AND revision=$3 RETURNING remark,revision`, m.Upstream, m.Remark, *m.Revision).Scan(&out.Remark, &out.Revision)
		}
		if err == sql.ErrNoRows {
			return service.ErrUpstreamURLRemarkConflict
		}
		return err
	})
	return out, err
}

// Include only identities visible on this page, keeping the DTO bounded by the
// existing catalog, usage rows and account-detail limits. Unknown buckets cannot
// acquire notes. The metadata shares the financial snapshot's read transaction.
func readProfitURLRemarks(ctx context.Context, tx *sql.Tx, usage json.RawMessage, f service.UpstreamOperationsFilter) (map[string]service.UpstreamURLRemark, error) {
	var visible struct {
		Catalog  []struct{ Key string } `json:"catalog"`
		Rows     []struct{ Key string } `json:"rows"`
		Accounts []struct{ Key string } `json:"accounts"`
		Members  []struct{ Key string } `json:"members"`
	}
	if err := json.Unmarshal(usage, &visible); err != nil {
		return nil, err
	}
	out := map[string]service.UpstreamURLRemark{}
	add := func(key string) {
		if service.ValidateUpstreamURLRemarkIdentity(key) == nil {
			out[key] = service.UpstreamURLRemark{Upstream: key}
		}
	}
	for _, row := range visible.Catalog {
		add(row.Key)
	}
	if f.Dimension == "upstream" {
		for _, row := range visible.Rows {
			add(row.Key)
		}
	}
	for _, row := range visible.Accounts {
		add(row.Key)
	}
	for _, row := range visible.Members {
		add(row.Key)
	}
	keys := make([]string, 0, len(out))
	for key := range out {
		keys = append(keys, key)
	}
	raw, err := json.Marshal(keys)
	if err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT upstream,remark,revision FROM upstream_url_channel_remarks WHERE upstream IN (SELECT jsonb_array_elements_text($1::jsonb))`, string(raw))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var note service.UpstreamURLRemark
		if err = rows.Scan(&note.Upstream, &note.Remark, &note.Revision); err != nil {
			return nil, err
		}
		out[note.Upstream] = note
	}
	return out, rows.Err()
}

var _ service.UpstreamURLRemarkRepository = (*channelRepository)(nil)
