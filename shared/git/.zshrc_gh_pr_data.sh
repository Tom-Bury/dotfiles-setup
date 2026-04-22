#!/usr/bin/env zsh

ghprdata() {
  # 1. Identity & Validation
  BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null)
  PR_JSON=$(gh pr list --head "$BRANCH" --state open --json number)
  PR_COUNT=$(echo "$PR_JSON" | jq '. | length')

  if [ "$PR_COUNT" -ne 1 ]; then
    echo "Error: Found $PR_COUNT open PRs for branch '$BRANCH'." >&2
    return 1
  fi

  PR_NUMBER=$(echo "$PR_JSON" | jq -r '.[0].number')

  # 2. Construct and Output JSON
  {
    echo '{"metadata":'
    # The --json flag already whitelists what we need for the PR metadata
    gh pr view "$PR_NUMBER" --json title,body,author
    
    echo ',"change_requests":'
    # We use a whitelist jq map to extract ONLY the path, line, body, and author login and reactions
    gh api "repos/:owner/:repo/pulls/$PR_NUMBER/comments" | \
    jq 'map(select(.line != null) | {path: .path, line: .line, body: .body, author: .user.login, good_comment: (.reactions.["+1"] > 0), bad_comment: (.reactions["-1"] > 0) })'
    
    echo '}'
  } | jq .
}

alias ghpr='ghprdata'