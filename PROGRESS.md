# 独立上游渠道实体与账号绑定（PR #593，完整范围）

- 唯一 writer：本 worktree `sub2api-profit-url-channel-remarks` / branch `feat/profit-url-channel-remarks`；base production `08d69a522c79a182fd681ead8fa4d848d9e1554e`。
- 已逐条消费用户更正：**URL 备注方案不是完整交付**。独立渠道内部 ID 才是关联键；可读唯一标识/名称和备注不是外键，绑定不要求可识别/可获取 URL。
- 同一 Draft PR：https://github.com/ranxi2001/sub2api/pull/593 。旧 head `e6d2022cf` 的验证保留为历史证据，不再等待或重复独立验收。当前任务恢复实现，旧冻结记录已失效。当前 head `bac3bddf8c11a92d6b98f7fa36cfbc0ba9f80e83`。
- 已重新读取 DEV_GUIDE 和 PR 当前正文。当前远端标题已是渠道实体，但读取到的正文仍含旧 URL 增量及“冻结”描述；后续保留现有模板小节，仅准确更新完整范围状态/head/证据/未完成清单，不退回简版。
- 披露：Codex / gpt-6.1-sol / xhigh；Hermes / gpt-6.1-sol / 推理档位未知（父级提供的参与声明）。不另启 writer。

## 设计与完整待办

- [x] 独立 `upstream_channels` 对象：不可变内部 ID、不可变唯一可读 code、可编辑唯一名称、备注、revision CAS。
- [x] 独立账号关联：account FK + nullable channel FK + 每账号 binding revision；所有账号类型（包括未知/动态/不支持 URL）可绑定；一次最多 100 个明确账号，显式确认，多账号影响可见。
- [x] 管理员 CRUD / list / GET / binding preview+readback API、observer/普通用户拒绝、审计正文省略及目标摘要。
- [x] 删除仅允许无绑定渠道并要求版本与确认；有绑定返回冲突，先执行明确有界解绑；FK RESTRICT，不级联删账号/usage，不静默重归属。
- [x] 账户管理新建、编辑、批量选择器、当前归属展示与管理入口；失败保留选择/草稿，真实 API 保存后 GET 读回，避免账号已建却因绑定失败重复建账号。
- [x] 渠道管理入口可新建/编辑/删除、分页账号成员和明确有界解绑；利润页直接可管理备注/进入渠道管理。
- [x] 利润聚合优先显式 `channel:<内部ID>`，未绑定保留安全有效 URL / unknown fallback，清晰标注未绑定；同渠道不同 URL / 不同渠道相同 URL / 改名改 URL / 翻页关联稳定。
- [x] 归属明确为当前绑定统计，非历史路由；不回写 usage 或历史账本。利润公式保持 `usage.total_cost × 当前账号成本倍率`，不额外叠加模型成本。
- [ ] 先 RED 再 GREEN：真实前端组件/实际 client；真实 HTTP→私有迁移 DB→GET→账户绑定→利润跨页/改名/解绑；CAS、删除、权限及隐私边界。
- [ ] zh/en、精确页面与账号新建/编辑/批量交互截图；远端实体专属 CI，保留旧增量回归，不删 skip 弱化测试。
- [ ] 尽快提交第一批实体实码并 push 同 branch，Draft body 明确未完成项；继续修复当前 push/PR filter=all 和 raw logs。
- [ ] 完整实现/新验收全部完成才冻结交父级；父级审查不是缺失实现的阻塞。

## 约束与证据

- 本机禁止 Go 编译/动态、Docker、全量 frontend build/lint/typecheck；允许 gofmt、静态检查、轻量定向 Vitest。完整动态仅远端 CI。
- 不改模型定价 channel、账号路由/凭据/倍率、利润公式、usage、财务账本/支付。邻居和旧模板复用文件/工作树不动。
- 不 merge/release/deploy，不访问生产 SSH/DB/Redis，不调用付费模型。
- 产物：`/data/development/artifacts/sub2api-profit-url-channel-remarks/`；旧 ACCEPTANCE/freeze 仅证明旧 URL 增量，不代表完整需求验收。新实体证据使用 `entity-*` 文件。

## 当前实现与验证状态

- 第一批实码 `0c3bf2d42` 已 push：对象/关联迁移、API、CAS/删除、普通账号创建事务绑定、利润优先聚合。
- 第二批：账户新建/编辑/批量选择器、独立实体管理/分页解绑、利润页管理和显式来源、账号 GET/list DTO 归属。修改成员会递增实体 revision，影响确认可防并发变化。
- 旧完整利润 CI 首批发现历史 usage 的模型定价 channel_id 与新投影重名，已改内部 explicit_channel_id，未改利润公式；修复待远端确认。
- 新页面先 RED（实体入口缺失），实现后旧账号+利润7 spec 324/324通过；新增管理/绑定6项通过。真实 HTTP→私有DB→GET→账户创建/绑定→利润跨页/改名/解绑回归已写，专属CI保留旧回归并新增丢弃绑定变异RED→GREEN。后端/完整前端 pending远端；截图正在补。

- `b052fd0bd` 专属后端 push job 123035226410 已完成真实实体绑定变异 RED→GREEN 与账号管理/利润跨页全链路。首批前端 typecheck 两处 nullable 已修复；Go lint QF1008 已按 raw 修正。旧响应在账号选择变化时的隔离追加专属回归，不允许同数量的新选择接收旧集合。

- 50f67c3b8 修复显式渠道下钻必须按 `channel_id` 外键过滤（同渠道不同URL不漏）；定向前端 7 spec 326/326 通过。其首个 security raw 的 undefined strconv 已在 bac3bddf8 修复。当前 head 的完整 push/PR/专属 channel CI pending。
- 9c82752b7 专属 push/PR 的实体浏览器产物已归档于 `entity-browser-accepted-9c`：真实 ProfitOperationsView / UpstreamChannelsView / AccountsView，包含实体改名备注读回、账号新建确认、编辑解绑、批量绑定；verification 为 current head 9c82752b7、0 live API、0 paid requests、0 browser errors。截图不含真实账号/凭据。

- 显式利润下钻支持 `channel_id`（也规范化 `channel:<id>`），按绑定FK跨不同URL过滤，不再把实体当URL文本；加入服务解析/冲突测试和真实 repository `Total=1 / Requests=4` 跨实体下钻回归。相关4个定向spec 12/12本机通过。当前最终尝试head `14c00451b16340662259d312c18306b7e0d66e8a`，完整push/PR CI pending。
- 最新完整 filter=all head 已滚动至 e0f095ae4；CI 实测暴露 `upstream_channels_test.go` 同名测试被重复追加，已于后续 head 删除重复。旧上游 Operations RED workflow 会将整个 service 文件替换为基线，下一提交为该测试夹具保留 channel_id DTO 并剔除已迁移的 duplicate multiplier helper，不改变原两项真实RED assertion。
