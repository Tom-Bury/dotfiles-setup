---
name: fix-ci
description: >
  The current branch has an open PR with failing CI checks.
  This skill fixes CI failures using JSON data provided by the user, typically piped from
  shared/git/gh_ci_status.sh --json.
  Use when user says "fix CI", "CI is failing", "address failing checks", "handle CI failures",
  or invokes /fix-ci.
---

## Input

Input is a JSON object representing failing CI checks for the current PR
Expected format:

```json
{
  "pr": "123",
  "results": [
    {
      "name": "Name of the job",
      "detailsUrl": "https://github.com/.../job/...",
      "runId": "123456789",
      "jobId": "123456789",
      "status": "ok",
      "run": {
        "conclusion": "failure",
        "event": "pull_request",
        "headBranch": "feat/foo",
        "headSha": "...",
        "name": "Name of the workflow",
        "status": "completed",
        "url": "https://github.com/.../actions/runs/...",
        "workflowName": "name of the workflow"
      },
      "logSnippet": "...",
      "logTail": "..."
    }
  ]
}
```

- `pr`: PR number.
- `results`: failing or problematic CI jobs.
- `name`: job name.
- `detailsUrl`: GitHub Actions job URL.
- `runId` / `jobId`: IDs that may be used with `gh run view` or `gh api` when more detail is needed.
- `status`: status of CI data extraction, not necessarily the job result.
- `run`: workflow/run metadata.
- `logSnippet`: failure-focused log excerpt.
- `logTail`: end of job log.

## Instructions

When this skill is invoked, follow these steps:

### 0. Validate the input

Ensure the user provided the CI JSON in the same prompt that invoked this skill.
NEVER assume CI data is available from prior chat history, files, or command output.
If the input is missing or malformed, respond with:

`No CI data found. Please provide CI JSON. (use ghci --json)`

If `results` is empty, respond that no failing CI checks were found and stop.

### 1. Triage failures

For each item in `results`:

- Read `name`, `run.workflowName`, `run.conclusion`, `detailsUrl`, `logSnippet`, and `logTail`.
- Identify the likely failure type: tests, lint, typecheck, formatting, build, dependency/install, flaky infra, timeout, or unknown.
- Prefer `logSnippet` for root cause. Use `logTail` for final error context.
- Group duplicate failures by same error message or same underlying cause.

### 2. Inspect the codebase

Use repository files and local commands to confirm the cause:

- Inspect relevant source, tests, configs, and package scripts.
  Ususally will mean files like `.github/workflows/...` and relevant code/test files.
- Run targeted local checks when possible. Prefer narrow commands over full expensive suites.
- If output may be large, summarize/filter it with a small script rather than dumping logs.

### 3. Fix the root cause

Make code/config/test changes that address the CI failure.
Do not blindly silence failures.

Preferred fixes by category:

- **Test failure:** fix product code if behavior regressed; update tests only when expected behavior intentionally changed.
- **Lint/format:** apply the smallest style-safe change or run the project formatter if clearly available.
- **Typecheck:** preserve type safety; avoid `any`/suppression unless justified.
- **Build:** fix missing imports, exports, env assumptions, bundler/config issues.
- **Dependency/install:** fix lockfile/config/package metadata only when clearly required.
- **Timeout/flaky/infra:** make deterministic fix if code-related; otherwise explain likely external flake and provide rerun guidance.

### 4. Verify

After changes, run the most relevant local verification command(s). If unable to run, explain why.

If CI JSON includes multiple independent failures, address all that share the same root cause. If failures are unrelated and large, fix the highest-signal blocker first and clearly list remaining work.

### 5. Output

Summarize:

- CI job(s) addressed.
- Root cause.
- Files changed.
- Verification command(s) run and result.
- Any remaining CI failures or recommended rerun.

## Network and GitHub access

Use the provided JSON first. Do not fetch logs from GitHub unless the supplied snippets are insufficient and network/GitHub access is available.
If more logs are required but access is blocked or unavailable, ask the user to rerun `ghci --json --max-lines <N>` with a larger value and paste the output.
