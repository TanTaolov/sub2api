# URL 渠道备注进度

- 专属 worktree / branch：`sub2api-profit-url-channel-remarks` / `feat/profit-url-channel-remarks`；base `08d69a522c79a182fd681ead8fa4d848d9e1554e`，目标 `production`。
- 已读 `DEV_GUIDE.md`；本 worktree 与父目录未发现 `AGENTS.md`。不操作邻居 worktree。
- 已核对精确入口 `ProfitOperationsView → ProfitAnalytics`，路由要求管理员，scoped observer 禁止管理。
- 复用 `ResolveOperationsUpstream` 的有效 URL identity；拒绝凭据、query、fragment；保留有意义的端口、path 大小写和转义边界，保留平台已有尾斜杠规则。
- 既有 channel operations 表绑定模型定价 channel / 手工分组；账号 notes 绑定账号，均不适合作为共享 URL 备注。采用独立最小表，不修改账号、路由、财务数据。
- 验证顺序：先跑既有定向前端回归，再补组件回归并观察 RED，然后实现。Go 动态、Docker、全量前端 build/lint/typecheck 仅远端 CI。
- 产物：`/data/development/artifacts/sub2api-profit-url-channel-remarks`。

## 待验收

- [x] 现有前端基线 5/5；新增先 RED（4 失败），实现后合计 GREEN 10/10，日志已存产物目录
- [x] 已实现管理员 API、revision CAS / 清空 tombstone、正文省略审计；后端动态验收 pending
- [x] 已补 API → 独立迁移 PostgreSQL → GET / snapshot 回归；执行 pending 远端 CI
- [x] 精确利润页编辑、保存、取消、失败草稿、清空、Unicode 长度、刷新组件测试通过；浏览器 pending 远端 CI
- [ ] 初步实现提交 / push / Draft PR（验收项保持 pending）
- [ ] 所有 push / PR checks（filter=all）及 raw logs 审核、失败修复
- [ ] 最终 head 冻结，交父级审查；不 merge / release / deploy

初步实现已完成：新专属 workflow 含后端持久化变异 RED→GREEN、既有利润/URL 回归及浏览器。未在本机运行 Go 编译/动态测试、Docker 或全量前端 build/lint/typecheck。`gofmt`、`git diff --check`、浏览器脚本语法检查通过。即刻创建 Draft PR，完整动态验收 pending。
