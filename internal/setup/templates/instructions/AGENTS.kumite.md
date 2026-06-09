<!-- kumite:begin -->
# Kumite Agent Index

This file is the primary entry point for agents working on this project. It should stay concise, curated, and useful as an index. Do not duplicate the deeper memory files here; link to them and summarize only the routing information an agent needs to find the right context.

## Start Here

1. At the start of each new task, ask the user whether they want the full Kumite workflow or a normal Pi workflow for a faster fix. If the user already asked for Kumite, subagents, planning, review, or a chain, treat that as choosing Kumite.
2. Use this file to decide which project memory files to read.
3. Read only the memory sections and source files relevant to the task.
4. Keep durable knowledge in the project wiki under `.kumite/memory/`, not in chat transcripts, Memo alone, or generated handoff files.

## Project Wiki Map

- `agents.md`: this curated agent index and workflow entry point.
- `.kumite/memory/architecture.md`: overall architecture plus the `Architecture Documentation Index` of domain-specific architecture docs.
- `.kumite/memory/business-rules.md`: glossary, `Business Documentation Index`, product/domain rule index, and links to domain-specific business memory.
- `.kumite/memory/code-standards.md`: coding, testing, review, and dependency standards. This file requires explicit user approval before changes.
- `.kumite/memory/project-status.md`: durable project status, completed work, active decisions, risks, and next steps.
- `.kumite/plans/`: dated spec plans and Gherkin acceptance artifacts.
- `.kumite/handoffs/current/`: latest scout, planner, implementer, reviewer, and rework handoffs.

## Memory Growth Rules

- Documentation should grow only when information is stable, relevant, and useful for future retrieval.
- Every feature, question, or request may reveal useful knowledge, but not every detail belongs in memory.
- Prefer domain-specific memory files when details would bloat a top-level memory file.
- Keep top-level memory files as curated indexes plus high-level summaries.
- Do not create duplicate standards files. Use the established standards file, currently `.kumite/memory/code-standards.md`.

## Architecture Memory

- Start with `.kumite/memory/architecture.md`.
- That file should describe the overall architecture and maintain an index of architecture docs by project domain.
- Domain-specific architecture details should live under `.kumite/memory/architecture/<domain>/`, for example `.kumite/memory/architecture/login/login-architecture.md`.
- Add a domain architecture file only when the domain has stable architecture details worth retrieving later.

## Business And Product Memory

- Start with `.kumite/memory/business-rules.md`.
- The glossary in that file should be treated as high-priority context for planning.
- Domain-specific business/product details should live under `.kumite/memory/business-rules/<domain>/`, for example `.kumite/memory/business-rules/billing/billing-rules.md`.
- Add domain business files only for stable terminology, rules, policies, workflows, or product decisions that future agents need.

## Interactive Kumite Flow

- Prefer the dynamic orchestrator flow for interactive sessions so planner questions and plan-driven parallelism can happen before implementation.
- After the user chooses the full Kumite workflow, delegate to scout quickly. `subagent list` is forbidden before scout; Pi already knows project agents from `.pi/agents/`. Do not save a Memo prompt and do not do broad parent-session reconnaissance before scout. At most confirm context-mode/Memo availability and create a minimal visible todo state, then run scout.
- Run scout, then launch planner in `grill-discovery` mode before writing final plan/Gherkin artifacts. When delegating scout, require the scout handoff mirror at `.kumite/handoffs/current/scout-context.md`; this handoff write is allowed and is not a source-code modification.
- The planner must use the `kumite-grill-with-docs` protocol to discover plan gaps from the scout handoff, project memory, and targeted project context. In `grill-discovery` mode, planner must return `STATUS: DRAFT_PLAN_FOR_GRILL` with a compact draft plan plus concrete grill questions or assumptions to confirm. It must not write final spec or Gherkin artifacts in this mode.
- Ask the draft-plan grill questions in the parent session with `ask_user_question`, including whether to accept, refine, or reject the draft plan. Pass the answers back to planner. Repeat this planner grill/refine loop while planner returns `STATUS: DRAFT_PLAN_FOR_GRILL` or `STATUS: NEED_USER_INPUT`.
- If running agents manually, the planner phase must use the `kumite-grill-with-docs` protocol before final plan/Gherkin writing and before implementation starts.
- If blocking planning questions are cancelled, declined, or left unanswered, stop with `STATUS: BLOCKED` and the unanswered questions. Do not write final plan/Gherkin artifacts, continue to implementation, or review work that depends on those answers.
- The planner must produce an `Execution Plan` and `Parallelization Plan` that specify implementer count, reviewer rounds, workstream ownership, safe parallelism, worktree use, and merge order.
- The orchestrator must follow the planner's execution plan. Do not run fixed extra reviewers or implementers after a pass unless the planner/reviewer evidence requires it.
- After any child subagent returns a completed result, immediately inspect its status/artifact, mark the corresponding todo complete, and start the next planned step. Ignore stale `subagent needs attention` notifications for runs that already report `Status: completed`; do not wait for additional child activity after `STATUS: COMPLETE`, `STATUS: IMPLEMENTED`, `STATUS: PASS`, `STATUS: REWORK_REQUIRED`, or `STATUS: BLOCKED`.
- After child completion, call `kumite_next_step` when available. If it reports `agent: reviewer` after an implementer handoff, launch reviewer immediately instead of waiting or continuing to reason about the completed implementer state.
- Child agents must not leave the parent blocked by waiting on intercom/contact-supervisor. If an implementer or rework agent needs a decision, it must return `STATUS: BLOCKED` or `STATUS: PARTIAL` inline with the exact question and handoff path so the parent can ask the user and relaunch.
- When reviewer returns `STATUS: REWORK_REQUIRED`, the next implementer run must consume only the reviewer's `REWORK_TASK`, then reviewer must run again if the planner's review budget allows it.

## Saved Chain

- `/run-chain kumite-loop -- <task>` is the noninteractive fallback for unattended work. It may defer planner questions because saved-chain subagents cannot reliably interview the user mid-step.
- If the user explicitly says to use `kumite-loop`, start `/run-chain kumite-loop -- <task>` or execute the saved chain steps through `subagent`.
- In noninteractive/tool-driven sessions, saved chains are not executable agents. Do not call `subagent({agent: "kumite-loop"})`.
- For direct tool execution, inspect the saved chain with `subagent({action: "get", chainName: "kumite-loop", agentScope: "both"})` if needed. Before executing it, replace every literal `{task}` in every returned step task with the actual user task. Do not pass a chain containing `{task}` placeholders to `subagent`.
- After placeholder replacement, execute the concrete steps as `subagent({chain: [...], task: "<task>", clarify: false, agentScope: "both"})` unless the user explicitly requests a timeout.
- The scout handoff must include an `Original task` section. Later steps must treat `scout-context.md` as the durable task artifact; do not wait for a separate `task.md` unless the user provided one.
- Saved-chain child outputs may be stored in a pi-subagents temporary chain-run directory. Kumite agents must also mirror handoffs into `.kumite/handoffs/current/` so the project keeps durable copies across sessions.
- In the interactive dynamic flow, do not use `output`, `outputMode: "file-only"`, or file-only routing for implementer subagent calls. The implementer should write its durable handoff file itself and return a compact inline result so the parent can immediately continue to reviewer.

## Parallel Work

- The planner owns the `Parallelization Plan`. It must name each workstream, scenario IDs, owned files, off-limits files, dependencies, tests/checks, `parallelSafe`, `worktree`, and merge order.
- The saved `kumite-loop` chain is the noninteractive serial fallback. It should follow the planner's merge order when multiple workstreams exist.
- The orchestrator may dynamically fan out implementers only for workstreams marked `parallelSafe: true` with clear file ownership and no unresolved dependencies.
- Parallel implementer calls should use `worktree: true` only when the planner marked that workstream safe for a separate worktree.
- The reviewer owns the final serial verdict over the merged result and must verify workstream ownership, off-limits files, merge order, and integration before passing the task.

## Curation Rules

- Curator updates happen only after the user has classified the feature as fully ready.
- Before updating memory, curator must review the proposed plan/spec file, reviewer response, and final implementation context.
- Curator decides which project memory files should be updated: `agents.md`, `.kumite/memory/architecture.md`, `.kumite/memory/business-rules.md`, domain-specific architecture or business files, `.kumite/memory/project-status.md`, or other relevant Kumi memory files.
- `.kumite/memory/code-standards.md` requires explicit user approval before any change. If standards should change, curator must propose the change and ask for approval before editing.
- Curator should update indexes when adding or moving domain-specific memory files.

## Context And Memo

- Markdown files under `.kumite/memory/` are canonical project memory.
- Memo is the local MCP memory/index layer for curated decisions, prior handoffs, project knowledge, and code intelligence. Use Memo to retrieve or save compact supporting memory; do not let it replace the markdown files.
- context-mode is for automatic routing, context-saving tool execution, indexed lookup, and session snapshots when available.
- Verify context-mode after setup/restart with `ctx stats`. In Kumite agents, use `ctx_stats` to confirm availability before full loops and reviews.
- Prefer `ctx_batch_execute`, `ctx_execute`, `ctx_execute_file`, `ctx_index`, `ctx_search`, and `ctx_fetch_and_index` for broad searches, large command output, large local files, and reusable external docs. Use ordinary tools for small focused reads/edits.

## Coordination

- Use `todo` for visible orchestration state. Include owners, dependencies, and artifact paths in task metadata.
- Child agents must return `STATUS: BLOCKED` or `STATUS: PARTIAL` inline for blocked decisions or meaningful plan-changing findings. Do not leave the parent waiting on `contact_supervisor`/intercom.
- Do not use intercom for routine status that can be returned in the child result.
<!-- kumite:end -->
