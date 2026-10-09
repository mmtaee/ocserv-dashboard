# Testing and Verification

## Test Locations

Project tests belong under `tests/model/`, `tests/units/`, and `tests/integration/`. Do not place new `*_test.go` files beside production packages unless the repository already has an explicit exception.

## Test Scope

Unit tests should cover usecase behavior, business validation, error paths, repository contracts where practical, and edge cases introduced by the change. Mock external dependencies at architectural boundaries.

Integration tests should cover behavior requiring actual component integration, including database or API behavior when repository support exists. Do not turn every test into an integration test.

For bug fixes, add a regression test that fails before the fix and passes afterward whenever practical.

## Required Verification

After modifying Go code, run applicable checks. At minimum:

```bash
gofmt -w <modified-go-files>
go test ./...
```

When supported, also run `go vet ./...`. Prefer established Makefile targets, Taskfile commands, CI scripts, or lint commands when the repository defines them.

Run `go mod tidy` after dependency changes only; unnecessary runs can create unrelated diffs.

## Completion Criteria

- Requested behavior is implemented.
- Architecture boundaries are preserved.
- No unnecessary duplication was introduced.
- New error codes and migrations are uniquely registered when applicable.
- Swagger matches actual behavior.
- Relevant tests exist.
- Modified Go code is formatted.
- Relevant checks ran, or failures are explicitly reported.

Completion reports should summarize changes, important design decisions, checks performed, and remaining issues or assumptions.
