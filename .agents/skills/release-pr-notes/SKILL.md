---
name: release-pr-notes
description: Create or update Ocserv Dashboard RELEASE.md and PULL_NOTE.md from a Git range, verified issue evidence, and the established v5 release-note format.
---

# Release and PR notes

Use this skill when the task asks to create, update, or review `RELEASE.md` or `PULL_NOTE.md` for this repository. These are local drafting artifacts and may be intentionally ignored by Git.

## Gather evidence first

- For a release note, identify the requested base tag and inspect `<tag>..HEAD`. If no tag is specified, ask or state the range used.
- For a pull note, inspect the requested branch/range and current working-tree changes.
- Use commit diffs and test output for implementation and validation claims. Do not infer that an issue is fixed from its label, title, or TODO entry.
- When issue references are requested, verify their current titles and labels from GitHub. Separate fixed issues from related or unresolved references.

## RELEASE.md

Follow the public v5.0 release structure when applicable:

1. `# Ocserv Dashboard <release label>`
2. `## Highlights`
3. `## Docker images` when image publishing changes
4. `## Security and compatibility` when applicable
5. `## Documentation` when applicable
6. `## Fixed issues` only for issues directly resolved in the range
7. `## Validation`

Write for operators and users. State expected image locations and tag behavior, but never say an image was published unless a registry or completed workflow verifies it. Preserve exact Git tag examples and clearly distinguish stable releases from prereleases.

## PULL_NOTE.md

Write for reviewers. Include a suggested PR title, a suggested merge-commit message, a compact change summary, fixed-issue references, validation, and upgrade or deployment notes when relevant. Keep related backlog issue references separate from fixes.

## Completion checks

- Keep the note scoped to the selected Git range.
- Keep `RELEASE.md` and `PULL_NOTE.md` in the requested ignore files when they are local-only artifacts.
- Run `git diff --check` after edits.
- Do not stage, commit, push, create a GitHub release, or modify application code unless explicitly requested.
