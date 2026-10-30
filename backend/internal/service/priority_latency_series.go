package service

import (
	"context"
	"errors"
	"time"
)

// PriorityLatencySeriesQuery selects the accounts and window used by the admin
// latency panel. AccountIDs are limited to the last observed priority snapshot;
// this endpoint is a diagnostic view and never changes routing state.
type PriorityLatencySeriesQuery struct {
	AccountIDs    []int64
	Model         string
	GroupID       *int64
	Since         time.Time
	BucketMinutes int
	SlowTTFTMs    int
}

type PriorityLatencyBucket struct {
	AccountID    int64     `json:"account_id"`
	Bucket       time.Time `json:"bucket"`
	Samples      int       `json:"samples"`
	P50TTFTMs    float64   `json:"p50_ttft_ms"`
	P90TTFTMs    float64   `json:"p90_ttft_ms"`
	P99TTFTMs    float64   `json:"p99_ttft_ms"`
	SlowSamples  int       `json:"slow_samples"`
	ErrorSamples int       `json:"error_samples"`
}

type PriorityLatencySeriesReader interface {
	ReadPriorityLatencySeries(context.Context, PriorityLatencySeriesQuery) ([]PriorityLatencyBucket, error)
}

type PriorityLatencyAccountSeries struct {
	AccountID    int64                   `json:"account_id"`
	AccountName  string                  `json:"account_name"`
	Samples      int                     `json:"samples"`
	P50TTFTMs    float64                 `json:"p50_ttft_ms"`
	P90TTFTMs    float64                 `json:"p90_ttft_ms"`
	P99TTFTMs    float64                 `json:"p99_ttft_ms"`
	SlowSamples  int                     `json:"slow_samples"`
	ErrorSamples int                     `json:"error_samples"`
	ErrorRate    float64                 `json:"error_rate"`
	Buckets      []PriorityLatencyBucket `json:"buckets"`
}

type PriorityLatencySeries struct {
	At            time.Time                      `json:"at"`
	Model         string                         `json:"model"`
	GroupID       *int64                         `json:"group_id"`
	BucketMinutes int                            `json:"bucket_minutes"`
	Accounts      []PriorityLatencyAccountSeries `json:"accounts"`
}

var ErrPriorityLatencySeriesUnavailable = errors.New("priority latency series unavailable")

// PriorityLatencySeries returns per-account latency buckets for the last
// observed priority candidate pool. The summary fields come from the same
// history window as the priority score; the buckets power the trend view.
func (s *OpenAIGatewayService) PriorityLatencySeries(ctx context.Context, minutes, bucketMinutes int) (*PriorityLatencySeries, error) {
	if s == nil || s.usageLogRepo == nil {
		return nil, ErrPriorityLatencySeriesUnavailable
	}
	reader, ok := s.usageLogRepo.(PriorityLatencySeriesReader)
	if !ok {
		return nil, ErrPriorityLatencySeriesUnavailable
	}
	snapshot := s.PrioritySchedulingSnapshot()
	if snapshot == nil || len(snapshot.Candidates) == 0 {
		return nil, ErrPriorityLatencySeriesUnavailable
	}
	ids := make([]int64, 0, len(snapshot.Candidates))
	summary := make(map[int64]PrioritySchedulingScore, len(snapshot.Candidates))
	for _, candidate := range snapshot.Candidates {
		ids = append(ids, candidate.AccountID)
		summary[candidate.AccountID] = candidate
	}
	cfg := s.prioritySchedulingRuntimeConfig()
	var group *int64
	if snapshot.GroupID != nil {
		value := *snapshot.GroupID
		group = &value
	}
	rows, err := reader.ReadPriorityLatencySeries(ctx, PriorityLatencySeriesQuery{
		AccountIDs:    ids,
		Model:         snapshot.Model,
		GroupID:       group,
		Since:         time.Now().Add(-time.Duration(minutes) * time.Minute),
		BucketMinutes: bucketMinutes,
		SlowTTFTMs:    cfg.TargetTTFTMs,
	})
	if err != nil {
		return nil, err
	}
	byAccount := make(map[int64][]PriorityLatencyBucket, len(ids))
	for _, row := range rows {
		byAccount[row.AccountID] = append(byAccount[row.AccountID], row)
	}
	series := &PriorityLatencySeries{At: snapshot.At, Model: snapshot.Model, GroupID: group, BucketMinutes: bucketMinutes, Accounts: make([]PriorityLatencyAccountSeries, 0, len(ids))}
	for _, id := range ids {
		score := summary[id]
		account := PriorityLatencyAccountSeries{
			AccountID:    id,
			AccountName:  score.AccountName,
			Samples:      score.Samples,
			P50TTFTMs:    score.P50TTFTMs,
			P90TTFTMs:    score.P90TTFTMs,
			P99TTFTMs:    score.P99TTFTMs,
			SlowSamples:  score.SlowSamples,
			ErrorSamples: score.ErrorSamples,
			ErrorRate:    score.PrioritySchedulingSignal.UpstreamErrorRate,
			Buckets:      byAccount[id],
		}
		series.Accounts = append(series.Accounts, account)
	}
	return series, nil
}
