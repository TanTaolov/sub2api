//go:build integration

package repository_test

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

// Swap leases exactly after A's acquisition and before A's registration reread.
// The real service must keep A's acquisition proof rather than adopting B's.
type takeoverConnectionRepository struct {
	service.OpenAIOAuthReauthRepository
	service.OpenAIAuthConnectionRepository
	takeover func()
	done     atomic.Bool
}

func (r *takeoverConnectionRepository) ListAuthConnections(ctx context.Context, id int64) ([]*service.OpenAIAuthConnection, error) {
	rows, err := r.OpenAIAuthConnectionRepository.ListAuthConnections(ctx, id)
	if err == nil && len(rows) > 0 && rows[0].LeaseOwner != "" && r.done.CompareAndSwap(false, true) {
		r.takeover()
		return r.OpenAIAuthConnectionRepository.ListAuthConnections(ctx, id)
	}
	return rows, err
}
func TestIssue594LeaseTakeoverCannotAdoptAnotherOwnersProof(t *testing.T) {
	for _, action := range []string{"disable", "unlink", "failure"} {
		t.Run(action, func(t *testing.T) {
			f := newConnectionHTTPFixture(t)
			ctx := context.Background()
			c := connectionResponse(t, f.request("POST", "", `{"kind":"api_key","label":"lease takeover","api_key":"synthetic-takeover-key"}`))
			base := repository.NewOpenAIOAuthReauthRepository(f.db)
			wrapper := &takeoverConnectionRepository{OpenAIOAuthReauthRepository: base, OpenAIAuthConnectionRepository: base.(service.OpenAIAuthConnectionRepository)}
			var ownerA string
			wrapper.takeover = func() {
				require.NoError(t, f.db.QueryRow(`SELECT lease_owner FROM openai_auth_connections WHERE id=$1`, c.ID).Scan(&ownerA))
				require.NotEmpty(t, ownerA)
				_, e := f.db.Exec(`UPDATE openai_auth_connections SET lease_until=clock_timestamp()-INTERVAL '1 second' WHERE id=$1`, c.ID)
				require.NoError(t, e)
				acquired, e := f.repo.AcquireAuthConnectionLease(ctx, f.accountID, c.ID, "worker-B")
				require.NoError(t, e)
				require.True(t, acquired)
			}
			svc := service.NewOpenAIOAuthReauthService(wrapper, connectionDBReader{f.accounts}, f.accounts.(service.OpenAIOAuthReauthCredentialUpdater), service.NewOpenAIOAuthService(nil, repository.NewOpenAIOAuthClient()), f.encryptor, true, nil, nil)
			var err error
			if action == "failure" {
				err = svc.FailAuthConnection(ctx, &service.Account{ID: f.accountID, AuthConnectionID: c.ID, AuthConnectionRevision: c.Revision}, 401)
			} else {
				_, err = svc.UpdateAuthConnection(ctx, f.accountID, c.ID, service.UpdateOpenAIAuthConnectionInput{Action: action, Revision: c.Revision})
			}
			require.True(t, wrapper.done.Load(), "must exercise A lease → expiry → B takeover → A reread")
			require.Error(t, err, "A cannot commit using owner/epoch obtained from B's snapshot")
			rows, e := f.repo.ListAuthConnections(ctx, f.accountID)
			require.NoError(t, e)
			require.Equal(t, "active", rows[0].Status)
			require.Equal(t, c.Revision, rows[0].Revision)
			require.Equal(t, "worker-B", rows[0].LeaseOwner)
			require.NoError(t, f.repo.ReleaseAuthConnectionLease(ctx, f.accountID, c.ID, "worker-B"))
		})
	}
}
