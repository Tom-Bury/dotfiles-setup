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
    [
      {
        "path": "file/path/where/comment/is/made.txt",
        "line": 123,
        "body": "This is the comment left by the reviewer on a specific line of code.",
        "author": "Alice",
        "good_comment": true,
        "bad_comment": false
      },
      {
        "path": "file/path/where/comment/is/made.txt",
        "line": 123,
        "body": "This comment is a response to Alice's comment; so it's in the same thread.",
        "author": "Bob",
        "good_comment": false,
        "bad_comment": false
      }
    ],
    [
      {
        "path": "file/path/where/comment/is/made.txt",
        "line": 45,
        "body": "This is another comment left by a different reviewer on a different line.",
        "author": "Charlie",
        "good_comment": false,
        "bad_comment": true
      }
    ]
  ]
}
```

- `metadata`: contains information about the PR itself, including the author, body, and title.
- `change_requests`: an array of comment threads left by reviewers on specific lines of code, along with the author of the comment and whether it has received positive or negative reactions.
  When a thread has multiple comments, it means there was discussion.
  Each comment in the thread should be evaluated to determine the appropriate action.
  When the `line` is `null`, it means the comment was made on the file in general and not on a specific line of code.

## Instructions

When this skill is invoked, follow these steps:

### 0. Validate the input

Ensure you have been provided with the necessary JSON input by the user.
NEVER assume you have the PR feedback data or look for it in files or history.
It must ALWAYS be provided in the message invoking the skill.
If the input is missing or malformed, respond with "No PR feedback found. Please provide the PR feedback JSON. (use `ghprdata` command to get it)".

### 1. Evaluate the Feedback Type

Process each comment thread in `change_requests`.
Analyze the `body` and metadata flags for each comment to determine the action:

- **Code Change:** Required when the comment identifies a valid defect or requested change, especially when `good_comment` is `true` or the comment contains 🛑 (Blocker) or 🔧 (Suggestion).
- **Code Comment (Response):** If no code change is made, add a concise justification, clarification, or counter-argument at the relevant line as a placeholder for the user to sync back to the PR. Use the comment syntax appropriate for the target file (for example, `//` for JavaScript, `#` for Python, `<!-- -->` for Markdown, or `/* */` where appropriate).
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

For each `change_request`, produce one of:

1. **A code diff** applying the requested fix.
2. **A source-code comment** at the specified line, using the target file's native comment syntax:
   `[comment marker] PR-REPLY (@[Author]): [Your justification/question/counter-argument]`
3. **No action**, when the feedback is invalid, out of scope, or purely positive. Briefly explain why.

Choose comment syntax from the file type and surrounding code. Keep the reply clearly marked `PR-REPLY` so it can be copied into the PR and removed later. Do not use comment syntax that changes program behavior or breaks formatting.

---

### Example Internal Logic

> _Reviewer Alice left a 🛑 on line 12. `good_comment` is true._
> **Action:** Generate code fix for line 12.
>
> _Reviewer Bob left a 🙃 on line 45. `bad_comment` is true._
> **Action:** Leave code unchanged and add a brief, correctly-syntaxed `PR-REPLY` comment only if a response is useful.
