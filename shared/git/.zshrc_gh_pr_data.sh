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
  OWNER_REPO=$(gh repo view --json nameWithOwner -q .nameWithOwner)
  OWNER=$(echo "$OWNER_REPO" | cut -d'/' -f1)
  REPO=$(echo "$OWNER_REPO" | cut -d'/' -f2)

  # 2. Construct and Output JSON
  {
    echo '{"metadata":'
    gh pr view "$PR_NUMBER" --json title,body,author
    
    echo ',"change_requests":'
    # GraphQL to fetch threads, filtered and grouped via jq
    gh api graphql -F owner="$OWNER" -F repo="$REPO" -F pr=$PR_NUMBER -f query='
      query($owner: String!, $repo: String!, $pr: Int!) {
        repository(owner: $owner, name: $repo) {
          pullRequest(number: $pr) {
            reviewThreads(first: 100) {
              nodes {
                isResolved
                comments(first: 50) {
                  nodes {
                    path
                    line
                    body
                    author { login }
                    reactions(content: THUMBS_UP) { totalCount }
                    bad_reactions: reactions(content: THUMBS_DOWN) { totalCount }
                  }
                }
              }
            }
          }
        }
      }' | jq '
        [ .data.repository.pullRequest.reviewThreads.nodes[] 
          | select(.isResolved == false) 
          | [ .comments.nodes[] | {
              path: .path, 
              line: .line, 
              body: .body, 
              author: .author.login, 
              good_comment: (.reactions.totalCount > 0), 
              bad_comment: (.bad_reactions.totalCount > 0) 
            } ]
        ]'
    
    echo '}'
  } | jq .
}

alias ghpr='ghprdata'