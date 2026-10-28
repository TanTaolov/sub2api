# 独立上游渠道对象与当前利润归属

本 PR 完整范围以本节为准，旧 URL 备注仅兼容历史增量，不能替代渠道实体。

渠道有不可变内部 ID / 唯一 code、可编辑唯一名称和备注，账号通过独立外键关系显式归属。绑定不依赖 URL 识别或获取能力；一个渠道可包含不同 URL、未知 URL 和不支持获取的账号，不同渠道也可包含同 URL。

- 管理入口 `/admin/upstream-channels`，账户管理和精确利润页均可进入。利润页直接管理渠道；名称/备注保存前显示成员数量并确认，保存后 GET 读回。
- 账号新建选择渠道并确认，普通新建和绑定在同一事务；各导入流程仅对新建账号追加显式绑定，不覆盖已有账号归属。编辑/批量弹窗独立提供归属保存：最多 100 个明确账号，读取当前版本与归属，确认影响后原子 CAS 保存并 GET 读回；全筛选模式不做无限绑定。
- 修改账号名称、URL、分页位置或渠道显示名不会改变内部 ID 关联。绑定变动会递增渠道版本，使先前确认的影响人数失效。
- 删除只允许无绑定对象且必须确认 revision；外键 RESTRICT，不级联删除账号/usage。有成员先在分页成员列表明确选择并确认解绑，读取新版本后删除。解绑保留每账号版本 tombstone，防旧编辑复活。
- 利润优先 `channel:<内部ID>` 聚合；未绑定才回退现有安全 URL 或 unknown。展示显式/未绑定来源。统计代表**当前绑定配置**，不是历史路由快照；绑定、改名、解绑不重写原 usage。成本仍是已保存 total_cost × 当前账号成本倍率，未叠加模型成本。
- migration 313 增加独立对象和关联表；不从旧 URL/备注猜测归属。312 保留兼容元数据，旧对象不自动绑定。
- API `/api/v1/admin/upstream-channels`、`/:id`、`/:id/members` 及 `/api/v1/admin/upstream-channel-bindings`。仅管理员可管理，沿用写审计，正文省略，仅记录 ID/版本/有界数量/目标摘要。

以下为此前已验证增量的记录（不代表完整实体需求验收）：

# 独立利润运维（stacked）

Depends-on #585；base `fix/upstream-url-channel-operations`。本PR差异只新增利润页面/API、财务聚合/来源、导航、财务测试和本说明。URL身份解析复用渠道PR，不另建渠道分组。

**修正旧混合#584口径：估算成本 = 保存的usage.total_cost × 当前独立成本倍率extra.cost_multiplier；收入 = 保存actual_cost；利润 = 扣费金额 − 估算成本。** 用户扣费倍率已包含在保存actual_cost，不再乘；account_rate_multiplier不是成本倍率。account_stats_cost可由管理员自定义统计价卡或不同模型价生成，不一定等于标准模型收费基数，故不用于此用户确认的成本公式。旧profit scorer仍用account_stats_cost是其调度评分口径，本PR不改评分或账本。

默认0.1明确调度估算，零倍率保留；缺失/无效/未知不当0，零收入毛利率未知，汇总由金额总额计算。金额使用历史保存值，缓存/按次已含其中，不按当前价卡重算。当前成本倍率/归属无历史快照，逐账号来源及配置updated_at明确为估算，订阅计价不等于现金收款。

- [ ] 本head成本口径RED→GREEN：标准10、统计8、账号扣费倍率999、成本倍率.25、用户扣费4 → 成本2.5，利润1.5。
- [ ] 财务零值/亏损/未知/分页/半开区间/只读HTTP→PG。
- [ ] 本head前端完整/浏览器及全套CI。


## URL 渠道备注

利润页 `/admin/profit-operations` 的「URL 渠道备注」可直接编辑、保存、取消和读取最新备注。最多 500 个 Unicode 字符，首尾空白去除；空白保存即清空。保存后通过独立 GET 确认，失败保留草稿；409 后需读取最新备注，核对当前保存内容再提交。页面刷新重新从服务器读取。

备注仅存于 `upstream_url_channel_remarks`（migration 312），键严格使用现有 `ResolveOperationsUpstream` 的安全有效 URL，包含有意义的 path / port / 转义边界及平台尾斜杠规则。同 URL 账号共用；未知或含凭据/query/fragment 的 URL 不可写。PUT 校验键确实存在于当前账号有效投影。清空保留 revision 防止旧版本复活。写入使用现有管理员认证、observer 拒绝与审计队列，审计正文整体省略，保留动作、操作者、结果、配置版本和安全 identity 的 SHA-256 摘要（复用 filter_hash）。

API：`GET/PUT /api/v1/admin/profit-operations/url-remark`；快照新增 `url_remarks`，只附加当前页面可见 identity 的元数据。不改变利润公式、历史 usage、账号路由或凭据、模型定价、账本结算支付。

远端专属验收：`.github/workflows/profit-url-remarks-regression.yml` 运行 DTO/权限测试、丢弃备注的持久化 RED、自有迁移 PostgreSQL 的真实 HTTP→DB→GET GREEN、既有利润/URL 回归、前端组件、完整 build 与精确页面离线浏览器。HTTP 测试使用独立 database 和显式 Ent runtime，不重置共享 fixture。
