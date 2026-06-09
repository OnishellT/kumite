
<!-- kumite:begin -->
# Kumite Pi Workflow

Use the kumite orchestration loop for non-trivial coding tasks in this project.

Default entry point:

1. Start with the project `orchestrator` subagent or the saved `kumite-loop` chain.
2. Keep parent context compact. Delegate detailed reading to `scout`, planning to `planner`, implementation to `implementer`, verification to `reviewer`, and confirmed documentation updates to `curator`.
3. Prefer file-backed handoffs under `.kumite/` and chain artifacts over large inline summaries.
4. When reviewer returns `STATUS: REWORK_REQUIRED`, the next implementer run must consume only the reviewer's `REWORK_TASK`, then reviewer must run again. Stop at `STATUS: PASS`, `STATUS: BLOCKED`, or after 2 rework rounds.

Saved chain:

- In interactive Pi sessions, run `/run-chain kumite-loop -- <task>` when the task needs the noninteractive scout -> planner-fallback -> implementer -> reviewer fallback.
- The saved chain is serial and compact. It does not interview the user; planner-fallback records deferred questions and conservative assumptions instead.
- If the user explicitly says to use `kumite-loop`, do not simulate the workflow manually in the parent session. Start `/run-chain kumite-loop -- <task>` or execute the saved chain steps through `subagent` as described below.
- In non-interactive/tool-driven sessions, saved chains are not executable agents. Do not call `subagent({agent: "kumite-loop"})`.
- For direct tool execution, inspect the saved chain with `subagent({action: "get", chainName: "kumite-loop", agentScope: "both"})` if needed. Before executing it, replace every literal `{task}` in every returned step task with the actual user task. Do not pass a chain containing `{task}` placeholders to `subagent`; direct tool execution does not reliably substitute them.
- After placeholder replacement, execute the concrete steps as `subagent({chain: [...], task: "<task>", clarify: false, agentScope: "both"})` unless the user explicitly requests a timeout.
- The scout handoff must include an `Original task` section. Later steps must treat `scout-context.md` as the durable task artifact; do not wait for a separate `task.md` unless the user provided one.
- Saved-chain child outputs may be stored in a pi-subagents temporary chain-run directory. Kumite agents must also mirror handoffs into `.kumite/handoffs/current/` so the project keeps durable copies across sessions.
- In scratch projects without a `.git` directory, reviewer should record Git diff checks as `SKIPPED: no git worktree`, run the remaining verification/static-analysis checks, and not block solely on missing Git metadata.
- If running the agents manually instead of through the saved chain, enforce rework from reviewer `STATUS` and `REWORK_TASK` output.
- If running the agents manually, the planner phase must use the `kumite-grill-with-docs` protocol: write a first-draft plan, challenge it against memory/docs, and include `Questions asked`, `Deferred questions`, or `No blocking questions` before implementation starts.
- For small edits, the parent may delegate directly to the narrowest specialist.

Parallel work:

- The planner owns the `Parallelization Plan`. It must name each workstream, scenario IDs, owned files, off-limits files, dependencies, tests/checks, `parallelSafe`, `worktree`, and merge order.
- The saved `kumite-loop` chain is the serial-safe fallback. It should follow the planner's merge order when multiple workstreams exist.
- The orchestrator may dynamically fan out implementers only for workstreams marked `parallelSafe: true` with clear file ownership and no unresolved dependencies.
- Parallel implementer calls should use `worktree: true` only when the planner marked that workstream safe for a separate worktree.
- The reviewer owns the final serial verdict over the merged result and must verify workstream ownership, off-limits files, merge order, and integration before passing the task.

Memory and context roles:

- Markdown files under `.kumite/memory/` are canonical project memory.
- Memo is the local MCP memory/index layer for curated decisions, prior handoffs, project knowledge, and code intelligence. Use Memo to retrieve or save compact supporting memory; do not let it replace the markdown files.
- context-mode is for automatic routing, context-saving tool execution, indexed lookup, and session snapshots when available.
- Verify context-mode after setup/restart with `ctx stats`. In kumite agents, use `ctx_stats` to confirm availability before full loops and reviews.
- Prefer `ctx_batch_execute`, `ctx_execute`, `ctx_execute_file`, `ctx_index`, `ctx_search`, and `ctx_fetch_and_index` for broad searches, large command output, large local files, and reusable external docs. Use ordinary tools for small focused reads/edits.
- Do not use context-mode or Memo as canonical memory; `.kumite/memory/*.md` remains the source of truth. Context-mode is an index/search and context-saving aid; Memo is searchable supporting memory and project/code intelligence.

Coordination:

- Use `todo` for visible orchestration state. Include owners, dependencies, and artifact paths in task metadata.
- Child agents may use `contact_supervisor` through pi-intercom only for blocked decisions, structured interview requests, or meaningful progress updates that change the plan.
- Do not use intercom for routine status that can be returned in the child result.
<!-- kumite:end -->
