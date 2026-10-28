//go:build integration

package repository_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/Wei-Shaw/sub2api/ent/runtime"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler/admin"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestProfitURLRemarksHTTPPostgresPersistenceIdentityAuditAndConcurrency(t *testing.T) {
	ctx := context.Background()
	db, client := repository.ProfitURLRemarksTestDatabase(t)
	userRepo := repository.NewUserRepository(client, db)
	cfg := &config.Config{JWT: config.JWTConfig{Secret: "profit-url-remarks-isolated-fixture", ExpireHour: 1}}
	auth := service.NewAuthService(client, userRepo, nil, nil, cfg, nil, nil, nil, nil, nil, nil, nil, nil)
	var adminID, userID int64
	require.NoError(t, db.QueryRow(`INSERT INTO users(email,password_hash,role,status) VALUES('remarks-admin@test.invalid','fixture','admin','active') RETURNING id`).Scan(&adminID))
	require.NoError(t, db.QueryRow(`INSERT INTO users(email,password_hash,role,status) VALUES('remarks-user@test.invalid','fixture','user','active') RETURNING id`).Scan(&userID))
	adminToken, err := auth.GenerateToken(ctx, &service.User{ID: adminID, Email: "remarks-admin@test.invalid", Role: service.RoleAdmin})
	require.NoError(t, err)
	userToken, err := auth.GenerateToken(ctx, &service.User{ID: userID, Email: "remarks-user@test.invalid", Role: service.RoleUser})
	require.NoError(t, err)
	upstream := "https://remarks.example:8443/Tenant/v1"
	urls := []string{"HTTPS://REMARKS.EXAMPLE:8443/Tenant/v1/", upstream, "https://remarks.example:9443/Tenant/v1", "https://remarks.example:8443/Other/v1", "https://remarks.example:8443/tenant/v1", "https://remarks.example:8443/Tenant/%2F", "https://name:secret@remarks.example/v1?api_key=query-secret"}
	ids := make([]int64, len(urls))
	for i, raw := range urls {
		credentials, _ := json.Marshal(map[string]string{"base_url": raw, "api_key": "credential-canary"})
		require.NoError(t, db.QueryRow(`INSERT INTO accounts(name,platform,type,credentials,extra) VALUES('duplicate display name','openai','apikey',$1::jsonb,'{"cost_multiplier":0.25}') RETURNING id`, string(credentials)).Scan(&ids[i]))
	}
	var key int64
	require.NoError(t, db.QueryRow(`INSERT INTO api_keys(user_id,key,name) VALUES($1,'fixture-remarks-key','fixture') RETURNING id`, userID).Scan(&key))
	now := time.Now().UTC().Truncate(time.Second)
	for i := 0; i < 2; i++ {
		_, err = db.Exec(`INSERT INTO usage_logs(user_id,api_key_id,account_id,request_id,model,total_cost,actual_cost,account_stats_cost,account_rate_multiplier,created_at) VALUES($1,$2,$3,$4,'fixture',10,4,8,999,$5)`, userID, key, ids[i], fmt.Sprintf("remark-usage-%d", i), now.Add(-time.Minute))
		require.NoError(t, err)
	}
	// Snapshot full rows, including credentials, balances and pricing data. Only
	// the independent note table and existing admin audit may change on writes.
	immutable := func() string {
		var raw string
		require.NoError(t, db.QueryRow(`SELECT jsonb_build_object('accounts',(SELECT jsonb_agg(to_jsonb(a) ORDER BY id) FROM accounts a),'users',(SELECT jsonb_agg(to_jsonb(u) ORDER BY id) FROM users u),'usage',(SELECT jsonb_agg(to_jsonb(u) ORDER BY id) FROM usage_logs u),'keys',(SELECT jsonb_agg(to_jsonb(k) ORDER BY id) FROM api_keys k),'channels',(SELECT jsonb_agg(to_jsonb(c) ORDER BY id) FROM channels c),'groups',(SELECT jsonb_agg(to_jsonb(g) ORDER BY id) FROM groups g))::text`).Scan(&raw))
		return raw
	}
	before := immutable()
	audit := service.NewAuditLogService(repository.NewAuditLogRepository(db), nil)
	audit.StartWriter()
	t.Cleanup(audit.Stop)
	handler := func() *admin.ChannelHandler {
		return admin.NewChannelHandler(service.NewChannelService(repository.NewChannelRepository(db), nil, nil, nil, nil), nil, nil)
	}
	router := func() *gin.Engine {
		r := gin.New()
		g := r.Group("/api/v1/admin")
		g.Use(gin.HandlerFunc(middleware.NewAdminAuthMiddleware(auth, service.NewUserService(userRepo, nil, nil, nil), nil, nil)))
		g.Use(gin.HandlerFunc(middleware.NewAuditLogMiddleware(audit)))
		g.Use((&admin.AccountHandler{}).AuthorizeObserver)
		h := handler()
		g.GET("/profit-operations/snapshot", h.ProfitOperations)
		g.GET("/profit-operations/url-remark", h.ReadProfitURLRemark)
		g.PUT("/profit-operations/url-remark", h.SaveProfitURLRemark)
		return r
	}
	r := router()
	request := func(method, path, body, token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, "/api/v1/admin/profit-operations/"+path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}
	mutation := func(upstream, remark string, revision int64) string {
		raw, e := json.Marshal(map[string]any{"upstream": upstream, "remark": remark, "revision": revision})
		require.NoError(t, e)
		return string(raw)
	}
	read := func(upstream string) service.UpstreamURLRemark {
		w := request("GET", "url-remark?upstream="+url.QueryEscape(upstream), "", adminToken)
		require.Equal(t, 200, w.Code, w.Body.String())
		var body struct{ Data service.UpstreamURLRemark }
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		return body.Data
	}
	for _, method := range []string{"GET", "PUT"} {
		path := "url-remark?upstream=" + url.QueryEscape(upstream)
		require.Equal(t, 401, request(method, path, mutation(upstream, "unauthorized", 0), "").Code)
		require.Equal(t, 403, request(method, path, mutation(upstream, "unauthorized", 0), userToken).Code)
	}
	require.Equal(t, int64(0), read(upstream).Revision)
	w := request("PUT", "url-remark", mutation(upstream, "  共享渠道备注  ", 0), adminToken)
	require.Equal(t, 200, w.Code, w.Body.String())
	var persisted string
	var revision int64
	require.NoError(t, db.QueryRow(`SELECT remark,revision FROM upstream_url_channel_remarks WHERE upstream=$1`, upstream).Scan(&persisted, &revision))
	require.Equal(t, "共享渠道备注", persisted, "API must persist the submitted note, not just return optimistic success")
	require.Equal(t, int64(1), revision)
	require.Equal(t, service.UpstreamURLRemark{Upstream: upstream, Remark: persisted, Revision: 1}, read(upstream))
	// New handlers/repositories emulate a server/page refresh; no process cache.
	r = router()
	require.Equal(t, "共享渠道备注", read(upstream).Remark)
	for i := 0; i < 2; i++ {
		path := fmt.Sprintf("snapshot?from=%s&until=%s&account_id=%d", now.Add(-time.Hour).Format(time.RFC3339), now.Format(time.RFC3339), ids[i])
		w = request("GET", path, "", adminToken)
		require.Equal(t, 200, w.Code, w.Body.String())
		var body struct {
			Data service.ProfitOperationsSnapshot
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.Equal(t, "共享渠道备注", body.Data.URLRemarks[upstream].Remark, "same effective URL shares one note across accounts")
		require.NotContains(t, w.Body.String(), "credential-canary")
		var usage struct {
			Summary struct {
				Profit float64 `json:"estimated_profit"`
			}
		}
		require.NoError(t, json.Unmarshal(body.Data.Usage, &usage))
		require.Equal(t, 1.5, usage.Summary.Profit)
	}
	for _, other := range urls[2:6] {
		require.Empty(t, read(other).Remark, "meaningful port/path/case/escaped boundaries must remain separate")
	}
	w = request("PUT", "url-remark", mutation(urls[2], "独立端口", 0), adminToken)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Equal(t, "共享渠道备注", read(upstream).Remark)
	require.Equal(t, "独立端口", read(urls[2]).Remark)
	for _, bad := range []string{urls[0], urls[6], "unknown:invalid_url", "https://absent.example", "https://remarks.example?token=query-secret"} {
		w = request("PUT", "url-remark", mutation(bad, "invalid", 0), adminToken)
		require.Equal(t, 400, w.Code, w.Body.String())
		require.NotContains(t, w.Body.String(), "query-secret")
	}
	for _, body := range []string{`{}`, `{"upstream":"https://remarks.example:8443/Tenant/v1","revision":1}`, mutation(upstream, strings.Repeat("中", 501), 1)} {
		require.Equal(t, 400, request("PUT", "url-remark", body, adminToken).Code)
	}
	// Concurrent edits have one winner, including revision-zero creation races.
	race := func(key string, rev int64) {
		var wg sync.WaitGroup
		codes := make(chan int, 2)
		for i := 0; i < 2; i++ {
			body := mutation(key, fmt.Sprintf("并发-%d", i), rev)
			wg.Add(1)
			go func() { defer wg.Done(); codes <- request("PUT", "url-remark", body, adminToken).Code }()
		}
		wg.Wait()
		close(codes)
		got := []int{}
		for code := range codes {
			got = append(got, code)
		}
		require.ElementsMatch(t, []int{200, 409}, got)
		require.Equal(t, rev+1, read(key).Revision)
	}
	race(upstream, 1)
	race(urls[3], 0)
	w = request("PUT", "url-remark", mutation(upstream, " \n\t ", 2), adminToken)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Empty(t, read(upstream).Remark)
	require.Equal(t, int64(3), read(upstream).Revision)
	require.Equal(t, 409, request("PUT", "url-remark", mutation(upstream, "stale recreate", 0), adminToken).Code)
	require.Equal(t, 409, request("PUT", "url-remark", mutation(upstream, "stale edit", 2), adminToken).Code)
	require.Equal(t, before, immutable(), "analytics notes must not mutate accounts, routing, pricing or financial history")
	// Rename a member: note identity is independent from account names and pages.
	_, err = db.Exec(`UPDATE accounts SET name='renamed member' WHERE id=$1`, ids[0])
	require.NoError(t, err)
	require.Equal(t, "独立端口", read(urls[2]).Remark)
	audit.Stop()
	var count int
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM audit_logs WHERE actor_user_id=$1 AND action='admin.profit_url_remark.update' AND status_code=200`, adminID).Scan(&count))
	require.GreaterOrEqual(t, count, 5)
	var auditRaw string
	require.NoError(t, db.QueryRow(`SELECT COALESCE(jsonb_agg(to_jsonb(a))::text,'[]') FROM audit_logs a WHERE actor_user_id=$1`, adminID).Scan(&auditRaw))
	require.Contains(t, auditRaw, "config_version")
	for _, secret := range []string{"query-secret", "credential-canary", "name:secret", "共享渠道备注", "并发-"} {
		require.NotContains(t, auditRaw, secret, "request body must be omitted even on validation failure")
	}
	var invalidRows int
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM upstream_url_channel_remarks WHERE upstream LIKE '%?%' OR upstream LIKE '%@%' OR upstream LIKE 'unknown:%'`).Scan(&invalidRows))
	require.Zero(t, invalidRows)
}
