---
applyTo: '**'
---

# AI Instructions

This project is governed by an AI Constitution.

The AI Constitution is located at:

.github/instructions/AI_CONSTITUTION.txt

This file is a **constitutional authority** and MUST be read and applied
before generating or modifying any of the following:

- Code
- Configuration
- Database behavior or schema
- Runtime logic
- Infrastructure or deployment logic

If the AI Constitution file cannot be found or read at the path above,
YOU MUST STOP and report the failure.  
Do NOT proceed under assumptions or partial rules.

Scope:
- The Constitution applies to any request that edits code, configuration,
  database behavior, or runtime logic.
- For non-code tasks (documentation, prompt authoring),
  the Constitution is inert unless explicitly invoked.

Hierarchy (strict):
1. .github/instructions/AI_CONSTITUTION.txt
2. User task prompt
3. Project documentation (README.md, README/**, TODO.md)

Do not infer, invent, or substitute rules.
If any conflict exists, defer to the Constitution.
