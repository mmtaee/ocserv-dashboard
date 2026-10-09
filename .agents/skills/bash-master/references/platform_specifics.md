# Platform-Specific Bash Scripting

Comprehensive guide to handling platform differences in bash scripts across Linux, macOS, Windows (Git Bash/WSL), and containers.

---


## ⚠️ WINDOWS GIT BASH / MINGW PATH CONVERSION

**CRITICAL REFERENCE:** For complete Windows Git Bash path conversion and shell detection guidance, see:

**📄 [windows-git-bash-paths.md](./windows-git-bash-paths.md)**

This comprehensive guide covers:
- **Automatic path conversion behavior** (Unix → Windows)
- **MSYS_NO_PATHCONV and MSYS2_ARG_CONV_EXCL** usage
- **cygpath** manual conversion tool
- **Shell detection methods** ($OSTYPE, uname, $MSYSTEM)
- **Claude Code specific issues** (#2602 snapshot path conversion)
- **Common problems and solutions**
- **Cross-platform scripting patterns**

**Git Bash path conversion is the #1 source of Windows bash scripting issues.** Always consult the dedicated guide when working with Windows/Git Bash.

---

## WARNING: WINDOWS GIT BASH / MINGW PATH CONVERSION

**CRITICAL REFERENCE:** For complete Windows Git Bash path conversion and shell detection guidance, see:

**[windows-git-bash-paths.md](./windows-git-bash-paths.md)**

This comprehensive guide covers:
- **Automatic path conversion behavior** (Unix to Windows)
- **MSYS_NO_PATHCONV and MSYS2_ARG_CONV_EXCL** usage
- **cygpath** manual conversion tool
- **Shell detection methods** ($OSTYPE, uname, $MSYSTEM)
- **Claude Code specific issues** (#2602 snapshot path conversion)
- **Common problems and solutions**
- **Cross-platform scripting patterns**

**Git Bash path conversion is the #1 source of Windows bash scripting issues.** Always consult the dedicated guide when working with Windows/Git Bash.

---

## Table of Contents

1. [Platform Detection](#platform-detection)
2. [Linux Specifics](#linux-specifics)
3. [macOS Specifics](#macos-specifics)
4. [Windows (Git Bash)](#windows-git-bash) - **See windows-git-bash-paths.md for complete guide** - **See windows-git-bash-paths.md for complete guide**
5. [Windows (WSL)](#windows-wsl)
6. [Container Environments](#container-environments)
7. [Cross-Platform Patterns](#cross-platform-patterns)
8. [Command Compatibility Matrix](#command-compatibility-matrix)

---

## 🚨 CRITICAL GUIDELINES

### Windows File Path Requirements

**MANDATORY: Always Use Backslashes on Windows for File Paths**

When using Edit or Write tools on Windows, you MUST use backslashes (`\`) in file paths, NOT forward slashes (`/`).

**Examples:**
- ❌ WRONG: `D:/repos/project/file.tsx`
- ✅ CORRECT: `D:\repos\project\file.tsx`

This applies to:
- Edit tool file_path parameter
- Write tool file_path parameter
- All file operations on Windows systems


### Documentation Guidelines

**NEVER create new documentation files unless explicitly requested by the user.**

- **Priority**: Update existing README.md files rather than creating new documentation
- **Repository cleanliness**: Keep repository root clean - only README.md unless user requests otherwise
- **Style**: Documentation should be concise, direct, and professional - avoid AI-generated tone
- **User preference**: Only create additional .md files when user specifically asks for documentation


---

## Platform Detection

### Comprehensive Detection Script

```bash
#!/usr/bin/env bash

detect_os() {
    case "$OSTYPE" in
        linux-gnu*)
            if grep -qi microsoft /proc/version 2>/dev/null; then
                echo "wsl"
            else
                echo "linux"
            fi
            ;;
        darwin*)
            echo "macos"
            ;;
        msys*|mingw*|cygwin*)
            echo "gitbash"
            ;;
        *)
            echo "unknown"
            ;;
    esac
}

detect_distro() {
    # Only for Linux
    if [[ -f /etc/os-release ]]; then
        # shellcheck source=/dev/null
        source /etc/os-release
        echo "$ID"
    elif [[ -f /etc/redhat-release ]]; then
        echo "rhel"
    elif [[ -f /etc/debian_version ]]; then
        echo "debian"
    else
        echo "unknown"
    fi
}

detect_container() {
    if [[ -f /.dockerenv ]]; then
        echo "docker"
    elif grep -q docker /proc/1/cgroup 2>/dev/null; then
        echo "docker"
    elif [[ -n "$KUBERNETES_SERVICE_HOST" ]]; then
        echo "kubernetes"
    else
        echo "none"
    fi
}

# Usage
OS=$(detect_os)
DISTRO=$(detect_distro)
CONTAINER=$(detect_container)

echo "OS: $OS"
echo "Distro: $DISTRO"
echo "Container: $CONTAINER"
```

### Environment Variables for Detection

```bash
# Check various environment indicators
check_environment() {
    echo "OSTYPE: $OSTYPE"
    echo "MACHTYPE: $MACHTYPE"
    echo "HOSTTYPE: $HOSTTYPE"

    # Kernel info
    uname -s    # Operating system name
    uname -r    # Kernel release
    uname -m    # Machine hardware
    uname -p    # Processor type

    # More detailed
    uname -a    # All information
}

# Platform-specific variables
# Linux:   OSTYPE=linux-gnu
# macOS:   OSTYPE=darwin20.0
# Git Bash: OSTYPE=msys
# Cygwin:  OSTYPE=cygwin
# WSL:     OSTYPE=linux-gnu (but with Microsoft in /proc/version)
```

---

## Linux Specifics

### Linux-Only Features

```bash
# /proc filesystem
get_process_info() {
    local pid=$1

    if [[ -d "/proc/$pid" ]]; then
        echo "Command: $(cat /proc/$pid/cmdline | tr '\0' ' ')"
        echo "Working dir: $(readlink /proc/$pid/cwd)"
        echo "Executable: $(readlink /proc/$pid/exe)"
    fi
}

# systemd
check_systemd() {
    if command -v systemctl &> /dev/null; then
        systemctl status my-service
        systemctl is-active my-service
        systemctl is-enabled my-service
    fi
}

# cgroups
check_cgroups() {
    if [[ -d /sys/fs/cgroup ]]; then
        cat /sys/fs/cgroup/memory/memory.limit_in_bytes
    fi
}

# inotify for file watching
watch_directory() {
    if command -v inotifywait &> /dev/null; then
        inotifywait -m -r -e modify,create,delete /path/to/watch
    fi
}
```

### Distribution-Specific Commands

```bash
# Package management
install_package() {
    local package=$1

    if command -v apt-get &> /dev/null; then
        # Debian/Ubuntu
        sudo apt-get update
        sudo apt-get install -y "$package"
    elif command -v yum &> /dev/null; then
        # RHEL/CentOS
        sudo yum install -y "$package"
    elif command -v dnf &> /dev/null; then
        # Fedora
        sudo dnf install -y "$package"
    elif command -v pacman &> /dev/null; then
        # Arch
        sudo pacman -S --noconfirm "$package"
    elif command -v zypper &> /dev/null; then
        # openSUSE
        sudo zypper install -y "$package"
    elif command -v apk &> /dev/null; then
        # Alpine
        sudo apk add "$package"
    else
        echo "Error: No supported package manager found" >&2
        return 1
    fi
}

# Service management
manage_service() {
    local action=$1
    local service=$2

    if command -v systemctl &> /dev/null; then
        # systemd (most modern distros)
        sudo systemctl "$action" "$service"
    elif command -v service &> /dev/null; then
        # SysV init
        sudo service "$service" "$action"
    else
        echo "Error: No supported service manager found" >&2
        return 1
    fi
}
```

### GNU Coreutils (Linux Standard)

```bash
# GNU-specific features
# These work on Linux but may not work on macOS/BSD

# sed with -i (in-place editing)
sed -i 's/old/new/g' file.txt         # Linux
sed -i '' 's/old/new/g' file.txt      # macOS requires empty string

# date with flexible parsing
date -d "yesterday" +%Y-%m-%d         # Linux
date -v-1d +%Y-%m-%d                  # macOS

# stat with -c format
stat -c "%s" file.txt                 # Linux (file size)
stat -f "%z" file.txt                 # macOS

# readlink with -f (canonicalize)
readlink -f /path/to/file             # Linux
# macOS doesn't have -f, use greadlink or:
python -c "import os; print(os.path.realpath('$file'))"

# GNU find with -printf
find . -type f -printf "%p %s\n"      # Linux
find . -type f -exec stat -f "%N %z" {} \;  # macOS
```

---

## macOS Specifics

### BSD vs GNU Commands

```bash
# Detect and use GNU versions if available
setup_commands_macos() {
    # Install GNU commands: brew install coreutils gnu-sed gnu-tar findutils
    if command -v gsed &> /dev/null; then
        SED=gsed
    else
        SED=sed
    fi

    if command -v ggrep &> /dev/null; then
        GREP=ggrep
    else
        GREP=grep
    fi

    if command -v greadlink &> /dev/null; then
        READLINK=greadlink
    else
        READLINK=readlink
    fi

    if command -v gdate &> /dev/null; then
        DATE=gdate
    else
        DATE=date
    fi

    if command -v gstat &> /dev/null; then
        STAT=gstat
    else
        STAT=stat
    fi

    export SED GREP READLINK DATE STAT
}

# Usage
setup_commands_macos
$SED -i 's/old/new/g' file.txt  # Works on both platforms
```

### macOS-Specific Features

```bash
# macOS filesystem (case-insensitive by default on APFS/HFS+)
check_case_sensitivity() {
    touch /tmp/test_case
    if [[ -f /tmp/TEST_CASE ]]; then
        echo "Filesystem is case-insensitive"
    else
        echo "Filesystem is case-sensitive"
    fi
    rm -f /tmp/test_case /tmp/TEST_CASE
}

# macOS extended attributes
# Set extended attribute
xattr -w com.example.myattr "value" file.txt

# Get extended attribute
xattr -p com.example.myattr file.txt

# List all extended attributes
xattr -l file.txt

# Remove extended attribute
xattr -d com.example.myattr file.txt

# macOS Spotlight
# Disable indexing for directory
mdutil -i off /path/to/directory

# Search with mdfind (Spotlight from command line)
mdfind "kMDItemFSName == 'filename.txt'"

# macOS clipboard
# Copy to clipboard
echo "text" | pbcopy

# Paste from clipboard
pbpaste

# macOS notifications
# Display notification
osascript -e 'display notification "Build complete" with title "Build Status"'

# macOS open command
# Open file with default application
open file.pdf

# Open URL
open https://example.com

# Open current directory in Finder
open .
```

### Homebrew Package Management

```bash
# Check if Homebrew is installed
if command -v brew &> /dev/null; then
    # Install package
    brew install package-name

    # Update Homebrew
    brew update

    # Upgrade packages
    brew upgrade

    # Search for package
    brew search package-name

    # Get package info
    brew info package-name
fi
```

---

## Windows (Git Bash)

### Git Bash Environment

```bash
# Git Bash uses MSYS2 runtime
# Provides Unix-like environment on Windows

# Path handling
convert_path() {
    local path=$1

    if command -v cygpath &> /dev/null; then
        # Convert Unix path to Windows
        windows_path=$(cygpath -w "$path")
        echo "$windows_path"

        # Convert Windows path to Unix
        unix_path=$(cygpath -u "C:\\Users\\user\\file.txt")
        echo "$unix_path"
    else
        # Manual conversion (Git Bash)
        # /c/Users/user → C:\Users\user
        echo "${path//\//\\}" | sed 's/^\\//'
    fi
}

# Git Bash path conventions
# C:\Users\user → /c/Users/user
# D:\data → /d/data

# Home directory
echo "$HOME"           # /c/Users/username
echo "$USERPROFILE"    # Windows-style path

# Temp directory
echo "$TEMP"           # Windows temp
echo "$TMP"            # Windows temp
echo "/tmp"            # Git Bash temp (usually C:\Users\username\AppData\Local\Temp)
```

### Limited Features in Git Bash

```bash
# Features NOT available in Git Bash:

# 1. No systemd
# Use Windows services instead:
# sc query ServiceName
# net start ServiceName

# 2. Limited signal support
# SIGTERM works, but some signals behave differently

# 3. No /proc filesystem
# Use wmic or PowerShell:
# wmic process get processid,commandline

# 4. Process handling differences
# ps command is available but limited
ps -W  # Show Windows processes

# 5. File permissions are simulated
# chmod works but doesn't map directly to Windows ACLs

# 6. Symbolic links require administrator privileges
# Or Developer Mode enabled in Windows 10+
```

### Windows-Specific Workarounds

```bash
# Run PowerShell commands from Git Bash
run_powershell() {
    local command=$1
    powershell.exe -Command "$command"
}

# Example: Get Windows version
run_powershell "Get-ComputerInfo | Select-Object WindowsVersion"

# Run cmd.exe commands
run_cmd() {
    local command=$1
    cmd.exe /c "$command"
}

# Example: Set Windows environment variable
run_cmd "setx MY_VAR value"

# Check if running with admin privileges
is_admin() {
    net session &> /dev/null
    return $?
}

if is_admin; then
    echo "Running with administrator privileges"
else
    echo "Not running as administrator"
fi

# Windows line endings (CRLF vs LF)
fix_line_endings() {
    local file=$1

    # Convert CRLF to LF
    dos2unix "$file"

    # Or with sed
    sed -i 's/\r$//' "$file"

    # Convert LF to CRLF
    unix2dos "$file"

    # Or with sed
    sed -i 's/$/\r/' "$file"
}
```

### Git Bash Best Practices

```bash
# Always handle spaces in Windows paths
process_file() {
    local file="$1"  # Always quote!

    # Windows paths often have spaces
    # C:\Program Files\...
}

# Use forward slashes when possible
cd /c/Program\ Files/Git  # Works
cd "C:\Program Files\Git" # Also works, but...
cd C:\\Program\ Files\\Git # Avoid

# Set Git config for line endings
git config --global core.autocrlf true  # Windows
git config --global core.autocrlf input # Linux/macOS

# Check Git Bash version
bash --version
uname -a  # Shows MINGW or MSYS
```

---

## Windows (WSL)

### WSL1 vs WSL2

```bash
# Detect WSL version
detect_wsl_version() {
    if grep -qi microsoft /proc/version; then
        if [[ $(uname -r) =~ microsoft ]]; then
            echo "WSL 1"
        elif [[ $(uname -r) =~ WSL2 ]]; then
            echo "WSL 2"
        else
            # Check kernel version
            if [[ $(uname -r) =~ ^4\. ]]; then
                echo "WSL 1"
            else
                echo "WSL 2"
            fi
        fi
    else
        echo "Not WSL"
    fi
}

# WSL1 limitations:
# - No full syscall compatibility
# - File I/O slower on Windows filesystem
# - No Docker/containers (needs WSL2)

# WSL2 improvements:
# - Full Linux kernel
# - Better filesystem performance
# - Docker/container support
# - Near-native Linux performance
```

### Windows Filesystem Access

```bash
# Access Windows drives from WSL
# Mounted at /mnt/c, /mnt/d, etc.

# List Windows drives
ls /mnt/

# Access Windows user directory
WINDOWS_HOME="/mnt/c/Users/$USER"
cd "$WINDOWS_HOME"

# File permissions on Windows filesystem
# Files on /mnt/c are owned by root but accessible
# Permissions are simulated

# Best practice: Use WSL filesystem for Linux files
# Use /home/username, not /mnt/c/...
# Much faster, especially in WSL1
```

### WSL Interoperability

```bash
# Run Windows executables from WSL
# .exe files are automatically executable

# Run Windows commands
cmd.exe /c dir
notepad.exe file.txt
explorer.exe .  # Open current directory in Windows Explorer

# Run PowerShell
powershell.exe -Command "Get-Date"

# Pipe between Linux and Windows
cat file.txt | clip.exe  # Copy to Windows clipboard

# Environment variables
# Windows environment is accessible with WSLENV

# Share environment variable from Windows to WSL
# In PowerShell:
# $env:WSLENV = "MYVAR/p"
# This converts Windows paths to WSL paths
```

### WSL-Specific Configuration

```bash
# /etc/wsl.conf configuration
cat > /etc/wsl.conf << 'EOF'
[automount]
enabled = true
root = /mnt/
options = "metadata,umask=22,fmask=11"

[network]
generateHosts = true
generateResolvConf = true

[interop]
enabled = true
appendWindowsPath = true
EOF

# Apply: wsl.exe --shutdown (from PowerShell)

# Network differences
# WSL1: Shares network with Windows
# WSL2: NAT network, different IP

# Get WSL IP address
ip addr show eth0 | grep -oP '(?<=inet\s)\d+(\.\d+){3}'

# Access Windows services from WSL2
# Use Windows IP, not localhost
# Or use: localhost (WSL2 has localhost forwarding)
```

---

## Container Environments

### Docker Considerations

```bash
# Minimal base images often lack bash
# alpine: Only has /bin/sh by default
# debian:slim: Has bash
# ubuntu: Has bash

# Check if bash is available
if [ -f /bin/bash ]; then
    exec /bin/bash "$@"
else
    exec /bin/sh "$@"
fi

# Container detection
is_docker() {
    if [[ -f /.dockerenv ]] || grep -q docker /proc/1/cgroup 2>/dev/null; then
        return 0
    else
        return 1
    fi
}

# PID 1 problem in containers
# Your script might be PID 1, which means:
# - Zombie process reaping is your responsibility
# - Signals behave differently

# Solution: Use tini or dumb-init
# Or handle signals explicitly
handle_sigterm() {
    # Forward to child processes
    kill -TERM "$child_pid" 2>/dev/null
    wait "$child_pid"
    exit 0
}

trap handle_sigterm SIGTERM

# Start main process
main_process &
child_pid=$!
wait "$child_pid"
```

### Kubernetes Considerations

```bash
# Kubernetes-specific environment variables
if [[ -n "$KUBERNETES_SERVICE_HOST" ]]; then
    echo "Running in Kubernetes"

    # Access Kubernetes API
    KUBE_TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
    KUBE_CA=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt

    # Get pod name
    POD_NAME=${POD_NAME:-$(hostname)}

    # Get namespace
    NAMESPACE=$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace)
fi

# Health checks
# Kubernetes expects:
# - HTTP probe on specific port
# - Or command that exits 0 for success

# Liveness probe handler
handle_health_check() {
    # Check if application is healthy
    if check_health; then
        exit 0
    else
        exit 1
    fi
}

# Readiness probe handler
handle_readiness_check() {
    # Check if ready to serve traffic
    if is_ready; then
        exit 0
    else
        exit 1
    fi
}

# Graceful shutdown for rolling updates
# Kubernetes sends SIGTERM, waits (default 30s), then SIGKILL
trap 'graceful_shutdown' SIGTERM

graceful_shutdown() {
    echo "Received SIGTERM, shutting down gracefully..."

    # Stop accepting new connections
    # Finish processing existing requests
    # Close connections
    # Exit

    exit 0
}
```

### Container Best Practices

```bash
# Don't assume specific users/groups exist
# Many containers run as non-root or random UID

# Check current user
if [[ $EUID -eq 0 ]]; then
    echo "Running as root"
else
    echo "Running as user $EUID"
fi

# Handle arbitrary UIDs (OpenShift)
# Files in mounted volumes may not be owned by container user
# Solution: Add current user to group, use group permissions

# Minimal dependencies
# Container images should be small
# Don't install unnecessary packages

# Use absolute paths or set PATH explicitly
export PATH=/usr/local/bin:/usr/bin:/bin

# Environment variables for configuration
# Don't hardcode values, use env vars
DATABASE_URL=${DATABASE_URL:-postgres://localhost/db}

# Logging to stdout/stderr
# Container orchestrators capture these
echo "Log message"       # To stdout
echo "Error message" >&2 # To stderr

# Don't write to filesystem (except for tmpfs)
# Containers are ephemeral
# Use volumes for persistent data
```

---

## Cross-Platform Patterns

### Portable Command Wrapper

```bash
# Create wrappers for platform-specific commands
setup_portable_commands() {
    local os
    os=$(detect_os)

    case "$os" in
        linux)
            SED=sed
            READLINK="readlink -f"
            DATE=date
            STAT="stat -c"
            GREP=grep
            ;;
        macos)
            # Prefer GNU versions if available
            SED=$(command -v gsed || echo sed)
            READLINK=$(command -v greadlink || echo "echo")  # No -f on BSD
            DATE=$(command -v gdate || echo date)
            STAT=$(command -v gstat || echo stat)
            GREP=$(command -v ggrep || echo grep)
            ;;
        gitbash)
            SED=sed
            READLINK=readlink  # Git Bash has GNU tools
            DATE=date
            STAT=stat
            GREP=grep
            ;;
    esac

    export SED READLINK DATE STAT GREP
}

# Use the wrappers
setup_portable_commands
$SED -i 's/old/new/g' file.txt
```

### Cross-Platform Temp Files

```bash
# Portable temporary file creation
create_temp_file() {
    # Works on all platforms
    local temp_file
    temp_file=$(mktemp) || {
        # Fallback if mktemp doesn't exist
        temp_file="/tmp/script.$$.$RANDOM"
        touch "$temp_file"
    }

    echo "$temp_file"
}

# Portable temporary directory
create_temp_dir() {
    local temp_dir
    temp_dir=$(mktemp -d) || {
        # Fallback
        temp_dir="/tmp/script.$$.$RANDOM"
        mkdir -p "$temp_dir"
    }

    echo "$temp_dir"
}

# Clean up temp files on exit
TEMP_DIR=$(create_temp_dir)
trap 'rm -rf "$TEMP_DIR"' EXIT
```

### Cross-Platform File Paths

```bash
# Normalize paths across platforms
normalize_path() {
    local path="$1"

    # Remove trailing slashes
    path="${path%/}"

    # Convert backslashes to forward slashes (Windows)
    path="${path//\\//}"

    # Resolve . and ..
    # Use Python for reliable normalization
    if command -v python3 &> /dev/null; then
        path=$(python3 -c "import os; print(os.path.normpath('$path'))")
    elif command -v python &> /dev/null; then
        path=$(python -c "import os; print(os.path.normpath('$path'))")
    fi

    echo "$path"
}

# Get absolute path (cross-platform)
get_absolute_path() {
    local path="$1"

    # Try readlink -f (Linux, Git Bash)
    if readlink -f "$path" &> /dev/null; then
        readlink -f "$path"
    # Try realpath (most platforms)
    elif command -v realpath &> /dev/null; then
        realpath "$path"
    # Fallback to Python
    elif command -v python3 &> /dev/null; then
        python3 -c "import os; print(os.path.abspath('$path'))"
    # Fallback to cd
    elif [[ -d "$path" ]]; then
        (cd "$path" && pwd)
    else
        (cd "$(dirname "$path")" && echo "$(pwd)/$(basename "$path")")
    fi
}
```

### Cross-Platform Process Management

```bash
# Find process by name (cross-platform)
find_process() {
    local process_name="$1"

    if command -v pgrep &> /dev/null; then
        pgrep -f "$process_name"
    else
        ps aux | grep "$process_name" | grep -v grep | awk '{print $2}'
    fi
}

# Kill process by name (cross-platform)
kill_process() {
    local process_name="$1"

    if command -v pkill &> /dev/null; then
        pkill -f "$process_name"
    else
        local pids
        pids=$(find_process "$process_name")
        if [[ -n "$pids" ]]; then
            kill $pids
        fi
    fi
}
```

---

## Command Compatibility Matrix

| Command | Linux | macOS | Git Bash | Notes |
|---------|-------|-------|----------|-------|
| `sed -i` | ✓ | ✓* | ✓ | macOS needs `sed -i ''` |
| `date -d` | ✓ | ✗ | ✓ | macOS uses `-v` |
| `readlink -f` | ✓ | ✗ | ✓ | macOS needs `greadlink` |
| `stat -c` | ✓ | ✗ | ✓ | macOS uses `-f` |
| `grep -P` | ✓ | ✗ | ✓ | macOS doesn't support PCRE |
| `find -printf` | ✓ | ✗ | ✓ | macOS doesn't have `-printf` |
| `xargs -r` | ✓ | ✗ | ✓ | macOS doesn't have `-r` |
| `ps aux` | ✓ | ✓ | ✓* | Git Bash has limited output |
| `ls --color` | ✓ | ✗ | ✓ | macOS uses `-G` |
| `du -b` | ✓ | ✗ | ✓ | macOS doesn't support bytes |
| `mktemp` | ✓ | ✓ | ✓ | Works on all platforms |
| `timeout` | ✓ | ✗ | ✓ | macOS needs `gtimeout` |

**Legend:**
- ✓ = Supported
- ✗ = Not supported
- ✓* = Supported with limitations

---

## Testing Across Platforms

```bash
# Test script on multiple platforms
test_platforms() {
    local script="$1"

    echo "Testing on current platform: $(detect_os)"
    bash -n "$script" || {
        echo "Syntax error!"
        return 1
    }

    # Run ShellCheck
    if command -v shellcheck &> /dev/null; then
        shellcheck "$script" || return 1
    fi

    # Run the script
    bash "$script" || return 1

    echo "Tests passed on $(detect_os)"
}

# CI/CD matrix testing
# Use GitHub Actions, GitLab CI, etc. to test on multiple platforms
```

**Example GitHub Actions matrix:**
```yaml
jobs:
  test:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
    steps:
      - uses: actions/checkout@v3
      - name: Test script
        run: bash test.sh
```

---

## Summary

**Key takeaways for cross-platform bash scripts:**

1. **Always detect the platform** before using platform-specific features
2. **Use portable commands** or provide fallbacks
3. **Test on all target platforms** (CI/CD with matrix builds)
4. **Avoid platform-specific assumptions** (file paths, users, services)
5. **Use ShellCheck** to catch portability issues
6. **Prefer POSIX compliance** when possible for maximum portability
7. **Document platform requirements** in script comments
8. **Provide GNU alternatives** on macOS when needed
9. **Handle path differences** carefully (especially Windows)
10. **Test in containers** if that's your deployment target

For maximum portability: stick to POSIX shell (`#!/bin/sh`) and avoid bashisms unless you control the deployment environment.
