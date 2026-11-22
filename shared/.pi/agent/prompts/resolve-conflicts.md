---
description: Resolve Git conflicts during a merge, rebase, cherry-pick, or restack
---

# Resolve Git conflicts

Preserve the intent of both changes, not just the text on one side.
A restack may replay several commits. Resolve the current stopped step against its current base; do not assume the final stack state is already present.

## Inspect before editing

1. Read the repository instructions and check the current working directory and write limits.
2. Run `git status` and `git diff --name-only --diff-filter=U` to identify the operation and unmerged paths.
3. Check for unrelated staged, unstaged, and untracked work. Preserve it. If you cannot safely distinguish it from the conflict, stop and ask.
4. For each conflict, read the surrounding code, relevant tests, and the unmerged versions. Use `git ls-files -u` and, where present, `git show :1:<path>`, `git show :2:<path>`, and `git show :3:<path>` for the common ancestor and both sides. Some stages may be absent for add/delete conflicts.
5. Inspect relevant history and the commit being replayed when needed. For a rebase, `git rebase --show-current-patch` can show the current patch. Do not assume the branch name or base from memory.

During a rebase, Git's `ours` usually means the rebased-onto side, and `theirs` means the commit being replayed. This is the reverse of what many people expect. Check the actual contents and operation before choosing either side. A restack tool may manage several operations; do not assume its state or commands match plain Git.

## Resolve and verify

1. Determine what each side intended and how the current code should behave. Consider whether an upstream change already implements the replayed change.
2. Edit only the files needed for the resolution. Integrate compatible changes. Do not blindly accept an entire side or restore a deleted file without understanding why it was deleted.
3. For rename/delete, binary, generated, or lockfile conflicts, inspect the available stages and project conventions. Regenerate files only with an appropriate, permitted tool. Ask when the intended result is unclear.
4. Check the result for conflict markers and review the diff in context. Add or adjust focused tests when needed. Run relevant checks if permitted, and report any that cannot run. A search for marker text alone cannot prove that a conflict is resolved.
5. Check `git status` and inspect the resolution again. Unmerged entries can remain until the result is staged.

## Git operation boundaries

Follow the current repository's Git policy and the user's explicit authorization. A request to resolve conflicts does not by itself authorize staging, continuing a rebase or restack, skipping commits, aborting, resetting, or rewriting history when the policy requires explicit permission.

If staging is authorized, stage only resolved paths after checking the result. Do not use `git add -A` or stage unrelated work. Verify that no unmerged entries remain. If continuation is also explicitly authorized, use the command appropriate to the active operation or stacking tool. Stop at each new conflict and repeat this process. Never skip or abort without explicit approval.

Report the files resolved, the decisions made, checks run, remaining unmerged paths, and whether the operation is still paused. If a decision needs product context or either side's intent is uncertain, explain the options and ask rather than guessing.
