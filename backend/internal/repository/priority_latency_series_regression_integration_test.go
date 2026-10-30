//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestPriorityLatencySeriesErrorsOnlyBucket(t *testing.T) {
	ctx := context.Background()
	tx := testEntTx(t)
	client := tx.Client()
	repo := newUsageLogRepositoryWithSQL(client, tx)
	account := mustCreateAccount(t, client, &service.Account{Name: "errors-only", Platform: service.PlatformOpenAI})
	groupID := int64(918273)
	createdAt := time.Now().UTC().Truncate(time.Second).Add(-time.Minute)
	requestID := fmt.Sprintf("latency-errors-only-%d", time.Now().UnixNano())
	_, err := tx.ExecContext(ctx, `
		INSERT INTO ops_error_logs(account_id,group_id,request_id,model,requested_model,request_path,error_phase,error_type,severity,status_code,created_at)
		VALUES($1,$2,$3,'fixture-model','fixture-model','/v1/responses','upstream','http_error','error',503,$4)
	`, account.ID, groupID, requestID, createdAt)
	require.NoError(t, err)
	group := groupID
	got, err := repo.ReadPriorityLatencySeries(ctx, service.PriorityLatencySeriesQuery{
		AccountIDs: []int64{account.ID}, Model: "fixture-model", GroupID: &group,
		Since: createdAt.Add(-time.Minute), BucketMinutes: 5, SlowTTFTMs: 1000,
	})
	require.NoError(t, err)
	require.Len(t, got, 1, "an upstream error without a usage row must still produce its bucket")
	require.Equal(t, account.ID, got[0].AccountID)
	require.Zero(t, got[0].Samples)
	require.Equal(t, 1, got[0].ErrorSamples)
}
