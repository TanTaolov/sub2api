package service

import (
	"context"
	"encoding/json"

	dbaccount "github.com/Wei-Shaw/sub2api/ent/account"
	dbproxy "github.com/Wei-Shaw/sub2api/ent/proxy"
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"
)

func timingTerminal(event string) string {
	switch event {
	case "response.completed", "response.done":
		return "completed"
	case "response.failed", "error":
		return "failed"
	case "response.incomplete":
		return "incomplete"
	}
	return ""
}

// Optional repository capability avoids changing billing's existing contract.
func (s *UsageService) RequestTimings(ctx context.Context, id int64) ([]json.RawMessage, error) {
	if _, err := s.GetByID(ctx, id); err != nil {
		return nil, err
	}
	if repo, ok := s.usageRepo.(interface {
		RequestTimings(context.Context, int64) ([]json.RawMessage, error)
	}); ok {
		return repo.RequestTimings(ctx, id)
	}
	return []json.RawMessage{}, nil
}

// TimingNames 解析耗时详情中出现的账号 ID 与代理 ID 并返回对应名称，仅用于诊断展示。
// 已软删除的账号和代理仍保留名称，便于回溯历史请求；查询失败时对应表为空，前端回退显示 ID。
func (s *UsageService) TimingNames(ctx context.Context, traces []json.RawMessage) (accounts, proxies map[int64]string) {
	accounts, proxies = map[int64]string{}, map[int64]string{}
	if s.entClient == nil {
		return accounts, proxies
	}
	accountIDs, proxyIDs := timingAttemptIDs(traces)
	ctx = mixins.SkipSoftDelete(ctx)
	if len(accountIDs) > 0 {
		if rows, err := s.entClient.Account.Query().Where(dbaccount.IDIn(accountIDs...)).All(ctx); err == nil {
			for _, a := range rows {
				accounts[a.ID] = a.Name
			}
		}
	}
	if len(proxyIDs) > 0 {
		if rows, err := s.entClient.Proxy.Query().Where(dbproxy.IDIn(proxyIDs...)).All(ctx); err == nil {
			for _, p := range rows {
				proxies[p.ID] = p.Name
			}
		}
	}
	return accounts, proxies
}

// timingAttemptIDs 收集所有尝试中去重后的正数账号 ID 与代理 ID，无法解析的记录直接跳过。
func timingAttemptIDs(traces []json.RawMessage) (accountIDs, proxyIDs []int64) {
	seenAccounts, seenProxies := map[int64]struct{}{}, map[int64]struct{}{}
	add := func(id int64, seen map[int64]struct{}, out *[]int64) {
		if id <= 0 {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		*out = append(*out, id)
	}
	for _, raw := range traces {
		var trace struct {
			Attempts []struct {
				AccountID int64 `json:"account_id"`
				ProxyID   int64 `json:"proxy_id"`
			} `json:"attempts"`
		}
		if json.Unmarshal(raw, &trace) != nil {
			continue
		}
		for _, attempt := range trace.Attempts {
			add(attempt.AccountID, seenAccounts, &accountIDs)
			add(attempt.ProxyID, seenProxies, &proxyIDs)
		}
	}
	return accountIDs, proxyIDs
}
