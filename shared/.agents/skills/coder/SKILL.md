---
name: coder
description: >
  Workflow for preserving context window space when analyzing large datasets,
  logs, raw HTTP responses, or command output. Instructs the agent to compute,
  count, filter, and summarize data with standalone Node.js scripts, returning
  only final parsed output or summaries. Does not require scripts for ordinary
  source-code inspection or small files.
---

## Purpose

To preserve the LLM context window when raw data would be large.
Program analysis of large datasets, logs, HTTP responses, and command output rather
than loading them into context. Ordinary source-code inspection may use file-reading
tools directly.

---

## Core directive: compute, don't read

When faced with large datasets, logs, raw web responses, or command output, write
and execute a Node.js script to process the information and extract only what you
need. Do not use this workflow mechanically for ordinary source-code inspection or
small files.

### The Standard Workflow:

1. **Write:** Create a temporary JavaScript file outside the repository when possible (for example, `/tmp/analyze_data.js`).
2. **Execute:** Run the script via your shell tool (`node /tmp/analyze_data.js`).
3. **Output:** The script must use `console.log()` to output ONLY the final answer, summary, or strictly filtered data.
4. **Clean up:** Remove temporary scripts after execution when practical.

### Environment & Constraints:

- **Language:** Pure JavaScript (Node.js).
- **Dependencies:** Use built-in modules ONLY (`fs`, `path`, `child_process`).
  Do not assume `npm` packages are available unless explicitly told.
- **Reliability:** Always wrap your logic in `try/catch` blocks and gracefully handle `null`/`undefined` data.
- **Efficiency:** Write one comprehensive script instead of making ten individual tool calls.

---

## Strictly Forbidden Actions

The following actions are explicitly blocked to prevent context bloat and raw data dumping.

| ❌ FORBIDDEN ACTION                     | ⚠️ REASON                                                                  | ✅ REQUIRED ALTERNATIVE                                                                                                       |
| :-------------------------------------- | :------------------------------------------------------------------------- | :---------------------------------------------------------------------------------------------------------------------------- |
| **`curl` / `wget` for data extraction** | Dumps raw, unparsed HTTP/HTML directly into your context.                  | Write a Node.js script using `fetch()` to retrieve the data, parse it, and `console.log()` only the specific fields you need. |
| **Reading large files into context**    | Using standard file-read tools on large logs/JSONs exceeds context limits. | Write a Node script using `fs.readFileSync`, parse the data programmatically, and output only the summarized result.          |
| **Inline HTTP (`node -e "fetch..."`)**  | Prone to quoting errors and hard to debug.                                 | Write the code to a proper `.js` file first, then execute it.                                                                 |

---

## Restricted Commands: Intent-Based Routing

Assess your intent before using standard shell commands.
If the output will be large, you must programmatic redirect.

### 1. `bash` / Shell Commands

- **WHEN TO USE RAW SHELL:** Use directly for low-output file-system or environment operations (`git`, `mkdir`, `rm`, `mv`, `cd`, `ls`) and commands with small, bounded output.
- **WHEN OUTPUT MAY BE LARGE:** Capture and summarize it with a Node.js script using `child_process.execSync`; log only relevant lines, counts, or fields. Do not apply this mechanically to commands whose output is known to be small.

### 2. Search Utilities (`grep` / `find`)

- **IF LARGE RESULTS ARE EXPECTED:** Do not run these directly in the shell tool.
  Write a Node script to execute the search, parse the output, and return a summarized count or the most relevant matches. For a small, bounded search, run the command directly.

---

## Example Scenario

When asked to "Find the most common error in a 50MB log file", DO NOT try to read the file.
Do this:

**1. Create `parse_logs.js`:**

```javascript
const fs = require("fs");
try {
  const data = fs.readFileSync("app.log", "utf8");
  const lines = data.split("\n");
  const errors = {};

  lines.forEach((line) => {
    if (line.includes("ERROR")) {
      const match = line.match(/ERROR: (.*)/);
      if (match) {
        const msg = match[1];
        errors[msg] = (errors[msg] || 0) + 1;
      }
    }
  });

  // Sort and output ONLY the top 3 errors
  const topErrors = Object.entries(errors)
    .sort((a, b) => b[1] - a[1])
    .slice(0, 3);

  console.log("Top 3 Errors:");
  console.log(JSON.stringify(topErrors, null, 2));
} catch (err) {
  console.error("Script failed:", err.message);
}
```

**2. Execute:**

```bash
node parse_logs.js
```

**3. Read context-safe stdout.**

The output will be a concise summary of the most common errors without ever loading the entire log into your context, preserving your LLM's ability to reason effectively.
