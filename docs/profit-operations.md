# 独立上游渠道对象与当前利润归属

利润运营页 `/admin/profit-operations` 和账户管理共同使用显式上游渠道实体：不可变内部 ID、不可变唯一 code、唯一可编辑名称、备注及 revision。账号通过外键显式归属；同渠道可有不同 URL，不同渠道可有相同 URL，未知/不支持获取的账号可手动绑定。

渠道管理入口 `/admin/upstream-channels`，账户新建/编辑/批量选择以及利润页均可管理。批量操作须明确选中最多100个账号、确认成员影响并按版本 CAS；保存后 GET 读回。删除要求无绑定成员并确认版本；先明确解绑，外键 RESTRICT，不级联删除账号或 usage。账号 hard-delete 只删自身绑定记录并递增 channel revision，历史 usage 保留。

利润按显式 channel ID 优先聚合；未绑定账号才按安全有效 URL 聚类，否则进入 unknown。归属是**当前配置**，不是历史路由快照；不回填历史用量。成本口径不变：估算成本 = 已保存 `usage_logs.total_cost` × 当前账号统一成本倍率 `extra.cost_multiplier`（未设置沿用既有默认 0.1），收入 = 已保存 `actual_cost`，利润 = 收入 − 估算成本；不额外叠加模型成本、账号扣费倍率或账本数据。

Migration 312 创建独立的 `upstream_channels` / `upstream_channel_bindings` 对象与关联表；不按 URL、展示名称或历史账本推断关系。旧分支URL备注迁移/API/UI已从本PR撤除，不回填其数据。

## 验收记录

- 当前 Draft PR #593 在唯一 branch 上实现全部上述范围。远端真实 HTTP → 私有迁移 PostgreSQL → GET / 账户创建绑定 / 利润跨页与 ID 下钻回归，以及精确利润页、渠道页、AccountsView 的合成浏览器截图由专属 workflow 执行。
- 当前唯一剩余测试是完整 Go unit/integration push 和 PR 两项；本机遵守约束，未执行 Go 动态、Docker、完整前端 build/lint/typecheck。完整前端构建、频道绑定后端与专属实体回归都已在当前 head 对应远端检查中完成/等待最终完整 CI。
- UI/API 使用虚构数据；浏览器阻断真实 `/api/` 和外部网络，无付费模型请求。
- 不改变路由、凭据、倍率、利润公式、历史 usage 或财务账本，不 merge/release/deploy/生产访问。
