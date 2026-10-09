# Windows Git Bash / MINGW Path Conversion & Shell Detection

**CRITICAL KNOWLEDGE FOR BASH SCRIPTING ON WINDOWS**

This reference provides comprehensive guidance for handling path conversion and shell detection in Git Bash/MINGW/MSYS2 environments on Windows - essential knowledge for cross-platform bash scripting.

---

## Table of Contents

1. [Path Conversion in Git Bash/MINGW](#path-conversion-in-git-bashMINGW)
2. [Shell Detection Methods](#shell-detection-methods)
3. [Claude Code Specific Issues](#claude-code-specific-issues)
4. [Practical Solutions](#practical-solutions)
5. [Best Practices](#best-practices)

---

## Path Conversion in Git Bash/MINGW

### Automatic Conversion Behavior

Git Bash/MINGW automatically converts Unix-style paths to Windows paths when passing arguments to native Windows programs. Understanding this behavior is critical for writing portable scripts.

**Conversion Rules:**

```bash
# Unix → Windows path conversion
/foo → C:/Program Files/Git/usr/foo

# Path lists (colon-separated → semicolon-separated)
/foo:/bar → C:\msys64\foo;C:\msys64\bar

# Arguments with paths
--dir=/foo → --dir=C:/msys64/foo
```

### What Triggers Conversion

Automatic path conversion is triggered by:

```bash
# ✓ Leading forward slash (/) in arguments
command /c/Users/username/file.txt

# ✓ Colon-separated path lists
export PATH=/usr/bin:/usr/local/bin

# ✓ Arguments after - or , with path components
command --path=/tmp/data
```

### What's Exempt from Conversion

These patterns do NOT trigger automatic conversion:

```bash
# ✓ Arguments containing = (variable assignments)
VAR=/path/to/something

# ✓ Drive specifiers (C:)
C:/Windows/System32

# ✓ Arguments with ; (already Windows format)
PATH=C:\foo;C:\bar

# ✓ Arguments starting with // (Windows switches or UNC paths)
//server/share
command //e //s  # Command-line switches
```

### Control Environment Variables

**MSYS_NO_PATHCONV** (Git for Windows only):

```bash
# Disable ALL path conversion
export MSYS_NO_PATHCONV=1
command /path/to/file

# Per-command usage (recommended)
MSYS_NO_PATHCONV=1 command /path/to/file

# Value doesn't matter, just needs to be defined
MSYS_NO_PATHCONV=0  # Still disables conversion
```

**MSYS2_ARG_CONV_EXCL** (MSYS2 only):

```bash
# Exclude everything
export MSYS2_ARG_CONV_EXCL="*"

# Exclude specific prefixes
export MSYS2_ARG_CONV_EXCL="--dir=;/test"

# Multiple patterns (semicolon-separated)
export MSYS2_ARG_CONV_EXCL="--path=;--config=;/tmp"
```

**MSYS2_ENV_CONV_EXCL**:

```bash
# Prevents environment variable conversion
# Same syntax as MSYS2_ARG_CONV_EXCL
export MSYS2_ENV_CONV_EXCL="MY_PATH;CONFIG_DIR"
```

### Manual Conversion with cygpath

The `cygpath` utility provides precise control over path conversion:

```bash
# Convert Windows → Unix format
unix_path=$(cygpath -u "C:\Users\username\file.txt")
# Result: /c/Users/username/file.txt

# Convert Unix → Windows format
windows_path=$(cygpath -w "/c/Users/username/file.txt")
# Result: C:\Users\username\file.txt

# Convert to mixed format (forward slashes, Windows drive)
mixed_path=$(cygpath -m "/c/Users/username/file.txt")
# Result: C:/Users/username/file.txt

# Convert absolute path
absolute_path=$(cygpath -a "relative/path")

# Convert multiple paths
cygpath -u "C:\path1" "C:\path2"
```

**Practical cygpath usage:**

```bash
#!/usr/bin/env bash
# Cross-platform path handling

get_native_path() {
    local path="$1"

    # Check if running on Windows (Git Bash/MINGW)
    if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "mingw"* ]]; then
        # Convert to Windows format for native programs
        cygpath -w "$path"
    else
        # Already Unix format on Linux/macOS
        echo "$path"
    fi
}

# Usage
native_path=$(get_native_path "/c/Users/data")
windows_program.exe "$native_path"
```

### Common Workarounds

When automatic conversion causes issues:

**1. Use double slashes:**

```bash
# Problem: /e gets converted to C:/Program Files/Git/e
command /e /s

# Solution: Use double slashes
command //e //s  # Treated as switches, not paths
```

**2. Use dash notation:**

```bash
# Problem: /e flag converted to path
command /e /s

# Solution: Use dash notation
command -e -s
```

**3. Set MSYS_NO_PATHCONV temporarily:**

```bash
# Disable conversion for single command
MSYS_NO_PATHCONV=1 command /path/with/special/chars

# Or export for script section
export MSYS_NO_PATHCONV=1
command1 /path1
command2 /path2
unset MSYS_NO_PATHCONV
```

**4. Quote paths with spaces:**

```bash
# Always quote paths with spaces
command "/c/Program Files/App/file.txt"

# Or escape spaces
command /c/Program\ Files/App/file.txt
```

---

## Shell Detection Methods

### Method 1: $OSTYPE (Fastest, Bash-Only)

Best for: Quick platform detection in bash scripts

```bash
#!/usr/bin/env bash

case "$OSTYPE" in
    linux-gnu*)
        echo "Linux"
        ;;
    darwin*)
        echo "macOS"
        ;;
    cygwin*)
        echo "Cygwin"
        ;;
    msys*)
        echo "MSYS/Git Bash/MinGW"
        # Most common in Git for Windows
        ;;
    win*)
        echo "Windows (native)"
        ;;
    *)
        echo "Unknown: $OSTYPE"
        ;;
esac
```

**Advantages:**
- Fast (shell variable, no external command)
- Reliable for bash
- No forking required

**Disadvantages:**
- Bash-specific (not available in POSIX sh)
- Less detailed than uname

### Method 2: uname -s (Most Portable)

Best for: Maximum portability and detailed information

```bash
#!/bin/sh
# Works in any POSIX shell

case "$(uname -s)" in
    Darwin*)
        echo "macOS"
        ;;
    Linux*)
        # Check for WSL
        if grep -qi microsoft /proc/version 2>/dev/null; then
            echo "Windows Subsystem for Linux (WSL)"
        else
            echo "Linux (native)"
        fi
        ;;
    CYGWIN*)
        echo "Cygwin"
        ;;
    MINGW64*)
        echo "Git Bash 64-bit / MINGW64"
        ;;
    MINGW32*)
        echo "Git Bash 32-bit / MINGW32"
        ;;
    MSYS_NT*)
        echo "MSYS"
        ;;
    *)
        echo "Unknown: $(uname -s)"
        ;;
esac
```

**Common uname -s outputs:**

| Output | Platform | Description |
|--------|----------|-------------|
| `Darwin` | macOS | All macOS versions |
| `Linux` | Linux/WSL | Check `/proc/version` for WSL |
| `MINGW64_NT-10.0-*` | Git Bash | Git for Windows (64-bit) |
| `MINGW32_NT-10.0-*` | Git Bash | Git for Windows (32-bit) |
| `CYGWIN_NT-*` | Cygwin | Cygwin environment |
| `MSYS_NT-*` | MSYS | MSYS environment |

**Advantages:**
- Works in any POSIX shell
- Detailed system information
- Standard on all Unix-like systems

**Disadvantages:**
- Requires forking (slower than $OSTYPE)
- Output format varies by OS version

### Method 3: $MSYSTEM (MSYS2/Git Bash Specific)

Best for: Detecting MINGW subsystem type

```bash
#!/usr/bin/env bash

case "$MSYSTEM" in
    MINGW64)
        echo "Native Windows 64-bit environment"
        # Build native Windows 64-bit applications
        ;;
    MINGW32)
        echo "Native Windows 32-bit environment"
        # Build native Windows 32-bit applications
        ;;
    MSYS)
        echo "POSIX-compliant environment"
        # Build POSIX applications (depend on msys-2.0.dll)
        ;;
    "")
        echo "Not running in MSYS2/Git Bash"
        ;;
    *)
        echo "Unknown MSYSTEM: $MSYSTEM"
        ;;
esac
```

**MSYSTEM Values:**

| Value | Purpose | Path Conversion | Libraries |
|-------|---------|-----------------|-----------|
| `MINGW64` | Native Windows 64-bit | Automatic | Windows native (mingw-w64) |
| `MINGW32` | Native Windows 32-bit | Automatic | Windows native (mingw) |
| `MSYS` | POSIX environment | Minimal | POSIX (msys-2.0.dll) |

**WARNING:** Never set `$MSYSTEM` manually outside of MSYS2/Git Bash shells! It's automatically set by the environment and changing it can break the system.

**Advantages:**
- Precise subsystem detection
- Important for build systems
- Fast (environment variable)

**Disadvantages:**
- Only available in MSYS2/Git Bash
- Not set on other platforms

### Comprehensive Detection Function

Combine all methods for robust detection:

```bash
#!/usr/bin/env bash

detect_platform() {
    local platform=""
    local details=""

    # Check MSYSTEM first (most specific for Git Bash)
    if [[ -n "${MSYSTEM:-}" ]]; then
        platform="gitbash"
        details="$MSYSTEM"
        echo "platform=$platform subsystem=$MSYSTEM"
        return 0
    fi

    # Check OSTYPE
    case "$OSTYPE" in
        linux-gnu*)
            # Distinguish WSL from native Linux
            if grep -qi microsoft /proc/version 2>/dev/null; then
                platform="wsl"
                if [[ -n "${WSL_DISTRO_NAME:-}" ]]; then
                    details="$WSL_DISTRO_NAME"
                fi
            else
                platform="linux"
            fi
            ;;
        darwin*)
            platform="macos"
            ;;
        msys*|mingw*|cygwin*)
            platform="gitbash"
            ;;
        *)
            # Fallback to uname
            case "$(uname -s 2>/dev/null)" in
                MINGW*|MSYS*)
                    platform="gitbash"
                    ;;
                CYGWIN*)
                    platform="cygwin"
                    ;;
                *)
                    platform="unknown"
                    ;;
            esac
            ;;
    esac

    echo "platform=$platform${details:+ details=$details}"
}

# Usage
platform_info=$(detect_platform)
echo "$platform_info"
```

---

## Claude Code Specific Issues

### Issue #2602: Snapshot Path Conversion Failure

**Problem:**
```text
/usr/bin/bash: line 1: C:UsersDavid...No such file
```

**Root Cause:**
- Node.js `os.tmpdir()` returns Windows paths (e.g., `C:\Users\...`)
- Git Bash expects Unix paths (e.g., `/c/Users/...`)
- Automatic conversion fails due to path format mismatch

**Solution (Claude Code v1.0.51+):**

Set environment variable before starting Claude Code:

```powershell
# PowerShell
$env:CLAUDE_CODE_GIT_BASH_PATH = "C:\Program Files\git\bin\bash.exe"
```

```cmd
# CMD
set CLAUDE_CODE_GIT_BASH_PATH=C:\Program Files\git\bin\bash.exe
```

```bash
# Git Bash (add to ~/.bashrc)
export CLAUDE_CODE_GIT_BASH_PATH="C:\\Program Files\\git\\bin\\bash.exe"
```

**Note:** Versions 1.0.72+ reportedly work without modifications, but setting the environment variable ensures compatibility.

### Other Known Issues

**Drive letter duplication:**
```bash
# Problem
cd D:\dev
pwd
# Output: D:\d\dev  (incorrect)

# Solution: Use Unix-style path in Git Bash
cd /d/dev
pwd
# Output: /d/dev
```

**Spaces in paths:**
```bash
# Problem: Unquoted path with spaces
cd C:\Program Files\App  # Fails

# Solution: Always quote paths with spaces
cd "C:\Program Files\App"
cd /c/Program\ Files/App
```

**VS Code extension Git Bash detection:**

VS Code may not auto-detect Git Bash. Configure manually in settings:

```json
{
  "terminal.integrated.defaultProfile.windows": "Git Bash",
  "terminal.integrated.profiles.windows": {
    "Git Bash": {
      "path": "C:\\Program Files\\Git\\bin\\bash.exe"
    }
  }
}
```

---

## Practical Solutions

### Cross-Platform Path Handling Function

```bash
#!/usr/bin/env bash

# Convert path to format appropriate for current platform
normalize_path_for_platform() {
    local path="$1"

    case "$OSTYPE" in
        msys*|mingw*)
            # On Git Bash, convert to Unix format if Windows format provided
            if [[ "$path" =~ ^[A-Z]:\\ ]]; then
                # Windows path detected, convert to Unix
                path=$(cygpath -u "$path" 2>/dev/null || echo "$path")
            fi
            ;;
        *)
            # On Linux/macOS, path is already correct
            ;;
    esac

    echo "$path"
}

# Convert path to native format for external programs
convert_to_native_path() {
    local path="$1"

    case "$OSTYPE" in
        msys*|mingw*)
            # Convert to Windows format for native Windows programs
            cygpath -w "$path" 2>/dev/null || echo "$path"
            ;;
        *)
            # Already native on Linux/macOS
            echo "$path"
            ;;
    esac
}

# Example usage
input_path="/c/Users/username/file.txt"
normalized=$(normalize_path_for_platform "$input_path")
echo "Normalized: $normalized"

native=$(convert_to_native_path "$normalized")
echo "Native: $native"
```

### Script Template for Windows Compatibility

```bash
#!/usr/bin/env bash
set -euo pipefail

# Detect if running on Git Bash/MINGW
is_git_bash() {
    [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "mingw"* ]]
}

# Handle path conversion based on platform
get_path() {
    local path="$1"

    if is_git_bash; then
        # Ensure Unix format in Git Bash
        if [[ "$path" =~ ^[A-Z]:\\ ]]; then
            cygpath -u "$path"
        else
            echo "$path"
        fi
    else
        echo "$path"
    fi
}

# Call Windows program from Git Bash
call_windows_program() {
    local program="$1"
    shift
    local args=("$@")

    if is_git_bash; then
        # Disable path conversion for complex arguments
        MSYS_NO_PATHCONV=1 "$program" "${args[@]}"
    else
        "$program" "${args[@]}"
    fi
}

# Main script logic
main() {
    local file_path="$1"

    # Normalize path
    file_path=$(get_path "$file_path")

    # Process file
    echo "Processing: $file_path"

    # Call Windows program if needed
    if is_git_bash; then
        local native_path
        native_path=$(cygpath -w "$file_path")
        call_windows_program notepad.exe "$native_path"
    fi
}

main "$@"
```

### Handling Command-Line Arguments

```bash
#!/usr/bin/env bash

# Parse arguments that might contain paths
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --path=*)
                local path="${1#*=}"
                # Disable conversion for this specific argument pattern
                MSYS2_ARG_CONV_EXCL="--path=" command --path="$path"
                shift
                ;;
            --dir)
                local dir="$2"
                # Use converted path
                local native_dir
                if command -v cygpath &>/dev/null; then
                    native_dir=$(cygpath -w "$dir")
                else
                    native_dir="$dir"
                fi
                command --dir "$native_dir"
                shift 2
                ;;
            *)
                shift
                ;;
        esac
    done
}
```

---

## Best Practices

### 1. Always Quote Paths

```bash
# ✗ WRONG - Breaks with spaces
cd $path

# ✓ CORRECT - Works with all paths
cd "$path"
```

### 2. Use cygpath for Reliable Conversion

```bash
# ✗ WRONG - Manual conversion is error-prone
path="${path//\\/\/}"
path="${path/C:/\/c}"

# ✓ CORRECT - Use cygpath
path=$(cygpath -u "$path")
```

### 3. Detect Platform Before Path Operations

```bash
# ✓ CORRECT - Platform-aware
if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "mingw"* ]]; then
    # Git Bash specific handling
    path=$(cygpath -u "$windows_path")
else
    # Linux/macOS handling
    path="$unix_path"
fi
```

### 4. Use MSYS_NO_PATHCONV Sparingly

```bash
# ✗ WRONG - Disables all conversion globally
export MSYS_NO_PATHCONV=1

# ✓ CORRECT - Per-command when needed
MSYS_NO_PATHCONV=1 command --flag=/value
```

### 5. Test on Target Platform

Always test scripts on Windows with Git Bash if that's a target platform:

```bash
# Test script
bash -n script.sh  # Syntax check
shellcheck script.sh  # Static analysis
bash script.sh  # Run on actual platform
```

### 6. Document Platform Requirements

```bash
#!/usr/bin/env bash
#
# Platform Support:
#   - Linux: Full support
#   - macOS: Full support
#   - Windows Git Bash: Requires Git for Windows 2.x+
#   - Windows WSL: Full support
#
# Known Issues:
#   - Path conversion may occur when calling Windows programs from Git Bash
#   - Use MSYS_NO_PATHCONV=1 if experiencing path-related errors
```

### 7. Use Forward Slashes in Git Bash

```bash
# ✓ PREFERRED - Works in all environments
cd /c/Users/username/project

# ✗ AVOID - Requires escaping or quoting
cd "C:\Users\username\project"
cd C:\\Users\\username\\project
```

### 8. Check for cygpath Availability

```bash
# Graceful fallback if cygpath not available
convert_path() {
    local path="$1"

    if command -v cygpath &>/dev/null; then
        cygpath -u "$path"
    else
        # Manual conversion as fallback
        echo "$path" | sed 's|\\|/|g' | sed 's|^\([A-Z]\):|/\L\1|'
    fi
}
```

---

## Quick Reference Card

### Path Conversion Control

| Variable | Scope | Effect |
|----------|-------|--------|
| `MSYS_NO_PATHCONV=1` | Git for Windows | Disables all conversion |
| `MSYS2_ARG_CONV_EXCL="pattern"` | MSYS2 | Excludes specific patterns |
| `MSYS2_ENV_CONV_EXCL="var"` | MSYS2 | Excludes environment variables |

### Shell Detection Variables

| Variable | Available | Purpose |
|----------|-----------|---------|
| `$OSTYPE` | Bash | Quick OS type detection |
| `$MSYSTEM` | MSYS2/Git Bash | Subsystem type (MINGW64/MINGW32/MSYS) |
| `$(uname -s)` | All POSIX | Detailed OS identification |

### cygpath Quick Reference

| Command | Purpose |
|---------|---------|
| `cygpath -u "C:\path"` | Windows → Unix format |
| `cygpath -w "/c/path"` | Unix → Windows format |
| `cygpath -m "/c/path"` | Unix → Mixed format (forward slashes) |
| `cygpath -a "path"` | Convert to absolute path |

### Common Issues & Solutions

| Problem | Solution |
|---------|----------|
| Path with spaces breaks | Quote the path: `"$path"` |
| Flag `/e` converted to path | Use `//e` or `-e` instead |
| Drive duplication `D:\d\` | Use Unix format: `/d/` |
| Windows program needs Windows path | Use `cygpath -w "$unix_path"` |
| Script fails in Claude Code | Set `CLAUDE_CODE_GIT_BASH_PATH` |

---

## Summary

Understanding Git Bash/MINGW path conversion is essential for writing robust cross-platform bash scripts that work on Windows. Key takeaways:

1. **Automatic conversion** happens for Unix-style paths in arguments
2. **Control conversion** using `MSYS_NO_PATHCONV` and `MSYS2_ARG_CONV_EXCL`
3. **Use cygpath** for reliable manual path conversion
4. **Detect platform** using `$OSTYPE`, `$MSYSTEM`, or `uname -s`
5. **Quote all paths** to handle spaces and special characters
6. **Test on target platforms** to catch platform-specific issues
7. **Document requirements** so users know what to expect

With this knowledge, you can write bash scripts that work seamlessly across Linux, macOS, Windows Git Bash, WSL, and other Unix-like environments.
