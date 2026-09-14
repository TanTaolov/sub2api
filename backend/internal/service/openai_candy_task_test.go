//go:build unit

package service

import (
	"io"
	"net/http"
	"testing"

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

	resp := newJSONResponse(http.StatusOK, `data: {"type":"response.completed"}`)
	upstream := &queuedHTTPUpstream{responses: []*http.Response{resp}}
	svc := &AccountTestService{httpUpstream: upstream}
	account := &Account{
		ID:          91,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Concurrency: 1,
		Credentials: map[string]any{"access_token": "test-token"},
	}

	err := svc.testOpenAIAccountConnection(ctx, account, "gpt-5.4", "", AccountTestModeCandy)
	require.NoError(t, err)
	require.Len(t, upstream.requests, 1)

	body, err := io.ReadAll(upstream.requests[0].Body)
	require.NoError(t, err)
	require.Equal(t, openAICandyTestPrompt, gjson.GetBytes(body, "input.0.content.0.text").String(),
		"糖果测试必须发送固定题面而不是探活提示词")
}
