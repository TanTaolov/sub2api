package repository

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
)

// All OAuth registrations retain their actual client. Signing keys are fetched
// before spending a one-time code or rotating refresh token.
func (s *openaiOAuthService) ConnectionToken(ctx context.Context, form url.Values, proxy, nonce, kind string) (*openai.TokenResponse, *openai.SIWCIdentity, error) {
	client, err := createOpenAIReqClient(proxy)
	if err != nil {
		return nil, nil, errors.New("connection proxy is unavailable")
	}
	keys, err := client.R().SetContext(ctx).Get("https://auth.openai.com/.well-known/jwks.json")
	if err != nil || keys == nil || !keys.IsSuccessState() {
		return nil, nil, errors.New("OpenAI signing keys unavailable")
	}
	clientID := form.Get("client_id")
	var token *openai.TokenResponse
	switch kind {
	case "siwc":
		if clientID == "" || clientID == openai.SIWCDynamicClientID {
			return nil, nil, errors.New("issued SIWC client is required")
		}
		form.Set("resource", openai.SIWCResource)
		token = new(openai.TokenResponse)
		resp, e := client.R().SetContext(ctx).SetFormDataFromValues(form).SetSuccessResult(token).Post(openai.SIWCTokenURL)
		if e != nil || resp == nil || !resp.IsSuccessState() {
			return nil, nil, errors.New("connection token exchange failed; reauthorize if required")
		}
	case "codex", "bps":
		expected := openai.ClientID
		if kind == "bps" {
			expected = openai.ExcelClientID
		}
		if clientID != expected {
			return nil, nil, errors.New("OAuth registration client mismatch")
		}
		if form.Get("grant_type") == "authorization_code" {
			token, err = s.ExchangeCode(ctx, form.Get("code"), form.Get("code_verifier"), form.Get("redirect_uri"), proxy, clientID)
		} else {
			token, err = s.RefreshTokenWithClientID(ctx, form.Get("refresh_token"), proxy, clientID)
		}
		if err != nil {
			return nil, nil, errors.New("connection token exchange failed; reauthorize if required")
		}
	default:
		return nil, nil, errors.New("unknown OAuth connection kind")
	}
	if token == nil || token.AccessToken == "" || token.ExpiresIn <= 0 || token.TokenType != "" && !strings.EqualFold(token.TokenType, "Bearer") {
		return nil, nil, errors.New("invalid connection token response")
	}
	var identity *openai.SIWCIdentity
	if token.IDToken != "" {
		identity, err = openai.VerifySIWCIdentity(token.IDToken, clientID, nonce, keys.Bytes())
		if err != nil {
			return nil, nil, err
		}
	} else if form.Get("grant_type") == "authorization_code" {
		return nil, nil, errors.New("connection identity token missing")
	}
	// An identity-only SIWC grant is retained for sign-in, with no inference
	// capabilities. It is never converted into Codex authorization.
	return token, identity, nil
}

func (s *openaiOAuthService) RevokeConnectionToken(ctx context.Context, token, clientID, proxy string) error {
	if token == "" || clientID == "" {
		return errors.New("renewable session is unavailable")
	}
	client, err := createOpenAIReqClient(proxy)
	if err != nil {
		return errors.New("revocation proxy unavailable")
	}
	discovery, err := client.R().SetContext(ctx).Get("https://auth.openai.com/.well-known/openid-configuration")
	if err != nil || discovery == nil || !discovery.IsSuccessState() {
		return errors.New("revocation discovery unavailable")
	}
	var config struct {
		RevocationEndpoint string `json:"revocation_endpoint"`
	}
	if json.Unmarshal(discovery.Bytes(), &config) != nil {
		return errors.New("invalid revocation discovery")
	}
	endpoint, err := url.Parse(config.RevocationEndpoint)
	if err != nil || endpoint.Scheme != "https" || endpoint.Host != "auth.openai.com" || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" {
		return errors.New("invalid revocation endpoint")
	}
	result, err := client.R().SetContext(ctx).SetFormData(map[string]string{"token": token, "token_type_hint": "refresh_token", "client_id": clientID}).Post(endpoint.String())
	if err != nil || result == nil || result.StatusCode != 200 {
		return errors.New("remote revocation was not confirmed")
	}
	return nil
}
