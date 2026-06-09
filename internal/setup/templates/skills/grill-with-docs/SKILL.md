---
name: kumite-grill-with-docs
description: Use during kumite planning to challenge a requested change against project memory, existing code, terminology, and documented decisions before writing implementation and Gherkin plan artifacts.
---

# Grill With Docs

Use this skill while planning non-trivial changes. Your job is to make the plan precise before implementation starts.

## Core Protocol

- Read the scout handoff, user task, and relevant project memory first.
- In interactive Kumite planning, run the grill discovery before writing final spec or Gherkin artifacts. Send a visible draft plan to the orchestrator first using `STATUS: DRAFT_PLAN_FOR_GRILL`; do not write `.kumite/plans/*`, `plan-handoff.md`, or `.kumite/handoffs/current/plan-handoff.md` until the user has accepted or refined that draft through the parent session.
- If a question can be answered by inspecting code or docs, inspect them instead of asking the user.
- Challenge the requested plan against the project's existing architecture, terminology, business rules, code standards, and status.
- In interactive Kumite planning, return planner-generated grill questions after reading memory and targeted project context. Do not treat one coarse pre-planner answer as the complete grill pass when the code/docs reveal deeper ambiguity.
- Ask deeper focused questions when behavior, scope, acceptance criteria, terminology, or risk cannot be resolved from the repository.
- Ask one decision at a time unless the available `ask_user_question` tool is clearly better for a small structured batch.
- For each question, include a recommended answer and the tradeoff behind it.
- If no blocking question is needed in an interactive run, ask the user to confirm the conservative assumptions that affect implementation or tests as part of `STATUS: DRAFT_PLAN_FOR_GRILL`. `No blocking questions` is valid only in non-interactive fallback after the draft-plan checkpoint has been accepted, or when the user already answered the relevant decision in the current task. Do not silently skip the grill pass.
- Do not wait indefinitely on intercom for routine planning questions. If a child planner cannot ask the user directly, return `STATUS: NEED_USER_INPUT` with exact questions, recommended answers, alternatives, tradeoffs, and why each answer affects the plan.

## Bounded Chain Use

When this skill is used by the kumite planner inside `kumite-loop`, stay artifact-first:

- Use the scout handoff as the primary source of code context.
- Do not perform broad code reconnaissance, tests, static analysis, package graph checks, or implementation experiments.
- Use only short targeted reads needed to resolve planning facts that scout omitted.
- If the user cannot answer during a non-interactive chain run, record `Deferred questions` with conservative assumptions instead of blocking unless the task cannot be planned.
- In non-interactive saved-chain runs, write the spec draft, Gherkin artifact, and plan handoff with deferred questions and conservative assumptions so unattended work can continue. This exception does not apply to the interactive dynamic workflow.

## Documentation Awareness

Prefer kumite memory docs when present:

- `.kumite/memory/architecture.md`
- `.kumite/memory/code-standards.md`
- `.kumite/memory/business-rules.md`
- `.kumite/memory/project-status.md`

Also check conventional project docs when they exist:

- `CONTEXT.md`
- `CONTEXT-MAP.md`
- `docs/adr/`
- package, module, or app-specific docs near the files being changed

When context-mode tools are available, use `ctx_search` for already indexed docs and `ctx_execute_file` for targeted extraction from large memory or reference files. Use normal reads for short files where full content is useful.

When Memo tools are available, use `memo_current_project` before relying on Memo data. Use `memo_recall`, `memo_knowledge_pack`, or `memo_knowledge_search` only when prior decisions, project history, or project knowledge could affect scope, acceptance criteria, terminology, or architecture constraints. Do not let Memo replace the `.kumite/memory/*.md` documents.

## Challenge Rules

- If the user uses a term that conflicts with the project glossary or code, call out the mismatch and ask which meaning should win.
- If language is fuzzy, propose a precise project term.
- Probe concrete scenarios and edge cases that affect implementation boundaries.
- Cross-check product claims against the code. If the code contradicts the request, surface that before planning.
- Do not write broad documentation during planning. The curator owns project memory document updates after user confirmation.

## Planning Output

After unresolved questions are answered or explicitly deferred, produce the normal kumite planner artifacts:

- A spec-driven plan under `.kumite/plans/`.
- A Gherkin-style artifact under `.kumite/plans/` with `Feature`, optional `Rule`, scenarios, and explicit `Why:` fields.
- A `Memory documents used` section in the spec plan and handoff. For each relevant memory file, state whether it was seed/empty or project-specific and name the constraint, glossary term, status item, or architecture rule it contributed.
- A `Grill questions and gaps` section in the spec plan and handoff. It must include `Questions asked`, `Deferred questions`, or `No blocking questions`.
- A handoff summary listing artifact paths, scenario IDs, implementation workstreams, test strategy, memory documents used, risks, and unresolved assumptions.

Before unresolved questions are answered and the draft plan is accepted/refined in the interactive dynamic workflow, return only a grill-discovery response:

- `STATUS: DRAFT_PLAN_FOR_GRILL`
- `Mode: grill-discovery`
- `Context read`
- `Draft plan` with proposed scope, non-goals, data/model/API or file ownership shape, test strategy, risks, assumptions, and likely execution/parallelization approach
- `Questions` with recommended answers, alternatives, tradeoffs, and why each answer matters

Do not write final plan or Gherkin artifacts in that state. The only allowed file write in this state is `.kumite/handoffs/current/draft-plan-for-grill.md`, clearly labeled as non-final.
