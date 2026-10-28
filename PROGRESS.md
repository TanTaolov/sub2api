# PR #593：独立上游渠道实体与多账号绑定

- Worktree / branch：`sub2api-profit-url-channel-remarks` / `feat/profit-url-channel-remarks`；base `production` @ `08d69a522c79a182fd681ead8fa4d848d9e1554e`。
- Draft PR：https://github.com/ranxi2001/sub2api/pull/593；同一 PR 按用户更正完成完整实体需求。
- 当前冻结实现 head：`9b514750d105d1d6992fb3a61bef5ad403ee208b`；最终 filter=all 已完成（30 success / 3 既有条件 skipped），工作区与远端一致。
- 已读取 DEV_GUIDE 和原 PR 八节模板。PR body 保留完整模板小节；按最新实际范围更新，不创建第二 PR。
- Agent 披露：Codex / `gpt-6.1-sol` / `xhigh`；Hermes / `gpt-6.1-sol` / 推理档位未知（父级提供）。未启动子 agent。

## 完整实现

- [x] `upstream_channels` 独立实体：BIGSERIAL 稳定 ID、唯一不可变 code、唯一可编辑名称、备注、revision。
- [x] `upstream_channel_bindings` 独立账号关联：允许任意账号类型，包括不可识别/不可获取 URL；每账号 revision CAS / NULL tombstone。
- [x] 管理员实体分页 CRUD、绑定列表、分页成员、确认绑定和有界解绑；observer / 普通用户拒绝；写审计正文省略。
- [x] 账号新建（同事务）/编辑/批量选择和读回；渠道与账号管理入口可见，保存前列出成员影响。显式创建账号失败后读回可重试而不重复创建。
- [x] 删除有成员渠道返回冲突；必须先明确解绑、确认并读取新版本后删除空对象。通道 FK RESTRICT，不级联删账号/usage。硬删除账号会在账号删除事务里清除其自身绑定并递增渠道 revision，历史 usage 不变。
- [x] 只有独立渠道是可写的渠道备注与身份。旧 URL-key 备注表/API/UI/迁移已撤除；不由 URL 或旧备注回填/猜测实体关系。URL 只作未绑定账号自动聚合 fallback。
- [x] 利润按显式 channel ID 聚合及下钻；同渠道可有不同 URL，不同渠道可有相同 URL。下钻筛选 channel FK 而不是 URL 字符串。未绑定 URL / unknown 明确标识；统计只是当前账号归属，不是历史路由。
- [x] 保持已确认利润口径：`usage_logs.total_cost × 当前 extra.cost_multiplier`；未设置沿用统一账户默认0.1。不叠加模型成本、账本费用倍率或改变账本/历史usage。
- [x] 中文/英文界面、独立 `/admin/upstream-channels`、账户新建/编辑/批量、精确利润页管理与直接入口，文档齐备。

## 完整验证（已验收 head bfb1f3804）

- 本机定向 Vitest 8 spec / 329 assertions；包含账号新建、编辑、批量、实体管理、绑定失败/读回/版本冲突、语言完整性。未运行 Go 编译/动态、Docker 或全量前端构建。
- 远端 filter=all：28 success / 3 既有条件 skipped（upstream v0.2.19 regression、CLA check/lock）；所有 check 全部完成，无失败。
- push / PR 通用 workflow 的 unit、integration、lint、安全及前端关键检查均通过；完整前端 build/typecheck 与实体专属测试均按实际 head 运行。
- 专属 PostgreSQL 变异绑定 RED 命中“explicit binding must persist independently of URL eligibility”；恢复后真实 Gin API → 私有 migration 312 DB → GET → 实际账户创建/绑定 → 利润跨页/改名/解绑/删除为 GREEN。同一实体跨不同 URL、两个实体共享相同 URL、无效 URL/动态 OAuth 都覆盖。
- 当前实体下钻 SQL 对所选 channel FK 过滤完整 cohort，防止其他账号错误落进 unknown。真实 repository 回归验证 selected total 1 / requests 4，且 JSON 历史 `usage_logs.channel_id`（模型定价）不参与实体归属。
- 浏览器：9c82752b7 的最终实体页面截图运行真实 ProfitOperationsView、UpstreamChannelsView、AccountsView 与账户新建/编辑/批量控件，使用 Axios synthetic adapter，禁止真实 `/api/`/外部访问；0 browser errors、0 live API、0 paid requests。实体浏览器产物与截图见 `/data/development/artifacts/sub2api-profit-url-channel-remarks/entity-browser-accepted-9c/`。旧 URL-key 编辑截图已废弃，不用于此次验收。
- 所有已完成及历史 raw CI 日志、filter=all 结果、修复原因与产物索引见 `/data/development/artifacts/sub2api-profit-url-channel-remarks/`。

## 最终提交与冻结门槛

- [ ] 本次收敛版文档（实体唯一身份、公式、旧 URL-key 路径已删除）提交到同一 branch。
- [ ] 最后提交 push / PR 的所有 checks `filter=all` 完成且 raw 日志复核；任何失败继续修复。
- [ ] 冻结最终 head、PR / 本地 / 远端一致且工作区干净后交父级审查。父级确认是交接点，不作为缺失功能阻塞。

完整推送 CI 后保持 Draft；不 merge / release / deploy。禁止生产 SSH/DB/Redis、付费模型；不修改模型定价 channel、路由、凭据、成本公式和历史账本。
