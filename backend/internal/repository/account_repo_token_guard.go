package repository

import "context"

// Presence, including a paused enrollment, transfers ownership of automatic
// credential recovery to Credential Operations. The legacy guard must not race
// its worker or undo an operator's pause by trying the other recovery path.
func (r *accountRepository) ListCredentialOperationsAccountIDs(ctx context.Context) ([]int64, error) {
	rows, err := r.sql.QueryContext(ctx, "SELECT account_id FROM account_token_guard_v2_accounts")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// deleteCredentialOperationsRecords 清除账号在凭证运营中的巡检登记与加密登录资料。
//
// accounts 走软删除（deleted_at），删除账号只更新 accounts 行，数据库的
// ON DELETE CASCADE 不会触发。若不显式清理，已删除的账号会一直留在凭证运营的
// 巡检列表里（显示为账号不可用），并继续保留它的加密密码 / TOTP。
// 该方法与账号删除在同一事务内执行，保证「账号已删除则凭证运营没有残留记录」。
func deleteCredentialOperationsRecords(ctx context.Context, exec sqlExecutor, accountID int64) error {
	for _, statement := range []string{
		"DELETE FROM account_token_guard_v2_accounts WHERE account_id = $1",
		"DELETE FROM openai_oauth_reauth_configs WHERE account_id = $1",
		"DELETE FROM openai_oauth_reauth_tasks WHERE account_id = $1",
	} {
		if _, err := exec.ExecContext(ctx, statement, accountID); err != nil {
			return err
		}
	}
	return nil
}
