package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProfitURLRemarksValidationAndIdentity(t *testing.T) {
	revision := int64(0)
	for _, upstream := range []string{"https://relay.example:8443/tenant/v1", "https://relay.example/Case/%2F", "https://relay.example/tenant/"} {
		m := UpstreamURLRemarkMutation{Upstream: upstream, Remark: "  共享渠道\n备注  ", Revision: &revision}
		require.NoError(t, ValidateUpstreamURLRemarkMutation(&m))
		require.Equal(t, upstream, m.Upstream, "use the resolver identity without applying a second platform-dependent slash trim")
		require.Equal(t, "共享渠道\n备注", m.Remark)
	}
	for _, upstream := range []string{"", "unknown:invalid_url", "https://user:password@relay.example/v1", "https://relay.example/v1?api_key=secret", "https://relay.example/v1?", "https://relay.example/v1#secret", "//relay.example", "masked://***", " https://relay.example"} {
		m := UpstreamURLRemarkMutation{Upstream: upstream, Revision: &revision}
		require.ErrorIs(t, ValidateUpstreamURLRemarkMutation(&m), ErrUpstreamURLRemarkValidation, upstream)
	}
	for _, remark := range []string{strings.Repeat("中", 501), "x\x00y", string([]byte{0xff})} {
		m := UpstreamURLRemarkMutation{Upstream: "https://relay.example", Remark: remark, Revision: &revision}
		require.ErrorIs(t, ValidateUpstreamURLRemarkMutation(&m), ErrUpstreamURLRemarkValidation)
	}
	m := UpstreamURLRemarkMutation{Upstream: "https://relay.example", Remark: strings.Repeat("😀", 500), Revision: &revision}
	require.NoError(t, ValidateUpstreamURLRemarkMutation(&m))
	m.Remark = "  \n\t "
	require.NoError(t, ValidateUpstreamURLRemarkMutation(&m))
	require.Empty(t, m.Remark)
	m.Revision = nil
	require.Error(t, ValidateUpstreamURLRemarkMutation(&m))
	m.Revision = new(int64(-1))
	require.Error(t, ValidateUpstreamURLRemarkMutation(&m))
}
