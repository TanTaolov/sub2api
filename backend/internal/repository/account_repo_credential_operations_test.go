package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// 账号是软删除：删除账号只更新 accounts 行，数据库外键级联不会触发。删除账号时必须
// 显式清除凭证运营的巡检登记与加密登录资料，否则已删除账号会一直留在巡检列表里。
func TestDeleteCredentialOperationsRecordsPurgesRegistrationAndLoginData(t *testing.T) {
	exec := &recordingSQLExecutor{result: rowsAffectedResult(1)}

	require.NoError(t, deleteCredentialOperationsRecords(context.Background(), exec, 42))

	require.Equal(t, []string{
		"DELETE FROM account_token_guard_v2_accounts WHERE account_id = $1",
		"DELETE FROM openai_oauth_reauth_configs WHERE account_id = $1",
		"DELETE FROM openai_oauth_reauth_tasks WHERE account_id = $1",
	}, exec.execQueries)
	require.Len(t, exec.execArgs, 3)
	for _, args := range exec.execArgs {
		require.Equal(t, []any{int64(42)}, args)
	}
}

// 清理失败必须向上返回，让账号删除事务回滚，而不是留下半清理状态。
func TestDeleteCredentialOperationsRecordsReportsFailure(t *testing.T) {
	exec := &recordingSQLExecutor{result: rowsAffectedResult(0), err: errors.New("delete failed")}

	require.Error(t, deleteCredentialOperationsRecords(context.Background(), exec, 42))
	require.Len(t, exec.execQueries, 1)
}
