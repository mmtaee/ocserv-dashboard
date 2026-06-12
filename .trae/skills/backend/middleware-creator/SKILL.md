---
name: middleware-creator
description: Create Echo v5 middlewares. **MANDATORY** ALWAYS read master-rules before implementation.
---

# Context
**CRITICAL**: You MUST read and strictly follow master-rules before starting any task. Every decision (naming, structure, errors, pagination) must align with these rules.

# Standards
- Path: `pkg/middlewares/`.
- Usage: Every middleware MUST have a usage comment (e.g., `// Usage: e.Use(middlewares.Name())`).
- Context: Use `*echo.Context` as per pointer rules.

# Tasks
1. Implement middleware in `pkg/middlewares/`.
2. Add usage sample comment.
3. Handle errors using `pkg/request` helpers.
