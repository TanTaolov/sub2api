# Issue #578 delivery log

- Workspace: `fix/issue-578-bps-usage-snapshot`, exclusive writer; baseline production `08d69a522c79a182fd681ead8fa4d848d9e1554e`.
- Read DEV_GUIDE.md and searched workspace/ancestor AGENTS.md (none found). Read full Issue #578 body and all comments (none), plus #582 scope. Repository Redis files and other dedicated workflows are excluded.
- User restrictions override DEV_GUIDE local dynamic checks: no local Go compilation/tests, Docker, full frontend build/lint/typecheck, production access or paid calls. Local gofmt/static/diff only; dynamic checks run on exact-head remote CI using synthetic HTTP fixtures.
- Acceptance: pending — header latest-value persistence with bounded writes, ordering and failure retry; BPS TTL credential routing/throttling; nonzero QueryUsage/RefreshQuota scheduler synchronization.
- Required controls: pending — thresholds/credits semantics, Spark independent windows, identity match, recovery CAS/new 429, BPS backoff, no stream replay, no unintended reset-credit consume.
- First action: writing behavior regressions before changing application code. Dedicated workflow will transplant these tests onto the pinned baseline and require assertion failures (build/setup failure is not RED evidence).
- Draft PR / exact head CI / independent parent review: pending.
- Evidence directory: `/data/development/artifacts/sub2api-issue-578`.
- Initial implementation committed/pushed as `99147b5253d8f83b015b75d5e939b39e58b8ff88`; Draft PR https://github.com/ranxi2001/sub2api/pull/591 targets production and contains `Fixes #578` plus pending acceptance checklist.
- Remote push run 40739626916 completed behavioral baseline RED step successfully; exact-head GREEN and race are still running. All push/PR check-runs retrieved with `filter=all` to evidence directory; no passing claim yet.
- Follow-up regressions added for credit/threshold semantics, Spark independence, malformed/mismatched payloads, explicit refresh persistence failure reporting, concurrent TTL reservation and queued header vs explicit query ordering. No existing assertion or skip removed.
- Follow-up committed/pushed as `978796e30`; dedicated GREEN service/admin step on initial head passed remotely, race pending. Added a loopback-only network-namespace runner for all dedicated Go test execution after dependency download, so unexpected endpoints cannot reach paid/model/mail services.
