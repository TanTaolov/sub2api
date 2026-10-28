package service

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

type codexPendingWrite struct {
	updates     map[string]any
	lastAttempt time.Time
	observedAt  time.Time
}

type codexWriteKey struct {
	repo AccountRepository
	id   int64
}

// Production repositories implement the shared, atomic observation CAS and
// return a cache publication error even if the database write already committed.
// In-memory/non-persistent adapters can retain the basic UpdateExtra contract.
type CodexUsageSnapshotRepository interface {
	UpdateCodexUsageSnapshot(context.Context, int64, map[string]any) (bool, error)
}

func (t *accountWriteThrottle) close() {
	t.mu.Lock()
	t.closed = true
	if t.cancel != nil {
		t.cancel()
	}
	t.mu.Unlock()
	t.workers.Wait()
}

// One retained latest update per account, at most one attempt per interval.
// Transient failures retry without another upstream response. Deleted accounts
// and shutdown terminate the worker and release all pending state.
func (t *accountWriteThrottle) submitCodex(repo AccountRepository, id int64, updates map[string]any, write func(context.Context, map[string]any) error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed || id <= 0 {
		return
	}
	if t.ctx == nil {
		t.ctx, t.cancel = context.WithCancel(context.Background())
	}
	if t.pending == nil {
		t.pending = make(map[codexWriteKey]*codexPendingWrite)
	}
	key := codexWriteKey{repo: repo, id: id}
	observedAt := codexUpdatesTime(updates)
	if pending := t.pending[key]; pending != nil {
		if !observedAt.Before(pending.observedAt) {
			if pending.updates == nil {
				pending.updates = make(map[string]any)
			}
			for k, v := range updates {
				pending.updates[k] = v
			}
			pending.observedAt = observedAt
		}
		return
	}
	pending := &codexPendingWrite{updates: updates, observedAt: observedAt}
	t.pending[key] = pending
	t.workers.Add(1)
	go func() {
		defer t.workers.Done()
		defer func() {
			t.mu.Lock()
			if t.pending[key] == pending {
				delete(t.pending, key)
			}
			t.mu.Unlock()
		}()
		for {
			t.mu.Lock()
			delay := time.Until(pending.lastAttempt.Add(t.minInterval))
			t.mu.Unlock()
			if delay > 0 {
				timer := time.NewTimer(delay)
				select {
				case <-t.ctx.Done():
					timer.Stop()
					return
				case <-timer.C:
				}
			}
			t.mu.Lock()
			if t.closed || len(pending.updates) == 0 {
				delete(t.pending, key)
				t.mu.Unlock()
				return
			}
			current := pending.updates
			pending.updates = nil
			pending.lastAttempt = time.Now()
			t.mu.Unlock()
			err := write(t.ctx, current)
			if t.ctx.Err() != nil {
				return
			}
			if errors.Is(err, ErrAccountNotFound) {
				slog.Info("openai_codex_snapshot_worker_stopped", "account_id", id, "reason", "account_deleted")
				return
			}
			if err != nil {
				// Do not emit opaque repository/upstream text: it may contain secrets.
				slog.Warn("openai_codex_snapshot_retry", "account_id", id, "error_code", "snapshot_persistence_failed")
				t.mu.Lock()
				for k, v := range pending.updates {
					current[k] = v
				}
				pending.updates = current
				t.mu.Unlock()
			}
		}
	}()
}

func codexUpdatesTime(updates map[string]any) time.Time {
	stamp, _ := updates["codex_usage_updated_at"].(string)
	parsed, _ := time.Parse(time.RFC3339Nano, stamp)
	return parsed
}

func persistCodexUsageUpdates(ctx context.Context, repo AccountRepository, id int64, updates map[string]any) (bool, error) {
	if writer, ok := repo.(CodexUsageSnapshotRepository); ok {
		return writer.UpdateCodexUsageSnapshot(ctx, id, updates)
	}
	if err := repo.UpdateExtra(ctx, id, updates); err != nil {
		return false, err
	}
	return true, nil
}
