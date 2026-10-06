package service

import (
	"context"
	"math/rand/v2"
	"sync"
	"time"
)

// AccountRandomProxyExtraKey 标记账号启用"每条请求随机代理"。
// 开启后账号仍绑定一个真实 proxy_id 作为兜底(后台刷新/OAuth/配额等按 ProxyID
// 回查代理的路径继续使用它)；网关请求和智慧测试路径在每次取账号时从已启用且
// 未过期的静态代理中随机挑选一个替换 account.Proxy。
const AccountRandomProxyExtraKey = "proxy_random_per_request"

const (
	accountRandomProxyPoolTTL      = 15 * time.Second
	accountRandomProxyRetryBackoff = 3 * time.Second
	accountRandomProxyLoadTimeout  = 2 * time.Second
)

// IsRandomProxyPerRequest 报告账号是否启用每条请求随机代理。
// spark 影子的代理恒继承母账号，不参与随机。
func (a *Account) IsRandomProxyPerRequest() bool {
	if a == nil || a.Extra == nil || a.IsCredentialShadow() {
		return false
	}
	enabled, _ := a.Extra[AccountRandomProxyExtraKey].(bool)
	return enabled
}

// accountRandomProxyPool 缓存可用于随机挑选的代理列表(短 TTL)，避免每条请求都查库。
type accountRandomProxyPool struct {
	repo ProxyRepository
	ttl  time.Duration

	loadMu   sync.Mutex
	mu       sync.RWMutex
	proxies  []Proxy
	loadedAt time.Time
}

func newAccountRandomProxyPool(repo ProxyRepository) *accountRandomProxyPool {
	if repo == nil {
		return nil
	}
	return &accountRandomProxyPool{repo: repo, ttl: accountRandomProxyPoolTTL}
}

func (p *accountRandomProxyPool) cached(now time.Time) ([]Proxy, bool, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	loaded := !p.loadedAt.IsZero()
	return p.proxies, loaded, loaded && now.Sub(p.loadedAt) < p.ttl
}

// snapshot 返回当前代理列表。过期时由单个调用方刷新，其余调用方继续使用旧列表；
// 仅在从未加载过时才同步等待首次加载。
func (p *accountRandomProxyPool) snapshot(ctx context.Context) []Proxy {
	list, loaded, fresh := p.cached(time.Now())
	if fresh {
		return list
	}
	if !p.loadMu.TryLock() {
		if loaded {
			return list
		}
		p.loadMu.Lock()
	}
	defer p.loadMu.Unlock()
	// 等锁期间可能已被其他调用方刷新。
	if list, _, fresh = p.cached(time.Now()); fresh {
		return list
	}

	loadCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), accountRandomProxyLoadTimeout)
	proxies, err := p.repo.ListActive(loadCtx)
	cancel()

	p.mu.Lock()
	defer p.mu.Unlock()
	if err != nil {
		// 保留旧列表，短暂退避后再重试，避免数据库故障时每条请求都打库。
		p.loadedAt = time.Now().Add(accountRandomProxyRetryBackoff - p.ttl)
		return p.proxies
	}
	p.proxies = proxies
	p.loadedAt = time.Now()
	return p.proxies
}

// pick 从已启用且未过期的代理中均匀随机挑选一个，返回副本。
func (p *accountRandomProxyPool) pick(ctx context.Context) *Proxy {
	if p == nil {
		return nil
	}
	list := p.snapshot(ctx)
	now := time.Now()
	var chosen *Proxy
	n := 0
	for i := range list {
		if !list[i].IsActive() || list[i].IsExpired(now) {
			continue
		}
		n++
		// 蓄水池抽样：单次遍历、无额外分配。
		if rand.IntN(n) == 0 {
			chosen = &list[i]
		}
	}
	if chosen == nil {
		return nil
	}
	picked := *chosen
	return &picked
}

// apply 为启用随机代理的账号替换本次请求使用的 Proxy。
// 只替换 account.Proxy，保留 account.ProxyID 为绑定的兜底代理，
// 防止请求路径上的账号回写把随机结果落库。没有可用代理时保留原绑定代理。
// 调用方必须传入本次请求独占的账号对象(缓存解码/数据库读取得到的新对象)。
func (p *accountRandomProxyPool) apply(ctx context.Context, account *Account) *Account {
	if p == nil || account == nil || account.ProxyID == nil || !account.IsRandomProxyPerRequest() {
		return account
	}
	if picked := p.pick(ctx); picked != nil {
		account.Proxy = picked
	}
	return account
}
