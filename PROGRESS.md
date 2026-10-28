# URL 渠道备注进度

- 专属 worktree / branch：`sub2api-profit-url-channel-remarks` / `feat/profit-url-channel-remarks`；base `08d69a522c79a182fd681ead8fa4d848d9e1554e`，目标 `production`。
- Draft PR：https://github.com/ranxi2001/sub2api/pull/593 。初步实现即提交 `008ca5059` / push / 创建 Draft，未等待全部 CI；后续持续同 branch 修复。
- 已读 `DEV_GUIDE.md`；本 worktree 与父目录未发现 `AGENTS.md`。未操作邻居 #578 / #582 worktree 或原 profit worktree。
- 精确页面：`/admin/profit-operations` → `ProfitOperationsView` → `ProfitAnalytics`，页面直接提供中文备注编辑、保存、取消和最新读回。

## 实现边界

- 复用 `ResolveOperationsUpstream` 的规范化有效 URL identity；拒绝凭据、query、fragment / 未知桶，保留有意义的 port、path 大小写、转义边界及平台尾斜杠规则。仅显示实际快照发现的 identity；写入时再次验证投影归属。
- 已检查账号 notes 与 channel description / channel operations：前者绑定账号，后者绑定模型定价 channel / 手工分组，均无法自然复用为 URL 共享备注，因此增加最小独立表 migration 312。
- 同 URL 的多个 / 新增 / 改名账号共享备注；保存按 revision CAS，空白保存清空并保留版本防 ABA。首尾空白去除，最多 500 Unicode 字符。
- 保存后独立 GET 确认；保存失败 / 409 / 读回失败均保留草稿，显式重新读取当前保存内容后再提交。
- 复用管理员认证、observer 拒绝及审计队列；审计正文整体省略，记录操作者、结果、配置版本及安全 identity SHA-256（filter_hash），不复制 URL / 备注文本。
- 仅分析元数据；不修改账号路由、凭据、定价倍率、利润公式、历史 usage、账本结算支付。未 merge / release / deploy，未访问生产 SSH / DB / Redis，未调用付费模型。

## 已获得的验证证据

- 本地先跑既有利润页 5/5 GREEN，再补新增组件测试得到 RED（4 失败），实现后合计 10/10 GREEN；没有删除 / skip / 弱化原测试。
- 本地仅依赖安装、定向 Vitest、gofmt、diff 检查和脚本语法检查；未运行 Go 编译 / 动态测试、Docker、全量 frontend build / lint / typecheck。
- `3dea0176a` 远端专属 PostgreSQL PR job `123003424834` 已通过：DTO / URL identity / 实际路由授权；丢弃持久化备注的变异 RED 确认命中 `API must persist the submitted note`；恢复后真实 HTTP → 私有迁移 PostgreSQL → GET / snapshot GREEN。涵盖同 URL、path / port / case / 转义边界、创建与更新并发、清空与旧版本、改名 / 新账号 / 分页、审计隐私及财务行不变。
- `3dea0176a` 远端专属前端 push job `123003396360` 已通过组件 10/10、完整 build/typecheck 和精确页面浏览器；桌面 / 移动截图、0 browser errors / 0 live API / 0 paid requests。
- 初始 head 专属后端失败 raw 日志确认 JWT fixture 缺失密码指纹，预期 403 实得 401；`3dea0176a` 改为 repository 完整用户签发，保留所有授权断言后通过。
- `8c70a2521` 补上审计目标 identity 摘要及断言，push/PR 专属回归均通过。
- 最终复核发现并修复前端 409 契约：apiClient 拦截器返回顶层 status；新增真实 API module + client interceptors + Vue editor 的 adapter 回归，先确认 2 项 RED，再兼容顶层 status / Axios response.status。原授权和冲突断言均保留，继续走同套 CI。

## 最终验收与冻结

- 专属 workflow：`.github/workflows/profit-url-remarks-regression.yml`，push + PR 均检查实际 head。Go / Docker / 完整前端动态检查仅远端执行，私有 database + 显式 Ent runtime，不清空共享 fixture。
- 全部 push / PR 检查（`check-runs?filter=all`）、原始 logs、每个 head 的状态和最终验收记录持续保存在 `/data/development/artifacts/sub2api-profit-url-channel-remarks/`，以 `runs.json`、`checks-<head>.json` 与 `ACCEPTANCE.md` 为交接证据。
- 最终必须等待所有检查完成并审核失败 raw 日志；保持 Draft，完成后冻结 branch 交父级审查，不 merge / release / deploy。CI 尚未完成时 PR body 保持 pending 清单；验收结果在最终交接记录列明。
