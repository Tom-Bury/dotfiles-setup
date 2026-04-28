---
name: coder
description: >
  Mandatory workflow for preserving context window space during data analysis. 
  Instructs the agent to compute, count, filter, and analyze data by writing 
  and executing standalone Node.js scripts, returning ONLY the final parsed 
  output or summaries to stdout. Prevents context bloat from raw data dumping 
  (e.g., large files, raw HTTP responses, or heavy shell command outputs).
---

## Purpose

To strictly preserve the LLM context window.
You must **PROGRAM** your analysis rather than reading raw data into your context and using internal reasoning to compute, count, filter, or analyze it.

---

## Core directive: compute, don't read

When faced with large datasets, logs, codebases, or web pages, you must write and execute Node.js scripts to process the information and extract only what you need.

### The Standard Workflow:

1. **Write:** Create a temporary JavaScript file (e.g., `analyze_data.js`) using your file-writing tool.
2. **Execute:** Run the script via your shell tool (`node analyze_data.js`).
3. **Output:** The script must use `console.log()` to output ONLY the final answer, summary, or strictly filtered data.

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

- **WHEN TO USE RAW SHELL:** ONLY for standard, low-output file system or environment operations (`git`, `mkdir`, `rm`, `mv`, `cd`, `ls`).
- **IF OUTPUT > 20 LINES:** Do not run it directly.
  Write a Node.js script using `child_process.execSync`, capture the output, filter it inside the script, and log only the relevant lines.

### 2. Search Utilities (`grep` / `find`)

- **IF LARGE RESULTS ARE EXPECTED:** Do NOT run these directly in the shell tool.
  Write a Node script to execute the grep/find command, parse the output string, and return a summarized count or the top 5 matches.

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
