package service

import (
	"context"
	"sync"
	"time"
)

type codexPendingWrite struct {
	updates     map[string]any
	lastAttempt time.Time
	observedAt  time.Time
}

// Submit keeps one latest snapshot per account, including after a failed write.
// Only its worker writes that account. Both successes and failures consume the
// interval, so bursts and unavailable storage cannot turn into unbounded writes.
func (t *accountWriteThrottle) submitCodex(id int64, updates map[string]any, write func(map[string]any) error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.pending == nil {
		t.pending = make(map[int64]*codexPendingWrite)
	}
	observedAt := codexUpdatesTime(updates)
	if pending := t.pending[id]; pending != nil {
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
	t.pending[id] = pending
	go func() {
		for {
			t.mu.Lock()
			delay := time.Until(pending.lastAttempt.Add(t.minInterval))
			t.mu.Unlock()
			if delay > 0 {
				time.Sleep(delay)
			}
			t.mu.Lock()
			if len(pending.updates) == 0 {
				delete(t.pending, id)
				t.mu.Unlock()
				return
			}
			current := pending.updates
			pending.updates = nil
			pending.lastAttempt = time.Now()
			t.mu.Unlock()
			err := write(current)
			t.mu.Lock()
			if err != nil {
				// Preserve failed fields, but never replace a newer queued field.
				for k, v := range pending.updates {
					current[k] = v
				}
				pending.updates = current
			}
			t.mu.Unlock()
		}
	}()
}

// A quota query/probe may finish while an older header snapshot is waiting for
// its interval. Serialize these writers and reject older observations so that
// the delayed header cannot undo a successful explicit refresh.
type codexWriteKey struct {
	repo AccountRepository
	id   int64
}
type codexWriteOrder struct {
	mu         sync.Mutex
	observedAt time.Time
}

var codexWriteOrders sync.Map

func codexUpdatesTime(updates map[string]any) time.Time {
	stamp, _ := updates["codex_usage_updated_at"].(string)
	parsed, _ := time.Parse(time.RFC3339Nano, stamp)
	return parsed
}

func persistCodexUsageUpdates(ctx context.Context, repo AccountRepository, id int64, updates map[string]any) (bool, error) {
	key := codexWriteKey{repo: repo, id: id}
	value, _ := codexWriteOrders.LoadOrStore(key, &codexWriteOrder{})
	order := value.(*codexWriteOrder)
	order.mu.Lock()
	defer order.mu.Unlock()
	observedAt := codexUpdatesTime(updates)
	if observedAt.Before(order.observedAt) {
		return false, nil
	}
	if err := repo.UpdateExtra(ctx, id, updates); err != nil {
		return false, err
	}
	order.observedAt = observedAt
	return true, nil
}
