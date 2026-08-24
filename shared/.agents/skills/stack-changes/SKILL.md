---
name: stack-changes
description: >
  Splits a large change into small, single-purpose units of review (stacked branches) — each making one argument (one diff, one thesis) and each buildable and testable on its own.
  The original branch is left completely untouched.
  Triggers on "split this PR", "break up this branch", "stack these changes", "one diff one thesis".
---

# Splitting Changes Into Stacked Branches

## Overview

### One diff, one thesis

Each change should make exactly one argument, and that argument should be buildable and testable by itself.
If a reviewer can't state the change's thesis in one sentence, it's carrying more than one.

This skill is about decomposition: given a large feature branch, how do you carve it into a stack of dependent branches where every change stands on its own.
The core move is finding the **natural seams** between theses.
The most common seam is **refactor-first**: the restructuring ships _before_ the feature, so the feature itself lands as a small, obvious diff.

### Core Rules for this Skill

1. **Plain Git Only:** No third-party stacking tools or custom scripts are used.
2. **Non-Destructive:** The original branch is **never** rebased, reset, or modified.
   All stacked changes are built on entirely new branches.
3. **Semantic Commits:** Every commit must use standard semantic commit formatting (e.g., `feat:`, `refactor:`, `fix:`).

## Procedure

1. **Inspect local context** — analyze the change shape on the current branch, recent history, merge base, and build/test hints.
   See [Inspect Local Context](#inspect-local-context-before-you-plan).
   If the user provides a base branch (e.g. `feat/some-branch`), use that instead of assuming `main`.
2. **Check safety preconditions** — ensure the working tree has no dirty tracked files before branch creation.
   If tracked files are dirty, stop and ask.
   Ignored/untracked local plan files are OK.
3. **Emit a Split Plan** — propose the stacked branches and commit messages, ordered refactor-first, and persist it to `STACK_PLAN.md`.
4. **Confirm the plan with the user _before touching git._** Tell the user they may edit branch names and commit messages in the plan.
   Revise until they agree on the shape and naming.
5. **Execute the mechanics** — build the new branches sequentially using plain git, migrating changes from the original branch while leaving it untouched.
   See [Execution Mechanics](#execution-mechanics-plain-git).
6. **Verify** — run the per-branch loop (checkout → build → test) to ensure every step in the stack works independently.
   Respect repository instructions that forbid or alter verification commands.
7. **Communicate the dependency** — write PR description templates mapping the stack dependency.

## When to Use

- A change is over ~400 lines, or touches many files for more than one reason.
- A new feature requires reshaping existing code first.
- A change mixes concerns: refactor + feature, behavior change + formatting.
- You need to land dependent changes and want the ordering to be unambiguous.

**When NOT to use:** A genuinely atomic, single-purpose change. Don't split a 30-line bug fix into a three-change stack.

## The One-Thing Rule (One Diff, One Thesis)

A change carries exactly one thesis when it passes these tests:

- **Statable in one sentence** — if the title needs an "and", it's two changes.
- **Understandable alone** — a reviewer gets the point without reading the rest of the stack.
- **Buildable alone** — checked out on its own, the revision compiles.
- **Testable alone** — tests pass independently at its position in the stack.

## The Decomposition Process

1. **Map the end state:** What does the final diff actually touch?
2. **Find the seams:** Split by thesis (refactor vs. behavior, independent vs. dependent, ownership boundaries).
   A **pure refactor** changes structure but not behavior (existing tests pass).
   A **behavior change** adds or alters what the code does (requires new/modified tests).
   Never mix them in one branch.
3. **Order by dependency:** Sort changes so each one only depends on changes before it.
   Refactors come first.
4. **Size & scope:** Target 100–400 lines per change with a coherent thesis.

## Inspect Local Context (before you plan)

Gather repo context using standard git commands.

1. **Current change shape:**
   Use the user-provided base branch when present.
   Otherwise identify the trunk (e.g. `main` or `master`) and use it as `<base>`.

   ```sh
   git status --short --branch
   git diff <base>...HEAD --stat
   git diff <base>...HEAD --name-only
   ```

2. **Recent history & base:**
   Find where the branch roots.

   ```sh
   git log --oneline -n 10
   git merge-base HEAD <base>
   ```

3. **Branch-name collision check:**
   Before proposing or creating branches, inspect existing stack branches.
   If proposed names already exist, ask whether to rename or reuse.
   Do not overwrite existing branches.

   ```sh
   git branch --list 'stack/*'
   ```

4. **Build/test hints:**
   Identify the manifest (`package.json`, `Cargo.toml`, `go.mod`, etc.) and read repository instructions to understand how to verify each branch.
   If repository instructions forbid a command (e.g. tests), do not run it; choose an allowed proof such as build/lint if available.

## The Split Plan

Before creating branches, emit a **Split Plan** in markdown and save it to `STACK_PLAN.md`.
Keep it local-only by adding `STACK_PLAN.md` to `.git/info/exclude`; do not edit tracked `.gitignore` unless the user explicitly approves.
Wait for user approval.

The plan is intentionally editable.
Ask the user to review and change proposed branch names and commit messages before execution.
The executed stack must follow the approved names exactly.

```markdown
## Split Plan

### Final goal

<one sentence: what the whole change accomplishes>

### Proposed stack

| #   | Branch Name          | Commit Message                 | Semantic Type | Depends on           | Files      | Test proof               |
| --- | -------------------- | ------------------------------ | ------------- | -------------------- | ---------- | ------------------------ |
| 1   | stack/1-extract-seam | refactor: extract routing seam | refactor      | <base>               | a.ts, b.ts | existing tests unchanged |
| 2   | stack/2-add-feature  | feat: add feature              | feat          | stack/1-extract-seam | c.ts       | new unit test            |
```

## Execution Mechanics (Plain Git)

Once the user approves the plan, generate the stack **without modifying the original branch**.

Before creating branches, run `git status --short --branch` again.
If there are dirty tracked files, stop and ask. Also re-check branch-name collisions.

If the original branch already contains clean semantic commits that exactly match the approved stack, prefer cherry-picking those commits onto the new stack branches.
If the commits mix concerns, or if the approved branch/commit names differ from the originals, migrate by file/hunk and create fresh commits with the approved messages.

For each step in the plan:

1.  **Checkout the correct base:**
    For step 1, base it on the approved base branch (`<base>`).
    For step N, base it on step N-1.

    ```sh
    git checkout -b stack/1-extract-seam <base>
    ```

2.  **Bring over the targeted changes:**
    Check out specific files or apply specific hunks from the original branch.
    _(Note: For partial file changes, use your file-editing capabilities to write the specific lines needed for this step, based on the original diff)._

    ```sh
    # To bring over an entire file from the original branch
    git checkout original-branch-name -- path/to/file.ts
    ```

3.  **Commit with the approved Semantic Message:**
    Ensure the commit message exactly matches the approved plan and strictly follows semantic conventions (e.g., `feat:`, `refactor:`, `fix:`, `chore:`).

    ```sh
    git add .
    git commit -m "refactor: extract routing seam"
    ```

4.  **Repeat for the next branch in the stack:**

    ```sh
    git checkout -b stack/2-add-feature stack/1-extract-seam
    # Migrate next set of changes, commit, etc.
    ```

By doing this, the original branch acts as a read-only reference, entirely preserving the user's initial, uncommitted, or heavily-committed work as a backup.

## Verify the Stack (per-revision)

Do not assume branches work; observe them working.

1. **Walk the stack bottom-up:**

   ```sh
   git checkout stack/1-extract-seam
   <build command>   # e.g., npm run build
   <test command>    # e.g., npm test, only if allowed by repo instructions
   ```

If verification triggers dependency or toolchain downloads and network is restricted, stop when it fails and ask the user to unblock network or provide an offline cache.

2. **Stop at the first red:**
   If a branch fails to build/test, the stack is invalid (likely missing a dependency from a later branch).
   Stop, inform the user, fix the decomposition, and re-verify.

3. **Summarize results:**
   Report each branch name, commit hash, actual commit title, base/dependency, verification result, current branch, and explicitly confirm the original branch was untouched.

## Verification Checklist

Before finalizing the skill execution, confirm:

- [ ] Original branch is completely untouched and intact.
- [ ] New branches map exactly to the approved Split Plan, including user-edited branch names.
- [ ] Every commit uses the approved semantic commit message.
- [ ] No change mixes a refactor with a behavior change.
- [ ] Each branch builds and passes tests independently, or uses the strongest allowed repo-specific verification.
- [ ] Existing branch-name collisions were checked and handled without overwrite.
- [ ] Final summary includes branch names, commit hashes, actual commit titles, dependencies, verification results, and current branch.
