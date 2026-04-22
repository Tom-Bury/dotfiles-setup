---
name: pr-feedback
description: >
  The current branch has an open PR and has received feedback from reviewers.
  This skill tackled that feedback which is given in JSON format as input.
  Tackling feedback is either:
  - addressing the comment by making a code change, or
  - a code-comment with a justification, clarifying question, or counter-argument etc on why reviewer suggestion isn't followed
  Use when user says "address PR feedback", "tackle PR comments", "handle code review", or invokes /pr-feedback.
---

## Input

Input is a JSON object representing the open PR and its feedback, with the following format:

```json
{
  "metadata": {
    "author": {
      "id": "...",
      "is_bot": false,
      "login": "JohnDoe",
      "name": "John Doe"
    },
    "body": "This is the user-written PR description.",
    "title": "This is the PR title."
  },
  "change_requests": [
    {
      "path": "file/path/where/comment/is/made.txt",
      "line": 123,
      "body": "This is the comment left by the reviewer on a specific line of code.",
      "author": "Alice",
      "good_comment": true,
      "bad_comment": false
    }
  ]
}
```

- `metadata`: contains information about the PR itself, including the author, body, and title.
- `change_requests`: an array of comments left by reviewers on specific lines of code, along with the author of the comment and whether it has received positive or negative reactions.

## Instructions

When this skill is invoked, follow these steps to process each item in `change_requests`.
If no JSON is provided, respond with "No PR feedback found. Please provide the PR feedback JSON. (use `ghprdata` command to get it)".

### 1. Evaluate the Feedback Type

Analyze the `body` and metadata flags to determine the action:

- **Code Change:** Required if `good_comment` is `true`, or if the comment contains 🛑 (Blocker) or 🔧 (Suggestion) that you deem valid.
- **Code Comment (Response):** If a code change is not made, you **must** leave a justification, clarification, or counter-argument in the code at the relevant line as a placeholder for the user to sync back to the PR.
- **Dismissal:** If `bad_comment` is `true`, or if the comment is a 🙃 (Nitpick) or 🍰 (Off-topic), prioritize moving forward without changes unless the fix is trivial.

### 2. Interpret Severity Emojis

Use the following priority matrix to handle comments:

| Emoji            | Category        | Action Required                                                                                                             |
| :--------------- | :-------------- | :-------------------------------------------------------------------------------------------------------------------------- |
| **🛑**           | **Blocker**     | **High.** Must address with a code change or a very strong justification.                                                   |
| **❓**           | **Blocker**     | **High.** Must provide a clarifying response or update code to be self-documenting.                                         |
| **🔧**           | **Suggestion**  | **Medium.** Default to implementing; especially if `good_comment` is true; requires pretty solid justification if skipping. |
| **🤔 / 🙃 / 🍰** | **Non-Blocker** | **Low.** Only address if it doesn't cause "scope creep." Usually safe to ignore. Implement when trivial.                    |
| **👍 / 🎉 / 💯** | **Positive**    | **None.** No action required.                                                                                               |

### 3. Critical Thinking & Autonomy

Do not blindly follow all instructions. You are an autonomous collaborator:

- If a comment is marked `bad_comment`, provide a polite but firm justification for why the change isn't being made.
- Protect against **feature/scope creep**. If a comment suggests a large change, suggest creating a follow-up ticket/PR instead of modifying the current branch.

## Output Format

For each `change_request`, produce either:

1.  **A code diff** applying the requested fix.
2.  **A source code comment** at the specified `line` in the following format:
    `// PR-REPLY (@[Author]): [Your justification/question/counter-argument]`

---

### Example Internal Logic

> _Reviewer Alice left a 🛑 on line 12. `good_comment` is true._
> **Action:** Generate code fix for line 12.
>
> _Reviewer Bob left a 🙃 on line 45. `bad_comment` is true._
> **Action:** Ignore or add a brief justification comment if the user insists on a response.
