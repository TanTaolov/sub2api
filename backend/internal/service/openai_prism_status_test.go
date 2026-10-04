package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestPrismBrowserStatusReportsGatewayConfiguration(t *testing.T) {
	gateway := func(enabled bool, baseURL, key string) *OpenAIGatewayService {
		cfg := &config.Config{}
		cfg.Gateway.PrismBrowser = config.GatewayPrismBrowserConfig{Enabled: enabled, BaseURL: baseURL, APIKey: key}
		return &OpenAIGatewayService{cfg: cfg}
	}
	tests := []struct {
		name         string
		service      *OpenAIGatewayService
		wantState    PrismBrowserState
		wantEndpoint string
		wantKey      bool
	}{
		{name: "disabled switch never reports an endpoint", service: gateway(false, "http://127.0.0.1:8319/v1", "fixture-key"), wantState: PrismBrowserStateDisabled, wantKey: true},
		{name: "non loopback address is rejected", service: gateway(true, "https://adapter.example/v1", "fixture-key"), wantState: PrismBrowserStateEndpointInvalid, wantKey: true},
		{name: "wrong path is rejected", service: gateway(true, "http://127.0.0.1:8319/responses", "fixture-key"), wantState: PrismBrowserStateEndpointInvalid, wantKey: true},
		{name: "missing bridge key blocks the path", service: gateway(true, "http://127.0.0.1:8319/v1", "   "), wantState: PrismBrowserStateKeyMissing, wantEndpoint: "http://127.0.0.1:8319/v1/responses"},
		{name: "loopback endpoint with key is ready", service: gateway(true, "http://127.0.0.1:8319/v1", "fixture-key"), wantState: PrismBrowserStateReady, wantEndpoint: "http://127.0.0.1:8319/v1/responses", wantKey: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.service.PrismBrowserStatus()
			require.Equal(t, tt.wantState, got.State)
			require.Equal(t, tt.wantEndpoint, got.Endpoint)
			require.Equal(t, tt.wantKey, got.APIKeyConfigured)
			require.Equal(t, PrismBrowserSupportedModels(), got.Models)
		})
	}

	// The status view must never claim readiness when the service is missing.
	var missing *OpenAIGatewayService
	require.Equal(t, PrismBrowserStateDisabled, missing.PrismBrowserStatus().State)
	require.Equal(t, PrismBrowserSupportedModels(), missing.PrismBrowserStatus().Models)
}
