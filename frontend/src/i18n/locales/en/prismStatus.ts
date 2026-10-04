export default {
  title: 'Prism status',
  description: 'Prism requests run through the loopback adapter, and the gateway only forwards by account switch and model scope. This page shows the configuration in effect on the gateway; it is not proof that the adapter or the accounts are usable right now.',
  enabled: 'Gateway switch',
  enabledHint: 'When off, Prism models on every account fall back to the ordinary Codex / Excel routing.',
  on: 'Enabled',
  off: 'Disabled',
  baseUrl: 'Adapter address',
  baseUrlHint: 'Only http://127.0.0.1:<port>/v1 or http://[::1]:<port>/v1 is accepted; no environment or account proxy is used.',
  notSet: 'Not configured',
  endpoint: 'Effective endpoint',
  endpointHint: 'The address the gateway actually calls; it ends with /v1/responses.',
  endpointUnavailable: 'Invalid address, no endpoint can be built',
  apiKey: 'Bridge key',
  apiKeyConfigured: 'Configured',
  apiKeyMissing: 'Not configured',
  apiKeyHint: 'Only whether the key is set is shown, never the secret; it must match the adapter PRISM_ADAPTER_API_KEY.',
  models: 'Model scope',
  modelsHint: 'Only requests whose account has Prism selected and whose mapped model is inside this scope reach the adapter; an explicit empty scope means no Prism routing.',
  states: {
    ready: 'Configuration ready',
    disabled: 'Prism is disabled on the gateway',
    endpoint_invalid: 'Adapter address is invalid',
    key_missing: 'Bridge key is not configured'
  },
  stateHints: {
    ready: 'The gateway forwards models in scope to the loopback adapter. Whether the adapter is running and whether the account has the entitlement still needs separate confirmation.',
    disabled: 'Even when an account has Prism selected, requests never reach the adapter and keep the ordinary Codex / Excel routing.',
    endpoint_invalid: 'The current address is not a valid loopback HTTP endpoint, so Prism requests fail before they are sent and never fall back to a native channel.',
    key_missing: 'The bridge key between the gateway and the adapter is empty, so Prism requests fail before they are sent.'
  },
  scopeNote: 'Configuration is read at process start, so a change needs a gateway restart. This page does not probe the adapter, does not send test requests, and never replays a failed Prism request; read the gateway log and the account test for the exact failure.'
}
