# Issue #578 delivery log

## Reopened after independent review (current work)

- Read the complete `parent-independent-review.md`. Prior green CI and FINAL_VALIDATION.md do **not** constitute acceptance: parent review blocks PR #591.
- P1 pending: replace process/repository-object-only ordering with an atomic shared persistence CAS and ordered cache publication; add two independent coordinator domains using the same real PostgreSQL/Redis and scheduler selection regression.
- P1 pending: audit every Codex window writer, remove DTO/time restamping bypasses in CachePostResetSnapshot and auto-reset persistFreshUsage, retain original observation/presence, and exercise actual query -> downstream seams under races and missing fields.
- P2 pending: stop/clean workers on ErrAccountNotFound, cancel waits and in-flight writes on service shutdown, preserve transient recovery, safely retire ordering state, and expose bounded non-sensitive errors; add lifecycle regressions.
- Keep same branch/Draft PR; no neighbor #592 repository Redis files/workflows or neighbor worktree edits. No local Go compilation/dynamic tests, Docker/full frontend checks, production access or merge/release/deploy.
- Preserve the parent's eight-section CONTRIBUTING PR body. Disclose Hermes independent review subAgent (`gpt-6.1-sol`, reasoning unknown). New-head remote checks/raw evidence and independent re-review remain required.
- New behavior tests written first: real QueryUsage/details -> auto-reset/post-reset stale DTO seam; permanent deletion and shutdown cancellation. Implementation now routes Codex windows through a PostgreSQL observation-version CASE/UPDATE, serializes cache publishers using transaction advisory locks plus fresh committed reads, and wraps scheduler rebuild/token-refresh publication without touching repository Redis files.
- Added dedicated shared-storage CI job: native isolated PostgreSQL/Redis inside loopback-only network namespace, two independent gateway/repository domains, actual 30-second delayed-header race, cache publication race and actual scheduler selection. The rejected head is tested RED with the same storage regressions. Dynamic results pending; this is initial review-fix increment, not acceptance.
- First review-fix increment `7a39e8594` pushed. Added follow-up coverage of the complete evaluateAccount / RunOpenAIQuotaResetPostProcess flows (query rejected, details delayed, missing usage), plus full stale account edits under the existing FOR UPDATE lock and stale scheduler publication. Native storage/RED/GREEN CI running; no local dynamic execution.

- Exclusive worktree/branch: `fix/issue-578-bps-usage-snapshot`; production baseline `08d69a522c79a182fd681ead8fa4d848d9e1554e`.
- Read DEV_GUIDE.md; searched ancestors and all hidden/non-hidden workspace paths for AGENTS.md (none). Read full Issue #578 body and all comments (none), and #582 scope. No repository Redis files or neighbor workflows changed.
- User restrictions override DEV_GUIDE local dynamic checks. Executed only local gofmt/static/diff checks; no local Go compilation/tests, Docker, full frontend checks, production access, model/paid/mail calls, merge, release or deploy.
- Artifact root: `/data/development/artifacts/sub2api-issue-578`.
- Draft PR: https://github.com/ranxi2001/sub2api/pull/591 -> production, `Fixes #578`.

## Implementation and regression sequence

1. Wrote four behavior regressions before changing application code: throttled 89 -> 100 latest-value persistence, blocked-write ordering/retry, nonzero query synchronization, and all-model BPS TTL refresh/routing/throttle.
2. Initial commit/push `99147b5253d8f83b015b75d5e939b39e58b8ff88`; created Draft PR immediately with individually pending acceptance checklist.
3. Header snapshots now keep one coalesced pending update per repository/account, serialize writes, retain failures for bounded-rate retry, and reject older observations after a newer query/probe. No new unconditional 100% cooldown or threshold bypass.
4. BPS TTL uses quota queries independently of WSv2, with atomic automatic-probe reservations. OAuth credentials are reused; Spark resolves parent credentials and persists only its independent bengalfox windows.
5. QueryUsage synchronously persists explicitly present canonical windows; missing/invalid/mismatched windows cannot fabricate zero usage or reset times. RefreshQuota embeds `usage_cache_persisted`, including false on storage failure. Complete-zero recovery retains the original generation CAS; nonzero queries never clear a 429.
6. `978796e30` added threshold/credits, Spark, malformed/identity, explicit handler refresh, stale query/header ordering and concurrent TTL regressions. `e2931e0a0` added loopback-only network namespaces after dependency download. `127609e06` scoped fallback queues and asserted retry timing. No existing skip or assertion removed/weakened.
7. Initial push/PR lint failed only on the new unchecked sync.Map type assertion. Read both raw logs; `e6e930f182954fe9baa53bb5376e9724409d72b0` replaced this with a typed map. No application changes after this source freeze.

## Remote evidence actually executed

- Initial dedicated push 40739626916 and PR 40739672226: SUCCESS, including baseline behavioral RED, exact-head GREEN and race count=3. Initial full unit/integration jobs also succeeded; only the above lint checks failed.
- Frozen source head `e6e930f182954fe9baa53bb5376e9724409d72b0`: dedicated push 40740834882 and PR 40740845997 both SUCCESS. Artifact head.txt matches the full SHA. Network evidence shows only loopback (127.0.0.1 / ::1).
- Each frozen-source dedicated run: baseline RED has four expected top-level failures (7 failure events with subtests), no build/setup failure; GREEN service has 159 pass events, admin 25; race count=3 has 75 pass events. No failure or skip in GREEN/race logs. Controls include BPS 429 independent cooldown/no replay, recovery CAS/rearmed 429, identity matching, Spark, threshold=100 vs auto_pause=1.0, and reset-credit disabled/ignored-window behavior.
- Frozen-source push and PR lint raw logs both report `0 issues` (jobs 123001668653 and 123001708505). Other completed checks passed or retained their existing workflow skips. The two full unit/integration jobs are still running at this log commit, so their final result remains pending here.
- All push/PR check-runs fetched with `filter=all` for every pushed head. Raw workflow archives, individual failed/successful lint logs and dedicated evidence downloaded. `audit_ci.py`, `audit-summary.json`, `raw/` and `e6e930f18-evidence/` record the audit. Superseded intermediate runs were cancelled to free runners; cancellation is not counted as a pass. Final delivery-head audit will be kept in the same artifact directory and PR body.
- Local gofmt and git diff --check passed. Source hashes recorded in `all-source.sha256`; only this progress document changes in the delivery log commit. Exact delivery-head CI is required again; prior-head evidence is not represented as a final-head pass.

## Handoff gates

- Source frozen at the above SHA; final delivery commit only records this log. Continue this branch only if CI finds a real failure.
- Pending at log commit: exact delivery-head CI and final all-runs raw-log audit, then independent parent review. Final PR body/CI links carry the completed validation result.
- Independent parent review remains a separate gate. PR stays Draft; no merge/release/deploy authorized or performed.
