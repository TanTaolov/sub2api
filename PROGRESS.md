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

- [x] Stable account identity → typed connections: issuer/client/verified subject, workspace/resource/audience/scopes, host/session; no email-only identity merge.
- [x] Independent protected credentials, real client refresh, CAS and cross-instance coordination; reauthorization/disable/unbind cannot overwrite peers.
- [x] Real production DTO → service → repository → HTTP → management UI integration.
- [x] Capability/credential/transport separation: Codex/BPS/SIWC/API key; reject unknown/identity-only grants without silent fallback.
- [x] zh/en list/status/capabilities, add/reauthorize/select/disable/unbind; no secrets returned.
- [x] Scheduler/stickiness/cooldown/errors/usage connection attribution; preserve account-wide security, shared quotas and concurrency.
- [x] Repeatable lossless legacy mapping, defaults/rollback; old API and merged SIWC/BPS fixes preserved.
- [x] Remote RED before GREEN; actual adapters compile; real HTTP → repository → isolated PostgreSQL → GET.
- [ ] CI coexistence/concurrent rotation/reauthorization/single unbind/routing/rejection/legacy upgrade; no single-file mock substitute.
- [x] UI component tests and sanitized remote screenshots.
- [x] First real code plus RED committed/pushed promptly; one Draft PR to production, Refs #594, complete eight-section template and unfinished checklist.
- [ ] Final exact-head push/PR checks and statuses paginated; raw failures read/fixed; independent review recorded without fabricated parent verification.

## Delivery and evidence

- Initial clean baseline confirmed. No local dynamic checks. Implementation/CI pending.
- Draft PR/latest head/exact-head pass/fail/pending/skipped: pending first reviewable implementation.
- Regression-first commit `cc30c542f` and dedicated original-baseline RED workflow `a3b1adf9a` pushed before implementation. Raw results pending.
- First production slice: typed safe DTO, forward-only connection table, encrypted API-key creation/pending OAuth registrations, list/select/disable/enable/local-unlink HTTP→service→SQL, revision CAS, connection lease and unknown/identity-only dispatch rejection. OAuth callbacks, actual selected-connection inference, UI and integration acceptance remain incomplete.
- Draft PR: https://github.com/ranxi2001/sub2api/pull/595 (production, Refs #594), first implementation `1029ba36b`; OAuth lifecycle follow-up `6a38d2408`.
- Original production RED raw job `123258743040` read and archived: all three unknown/swic/identity-only cases fail because legacy dispatch returns Codex credentials, as expected. Exact-head GREEN remains pending.
- Audit correction: BPS ALSO already has an independent encrypted side-grant table (`OpenAIExcelOAuthRepository`), not only the primary credentials slot. Legacy mapping will preserve all original lifecycle owners instead of copying rotating tokens.
- Current work adds explicit selection policies, metadata/outbox propagation, capability routing/projection, per-connection failure isolation and stable read-only legacy views. HTTP→actual service→SQL→GET/rotation/unlink/CAS/lease integration added for isolated CI. UI, full OAuth integration/usage attribution and final checks remain pending.
- Existing baseline dependency security failure: GHSA-83rm-6c8r-g728 in vue-i18n/@intlify; raw failure read. No exception or false pass added.
- Historical evidence read: `efe10013b` actual HTTP/PostgreSQL lifecycle and cross-instance CAS/lease passed, no Issue594 skips; raw `green-efe10013b-raw.log`. `218bd3022` UI: 19 tests, targeted lint/typecheck and actual en/zh desktop/mobile screenshots passed; raw and images archived and inspected. These do not count as final-head evidence.
- Full actual OAuth adapter integration added for two SIWC registrations plus Codex/BPS, synthetic issuer-signed identities, identity-only rejection, cross-instance rotating refresh, reauthorization, single remote revoke and primary-grant preservation. Production gateway route/affinity/account-ban/usage-attribution integration added. Current results pending.
- Session affinity now persists connection identity under the existing tenant/session scope. Credential/status CAS does not alter peer tokens during selection. Reauthorization and revoke use the same owned DB lease as refresh; final send checks the live connection revision.
- To fix baseline GHSA-83rm-6c8r-g728 without audit exceptions: official advisory read, isolated CI generated vue-i18n/compiler 11.4.14 lock candidate, unrelated lock metadata preserved after review; frontend/security/dedicated CI Node updated to required 22. Exact-head validation pending.
- Raw `green-218bd3022-raw.log` exposed the new connection path's reuse of a legacy audience-array-only decoder. New connection metadata parsing now happens only after the real adapter verifies the signed token, accepts standard string/array audience, and requires an actual workspace claim instead of substituting a session ID.
- Raw security job on `a1edcb7f2` exposed an incorrect importer restoration while retaining unrelated lock metadata. Corrected to the remotely generated 11.4.14 importers; original xlsx integrity/platform metadata retained. No audit exception added.
- Source audit added authoritative DB selection-policy checks at preparation and final send so pre-selection cached account snapshots cannot silently use legacy credentials after another instance enables managed routing.
- Additional source review hardened encrypted records with connection/account/kind/subject binding; transplanting a ciphertext into a peer connection is rejected. Callback and revocation operations now have an overall timeout below the owned lease lifetime. Added actual DB regression for ciphertext transplant rejection.
- `dea23abe8` push and PR security scans pass after the i18n/lock fixes; other exact-head jobs pending. All earlier failed jobs' raw logs are archived, including the intentionally red test-only commit.
- BPS review: managed grants now bypass legacy account-group auto-move/auto-disable; model/credential failure cools the connection. Managed 429 follows conservative shared account policy. Its observed HTTP send boundary validates live policy and grant revision too. Legacy BPS behavior is unchanged.
- Observational attribution includes regional-metered Responses/Chat results while preserving all original money/ledger handling. Legacy registration labels now use the panel locale.
- Final regression additions exercise stale legacy snapshots after managed selection and a shared account rate-limit window through the real Gateway. Explicit return-to-legacy policy is asserted in the real DB. Local static review: 39 changed Go files gofmt-clean; 47 new zh/en locale keys match; complete eight-section PR body retained; diff clean. No local dynamic test was run.
- Full-CI raw review found two staticcheck style findings (capitalized `Codex` error, De Morgan expression) and the existing plugin snapshot field-classification tripwire. Fixed style and explicitly classified the new UUID/revision as non-secret readable metadata; no test guard was removed. The credential-operation seam's expected baseline RED cases were not real current-head regressions; its actual job failure was the already-fixed old lock importer.
- Actual adapter integration now includes two Codex and two BPS registrations with the same verified subject/client plus two SIWC registrations and one identity-only grant. Each retains its own token set/UUID; there is no uniqueness constraint that would collapse same-client connections. This closes the explicit per-kind coexistence coverage before final exact-head verification.

## Implementation completion and remaining gates

- Implementation checklist above describes delivered source, not a claim that pending exact-head CI passed. Whole-contract final validation remains unchecked until the final head completes all checks.
- Historical full dedicated GREEN at `dea23abe8`: actual adapter compilation and all four Issue594 integration cases passed; raw log archived/read. New per-kind coexistence, ciphertext binding, shared quota and latest fixes require final-head confirmation.
- Independent source review remains a human/parent gate; this sole writer's source audit is not represented as independent approval. Actual SIWC consent/eligibility/commercial approval likewise remains manual.
- Final local static evidence: actual dedicated workflow's six shell run blocks pass bash -n; changed Go files are gofmt-clean, zh/en keys agree, PR retains all eight sections, diff clean. No local compilation, dynamic test, generation, Docker or full frontend checks.
- Final lifecycle review found that disable/error-state writes could race a rotating refresh. Disable/enable now serialize under the same bounded lease; a rejection cannot discard an in-flight replacement. Managed OAuth 401 now expires only its access token and keeps its refresh grant, preserving the existing SIWC recovery behavior; API keys still require replacement. Added real-DB lease exclusion and actual-adapter 401 recovery regressions. This necessary fix supersedes the previous candidate head.
