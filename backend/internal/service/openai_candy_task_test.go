//go:build unit

package service

import (
	"io"
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai_compat"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestResolveOpenAITextTestPrompt(t *testing.T) {
	tests := []struct {
		name   string
		mode   string
		prompt string
		want   string
	}{
		{
			name:   "糖果测试固定题面且忽略调用方提示词",
			mode:   AccountTestModeCandy,
			prompt: "随便写点什么",
			want:   openAICandyTestPrompt,
		},
		{
			name: "糖果测试在提示词为空时仍发固定题面",
			mode: AccountTestModeCandy,
			want: openAICandyTestPrompt,
		},
		{
			name:   "常规模式沿用调用方提示词",
			mode:   AccountTestModeDefault,
			prompt: "hello",
			want:   "hello",
		},
		{
			name: "常规模式提示词为空时回退极短探活提示词",
			mode: AccountTestModeDefault,
			want: defaultOpenAITextTestPrompt,
		},
		{
			name:   "常规模式提示词去空白",
			mode:   AccountTestModeDefault,
			prompt: "  spaced  ",
			want:   "spaced",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, resolveOpenAITextTestPrompt(tt.mode, tt.prompt))
		})
	}
}

func TestOpenAICandyTestPromptIsFrozenQuestion(t *testing.T) {
	// 题面是固定题目的单点定义：改动这里等于改动「糖果测试」的语义，需同步前端提示文案。
	require.Contains(t, openAICandyTestPrompt, "三种口味的糖果")
	require.Contains(t, openAICandyTestPrompt, "五角星形")
	require.Contains(t, openAICandyTestPrompt, "最少取出多少个糖果")
}

func TestCreateOpenAITestPayloadUsesCandyPrompt(t *testing.T) {
	payload := createOpenAITestPayload("gpt-5.4", openAICandyTestPrompt, false)

	require.Equal(t, "gpt-5.4", payload["model"])
	require.Equal(t, true, payload["stream"])

	input, ok := payload["input"].([]map[string]any)
	require.True(t, ok)
	require.Len(t, input, 1)

	content, ok := input[0]["content"].([]map[string]any)
	require.True(t, ok)
	require.Len(t, content, 1)
	require.Equal(t, openAICandyTestPrompt, content[0]["text"])
}

func TestCreateOpenAITestPayloadEmptyPromptFallsBackToProbeText(t *testing.T) {
	payload := createOpenAITestPayload("gpt-5.4", "   ", false)
	input := payload["input"].([]map[string]any)
	content := input[0]["content"].([]map[string]any)

	require.Equal(t, defaultOpenAITextTestPrompt, content[0]["text"])
}

func TestAccountTestService_OpenAICandyModeSendsCandyQuestion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := newTestContext()

	resp := newJSONResponse(http.StatusOK, "data: {\"type\":\"response.completed\"}\n\n")
	upstream := &queuedHTTPUpstream{responses: []*http.Response{resp}}
	svc := &AccountTestService{httpUpstream: upstream}
	account := &Account{
		ID:          91,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Concurrency: 1,
		Credentials: map[string]any{"access_token": "test-token"},
	}

	err := svc.testOpenAIAccountConnection(ctx, account, "gpt-5.4", "ignored custom prompt", AccountTestModeCandy)
	require.NoError(t, err)
	require.Len(t, upstream.requests, 1)

	body, err := io.ReadAll(upstream.requests[0].Body)
	require.NoError(t, err)
	require.Equal(t, openAICandyTestPrompt, gjson.GetBytes(body, "input.0.content.0.text").String(),
		"糖果测试必须发送固定题面而不是探活提示词")
}

func TestAccountTestService_OpenAICandyModeAPIKeySendsCandyQuestion(t *testing.T) {
	for _, testCase := range []struct {
		name               string
		responsesSupported bool
		response           string
		path               string
		promptPath         string
	}{
		{
			name:               "responses",
			responsesSupported: true,
			response:           "data: {\"type\":\"response.completed\"}\n\n",
			path:               "/v1/responses",
			promptPath:         "input.0.content.0.text",
		},
		{
			name:       "chat completions",
			response:   "data: {\"choices\":[{\"delta\":{\"content\":\"answer\"},\"finish_reason\":\"stop\"}]}\n\ndata: [DONE]\n\n",
			path:       "/v1/chat/completions",
			promptPath: "messages.0.content",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			ctx, _ := newTestContext()
			upstream := &queuedHTTPUpstream{responses: []*http.Response{
				newJSONResponse(http.StatusOK, testCase.response),
			}}
			svc := &AccountTestService{httpUpstream: upstream, cfg: &config.Config{}}
			account := &Account{
				ID:       94,
				Platform: PlatformOpenAI,
				Type:     AccountTypeAPIKey,
				Credentials: map[string]any{
					"api_key":  "test-api-key",
					"base_url": "https://text-upstream.example/v1",
				},
				Extra: map[string]any{openai_compat.ExtraKeyResponsesSupported: testCase.responsesSupported},
			}

			err := svc.testOpenAIAccountConnection(ctx, account, "gpt-5.4", "ignored custom prompt", AccountTestModeCandy)
			require.NoError(t, err)
			require.Len(t, upstream.requests, 1)
			require.Equal(t, testCase.path, upstream.requests[0].URL.Path)
			body, err := io.ReadAll(upstream.requests[0].Body)
			require.NoError(t, err)
			require.Equal(t, openAICandyTestPrompt, gjson.GetBytes(body, testCase.promptPath).String())
		})
	}
}

func TestAccountTestService_OpenAICandyModeRejectsImageModels(t *testing.T) {
	for _, accountType := range []string{AccountTypeOAuth, AccountTypeAPIKey} {
		for _, modelID := range []string{"gpt-image-2", "text-alias"} {
			t.Run(accountType+"/"+modelID, func(t *testing.T) {
				ctx, recorder := newTestContext()
				upstream := &queuedHTTPUpstream{}
				svc := &AccountTestService{httpUpstream: upstream}
				account := &Account{
					ID:       92,
					Platform: PlatformOpenAI,
					Type:     accountType,
					Credentials: map[string]any{
						"access_token": "test-token",
						"api_key":      "test-api-key",
						"model_mapping": map[string]any{
							"text-alias": "gpt-image-2",
						},
					},
				}

				err := svc.testOpenAIAccountConnection(ctx, account, modelID, "draw a cat", AccountTestModeCandy)
				require.ErrorContains(t, err, "Candy test requires a text model")
				require.Contains(t, recorder.Body.String(), `"type":"error"`)
				require.Contains(t, recorder.Body.String(), "image models are not supported")
				require.Empty(t, upstream.requests, "拒绝图片模型时不得发出上游请求")
			})
		}
	}
}

func TestAccountTestService_OpenAIDefaultModeKeepsImageGeneration(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		prompt string
		want   string
	}{
		{name: "custom prompt", prompt: "draw a cat", want: "draw a cat"},
		{name: "default prompt", want: defaultOpenAIImageTestPrompt},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			ctx, recorder := newTestContext()
			upstream := &queuedHTTPUpstream{responses: []*http.Response{
				newJSONResponse(http.StatusOK, `{"data":[{"b64_json":"aGVsbG8="}]}`),
			}}
			svc := &AccountTestService{httpUpstream: upstream, cfg: &config.Config{}}
			account := &Account{
				ID:       93,
				Platform: PlatformOpenAI,
				Type:     AccountTypeAPIKey,
				Credentials: map[string]any{
					"api_key":  "test-api-key",
					"base_url": "https://image-upstream.example/v1",
				},
			}

			err := svc.testOpenAIAccountConnection(ctx, account, "gpt-image-2", testCase.prompt, AccountTestModeDefault)
			require.NoError(t, err)
			require.Len(t, upstream.requests, 1)
			require.Equal(t, "/v1/images/generations", upstream.requests[0].URL.Path)
			body, err := io.ReadAll(upstream.requests[0].Body)
			require.NoError(t, err)
			require.Equal(t, testCase.want, gjson.GetBytes(body, "prompt").String())
			require.Contains(t, recorder.Body.String(), "data:image/png;base64,aGVsbG8=")
		})
	}
}
