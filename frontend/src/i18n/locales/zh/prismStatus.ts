export default {
  title: 'Prism 状态',
  description: 'Prism 请求由回环适配器执行，网关只按账号开关和模型范围转发。这里显示网关侧已生效的配置，不代表适配器或账号当前可用。',
  enabled: '网关总开关',
  enabledHint: '关闭时，所有账号的 Prism 模型都回到原有的 Codex / Excel 路由。',
  on: '已启用',
  off: '未启用',
  baseUrl: '适配器地址',
  baseUrlHint: '只接受 http://127.0.0.1:<端口>/v1 或 http://[::1]:<端口>/v1，不使用环境代理或账号代理。',
  notSet: '未配置',
  endpoint: '生效端点',
  endpointHint: '网关实际请求的地址，末尾为 /v1/responses。',
  endpointUnavailable: '地址无效，无法生成端点',
  apiKey: '桥接密钥',
  apiKeyConfigured: '已配置',
  apiKeyMissing: '未配置',
  apiKeyHint: '只显示是否配置，不显示密钥内容；需与适配器的 PRISM_ADAPTER_API_KEY 一致。',
  models: '模型范围',
  modelsHint: '只有账号编辑页勾选 Prism、且映射后模型在此范围内的请求才走适配器；显式空范围表示不走 Prism。',
  states: {
    ready: '配置就绪',
    disabled: '网关未启用 Prism',
    endpoint_invalid: '适配器地址无效',
    key_missing: '桥接密钥未配置'
  },
  stateHints: {
    ready: '网关会把范围内的模型转发给回环适配器；适配器是否在线、账号是否有对应权益仍需另行确认。',
    disabled: '账号即使勾选了 Prism，请求也不会走适配器，仍按原有 Codex / Excel 路由。',
    endpoint_invalid: '当前地址不满足回环 HTTP 要求，Prism 请求会在发出前失败，且不会回退原生通道。',
    key_missing: '网关与适配器之间的桥接密钥为空，Prism 请求会在发出前失败。'
  },
  scopeNote: '配置在进程启动时读取，修改后需要重启网关才生效。此页不探测适配器、不发送测试请求，也不会自动重放失败的 Prism 请求；具体失败原因以网关日志和账号测试为准。'
}
