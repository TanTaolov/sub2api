//go:build unit

package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type intelligenceRandomProxyAccountRepo struct {
	AccountRepository
	account *Account
}

func (r *intelligenceRandomProxyAccountRepo) GetByID(context.Context, int64) (*Account, error) {
	// 模拟数据库读取的独立账号对象，确保随机结果不会污染绑定代理。
	account := *r.account
	if account.Proxy != nil {
		proxy := *account.Proxy
		account.Proxy = &proxy
	}
	return &account, nil
}

func (r *intelligenceRandomProxyAccountRepo) UpdateExtra(context.Context, int64, map[string]any) error {
	return nil
}

type intelligenceRandomProxyRepo struct {
	ProxyRepository
	proxies []Proxy
	perCall [][]Proxy
	err     error
	calls   int
}

func (r *intelligenceRandomProxyRepo) ListActive(context.Context) ([]Proxy, error) {
	r.calls++
	if r.err != nil {
		return nil, r.err
	}
	if len(r.perCall) > 0 {
		proxies := r.perCall[0]
		r.perCall = r.perCall[1:]
		return proxies, nil
	}
	return r.proxies, nil
}

func intelligenceTestProxy(id int64) Proxy {
	return Proxy{ID: id, Protocol: "http", Host: fmt.Sprintf("proxy-%d.example", id), Port: 8080, Status: StatusActive}
}

func intelligenceRandomProxyAccount(platform string, enabled bool) *Account {
	account := stateProbeAccount()
	account.Platform = platform
	account.Status = StatusActive
	account.Schedulable = true
	account.Extra = map[string]any{AccountRandomProxyExtraKey: enabled}
	bound := intelligenceTestProxy(10)
	account.ProxyID = &bound.ID
	account.Proxy = &bound
	return account
}

func intelligenceRandomProxyService(account *Account, upstream HTTPUpstream, proxies *intelligenceRandomProxyRepo) *AccountTestService {
	snapshot := &SchedulerSnapshotService{}
	snapshot.SetRandomProxySource(proxies)
	return &AccountTestService{
		accountRepo:          &intelligenceRandomProxyAccountRepo{account: account},
		schedulerSnapshot:    snapshot,
		httpUpstream:         upstream,
		openaiGatewayService: &OpenAIGatewayService{httpUpstream: upstream},
		cfg:                  &config.Config{},
	}
}

const intelligenceOpenAIStream = `data: {"type":"response.output_text.delta","delta":"21"}

data: {"type":"response.completed","response":{"status":"completed"}}

`

const intelligenceClaudeStream = `data: {"type":"message_start","message":{"usage":{"input_tokens":1}}}

data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"21"}}

data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":1}}

data: {"type":"message_stop"}

`

func TestIntelligenceRandomProxyRespectsAccountModeAndFallback(t *testing.T) {
	selected := intelligenceTestProxy(20)
	bound := intelligenceTestProxy(10)
	disabled := intelligenceTestProxy(30)
	disabled.Status = StatusDisabled
	expired := intelligenceTestProxy(40)
	past := time.Now().Add(-time.Hour)
	expired.ExpiresAt = &past

	for _, platform := range []string{PlatformOpenAI, PlatformAnthropic} {
		t.Run(platform, func(t *testing.T) {
			for _, tc := range []struct {
				name           string
				enabled        bool
				proxies        []Proxy
				proxyErr       error
				noSnapshot     bool
				unbound        bool
				connectionOnly bool
				wantProxy      string
				wantLoads      int
			}{
				{name: "enabled", enabled: true, proxies: []Proxy{disabled, expired, selected}, wantProxy: selected.URL(), wantLoads: 1},
				{name: "disabled", proxies: []Proxy{selected}, wantProxy: bound.URL()},
				{name: "empty pool", enabled: true, wantProxy: bound.URL(), wantLoads: 1},
				{name: "no eligible proxy", enabled: true, proxies: []Proxy{disabled, expired}, wantProxy: bound.URL(), wantLoads: 1},
				{name: "pool unavailable", enabled: true, proxyErr: errors.New("proxy repository unavailable"), wantProxy: bound.URL(), wantLoads: 1},
				{name: "snapshot not injected", enabled: true, noSnapshot: true, proxies: []Proxy{selected}, wantProxy: bound.URL()},
				{name: "no binding", enabled: true, unbound: true, proxies: []Proxy{selected}},
				{name: "ordinary connection test", enabled: true, connectionOnly: true, proxies: []Proxy{selected}, wantProxy: bound.URL()},
			} {
				t.Run(tc.name, func(t *testing.T) {
					account := intelligenceRandomProxyAccount(platform, tc.enabled)
					if tc.unbound {
						account.ProxyID, account.Proxy = nil, nil
					}
					proxies := &intelligenceRandomProxyRepo{proxies: tc.proxies, err: tc.proxyErr}
					model, stream := "gpt-6-astra", intelligenceOpenAIStream
					if platform == PlatformAnthropic {
						model, stream = "claude-opus-5-5", intelligenceClaudeStream
					}
					upstream := &stateProbeUpstream{replies: []stateProbeReply{{status: http.StatusOK, body: stream}}}
					svc := intelligenceRandomProxyService(account, upstream, proxies)
					if tc.noSnapshot {
						svc.schedulerSnapshot = nil
					}
					c, recorder := newTestContext()
					var err error
					if tc.connectionOnly {
						err = svc.TestAccountConnection(c, account.ID, model, "", AccountTestModeDefault)
					} else {
						err = svc.TestPelicanAccountConnection(c, account.ID, model, "test question", "medium")
					}

					require.NoError(t, err)
					require.Contains(t, recorder.Body.String(), `"type":"test_complete","success":true`)
					require.Len(t, upstream.calls, 1)
					require.Equal(t, tc.wantProxy, upstream.calls[0].proxy)
					require.Equal(t, tc.wantLoads, proxies.calls)
					if !tc.unbound {
						require.Equal(t, bound.ID, *account.ProxyID)
						require.Equal(t, bound.URL(), account.Proxy.URL())
					}
				})
			}
		})
	}
}

func TestIntelligenceRandomProxyIsSelectedForEachSample(t *testing.T) {
	account := intelligenceRandomProxyAccount(PlatformOpenAI, true)
	first, second := intelligenceTestProxy(20), intelligenceTestProxy(21)
	proxies := &intelligenceRandomProxyRepo{perCall: [][]Proxy{{first}, {second}}}
	upstream := &stateProbeUpstream{replies: []stateProbeReply{
		{status: http.StatusOK, body: intelligenceOpenAIStream},
		{status: http.StatusOK, body: intelligenceOpenAIStream},
	}}
	svc := intelligenceRandomProxyService(account, upstream, proxies)
	// 禁用候选缓存，使两次选取各有唯一候选，不依赖概率断言。
	svc.schedulerSnapshot.randomProxies.ttl = 0

	for i := 0; i < 2; i++ {
		c, _ := newTestContext()
		require.NoError(t, svc.TestPelicanAccountConnection(c, account.ID, "gpt-6-astra", "test question", "medium"))
	}

	require.Len(t, upstream.calls, 2)
	require.Equal(t, first.URL(), upstream.calls[0].proxy)
	require.Equal(t, second.URL(), upstream.calls[1].proxy)
	require.Equal(t, 2, proxies.calls)
	require.Equal(t, int64(10), *account.ProxyID)
	require.Equal(t, "http://proxy-10.example:8080", account.Proxy.URL())
}

func TestIntelligenceRandomProxyScheduledQuestion(t *testing.T) {
	account := intelligenceRandomProxyAccount(PlatformOpenAI, true)
	selected := intelligenceTestProxy(20)
	proxies := &intelligenceRandomProxyRepo{proxies: []Proxy{selected}}
	upstream := &stateProbeUpstream{replies: []stateProbeReply{{status: http.StatusOK, body: intelligenceOpenAIStream}}}
	svc := intelligenceRandomProxyService(account, upstream, proxies)

	result, err := svc.RunPelicanBackground(context.Background(), account.ID, "gpt-6-astra", &PelicanTestConfig{
		QuestionKind: "candy", Prompt: CandyPrompt, ReasoningEffort: "medium", ParallelCount: 1,
	})

	require.NoError(t, err)
	require.Equal(t, "success", result.Status)
	require.Equal(t, "21", result.ResponseText)
	require.Len(t, upstream.calls, 1)
	require.Equal(t, selected.URL(), upstream.calls[0].proxy)
	require.Equal(t, int64(10), *account.ProxyID)
}

func TestIntelligenceRandomProxyStateProbeKeepsBothShotsOnOneProxy(t *testing.T) {
	for _, scheduled := range []bool{false, true} {
		for _, enabled := range []bool{false, true} {
			t.Run(fmt.Sprintf("scheduled=%t/enabled=%t", scheduled, enabled), func(t *testing.T) {
				account := intelligenceRandomProxyAccount(PlatformOpenAI, enabled)
				selected := intelligenceTestProxy(20)
				proxies := &intelligenceRandomProxyRepo{proxies: []Proxy{selected}}
				upstream := &stateProbeUpstream{replies: []stateProbeReply{
					stateProbeMint("ticket-1"),
					{status: http.StatusOK, body: stateProbeCompletedStream},
				}}
				svc := intelligenceRandomProxyService(account, upstream, proxies)
				svc.schedulerSnapshot.randomProxies.ttl = 0

				if scheduled {
					result, err := svc.RunPelicanBackground(context.Background(), account.ID, "gpt-6-astra", stateProbePlanConfig())
					require.NoError(t, err)
					require.Equal(t, "success", result.Status)
				} else {
					result, err := svc.ProbeOpenAICodexState(context.Background(), account.ID, "gpt-6-astra")
					require.NoError(t, err)
					require.Equal(t, OpenAICodexStateHealthy, result.Verdict)
				}

				wantProxy, wantLoads := account.Proxy.URL(), 0
				if enabled {
					wantProxy, wantLoads = selected.URL(), 1
				}
				require.Len(t, upstream.calls, 2)
				require.Equal(t, wantProxy, upstream.calls[0].proxy)
				require.Equal(t, wantProxy, upstream.calls[1].proxy)
				require.Equal(t, wantLoads, proxies.calls)
				require.Equal(t, int64(10), *account.ProxyID)
				require.Equal(t, "http://proxy-10.example:8080", account.Proxy.URL())
			})
		}
	}
}
