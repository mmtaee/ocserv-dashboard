---
name: bash-master
description: |
  Expert bash/shell scripting across environments where Bash is available
  (Linux, macOS, Git Bash on Windows, WSL, containers). This is a Bash-focused
  skill - it does not provide PowerShell parity; on Windows the target is the
  Bash interpreter (Git Bash / WSL), not native PowerShell.
  PROACTIVELY activate for: (1) ANY bash/shell script task, (2) System automation, (3) DevOps/CI/CD scripts, (4) Build/deployment automation, (5) Script review/debugging, (6) Converting commands to scripts.
  Provides: Google Shell Style Guide compliance, ShellCheck validation, Bash portability across Linux/macOS/WSL/Git Bash/containers, POSIX compliance, security hardening, error handling, performance optimization, testing with BATS, and production-ready patterns.
  Ensures professional-grade, secure, portable Bash scripts every time.
---

# Bash Scripting Mastery

## Scope and platform contract

This skill targets **Bash itself**, wherever Bash runs - Linux, macOS, WSL, Git Bash / MSYS2 on Windows, and Bash-based container images. It does not cover **native PowerShell**: a PowerShell script is a different language and should use `powershell-master`. On Windows, `bash-master` assumes the user is running Bash inside Git Bash, WSL, or a similar Bash environment, and addresses the MSYS path-translation quirks that result.

## Repository conventions

Project-level conventions (Windows backslashes in tool calls, documentation discipline, etc.) live in the agent body and the `windows-path-master` plugin. This skill focuses on Bash content; do not duplicate that boilerplate here.

## Quick reference

```bash
#!/usr/bin/env bash
set -euo pipefail  # Exit on error, undefined vars, pipe failures
IFS=$'\n\t'        # Safe word splitting
# Run shellcheck your_script.sh before deployment.
# Test on every target platform before production.
```

**Bash-portability quick check:**

```bash
# Linux/macOS:       Full bash features
# Git Bash (Windows): Most features, some system calls missing (no systemd, /proc differs)
# WSL:               Effectively Linux; /mnt/c for Windows filesystem
# Containers:        Depends on base image - alpine ships /bin/sh, not bash
# POSIX mode:        Use /bin/sh and avoid bashisms
```

## When to use this skill

**Always activate for:**

- Writing or modifying any bash/shell script
- Reviewing or refactoring existing scripts
- Debugging shell script failures
- DevOps automation, CI/CD pipelines, system administration
- Cross-environment Bash portability (Linux <-> macOS <-> WSL <-> Git Bash <-> container)

**Do not use this skill for:**

- PowerShell scripts - use `powershell-master`
- Batch (`.cmd`/`.bat`) scripting
- Generic command help unrelated to scripting

## Core principles

### 1. Safety first

Every script should open with the safety preamble:

```bash
#!/usr/bin/env bash
set -e            # Exit on any error
set -u            # Exit on undefined variable
set -o pipefail   # Catch failures mid-pipeline
set -E            # Inherit ERR trap into functions
IFS=$'\n\t'       # Avoid word splitting on spaces

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_NAME="$(basename "${BASH_SOURCE[0]}")"
```

### 2. POSIX vs Bash

| Need to run on any UNIX `sh` | `#!/bin/sh`, no `[[ ]]`, no arrays, no process substitution |
|--|--|
| Modern Linux/macOS with Bash | `#!/usr/bin/env bash`, prefer `[[ ]]`, arrays, regex |
| Alpine/minimal containers | Either install bash explicitly or write POSIX-compliant `sh` |

### 3. Quoting

```bash
# Always quote expansions
process "$file_path"        # correct
process $file_path          # word-splitting bug

# Arrays
files=("file 1.txt" "file 2.txt")
process "${files[@]}"       # each element kept separate
process "${files[*]}"       # joined as one string - usually wrong
```

### 4. ShellCheck

Run `shellcheck` on every script. Only disable warnings with a justification comment: `# shellcheck disable=SC2086 reason: intentional word splitting`.

See `references/best_practices.md` for the full quoting/style table and `references/patterns_antipatterns.md` for the common pitfalls.

## Platform-specific considerations

### Git Bash / MSYS2 (Windows)

Git Bash auto-converts Unix-style arguments to Windows paths. This is the largest single source of cross-platform Bash bugs on Windows.

```bash
# The conversion: /foo becomes C:/Program Files/Git/usr/foo
# Disable per-command:
MSYS_NO_PATHCONV=1 command /path/that/should/stay/unix

# Manual conversion
unix_path=$(cygpath -u "C:\Windows\System32")
win_path=$(cygpath -w "/c/Users/username")

# Detect Git Bash
if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "mingw"* ]]; then
    : # Git Bash
fi
case "${MSYSTEM:-}" in
    MINGW64|MINGW32|MSYS) : ;;  # MSYS2 / Git Bash environment
esac

# Flags that look like paths
command //e //s   # double-slash to suppress conversion
command -e -s     # or use dash-options
```

Full Git Bash + Windows path notes live in `references/windows-git-bash-paths.md`.

### Linux

GNU coreutils, `/proc`, systemd integration. Detect via `[[ "$OSTYPE" == "linux-gnu"* ]]`.

### macOS

BSD utilities behave differently from GNU. Most common gotchas: `sed -i ''` (empty string required), `date` flags differ, `readlink -f` not available on stock macOS.

```bash
if command -v gsed >/dev/null; then SED=gsed; else SED=sed; fi
```

### WSL

Effectively Linux. The Windows filesystem is mounted at `/mnt/c/`. Detect via `grep -qi microsoft /proc/version`.

### Containers

Alpine images ship only `/bin/sh` (BusyBox). Write POSIX-compliant scripts or `apk add bash`. Container init quirks: PID 1 must reap children and handle signals. Detect via `[ -f /.dockerenv ]` or `[ -n "$KUBERNETES_SERVICE_HOST" ]`.

### Portable platform-detection template

```bash
detect_platform() {
    case "$OSTYPE" in
        linux-gnu*)    echo "linux" ;;
        darwin*)       echo "macos" ;;
        msys*|cygwin*) echo "windows" ;;
        *)             echo "unknown" ;;
    esac
}
```

Full per-platform tables (BSD-vs-GNU coreutils flags, WSL networking, container init patterns) live in `references/platform_specifics.md`.

## Best practices (summary)

The full patterns - function design, error handling, input validation, argument parsing, logging - live in `references/in-depth-patterns.md`. The headline rules:

- One concern per function; locals declared first; validate input; return non-zero on error.
- Constants `UPPER_CASE`; locals `lower_case`; mark immutable values `readonly`.
- Always check exit codes (`if ! cmd`, `||`, traps, or a central `error_exit` helper).
- Validate every external input - empty, format, length, charset.
- Use `getopts` or a `case`-based argument parser; print usage and exit 1 on bad input.
- Use a leveled logger that writes to stderr.

## Security, performance, testing, debugging, advanced patterns

These each have dedicated sections in `references/in-depth-patterns.md`:

| Topic | What it covers |
|--|--|
| Security | Command-injection prevention, path-traversal guards, privilege management, secure temp files |
| Performance | Avoiding subshells, bash built-ins vs externals, process substitution, array ops |
| Testing | BATS unit tests, integration test patterns, CI/CD wiring |
| Debugging | `set -x`, `PS4`, conditional debug helpers, tracing and profiling |
| Advanced patterns | Safe config parsing, parallel processing, signal handling, retries with backoff |

Read that reference any time you need the canonical code template for one of those topics.

## Reference files

- [`references/platform_specifics.md`](references/platform_specifics.md) - Detailed platform differences and workarounds
- [`references/best_practices.md`](references/best_practices.md) - Comprehensive industry standards and guidelines
- [`references/patterns_antipatterns.md`](references/patterns_antipatterns.md) - Common patterns and pitfalls with solutions
- [`references/windows-git-bash-paths.md`](references/windows-git-bash-paths.md) - Git Bash / MSYS path-translation reference
- [`references/in-depth-patterns.md`](references/in-depth-patterns.md) - Function design, security, performance, testing, debugging, advanced patterns
- [`references/resources.md`](references/resources.md) - Official docs, style guides, tooling, and learning links

## Success criteria

A Bash script written with this skill should:

1. Pass `shellcheck` with no warnings
2. Begin with `set -euo pipefail`
3. Quote every variable expansion
4. Print usage on `-h`/`--help`
5. Decompose into testable functions
6. Handle empty input, missing files, and unexpected arguments
7. Run on every target platform (Linux/macOS/WSL/Git Bash/container) where it claims support
8. Match the Google Shell Style Guide
9. Clean up on exit (`trap EXIT`)
10. Be unit-tested with BATS where logic is non-trivial

```bash
# Pre-deployment checklist
shellcheck script.sh
bash -n script.sh
bats test/script.bats
./script.sh --help
DEBUG=true ./script.sh
```

## Troubleshooting

### Script fails on a different platform

- `checkbashisms script.sh` to surface non-portable constructs.
- `command -v tool` to verify a required tool is installed.
- Diff command flags between GNU and BSD (`sed --version` etc.).

### ShellCheck warnings

- Read the rule explanation (`shellcheck -W SC2086`).
- Fix the underlying issue; only disable a rule with a justification comment.

### Works interactively but fails in cron

- Cron has a minimal `PATH` - set `PATH` explicitly.
- Use absolute paths.
- Redirect stdout/stderr: `./script.sh >> /tmp/cron.log 2>&1`.

### Performance issues

- Profile with `time`.
- Enable `set -x` to find slow steps.
- Replace external invocations with Bash built-ins where possible.
