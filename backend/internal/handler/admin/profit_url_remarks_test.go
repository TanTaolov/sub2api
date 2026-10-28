package admin

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type profitURLRemarkRepo struct {
	service.ChannelRepository
	err error
}

func (r *profitURLRemarkRepo) ReadUpstreamURLRemark(_ context.Context, key string) (*service.UpstreamURLRemark, error) {
	return &service.UpstreamURLRemark{Upstream: key, Remark: "DB readback", Revision: 3}, r.err
}
func (r *profitURLRemarkRepo) SaveUpstreamURLRemark(_ context.Context, m service.UpstreamURLRemarkMutation) (*service.UpstreamURLRemark, error) {
	return &service.UpstreamURLRemark{Upstream: m.Upstream, Remark: m.Remark, Revision: *m.Revision + 1}, r.err
}
func TestProfitURLRemarksHTTPDTOValidationAndSafeErrors(t *testing.T) {
	repo := &profitURLRemarkRepo{}
	h := NewChannelHandler(service.NewChannelService(repo, nil, nil, nil, nil), nil, nil)
	r := gin.New()
	r.GET("/remark", h.ReadProfitURLRemark)
	r.PUT("/remark", h.SaveProfitURLRemark)
	for _, tc := range []struct {
		method, path, body string
		code               int
	}{
		{"GET", "/remark?upstream=https://relay.example", "", 200},
		{"GET", "/remark?upstream=https://user:secret@relay.example", "", 400},
		{"PUT", "/remark", `{"upstream":"https://relay.example","remark":"","revision":0}`, 200},
		{"PUT", "/remark", `{"upstream":"https://relay.example","revision":0}`, 400},
		{"PUT", "/remark", `{"upstream":"https://relay.example","remark":"note"}`, 400},
	} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		require.Equal(t, tc.code, w.Code, w.Body.String())
		if tc.method == "GET" && tc.code == 200 {
			require.Contains(t, w.Body.String(), `"remark":"DB readback"`)
			require.Contains(t, w.Body.String(), `"revision":3`)
		}
	}
	for _, tc := range []struct {
		err  error
		code int
	}{{service.ErrUpstreamURLRemarkConflict, 409}, {errors.New("raw SQL secret credentials"), 500}} {
		repo.err = tc.err
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("PUT", "/remark", strings.NewReader(`{"upstream":"https://relay.example","remark":"note","revision":2}`)))
		require.Equal(t, tc.code, w.Code)
		require.NotContains(t, w.Body.String(), "raw SQL secret")
	}
}
