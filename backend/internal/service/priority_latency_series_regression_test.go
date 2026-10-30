package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type priorityLatencySeriesRegressionReader struct {
	UsageLogRepository
	query PriorityLatencySeriesQuery
}

func (r *priorityLatencySeriesRegressionReader) ReadPriorityLatencySeries(_ context.Context, query PriorityLatencySeriesQuery) ([]PriorityLatencyBucket, error) {
	r.query = query
	return nil, nil
}

func TestPriorityLatencySeriesNoHistorySerializesEmptyBuckets(t *testing.T) {
	svc := &OpenAIGatewayService{usageLogRepo: &priorityLatencySeriesRegressionReader{}}
	cfg := DefaultPrioritySchedulingConfig()
	svc.priorityScheduling.latestInput = &prioritySnapshotInput{
		limited: true,
		request: OpenAIAccountScheduleRequest{RequestedModel: "fixture-model"},
		config:  cfg,
		candidates: []openAIAccountCandidateScore{
			{account: &Account{ID: 77, Name: "No history"}},
		},
	}

	got, err := svc.PriorityLatencySeries(context.Background(), 60, 5)
	require.NoError(t, err)
	require.Len(t, got.Accounts, 1)
	raw, err := json.Marshal(got)
	require.NoError(t, err)
	var response struct {
		Accounts []struct {
			Buckets json.RawMessage `json:"buckets"`
		} `json:"accounts"`
	}
	require.NoError(t, json.Unmarshal(raw, &response))
	require.JSONEq(t, `[]`, string(response.Accounts[0].Buckets), "the public API must represent an empty series as an array")
}

func TestPriorityLatencySeriesUsesSnapshotMeasurementAndPercentile(t *testing.T) {
	cfg := DefaultPrioritySchedulingConfig()
	cfg.LatencyPercentile = 50
	cfg.LatencyMeasurement = "attributed_v1"
	reader := &priorityLatencySeriesRegressionReader{}
	svc := &OpenAIGatewayService{usageLogRepo: reader}
	svc.priorityScheduling.latestInput = &prioritySnapshotInput{
		limited: true,
		request: OpenAIAccountScheduleRequest{RequestedModel: "fixture-model"},
		config:  cfg,
		candidates: []openAIAccountCandidateScore{
			{account: &Account{ID: 77, Name: "No history"}},
		},
	}

	_, err := svc.PriorityLatencySeries(context.Background(), 60, 5)
	require.NoError(t, err)
	raw, err := json.Marshal(reader.query)
	require.NoError(t, err)
	var query map[string]any
	require.NoError(t, json.Unmarshal(raw, &query))
	require.Equal(t, float64(50), query["LatencyPercentile"])
	require.Equal(t, "attributed_v1", query["LatencyMeasurement"])
}

func TestPriorityRecomputeDoesNotTreatHistoryErrorAsSuccess(t *testing.T) {
	cfg := DefaultPrioritySchedulingConfig()
	reader := &priorityReaderStub{err: context.DeadlineExceeded}
	svc := priorityGateway(cfg, reader)
	svc.priorityScheduling.latestInput = newPrioritySnapshotInput(
		OpenAIAccountScheduleRequest{RequestedModel: "fixture-model"}, cfg,
		[]openAIAccountCandidateScore{priorityCandidate(77, 1, 1)},
		time.Now(),
	)

	result, err := svc.RecomputePriorityScheduling(context.Background())
	require.Error(t, err, "a history error is not a successful recomputation")
	require.Nil(t, result)
}
