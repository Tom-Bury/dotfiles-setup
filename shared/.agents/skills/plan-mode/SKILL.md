---
name: plan-mode
description: >
  Ad-hoc planning workflow for coding agents. Generate a PLAN.md before implementation,
  ask the user to review and edit it, then execute the plan while crossing out completed
  steps. Supports fully automatic execution, stopping after each step, and user-defined
  stopping points in PLAN.md. Use when the user says "plan mode", "make a plan first",
  "write PLAN.md", "review plan before coding", or invokes /plan-mode.
---

Use this skill when work should be planned in a file before execution.

## Core Workflow

1. Create or update `PLAN.md` in the current working directory before making implementation changes.
2. Ask the user to review `PLAN.md` and edit it where needed.
3. Do not start implementation until the user confirms the plan is ready.
4. Execute plan steps in order.
5. After each completed step, edit `PLAN.md` and cross out that step with Markdown strikethrough (`~~step~~`).
6. Keep `PLAN.md` as the source of truth. If scope changes, update it before continuing.

## PLAN.md Format

Use a simple checklist:

```md
# Plan

Mode: automatic | step-by-step

- [ ] Step 1: Describe concrete action.
- [ ] Step 2: Describe concrete action. <!-- STOP -->
- [ ] Step 3: Describe concrete action.
```

Completion format:

```md
- [x] ~~Step 1: Describe concrete action.~~
```

## Execution Modes

Default to `step-by-step` unless the user explicitly asks for automatic execution.

### Step-by-step

- Complete one unchecked step.
- Cross it out in `PLAN.md`.
- Stop and ask the user whether to continue.

### Automatic

- Continue through unchecked steps without asking after each one.
- Still stop at explicit stopping points.
- Still update `PLAN.md` after every completed step.

## Stopping Points

Users may add stopping points in `PLAN.md` before execution.
Treat any of these as a stop marker:

- `<!-- STOP -->`
- `[STOP]`
- `STOP HERE`
- `PAUSE`
- `ASK USER`

When a stop marker is reached:

1. Finish any work for the current step only if the marker is on that step.
2. Cross out completed work.
3. Stop execution.
4. Ask the user for review/approval before continuing.

## Plan Quality Rules

Plans must be:

- Specific enough to execute without reinterpreting intent.
- Ordered by dependency.
- Small enough that each step can be completed and verified independently.
- Honest about unknowns. Include investigation steps when needed.
- Minimal. Do not over-plan trivial work unless the user requested plan mode.

## Safety Rules

- Never overwrite a user-edited `PLAN.md` without reading it first.
- Preserve user-added notes, edits, and stop markers.
- If the implementation reveals the plan is wrong, pause, update `PLAN.md`, and ask for confirmation unless the change is obviously minor.
- If tests or verification fail, add or update a plan step for the fix instead of hiding the failure.

## User Prompts

After writing the initial plan, say:

> `PLAN.md` created. Please review/edit it, then tell me to continue. Choose automatic or step-by-step execution if you care.

After each step in step-by-step mode, say:

> Step complete and crossed out in `PLAN.md`. Continue?
