<!-- kumite:begin -->
# Kumite Pi Context Bridge

Pi auto-loads `AGENTS.md`, but Kumite's canonical agent index is `agents.md`.

Read `agents.md` first. Treat it as the source of truth for Kumite workflow rules, project memory indexes, curation policy, and agent coordination. Do not duplicate project memory in this file.

Before any code-modifying or planning task, ask the user whether they want the full Kumite workflow or a normal Pi workflow for a faster fix, unless the user already explicitly chose one. Use `ask_user_question` when available. Do not inspect source, call Memo, create todos, or edit files before this choice is answered.

If the user chooses full Kumite, start scout quickly. `subagent list` is forbidden before scout; project agents are already loaded from `.pi/agents/`. Do not save a Memo prompt and do not do parent-session source reconnaissance before scout. Use only minimal setup such as context-mode/Memo availability checks and visible todos, then delegate to scout.

After scout returns, launch planner in `grill-discovery` mode before final plan/Gherkin writing. The planner must use `kumite-grill-with-docs` to inspect the scout handoff, relevant memory, and targeted project context, then return `STATUS: DRAFT_PLAN_FOR_GRILL` with a compact draft plan plus concrete grill questions or assumptions to confirm. Ask those draft-plan grill questions in the parent session with `ask_user_question`, pass the answers back to planner, and repeat until the user accepts/refines the draft and planner writes final artifacts or blocks. Do not continue to implementation while draft-plan grill questions are unanswered.

Child agents must not leave the parent blocked by waiting on intercom/contact-supervisor. If an implementer or rework agent needs a decision, it must return `STATUS: BLOCKED` or `STATUS: PARTIAL` inline with the exact question and handoff path so the parent can ask the user and relaunch.

When delegating scout, do not say plain “do not modify files.” Say “do not modify source, tests, manifests, or memory docs; writing `.kumite/handoffs/current/scout-context.md` is required.”
<!-- kumite:end -->
