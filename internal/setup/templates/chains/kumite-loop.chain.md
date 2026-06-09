---
name: kumite-loop
description: Non-interactive fallback chain that scouts, plans, implements once, and reviews once with file-backed handoffs.
---

## scout
phase: Context
label: Gather project context
as: scout_context
output: scout-context.md
outputMode: file-only
progress: true

This is a bounded saved-chain scout step. Do not read files, search, use context-mode, use Memo, or inspect source.

Immediately write a compact handoff to `.kumite/handoffs/current/scout-context.md`, then return the same handoff. The handoff must start with this exact original task:

{task}

Include these canonical memory paths and required planner sections:

- `.kumite/memory/architecture.md`: metadata, system overview, architecture rules, important flows.
- `.kumite/memory/code-standards.md`: general standards, language standards, review gates.
- `.kumite/memory/business-rules.md`: glossary first, then task-related rules.
- `.kumite/memory/project-status.md`: current state, active decisions, known risks, next steps.

If the task is a Go project and no exact paths are known, tell planner to assign implementer to inspect `go.mod`, `*.go`, and `*_test.go` under the implementation reconnaissance budget.

Report context-mode and Memo status as `not used by scout; planner should stay file-backed and may use one Memo retrieval pass when prior decisions/project knowledge matter; reviewer should use ctx_stats/ctx_batch_execute for high-output verification and Memo code/knowledge tools when useful`.

Return only: original task, memory paths, likely source/test path hints, risks, context-mode status, Memo status, durable handoff mirror status, and planner questions needed for the next step.

## planner-fallback
phase: Planning
label: Produce compact saved-chain plan
as: plan
reads: scout-context.md
output: plan-handoff.md
outputMode: file-only
progress: true

Produce compact saved-chain planning artifacts for the original task. This is a noninteractive fallback planner step; prefer completion over exhaustive planning.

The original user task for this chain step is: {task}

Do not read any files in this fallback planner step. Use only the inline original task above and the canonical memory paths below. The interactive planner handles richer memory-backed planning; this saved-chain fallback exists to keep unattended runs moving.

- `.kumite/memory/architecture.md`
- `.kumite/memory/code-standards.md`
- `.kumite/memory/business-rules.md`
- `.kumite/memory/project-status.md`

Do not use context-mode, Memo, web tools, supervisor/intercom, user-question tools, or read tools in this fallback planner step. Do not read scout handoff files, memory files, source files, test files, parser/evaluator/runtime files, package manifests, repository docs, run tests, run static-analysis commands, run package graph tools, or perform source reconnaissance in this planner step. If exact paths are unknown, record module-level ownership and make implementer resolve exact files under the implementation reconnaissance budget.

Do not use Memo in this fallback planner step. Record `Memo status: not used in noninteractive fallback planner`.

Immediately compose one compact handoff that starts with `STATUS: PLANNED`, then write it to:

- `.kumite/handoffs/current/plan-handoff.md`

The handoff itself must include compact `Spec Plan` and `Gherkin Scenarios` sections. Each Gherkin scenario must include a `Why:` field.

Return the same plan handoff content as the final answer so pi-subagents can save `plan-handoff.md` for the chain runner. Do not continue drafting privately after that one write. Do not write `.kumite/plans/current-kumite-loop-plan.md` or `.kumite/plans/current-kumite-loop.feature.md` in this fallback step. If planning cannot complete, still write `.kumite/handoffs/current/plan-handoff.md` and return a `STATUS: BLOCKED` handoff with the exact missing decision or file.

Run a noninteractive equivalent of the `kumite-grill-with-docs` planning protocol in compact form. Challenge the draft plan against architecture, code standards, business rules, project status, glossary, tests, data model, failure modes, rollout, and edge cases. Do not block for clarification unless the task is impossible to plan. Instead, record `Deferred questions` and conservative assumptions. If no blocking question is needed, record `No blocking questions` with the reason. The plan handoff must include `Grill questions and gaps`.

Include a `Parallelization Plan` even if the result is `serial-only`.

Include an `Execution Plan` that states the recommended number of implementer runs, whether work should run serially or in parallel, whether worktrees are useful, and how many reviewer rounds are justified by risk.

Return a handoff that includes `STATUS: PLANNED`, artifact paths, scenario IDs, safe parallel workstreams, merge order, owned files, off-limits files, test strategy, execution plan, memory documents used as `deferred to implementer/reviewer in noninteractive fallback`, `Memo status: not used in noninteractive fallback planner`, `Context-mode status: not used in noninteractive fallback planner`, durable handoff mirror status, and any user decisions still needed.

## implementer
phase: Implementation
label: Implement planned scenarios
as: implementation
reads: plan-handoff.md
output: implementation-summary.md
outputMode: file-only
progress: true

Implement the approved plan for the original task described in `scout-context.md` and `plan-handoff.md`; use `.kumite/handoffs/current/scout-context.md` and `.kumite/handoffs/current/plan-handoff.md` as fallbacks if the chain-run files are unavailable.

The original user task for this chain step is: {task}

Before editing, verify the plan and scout handoff match the original task above. If they describe another task, write `implementation-summary.md` and `.kumite/handoffs/current/implementation-summary.md` with `STATUS: BLOCKED`, the conflicting artifact path, and no source edits.

Use TDD, read the plan handoff, make the needed code and test changes, and keep edits scoped to assigned scenarios.

After reading the plan handoff and task, use at most six targeted source-inspection tool calls before the first test edit. If the right files are still unclear, write `implementation-summary.md` with `STATUS: BLOCKED`, the exact missing path/decision, and no source edits. Do not continue broad reconnaissance after this budget.

This saved chain is executing the first serial implementation pass only. If the plan lists multiple workstreams, implement the planner's serial-safe workstreams in merge order and respect every workstream's owned files and off-limits files. If the plan requires parallel work or more implementer passes than this saved chain can provide, implement the first safe slice and report the remaining workstreams as not run.

Record changed files, tests, commands, workstream IDs completed, and unresolved assumptions. Mirror the implementation summary to `.kumite/handoffs/current/implementation-summary.md` before returning.

## reviewer
phase: Review
label: Verify implementation round 1
as: review_round_1
reads: scout-context.md, plan-handoff.md, implementation-summary.md
outputMode: file-only
progress: true
output: review-summary-round-1.md

Review the implementation for the original task described in `scout-context.md` and `plan-handoff.md`; use `.kumite/handoffs/current/scout-context.md`, `.kumite/handoffs/current/plan-handoff.md`, and `.kumite/handoffs/current/implementation-summary.md` as fallbacks if the chain-run files are unavailable.

The original user task for this chain step is: {task}

Before reviewing implementation quality, verify that scout, plan, Gherkin, and implementation summaries match the original task above. If any required artifact describes another task, return `STATUS: REWORK_REQUIRED` with a `REWORK_TASK` to regenerate the mismatched planning/implementation artifact for the original task; do not pass a mismatched task.

Compare the code changes against the plan, Gherkin scenarios, code standards, tests, and static-analysis requirements. Use the reviewer agent's static-analysis-reviewer skill.

Use context-mode when available for high-output test/static-analysis commands and include `ctx_stats` status in the review summary.

Use Memo when available to check project knowledge, prior similar fixes, related tests, or code impact only when it reduces broad manual searching. Include Memo status and any Memo evidence used in the review summary.

If the plan includes parallel workstreams, verify ownership, off-limits files, merge order, and cross-workstream integration even when this saved chain implemented the work serially.

Use a bounded review protocol: read the handoffs, inspect the diff, compare plan/Gherkin, run one verification pass, then use at most three manual behavior probes. If the project has no `.git` directory, record Git diff checks as `SKIPPED: no git worktree`, run the remaining verification commands, and do not block solely on missing Git metadata. After that, immediately write `review-summary-round-1.md` and mirror it to `.kumite/handoffs/current/review-summary-round-1.md`. Do not continue open-ended exploratory review.

Return a strict review contract with:

- `STATUS: PASS`, `STATUS: REWORK_REQUIRED`, or `STATUS: BLOCKED`
- `REWORK_TASK:` with a narrow implementer task when status is `REWORK_REQUIRED`; otherwise `REWORK_TASK: none`
- `FOLLOW_UP_PLAN:` with any additional implementer/reviewer runs needed according to the planner's execution plan
- commands run and results
- manual verification steps
