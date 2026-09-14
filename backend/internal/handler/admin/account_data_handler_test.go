package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type dataResponse struct {
	Code int         `json:"code"`
	Data dataPayload `json:"data"`
}

type dataPayload struct {
	Type           string        `json:"type"`
	Version        int           `json:"version"`
	Proxies        []dataProxy   `json:"proxies"`
	Accounts       []dataAccount `json:"accounts"`
	SkippedShadows int           `json:"skipped_shadows"`
}

type dataProxy struct {
	ProxyKey string `json:"proxy_key"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Status   string `json:"status"`
}

type dataAccount struct {
	Name        string         `json:"name"`
	Platform    string         `json:"platform"`
	Type        string         `json:"type"`
	Credentials map[string]any `json:"credentials"`
	Extra       map[string]any `json:"extra"`
	ProxyKey    *string        `json:"proxy_key"`
	Concurrency int            `json:"concurrency"`
	Priority    int            `json:"priority"`
}

func setupAccountDataRouter() (*gin.Engine, *stubAdminService) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	adminSvc := newStubAdminService()

	h := NewAccountHandler(
		adminSvc,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	router.GET("/api/v1/admin/accounts/data", h.ExportData)
	router.POST("/api/v1/admin/accounts/data", h.ImportData)
	return router, adminSvc
}

func TestExportDataIncludesSecrets(t *testing.T) {
	router, adminSvc := setupAccountDataRouter()

	proxyID := int64(11)
	adminSvc.proxies = []service.Proxy{
		{
			ID:       proxyID,
			Name:     "proxy",
			Protocol: "http",
			Host:     "127.0.0.1",
			Port:     8080,
			Username: "user",
			Password: "pass",
			Status:   service.StatusActive,
		},
		{
			ID:       12,
			Name:     "orphan",
			Protocol: "https",
			Host:     "10.0.0.1",
			Port:     443,
			Username: "o",
			Password: "p",
			Status:   service.StatusActive,
		},
	}
	adminSvc.accounts = []service.Account{
		{
			ID:          21,
			Name:        "account",
			Platform:    service.PlatformOpenAI,
			Type:        service.AccountTypeOAuth,
			Credentials: map[string]any{"token": "secret"},
			Extra:       map[string]any{"note": "x"},
			ProxyID:     &proxyID,
			Concurrency: 3,
			Priority:    50,
			Status:      service.StatusDisabled,
		},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/accounts/data", nil)
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var resp dataResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.Empty(t, resp.Data.Type)
	require.Equal(t, 0, resp.Data.Version)
	require.Len(t, resp.Data.Proxies, 1)
	require.Equal(t, "pass", resp.Data.Proxies[0].Password)
	require.Len(t, resp.Data.Accounts, 1)
	require.Equal(t, "secret", resp.Data.Accounts[0].Credentials["token"])
}

func TestExportDataWithoutProxies(t *testing.T) {
	router, adminSvc := setupAccountDataRouter()

	proxyID := int64(11)
	adminSvc.proxies = []service.Proxy{
		{
			ID:       proxyID,
			Name:     "proxy",
			Protocol: "http",
			Host:     "127.0.0.1",
			Port:     8080,
			Username: "user",
			Password: "pass",
			Status:   service.StatusActive,
		},
	}
	adminSvc.accounts = []service.Account{
		{
			ID:          21,
			Name:        "account",
			Platform:    service.PlatformOpenAI,
			Type:        service.AccountTypeOAuth,
			Credentials: map[string]any{"token": "secret"},
			ProxyID:     &proxyID,
			Concurrency: 3,
			Priority:    50,
			Status:      service.StatusDisabled,
		},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/accounts/data?include_proxies=false", nil)
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var resp dataResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.Len(t, resp.Data.Proxies, 0)
	require.Len(t, resp.Data.Accounts, 1)
	require.Nil(t, resp.Data.Accounts[0].ProxyKey)
}

// TestExportDataExcludesSparkShadow 验证外审第5轮 P1/P2:导出时排除 spark 影子账号
// (影子无凭据、导入侧强制 credentials 非空,混入会产出无法还原的坏备份),并透出跳过计数。
func TestExportDataExcludesSparkShadow(t *testing.T) {
	router, adminSvc := setupAccountDataRouter()

	parentID := int64(21)
	adminSvc.accounts = []service.Account{
		{
			ID:          parentID,
			Name:        "mother",
			Platform:    service.PlatformOpenAI,
			Type:        service.AccountTypeOAuth,
			Credentials: map[string]any{"token": "secret"},
			Status:      service.StatusActive,
		},
		{
			ID:              22,
			Name:            "mother (Spark)",
			Platform:        service.PlatformOpenAI,
			Type:            service.AccountTypeOAuth,
			Credentials:     map[string]any{}, // 影子恒空凭据
			ParentAccountID: &parentID,        // 影子标记
			QuotaDimension:  service.QuotaDimensionSpark,
			Status:          service.StatusActive,
		},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/accounts/data?include_proxies=false", nil)
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var resp dataResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.Len(t, resp.Data.Accounts, 1, "影子应被排除,仅导出母账号")
	require.Equal(t, "mother", resp.Data.Accounts[0].Name)
	require.Equal(t, 1, resp.Data.SkippedShadows, "跳过的影子数量应透出")
}

func TestExportDataPassesAccountFiltersAndSort(t *testing.T) {
	router, adminSvc := setupAccountDataRouter()
	adminSvc.accounts = []service.Account{
		{ID: 1, Name: "acc-1", Status: service.StatusActive},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/admin/accounts/data?platform=openai&type=oauth&status=active&group=12&privacy_mode=blocked&search=keyword&sort_by=priority&sort_order=desc",
		nil,
	)
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	require.Equal(t, 1, adminSvc.lastListAccounts.calls)
	require.Equal(t, "openai", adminSvc.lastListAccounts.platform)
	require.Equal(t, "oauth", adminSvc.lastListAccounts.accountType)
	require.Equal(t, "active", adminSvc.lastListAccounts.status)
	require.Equal(t, int64(12), adminSvc.lastListAccounts.groupID)
	require.Equal(t, "blocked", adminSvc.lastListAccounts.privacyMode)
	require.Equal(t, "keyword", adminSvc.lastListAccounts.search)
	require.Equal(t, "priority", adminSvc.lastListAccounts.sortBy)
	require.Equal(t, "desc", adminSvc.lastListAccounts.sortOrder)
}

func TestExportDataSelectedIDsOverrideFilters(t *testing.T) {
	router, adminSvc := setupAccountDataRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/admin/accounts/data?ids=1,2&platform=openai&search=keyword&sort_by=priority&sort_order=desc",
		nil,
	)
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var resp dataResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.Len(t, resp.Data.Accounts, 2)
	require.Equal(t, 0, adminSvc.lastListAccounts.calls)
}

func TestImportDataReusesProxyAndSkipsDefaultGroup(t *testing.T) {
	router, adminSvc := setupAccountDataRouter()

	adminSvc.proxies = []service.Proxy{
		{
			ID:       1,
			Name:     "proxy",
			Protocol: "socks5",
			Host:     "1.2.3.4",
			Port:     1080,
			Username: "u",
			Password: "p",
			Status:   service.StatusActive,
		},
	}

	dataPayload := map[string]any{
		"data": map[string]any{
			"type":    dataType,
			"version": dataVersion,
			"proxies": []map[string]any{
				{
					"proxy_key": "socks5|1.2.3.4|1080|u|p",
					"name":      "proxy",
					"protocol":  "socks5",
					"host":      "1.2.3.4",
					"port":      1080,
					"username":  "u",
					"password":  "p",
					"status":    "active",
				},
			},
			"accounts": []map[string]any{
				{
					"name":        "acc",
					"platform":    service.PlatformOpenAI,
					"type":        service.AccountTypeOAuth,
					"credentials": map[string]any{"token": "x"},
					"proxy_key":   "socks5|1.2.3.4|1080|u|p",
					"concurrency": 3,
					"priority":    50,
				},
			},
		},
		"skip_default_group_bind": true,
	}

	body, _ := json.Marshal(dataPayload)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/data", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	require.Len(t, adminSvc.createdProxies, 0)
	require.Len(t, adminSvc.createdAccounts, 1)
	require.True(t, adminSvc.createdAccounts[0].SkipDefaultGroupBind)
}

func TestImportDataAutoBindProxy(t *testing.T) {
	past := time.Now().Add(-time.Hour)
	future := time.Now().Add(time.Hour)
	availableProxies := []service.Proxy{
		{ID: 11, Name: "z-last", Protocol: "http", Host: "127.0.0.1", Port: 8080, Status: service.StatusActive},
		{ID: 12, Name: "b-selected", Protocol: "http", Host: "127.0.0.2", Port: 8080, Status: service.StatusActive, ExpiresAt: &future},
		{ID: 13, Name: "a-expired", Protocol: "http", Host: "127.0.0.3", Port: 8080, Status: service.StatusActive, ExpiresAt: &past},
		{ID: 14, Name: "a-inactive", Protocol: "http", Host: "127.0.0.4", Port: 8080, Status: "inactive"},
	}
	explicitProxyKey := buildProxyKey("http", "127.0.0.1", 8080, "", "")
	emptyProxyKey := ""
	validPool := []any{
		map[string]any{"proxy_id": float64(41), "concurrency": float64(6)},
		map[string]any{"proxy_id": float64(42), "concurrency": float64(9)},
	}
	invalidPool := []any{
		map[string]any{"proxy_id": float64(0), "concurrency": float64(6)},
		map[string]any{"proxy_id": float64(42), "concurrency": float64(0)},
	}
	mixedPool := []any{invalidPool[0], validPool[0]}

	tests := []struct {
		name         string
		settingValue string
		platform     string
		proxyKey     *string
		pool         any
		proxies      []service.Proxy
		createErr    error
		wantProxyID  int64
		wantAutoBind bool
	}{
		{
			name:    "missing setting keeps import unchanged",
			proxies: availableProxies,
		},
		{
			name:         "disabled setting keeps import unchanged",
			settingValue: "false",
			proxies:      availableProxies,
		},
		{
			name:         "bind first active unexpired proxy by name before creation",
			settingValue: "true",
			proxies:      availableProxies,
			wantProxyID:  12,
			wantAutoBind: true,
		},
		{
			name:         "empty proxy key permits automatic binding",
			settingValue: "true",
			proxyKey:     &emptyProxyKey,
			proxies:      availableProxies,
			wantProxyID:  12,
			wantAutoBind: true,
		},
		{
			name:         "explicit proxy key preserves concurrency and extra",
			settingValue: "true",
			proxyKey:     &explicitProxyKey,
			proxies:      availableProxies,
			wantProxyID:  11,
		},
		{
			name:         "valid JSON proxy pool preserves all entries and concurrency",
			settingValue: "true",
			pool:         validPool,
			proxies:      availableProxies,
		},
		{
			name:         "explicit key and pool both remain unchanged",
			settingValue: "true",
			proxyKey:     &explicitProxyKey,
			pool:         validPool,
			proxies:      availableProxies,
			wantProxyID:  11,
		},
		{
			name:         "one valid pool entry preserves the original extra",
			settingValue: "true",
			pool:         mixedPool,
			proxies:      availableProxies,
		},
		{
			name:         "invalid pool entries permit automatic binding",
			settingValue: "true",
			pool:         invalidPool,
			proxies:      availableProxies,
			wantProxyID:  12,
			wantAutoBind: true,
		},
		{
			name:         "empty pool permits automatic binding",
			settingValue: "true",
			pool:         []any{},
			proxies:      availableProxies,
			wantProxyID:  12,
			wantAutoBind: true,
		},
		{
			name:         "no system proxies keeps import unchanged",
			settingValue: "true",
		},
		{
			name:         "only expired and inactive proxies keeps import unchanged",
			settingValue: "true",
			proxies:      availableProxies[2:],
		},
		{
			name:         "non OpenAI account keeps import unchanged",
			settingValue: "true",
			platform:     service.PlatformAnthropic,
			proxies:      availableProxies,
		},
		{
			name:         "failed creation does not count as automatic binding",
			settingValue: "true",
			proxies:      availableProxies,
			createErr:    errors.New("account creation failed"),
			wantProxyID:  12,
			wantAutoBind: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			adminSvc := newStubAdminService()
			adminSvc.proxies = test.proxies
			adminSvc.createAccountErr = test.createErr
			settingsRepo := &settingHandlerRepoStub{values: map[string]string{}}
			if test.settingValue != "" {
				settingsRepo.values[service.SettingKeyImportAutoBindProxy] = test.settingValue
			}
			handler := &AccountHandler{
				adminService:   adminSvc,
				settingService: service.NewSettingService(settingsRepo, nil),
			}
			platform := test.platform
			if platform == "" {
				platform = service.PlatformOpenAI
			}
			extra := map[string]any{
				"note":   "preserve this value",
				"custom": map[string]any{"enabled": true},
			}
			if test.pool != nil {
				extra[service.AccountProxyPoolExtraKey] = test.pool
			}
			originalExtra, err := json.Marshal(extra)
			require.NoError(t, err)
			request := DataImportRequest{Data: DataPayload{
				Proxies: []DataProxy{},
				Accounts: []DataAccount{{
					Name:        "imported-account",
					Platform:    platform,
					Type:        service.AccountTypeOAuth,
					Credentials: map[string]any{"access_token": "test-token"},
					Extra:       extra,
					ProxyKey:    test.proxyKey,
					Concurrency: 7,
				}},
			}}

			result, err := handler.importData(context.Background(), request)
			require.NoError(t, err)
			require.Len(t, adminSvc.createdAccounts, 1)
			input := adminSvc.createdAccounts[0]
			if test.wantProxyID == 0 {
				require.Nil(t, input.ProxyID)
			} else {
				require.NotNil(t, input.ProxyID)
				require.Equal(t, test.wantProxyID, *input.ProxyID)
			}
			if test.wantAutoBind {
				require.Equal(t, 30, input.Concurrency)
				createdInput := service.Account{Extra: input.Extra}
				require.Equal(t, []service.ProxyPoolEntry{{ProxyID: test.wantProxyID, Concurrency: 30}}, createdInput.ProxyPool())
				require.Equal(t, extra["note"], input.Extra["note"])
				require.Equal(t, extra["custom"], input.Extra["custom"])
			} else {
				require.Equal(t, 7, input.Concurrency)
				require.Equal(t, extra, input.Extra)
			}
			unchangedExtra, err := json.Marshal(request.Data.Accounts[0].Extra)
			require.NoError(t, err)
			require.Equal(t, originalExtra, unchangedExtra, "import must not mutate the source Extra")
			require.Zero(t, adminSvc.updateAccountCalls, "proxy configuration must be supplied only during creation")
			if test.createErr != nil {
				require.Equal(t, 1, result.AccountFailed)
				require.Zero(t, result.AccountCreated)
				require.Zero(t, result.AccountProxyBound)
				require.Len(t, result.Errors, 1)
				require.Equal(t, test.createErr.Error(), result.Errors[0].Message)
			} else {
				require.Equal(t, 1, result.AccountCreated)
				require.Zero(t, result.AccountFailed)
				require.Empty(t, result.Errors)
				if test.wantAutoBind {
					require.Equal(t, 1, result.AccountProxyBound)
				} else {
					require.Zero(t, result.AccountProxyBound)
				}
			}
		})
	}
}
