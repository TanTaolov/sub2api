# PR #593：独立上游渠道实体与多账号归属

- 专属 worktree `sub2api-profit-url-channel-remarks`，唯一 branch `feat/profit-url-channel-remarks`；base `production` @ `08d69a522c79a182fd681ead8fa4d848d9e1554e`。
- 同一 Draft PR：https://github.com/ranxi2001/sub2api/pull/593 。当前实现范围已消费全部用户更正：不可变内部 ID + 唯一不可变 code + 唯一名称 + 备注；多账号显式关联；未知/不支持 URL 可绑定；显式渠道优先利润归属。
- Codex / gpt-6.1-sol / xhigh；Hermes / gpt-6.1-sol / 推理档位未知（父级提供）。无子 Agent。
- 最新源码变更 head（待文档收尾后提交）：当前本地 `762f9a36e5aadb9f4a2224e4966a323aafcc6977`；PR 当前 head 同步读取 PR API。最新远端 CI 对应完整实体实现；旧URL-key移除变更会产生新的完整 head验收。

## 完整待办（实现）

- [x] 独立表 `upstream_channels`：稳定 BIGSERIAL ID，不可变唯一 code，唯一可编辑 name，remark，revision。
- [x] `upstream_channel_bindings` 独立 FK；所有账号类型可显式绑定，不解析 URL；channel / account 外键独立模型，模型定价 channel 不参与。
- [x] 账号新建同一事务绑定；编辑/批量选择单独提供最多100账号、确认影响、revision CAS、GET 读回；选择变化时丢弃旧异步响应。导入新增账号绑定，已有账号不被默默重归属。
- [x] 独立 `/admin/upstream-channels` 管理实体、分页成员、确认改名/备注、明确有界解绑。删除仅允许无成员实体且版本+确认匹配；FK RESTRICT，不删除账号/usage。硬删账号只清其 binding 且递增渠道 revision，usage保留。
- [x] `/admin/profit-operations` 显式 channel ID 优先聚合、显式 ID 下钻过滤跨URL、未绑定才回退安全URL/unknown，显示当前归属来源并说明非历史路由。
- [x] 清理 superseded URL-key备注方案：删除旧 URL note 表/迁移、写API、handler/repository/service、前端API/编辑组件、利润DTO/页面编辑区、旧 RED→GREEN 测试路径。迁移 312 现在只创建实体/绑定，不从URL备注猜测或迁移归属。
- [x] 统一成本公式：`saved usage.total_cost × 当前账号 extra.cost_multiplier`；保留既有0.1默认，不新增模型成本或扣费倍率。
- [x] zh/en实体文案、真实利润/账户/渠道浏览器截图脚本、实体专属CI。保留原PR八节模板和Codex/Hermes披露。

## 完整回归待验收

- [x] 新增实体入口 RED→GREEN；本地定向现有账号/利润、新建/编辑/批量/实体管理与语言测试 8 spec / 329 assertions通过（含后续新增筛选回归：7 spec / 326 assertions）。没有删除/skip/弱化原有测试。
- [x] 远端 `9c82752b7` 实体浏览器截图通过：真实 ProfitOperationsView、UpstreamChannelsView、AccountsView；实体备注读回、新账号确认、新建、编辑解绑、批量绑定均覆盖；0真实API、0付费、0 browser errors。产物 `entity-browser-accepted-9c/` 是旧UI兼容版证据，旧URL面板截图不能代表最新清理后UI。
- [x] 远端 `888c450f7` filter=all 30 checks全部完成，28 success / 2原有条件 skipped；后续改动导致的head过期，不能冒充最终验证。
- [ ] 在最终清理后head重新通过实体专属变异 RED→GREEN、HTTP→私有迁移DB→GET→账号创建/绑定→利润跨页/同URL不同实体/同渠道不同URL/ID下钻/改名/解绑/hard-delete历史不变。
- [ ] 最终清理后head真实浏览器产物、Accounts create/edit/bulk、利润/实体管理/zh-en完整build。
- [ ] 所有最新 push/PR `check-runs?filter=all` 和原始日志审核；当前检测需要以最终清理后head重新执行。
- [ ] 最终head所有远端检查完成后冻结分支交父级审查；父级审查不是缺失实现的阻塞。

## 当前最新实现补充

- 显式 `channel_id` 下钻复用同一 accounts cohort；usage scoped query也 join过滤后的 cohort，避免将非所选渠道请求塞入unknown桶。真实repository跨URL drill-down断言已加。
- 当前 `ProfitAnalytics.vue` 不显示或编辑 URL-key备注；渠道备注只由独立实体ID绑定与管理。
- 最终文档/清理提交及其触发的完整CI尚待执行。

## 约束与证据

- 本机严禁Go编译/动态、Docker、全量frontend build/lint/typecheck；只用定向Vitest、gofmt、diff/node语法检查。完整验证走远端CI。
- 不改账号路由/凭据/倍率、模型定价channel、利润公式/历史usage/账本支付。未merge/release/deploy，未访问生产SSH/DB/Redis、未调用付费模型。
- 产物目录 `/data/development/artifacts/sub2api-profit-url-channel-remarks/`；旧 `ACCEPTANCE.md` / `freeze.json` 只作URL旧版本历史证据，勿用作当前验收。实体日志/产物名均为 `entity-*`。
