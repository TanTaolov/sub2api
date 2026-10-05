//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/DATA-DOG/go-sqlmock"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/stretchr/testify/require"
)

type randomBulkProxyRepo struct {
	ProxyRepository
	proxies []Proxy
}

func (r *randomBulkProxyRepo) ListActive(context.Context) ([]Proxy, error) {
	return r.proxies, nil
}

type randomBulkAccountRepo struct {
	*accountRepoStubForBulkUpdate
	assignments      map[int64]int64
	shadows          map[int64][]*Account
	shadowUpdateErr  error
	shadowUpdateInTx bool
}

func (r *randomBulkAccountRepo) BulkUpdate(_ context.Context, ids []int64, updates AccountBulkUpdate) (int64, error) {
	r.bulkUpdateCalls++
	for _, id := range ids {
		if proxyID, ok := updates.ProxyAssignments[id]; ok {
			r.assignments[id] = proxyID
		}
	}
	return int64(len(ids)), nil
}

func (r *randomBulkAccountRepo) ListShadowsByParent(_ context.Context, id int64) ([]*Account, error) {
	return r.shadows[id], nil
}

func (r *randomBulkAccountRepo) Update(ctx context.Context, account *Account) error {
	r.shadowUpdateInTx = dbent.TxFromContext(ctx) != nil
	if r.shadowUpdateErr != nil {
		return r.shadowUpdateErr
	}
	r.updatedAccounts = append(r.updatedAccounts, account)
	return nil
}

func TestBulkRandomProxyPersistsRealIDsAndPropagatesToShadows(t *testing.T) {
	past := time.Now().Add(-time.Hour)
	parentID := int64(1)
	shadow := &Account{ID: 3, ParentAccountID: &parentID}
	repo := &randomBulkAccountRepo{
		accountRepoStubForBulkUpdate: &accountRepoStubForBulkUpdate{
			getByIDsAccounts: []*Account{{ID: 1}, {ID: 2}},
		},
		assignments: make(map[int64]int64),
		shadows:     map[int64][]*Account{parentID: {shadow}},
	}
	svc := &adminServiceImpl{accountRepo: repo, proxyRepo: &randomBulkProxyRepo{proxies: []Proxy{
		{ID: 4, Status: StatusDisabled},
		{ID: 5, Status: StatusActive, ExpiresAt: &past},
		{ID: 7, Status: StatusActive},
	}}}

	result, err := svc.BulkUpdateAccounts(context.Background(), &BulkUpdateAccountsInput{
		AccountIDs: []int64{1, 2}, RandomProxy: true,
	})

	require.NoError(t, err)
	require.Equal(t, 2, result.Success)
	require.Equal(t, 1, repo.bulkUpdateCalls)
	require.Equal(t, map[int64]int64{1: 7, 2: 7}, repo.assignments)
	require.Len(t, repo.updatedAccounts, 1)
	require.Equal(t, int64(7), *shadow.ProxyID)
}

func TestBulkRandomProxyRejectsInvalidOrEmptySelectionBeforeWrite(t *testing.T) {
	parentID := int64(9)
	for _, tc := range []struct {
		name, reason string
		proxyID      *int64
		accounts     []*Account
		proxies      []Proxy
	}{
		{name: "empty", reason: "RANDOM_PROXY_EMPTY", accounts: []*Account{{ID: 1}}},
		{name: "shadow", reason: "SPARK_SHADOW_PROXY_INHERITED", accounts: []*Account{{ID: 1, ParentAccountID: &parentID}}, proxies: []Proxy{{ID: 7, Status: StatusActive}}},
		{name: "missing", reason: "ACCOUNT_NOT_FOUND", proxies: []Proxy{{ID: 7, Status: StatusActive}}},
		{name: "conflict", reason: "RANDOM_PROXY_CONFLICT", accounts: []*Account{{ID: 1}}, proxyID: &parentID},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &randomBulkAccountRepo{
				accountRepoStubForBulkUpdate: &accountRepoStubForBulkUpdate{getByIDsAccounts: tc.accounts},
				assignments:                  make(map[int64]int64),
			}
			svc := &adminServiceImpl{accountRepo: repo, proxyRepo: &randomBulkProxyRepo{proxies: tc.proxies}}
			result, err := svc.BulkUpdateAccounts(context.Background(), &BulkUpdateAccountsInput{
				AccountIDs: []int64{1}, ProxyID: tc.proxyID, RandomProxy: true,
			})
			require.Nil(t, result)
			requireApplicationErrorReason(t, err, tc.reason)
			require.Zero(t, repo.bulkUpdateCalls)
		})
	}
}

func TestBulkRandomProxyRollsBackWhenShadowUpdateFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	t.Cleanup(func() { _ = client.Close() })
	mock.ExpectBegin()
	mock.ExpectRollback()

	parentID := int64(1)
	repo := &randomBulkAccountRepo{
		accountRepoStubForBulkUpdate: &accountRepoStubForBulkUpdate{getByIDsAccounts: []*Account{{ID: parentID}}},
		assignments:                  make(map[int64]int64),
		shadows:                      map[int64][]*Account{parentID: {{ID: 2, ParentAccountID: &parentID}}},
		shadowUpdateErr:              errors.New("shadow update failed"),
	}
	svc := &adminServiceImpl{entClient: client, accountRepo: repo, proxyRepo: &randomBulkProxyRepo{proxies: []Proxy{{ID: 7, Status: StatusActive}}}}

	result, err := svc.BulkUpdateAccounts(context.Background(), &BulkUpdateAccountsInput{AccountIDs: []int64{parentID}, RandomProxy: true})
	require.Nil(t, result)
	require.ErrorContains(t, err, "shadow update failed")
	require.True(t, repo.shadowUpdateInTx)
	require.Equal(t, 1, repo.bulkUpdateCalls)
	require.NoError(t, mock.ExpectationsWereMet())
}
