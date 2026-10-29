# Issue #594 — OpenAI multiple authorization connections

## Contract and boundaries

- Sole writer: Codex, `gpt-6.1-sol`, explicitly `xhigh` (user-supplied session declaration). Hermes: `gpt-6.1-sol`, reasoning unknown; requirements/Issue/parent work only; no parent verification claimed.
- Worktree `/data/development/projects/sub2api-openai-multi-auth`; branch `feat/openai-multi-auth-connections`; baseline `12a49ef4ef6a2f116389701ac8f1fe620f27bbf5`; artifacts `/data/development/artifacts/sub2api-openai-multi-auth`.
- Read full Issue #594 and paginated comments (none), CONTRIBUTING.md, DEV_GUIDE.md, all eight PR-template sections. No AGENTS.md found in worktree or ancestors.
- No other worktrees or PR #591/#593. Ordinary commit/push; production PR base. No rebase/force push/PR merge/release/deploy.
- No local Go compile/test/run/generate, Docker, full frontend build/lint/typecheck. Only gofmt, bash -n, static/diff, lightweight scripts. Dynamic/generated/UI evidence: isolated dedicated remote CI.
- No production, real credentials, paid inference, real mail, challenge bypass or impersonated registration. Actual consent/commercial SIWC approval stays manual.
- Preserve billing formulas, multipliers, ledgers, account security, concurrency and shared limits; unknown entitlement sharing must restrict rather than multiply quota.

## Existing implementation audit

- BPS already has a dedicated OAuth client/callback/transport; reuse it.
- SIWC #485/#491/#498/#511 merged history read and archived. #498 verifies ID tokens instead of decoding opaque access metadata; #511 adds encrypted shared sessions and eligibility diagnostics. Preserve these and #592 Redis 6.0 compatibility.
- `openai_siwc_credentials` currently keys one registration by account_id; Codex/BPS share the account credentials map. SIWC has account-scoped refresh locks/CAS and a route-selection boolean. This is not arbitrary concurrent registrations.
- Application-user auth identities/channels are sign-in identities, not inference grants. Never email-merge them into upstream connections. `swic` remains only the original typo for SIWC.
- Official SIWC sessions/inference docs read: separate issued client/token sets, serialized rotating refresh, public Responses API, store=false/stream=true, separate identity consent and shared plan limits.

## Acceptance checklist (unchecked = incomplete)

- [ ] Stable account identity → typed connections: issuer/client/verified subject, workspace/resource/audience/scopes, host/session; no email-only identity merge.
- [ ] Independent protected credentials, real client refresh, CAS and cross-instance coordination; reauthorization/disable/unbind cannot overwrite peers.
- [ ] Real production DTO → service → repository → HTTP → management UI integration.
- [ ] Capability/credential/transport separation: Codex/BPS/SIWC/API key; reject unknown/identity-only grants without silent fallback.
- [ ] zh/en list/status/capabilities, add/reauthorize/select/disable/unbind; no secrets returned.
- [ ] Scheduler/stickiness/cooldown/errors/usage connection attribution; preserve account-wide security, shared quotas and concurrency.
- [ ] Repeatable lossless legacy mapping, defaults/rollback; old API and merged SIWC/BPS fixes preserved.
- [ ] Remote RED before GREEN; actual adapters compile; real HTTP → repository → isolated PostgreSQL → GET.
- [ ] CI coexistence/concurrent rotation/reauthorization/single unbind/routing/rejection/legacy upgrade; no single-file mock substitute.
- [ ] UI component tests and sanitized remote screenshots.
- [ ] First real code plus RED committed/pushed promptly; one Draft PR to production, Refs #594, complete eight-section template and unfinished checklist.
- [ ] Final exact-head push/PR checks and statuses paginated; raw failures read/fixed; independent review recorded without fabricated parent verification.

## Delivery and evidence

- Initial clean baseline confirmed. No local dynamic checks. Implementation/CI pending.
- Draft PR/latest head/exact-head pass/fail/pending/skipped: pending first reviewable implementation.
