## Network policy

You are running in a Docker sandbox with restricted network access.
Network requests might fail. If network access is critical and fails, stop and ask the user to temporarily unblock it. Explain why access is needed.

## Git policy

Never use git in an active or mutating way unless the user explicitly asks in the current conversation.

Forbidden by default:

- creating, deleting, or switching branches
- committing
- pushing
- pulling
- fetching
- rebasing
- merging
- stashing
- resetting
- checking out files or commits
- changing remotes or config

Allowed by default: read-only git inspection, such as `git status`, `git diff`, `git log`, `git show`, and `git branch --show-current`.
