package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type priorityRecomputeBarrierReader struct {
	UsageLogRepository
	mu          sync.Mutex
	calls       int
	entered     chan int
	release     []chan struct{}
	releaseOnce []sync.Once
	results     []PrioritySchedulingSignal
	errors      []error
}

func newPriorityRecomputeBarrierReader(results []PrioritySchedulingSignal, errs []error) *priorityRecomputeBarrierReader {
	release := make([]chan struct{}, len(results))
	for i := range release {
		release[i] = make(chan struct{})
	}
	return &priorityRecomputeBarrierReader{entered: make(chan int, len(results)), release: release, releaseOnce: make([]sync.Once, len(results)), results: results, errors: errs}
}

func (r *priorityRecomputeBarrierReader) releaseRead(index int) {
	r.releaseOnce[index].Do(func() { close(r.release[index]) })
}

func (r *priorityRecomputeBarrierReader) callCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls
}

func (r *priorityRecomputeBarrierReader) ReadPrioritySchedulingSignals(ctx context.Context, query PrioritySchedulingQuery) (map[int64]PrioritySchedulingSignal, error) {
	r.mu.Lock()
	index := r.calls
	r.calls++
	r.mu.Unlock()
	if index >= len(r.release) {
		return nil, fmt.Errorf("unexpected history reader call %d", index+1)
	}
	r.entered <- index + 1
	select {
	case <-r.release[index]:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	if index < len(r.errors) && r.errors[index] != nil {
		return nil, r.errors[index]
	}
	result := make(map[int64]PrioritySchedulingSignal, len(query.AccountIDs))
	for _, accountID := range query.AccountIDs {
		result[accountID] = r.results[index]
	}
	return result, nil
}

func priorityRecomputeEpochFixture(reader *priorityRecomputeBarrierReader) (*OpenAIGatewayService, PrioritySchedulingConfig, OpenAIAccountScheduleRequest, []openAIAccountCandidateScore, PrioritySchedulingSignal) {
	cfg := DefaultPrioritySchedulingConfig()
	cfg.MinSamples = 1
	cfg.MinQualityPercent = 90
	cfg.TargetTTFTMs = 1000
	request := OpenAIAccountScheduleRequest{RequestedModel: "recompute-epoch-model"}
	candidates := []openAIAccountCandidateScore{priorityCandidate(8871, 1, 1)}
	old := PrioritySchedulingSignal{Samples: 20, P90TTFTMs: 6000, QualityPassed: 2, QualitySamples: 10}
	now := time.Now()
	key := priorityHistoryKey(request, cfg)
	svc := &OpenAIGatewayService{usageLogRepo: reader}
	svc.priorityScheduling.entries = map[string]*prioritySignalEntry{
		key: {
			signals:   map[int64]PrioritySchedulingSignal{8871: old},
			accounts:  map[int64]time.Time{8871: now},
			observed:  now,
			windowEnd: now,
			expires:   now.Add(time.Hour),
		},
	}
	svc.priorityScheduling.latestInput = newPrioritySnapshotInput(request, cfg, candidates, now)
	return svc, cfg, request, candidates, old
}

func assertPriorityRiskEvidence(t *testing.T, snapshot *PrioritySchedulingSnapshot, candidate openAIAccountCandidateScore, expected PrioritySchedulingSignal) {
	t.Helper()
	require.NotNil(t, snapshot)
	require.Len(t, snapshot.Candidates, 1)
	score := snapshot.Candidates[0]
	require.Equal(t, candidate.account.ID, score.AccountID)
	require.Equal(t, expected.Samples, score.Samples, "online scheduling must retain its last successful latency evidence")
	require.Equal(t, expected.P90TTFTMs, score.P90TTFTMs)
	require.Equal(t, "degraded", score.Tier)
	require.Contains(t, score.Reasons, "latency_above_target")
	require.Contains(t, score.Reasons, "quality_below_target")
}

func TestPriorityRecomputeRefreshFailurePreservesCachedRiskForScheduler(t *testing.T) {
	reader := newPriorityRecomputeBarrierReader(
		[]PrioritySchedulingSignal{{}},
		[]error{errors.New("fixture reader failure")},
	)
	svc, _, _, candidates, old := priorityRecomputeEpochFixture(reader)
	type recomputeResponse struct {
		result *PrioritySchedulingRecomputeResult
		err    error
	}
	done := make(chan recomputeResponse, 1)
	go func() {
		result, err := svc.RecomputePriorityScheduling(context.Background())
		done <- recomputeResponse{result: result, err: err}
	}()

	select {
	case call := <-reader.entered:
		require.Equal(t, 1, call)
	case <-time.After(time.Second):
		t.Fatal("recompute did not start its history reader")
	}
	t.Cleanup(func() { reader.releaseRead(0) })

	// Ordinary scheduler reads and scoring must keep using the last known risky
	// evidence while the diagnostic refresh is blocked.
	during := svc.PrioritySchedulingSnapshot()
	assertPriorityRiskEvidence(t, during, candidates[0], old)

	reader.releaseRead(0)
	select {
	case response := <-done:
		require.Error(t, response.err, "a failed refresh must not produce a successful recompute")
		require.Nil(t, response.result)
	case <-time.After(2 * time.Second):
		t.Fatal("recompute did not report the reader failure")
	}

	after := svc.PrioritySchedulingSnapshot()
	assertPriorityRiskEvidence(t, after, candidates[0], old)
}

func TestPriorityRecomputeWaitsForFreshReadAfterInflightReader(t *testing.T) {
	reader := newPriorityRecomputeBarrierReader(
		[]PrioritySchedulingSignal{
			{Samples: 20, P90TTFTMs: 3000, QualityPassed: 5, QualitySamples: 10},
			{Samples: 20, P90TTFTMs: 250, QualityPassed: 10, QualitySamples: 10},
		},
		nil,
	)
	svc, cfg, request, candidates, _ := priorityRecomputeEpochFixture(reader)
	key := priorityHistoryKey(request, cfg)
	svc.priorityScheduling.entries[key].expires = time.Time{}
	_ = svc.priorityHistory(request, cfg, candidates)
	select {
	case call := <-reader.entered:
		require.Equal(t, 1, call, "selection history read should already be in flight")
	case <-time.After(time.Second):
		t.Fatal("initial history reader did not start")
	}
	t.Cleanup(func() {
		reader.releaseRead(0)
		reader.releaseRead(1)
	})

	type recomputeResponse struct {
		result *PrioritySchedulingRecomputeResult
		err    error
	}
	done := make(chan recomputeResponse, 1)
	go func() {
		result, err := svc.RecomputePriorityScheduling(context.Background())
		done <- recomputeResponse{result: result, err: err}
	}()

	// Both the c0c baseline and the repair increment this production counter
	// under the force-refresh lock. Observe that boundary before releasing the
	// old reader; goroutine launch alone does not establish that ordering.
	require.Eventually(t, func() bool {
		svc.priorityScheduling.mu.Lock()
		defer svc.priorityScheduling.mu.Unlock()
		return svc.priorityScheduling.entries[key].recomputeGeneration > 0
	}, time.Second, time.Millisecond, "recompute must reach its force-refresh boundary before the old reader completes")

	reader.releaseRead(0)
	select {
	case call := <-reader.entered:
		require.Equal(t, 2, call, "the in-flight pre-recompute read cannot satisfy the forced refresh")
	case response := <-done:
		t.Fatalf("recompute finished before a post-request history read: result=%+v err=%v", response.result, response.err)
	case <-time.After(2 * time.Second):
		t.Fatal("no fresh history reader was launched after the old read completed")
	}

	select {
	case response := <-done:
		t.Fatalf("recompute returned before its own reader completed: result=%+v err=%v", response.result, response.err)
	default:
	}
	reader.releaseRead(1)
	select {
	case response := <-done:
		require.NoError(t, response.err)
		require.NotNil(t, response.result)
		require.Equal(t, 2, reader.callCount())
		require.Equal(t, float64(250), response.result.After.Candidates[0].P90TTFTMs)
	case <-time.After(2 * time.Second):
		t.Fatal("recompute did not complete after its own reader completed")
	}
}
