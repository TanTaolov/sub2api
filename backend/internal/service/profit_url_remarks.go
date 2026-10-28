package service

import (
	"context"
	"math"
	"strings"
	"unicode"
	"unicode/utf8"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var ErrUpstreamURLRemarkValidation = infraerrors.BadRequest("UPSTREAM_URL_REMARK_VALIDATION", "URL 渠道备注无效：仅支持有效上游网址，备注最多 500 字，必须提供版本")
var ErrUpstreamURLRemarkConflict = infraerrors.Conflict("UPSTREAM_URL_REMARK_CONFLICT", "备注已被其他管理员修改，请读取最新版本后重试")

type UpstreamURLRemark struct {
	Upstream string `json:"upstream"`
	Remark   string `json:"remark"`
	Revision int64  `json:"revision"`
}
type UpstreamURLRemarkMutation struct {
	Upstream string `json:"upstream"`
	Remark   string `json:"remark"`
	Revision *int64 `json:"revision"`
}
type UpstreamURLRemarkRepository interface {
	ReadUpstreamURLRemark(context.Context, string) (*UpstreamURLRemark, error)
	SaveUpstreamURLRemark(context.Context, UpstreamURLRemarkMutation) (*UpstreamURLRemark, error)
}

// Validate privacy without re-normalizing an effective identity: upstream
// protocol builders have different trailing-separator semantics.
func ValidateUpstreamURLRemarkIdentity(upstream string) error {
	if upstream == "" || strings.TrimSpace(upstream) != upstream || !utf8.ValidString(upstream) || strings.IndexFunc(upstream, unicode.IsControl) >= 0 {
		return ErrUpstreamURLRemarkValidation
	}
	if _, err := normalizeOperationsBaseURL(upstream); err != nil {
		return ErrUpstreamURLRemarkValidation
	}
	return nil
}
func ValidateUpstreamURLRemarkMutation(m *UpstreamURLRemarkMutation) error {
	if err := ValidateUpstreamURLRemarkIdentity(m.Upstream); err != nil {
		return err
	}
	if m.Revision == nil || *m.Revision < 0 || *m.Revision == math.MaxInt64 || !utf8.ValidString(m.Remark) {
		return ErrUpstreamURLRemarkValidation
	}
	m.Remark = strings.TrimSpace(m.Remark)
	if utf8.RuneCountInString(m.Remark) > 500 || strings.IndexFunc(m.Remark, func(r rune) bool { return unicode.IsControl(r) && r != '\n' && r != '\t' }) >= 0 {
		return ErrUpstreamURLRemarkValidation
	}
	return nil
}
func (s *ChannelService) ReadUpstreamURLRemark(ctx context.Context, upstream string) (*UpstreamURLRemark, error) {
	if err := ValidateUpstreamURLRemarkIdentity(upstream); err != nil {
		return nil, err
	}
	repo, ok := s.repo.(UpstreamURLRemarkRepository)
	if !ok {
		return nil, infraerrors.ServiceUnavailable("UPSTREAM_URL_REMARK_UNAVAILABLE", "URL 渠道备注暂不可用")
	}
	return repo.ReadUpstreamURLRemark(ctx, upstream)
}
func (s *ChannelService) SaveUpstreamURLRemark(ctx context.Context, m UpstreamURLRemarkMutation) (*UpstreamURLRemark, error) {
	if err := ValidateUpstreamURLRemarkMutation(&m); err != nil {
		return nil, err
	}
	repo, ok := s.repo.(UpstreamURLRemarkRepository)
	if !ok {
		return nil, infraerrors.ServiceUnavailable("UPSTREAM_URL_REMARK_UNAVAILABLE", "URL 渠道备注暂不可用")
	}
	return repo.SaveUpstreamURLRemark(ctx, m)
}
