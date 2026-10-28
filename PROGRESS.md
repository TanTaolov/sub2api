# Issue #578 delivery log

- Workspace: `fix/issue-578-bps-usage-snapshot`, exclusive writer; baseline production `08d69a522c79a182fd681ead8fa4d848d9e1554e`.
- Read DEV_GUIDE.md and searched workspace/ancestor AGENTS.md (none found). Read full Issue #578 body and all comments (none), plus #582 scope. Repository Redis files and other dedicated workflows are excluded.
- User restrictions override DEV_GUIDE local dynamic checks: no local Go compilation/tests, Docker, full frontend build/lint/typecheck, production access or paid calls. Local gofmt/static/diff only; dynamic checks run on exact-head remote CI using synthetic HTTP fixtures.
- Acceptance: pending — header latest-value persistence with bounded writes, ordering and failure retry; BPS TTL credential routing/throttling; nonzero QueryUsage/RefreshQuota scheduler synchronization.
- Required controls: pending — thresholds/credits semantics, Spark independent windows, identity match, recovery CAS/new 429, BPS backoff, no stream replay, no unintended reset-credit consume.
- First action: writing behavior regressions before changing application code. Dedicated workflow will transplant these tests onto the pinned baseline and require assertion failures (build/setup failure is not RED evidence).
- Draft PR / exact head CI / independent parent review: pending.
- Evidence directory: `/data/development/artifacts/sub2api-issue-578`.
