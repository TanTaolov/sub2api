package repository

import (
	"context"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

// ReadPriorityLatencySeries returns per-account latency buckets for the
// priority dashboard. The window is bounded by the caller; failed upstream
// attempts are counted separately and never treated as zero-latency successes.
func (r *usageLogRepository) ReadPriorityLatencySeries(ctx context.Context, query service.PriorityLatencySeriesQuery) ([]service.PriorityLatencyBucket, error) {
	slowTTFTMs := query.SlowTTFTMs
	if slowTTFTMs <= 0 {
		slowTTFTMs = 10000
	}
	bucketMinutes := query.BucketMinutes
	if bucketMinutes <= 0 {
		bucketMinutes = 5
	}
	rows, err := r.sql.QueryContext(ctx, `
 WITH buckets AS (
  SELECT account_id,
   date_bin(make_interval(mins => $5), created_at, TIMESTAMPTZ '2026-01-01') AS bucket,
   COUNT(*) FILTER (WHERE stream AND first_token_ms>0) AS samples,
   percentile_cont(0.5) WITHIN GROUP (ORDER BY first_token_ms) FILTER (WHERE stream AND first_token_ms>0) AS p50,
   percentile_cont(0.9) WITHIN GROUP (ORDER BY first_token_ms) FILTER (WHERE stream AND first_token_ms>0) AS p90,
   percentile_cont(0.99) WITHIN GROUP (ORDER BY first_token_ms) FILTER (WHERE stream AND first_token_ms>0) AS p99,
   COUNT(*) FILTER (WHERE stream AND first_token_ms>$6) AS slow_samples
  FROM usage_logs
  WHERE account_id=ANY($1) AND created_at >= $2 AND created_at <= NOW()
   AND COALESCE(NULLIF(requested_model,''),model)=$3
   AND group_id IS NOT DISTINCT FROM $4::bigint
  GROUP BY account_id, bucket
 ), errors AS (
  SELECT account_id,
   date_bin(make_interval(mins => $5), created_at, TIMESTAMPTZ '2026-01-01') AS bucket,
   COUNT(DISTINCT COALESCE(NULLIF(client_request_id,''), NULLIF(request_id,''), id::text)) AS error_samples
  FROM ops_error_logs
  WHERE account_id=ANY($1) AND created_at >= $2 AND created_at <= NOW()
   AND COALESCE(NULLIF(requested_model,''),model)=$3
   AND group_id IS NOT DISTINCT FROM $4::bigint
   AND error_phase='upstream'
  GROUP BY account_id, bucket
 )
 SELECT b.account_id, b.bucket, b.samples, COALESCE(b.p50,0), COALESCE(b.p90,0), COALESCE(b.p99,0), b.slow_samples, COALESCE(e.error_samples,0)
 FROM buckets b
 LEFT JOIN errors e ON e.account_id=b.account_id AND e.bucket=b.bucket
 ORDER BY b.account_id, b.bucket
 `, pq.Array(query.AccountIDs), query.Since, query.Model, query.GroupID, bucketMinutes, slowTTFTMs)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]service.PriorityLatencyBucket, 0, 128)
	for rows.Next() {
		var item service.PriorityLatencyBucket
		if err := rows.Scan(&item.AccountID, &item.Bucket, &item.Samples, &item.P50TTFTMs, &item.P90TTFTMs, &item.P99TTFTMs, &item.SlowSamples, &item.ErrorSamples); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
