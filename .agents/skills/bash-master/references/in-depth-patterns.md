# Bash in-depth patterns reference

Detailed patterns extracted from `SKILL.md` to keep the core skill at a navigable size. Covers function design, security, performance, testing, debugging, and advanced patterns. Pair with `best_practices.md`, `patterns_antipatterns.md`, and `platform_specifics.md` for full breadth.

## Function Design

```bash
# Good function structure
function_name() {
    # 1. Local variables first
    local arg1="$1"
    local arg2="${2:-default_value}"
    local result=""

    # 2. Input validation
    if [[ -z "$arg1" ]]; then
        echo "Error: arg1 is required" >&2
        return 1
    fi

    # 3. Main logic
    result=$(some_operation "$arg1" "$arg2")

    # 4. Output/return
    echo "$result"
    return 0
}

# Use functions, not scripts-in-scripts
# Benefits: testability, reusability, namespacing
```

## Variable Naming

```bash
# Constants: UPPER_CASE
readonly MAX_RETRIES=3
readonly CONFIG_FILE="/etc/app/config.conf"

# Global variables: UPPER_CASE or lower_case (be consistent)
GLOBAL_STATE="initialized"

# Local variables: lower_case
local user_name="john"
local file_count=0

# Environment variables: UPPER_CASE (by convention)
export DATABASE_URL="postgres://..."

# Readonly when possible
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
```

## Error Handling

```bash
# Method 1: Check exit codes explicitly
if ! command_that_might_fail; then
    echo "Error: Command failed" >&2
    return 1
fi

# Method 2: Use || for alternative actions
command_that_might_fail || {
    echo "Error: Command failed" >&2
    return 1
}

# Method 3: Trap for cleanup
cleanup() {
    local exit_code=$?
    rm -f "$TEMP_FILE"
    exit "$exit_code"
}
trap cleanup EXIT

# Method 4: Custom error handler
error_exit() {
    local message="$1"
    local code="${2:-1}"
    echo "Error: $message" >&2
    exit "$code"
}

# Usage
[[ -f "$config_file" ]] || error_exit "Config file not found: $config_file"
```

## Input Validation

```bash
validate_input() {
    local input="$1"

    if [[ -z "$input" ]]; then
        echo "Error: Input cannot be empty" >&2
        return 1
    fi

    if [[ ! "$input" =~ ^[a-zA-Z0-9_-]+$ ]]; then
        echo "Error: Input contains invalid characters" >&2
        return 1
    fi

    if [[ ${#input} -gt 255 ]]; then
        echo "Error: Input too long (max 255 characters)" >&2
        return 1
    fi

    return 0
}

read -r user_input
if validate_input "$user_input"; then
    process "$user_input"
fi
```

## Argument Parsing

```bash
usage() {
    cat <<EOF
Usage: $SCRIPT_NAME [OPTIONS] <command>

Options:
    -h, --help          Show this help
    -v, --verbose       Verbose output
    -f, --file FILE     Input file
    -o, --output DIR    Output directory

Commands:
    build               Build the project
    test                Run tests
EOF
}

main() {
    local verbose=false
    local input_file=""
    local output_dir="."
    local command=""

    while [[ $# -gt 0 ]]; do
        case "$1" in
            -h|--help) usage; exit 0 ;;
            -v|--verbose) verbose=true; shift ;;
            -f|--file) input_file="$2"; shift 2 ;;
            -o|--output) output_dir="$2"; shift 2 ;;
            -*)
                echo "Error: Unknown option: $1" >&2
                usage >&2; exit 1 ;;
            *) command="$1"; shift; break ;;
        esac
    done

    if [[ -z "$command" ]]; then
        echo "Error: Command is required" >&2
        usage >&2; exit 1
    fi

    case "$command" in
        build) do_build ;;
        test)  do_test ;;
        *)
            echo "Error: Unknown command: $command" >&2
            usage >&2; exit 1 ;;
    esac
}

main "$@"
```

## Logging

```bash
readonly LOG_LEVEL_DEBUG=0
readonly LOG_LEVEL_INFO=1
readonly LOG_LEVEL_WARN=2
readonly LOG_LEVEL_ERROR=3

LOG_LEVEL=${LOG_LEVEL:-$LOG_LEVEL_INFO}

log_debug() { [[ $LOG_LEVEL -le $LOG_LEVEL_DEBUG ]] && echo "[DEBUG] $*" >&2; }
log_info()  { [[ $LOG_LEVEL -le $LOG_LEVEL_INFO  ]] && echo "[INFO]  $*" >&2; }
log_warn()  { [[ $LOG_LEVEL -le $LOG_LEVEL_WARN  ]] && echo "[WARN]  $*" >&2; }
log_error() { [[ $LOG_LEVEL -le $LOG_LEVEL_ERROR ]] && echo "[ERROR] $*" >&2; }

log_with_timestamp() {
    local level="$1"
    shift
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] [$level] $*" >&2
}

log_info "Starting process"
log_error "Failed to connect to database"
```

## Security

### Command Injection Prevention

```bash
# NEVER use eval with user input - DANGEROUS
eval "$user_input"                       # WRONG
eval "var_$user_input=value"             # WRONG

# NEVER concatenate user input into commands
grep "$user_pattern" file.txt            # If pattern contains -e flag, injection possible

# Use arrays
grep_args=("$user_pattern" "file.txt")
grep "${grep_args[@]}"

# Use -- to separate options from arguments
grep -- "$user_pattern" file.txt
```

### Path Traversal Prevention

```bash
sanitize_path() {
    local path="$1"
    path="${path//..\/}"
    path="${path//\/..\//}"
    path="${path#/}"
    echo "$path"
}

is_safe_path() {
    local file_path="$1"
    local base_dir="$2"

    local real_path
    real_path=$(readlink -f "$file_path" 2>/dev/null) || return 1
    local real_base
    real_base=$(readlink -f "$base_dir" 2>/dev/null) || return 1

    [[ "$real_path" == "$real_base"/* ]]
}

if is_safe_path "$user_file" "/var/app/data"; then
    process_file "$user_file"
else
    echo "Error: Invalid file path" >&2
    exit 1
fi
```

### Privilege Management

```bash
# Refuse to run as root
if [[ $EUID -eq 0 ]]; then
    echo "Error: Do not run this script as root" >&2
    exit 1
fi

drop_privileges() {
    local user="$1"
    if [[ $EUID -eq 0 ]]; then
        exec sudo -u "$user" "$0" "$@"
    fi
}

run_as_root() {
    if [[ $EUID -ne 0 ]]; then
        sudo "$@"
    else
        "$@"
    fi
}
```

### Temporary File Handling

```bash
readonly TEMP_DIR=$(mktemp -d)
readonly TEMP_FILE=$(mktemp)

cleanup() {
    rm -rf "$TEMP_DIR"
    rm -f "$TEMP_FILE"
}
trap cleanup EXIT

# Secure temporary file (owner-only)
secure_temp=$(mktemp)
chmod 600 "$secure_temp"
```

## Performance Optimization

### Avoid Unnecessary Subshells

```bash
# SLOW - subshell per iteration
while IFS= read -r line; do
    count=$(echo "$count + 1" | bc)
done < file.txt

# FAST - bash arithmetic
count=0
while IFS= read -r line; do
    ((count++))
done < file.txt
```

### Use Bash Built-ins

```bash
# Slow: external commands
length=$(echo "$string" | wc -c)
upper=$(echo "$string" | tr '[:lower:]' '[:upper:]')

# Fast: bash built-ins
length=${#string}
upper=${string^^}

# String contains check
if [[ "$haystack" == *"$needle"* ]]; then
    echo "Found"
fi
```

### Process Substitution vs Pipes

```bash
# Pipes start a subshell - variables don't persist
count=0
echo "data" | while read -r line; do
    ((count++))            # changes lost in subshell
done

# Process substitution keeps the current shell
count=0
while read -r line; do
    ((count++))
done < <(echo "data")
```

### Array Operations

```bash
arr=("one" "two" "three")

length=${#arr[@]}
last_index=$((${#arr[@]} - 1))

arr+=("four")
unset 'arr[1]'              # remove by index
arr=("${arr[@]}")           # reindex after unset

# Iterate
for item in "${arr[@]}"; do
    echo "$item"
done

# Iterate with index
for i in "${!arr[@]}"; do
    echo "$i: ${arr[$i]}"
done
```

## Testing

### Unit Testing with BATS

```bash
# test/script.bats
#!/usr/bin/env bats

load '../script.sh'

@test "function returns correct value" {
    result=$(my_function "input")
    [ "$result" = "expected" ]
}

@test "function handles empty input" {
    run my_function ""
    [ "$status" -eq 1 ]
    [ "${lines[0]}" = "Error: Input cannot be empty" ]
}

@test "function validates input format" {
    run my_function "invalid@input"
    [ "$status" -eq 1 ]
}

# Run: bats test/script.bats
```

### Integration Testing

```bash
# integration_test.sh
#!/usr/bin/env bash
set -euo pipefail

setup() {
    export TEST_DIR=$(mktemp -d)
    export TEST_FILE="$TEST_DIR/test.txt"
}

teardown() {
    rm -rf "$TEST_DIR"
}

test_file_creation() {
    ./script.sh create "$TEST_FILE"
    if [[ ! -f "$TEST_FILE" ]]; then
        echo "FAIL: File was not created"
        return 1
    fi
    echo "PASS: File creation works"
}

main() {
    setup
    trap teardown EXIT
    test_file_creation || exit 1
    echo "All tests passed"
}

main
```

### CI/CD Integration

```yaml
# .github/workflows/test.yml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install shellcheck
        run: sudo apt-get install -y shellcheck

      - name: Run shellcheck
        run: find . -name "*.sh" -exec shellcheck {} +

      - name: Install bats
        run: |
          git clone https://github.com/bats-core/bats-core.git
          cd bats-core
          sudo ./install.sh /usr/local

      - name: Run tests
        run: bats test/
```

## Debugging Techniques

### Debug Mode

```bash
# Method 1: set -x
set -x
command1
command2
set +x

# Method 2: PS4 for richer trace output
export PS4='+(${BASH_SOURCE}:${LINENO}): ${FUNCNAME[0]:+${FUNCNAME[0]}(): }'
set -x

# Method 3: Conditional debugging
DEBUG=${DEBUG:-false}
debug() {
    if [[ "$DEBUG" == "true" ]]; then
        echo "[DEBUG] $*" >&2
    fi
}
# Usage: DEBUG=true ./script.sh
```

### Tracing and Profiling

```bash
trace() {
    echo "[TRACE] Function: ${FUNCNAME[1]}, Args: $*" >&2
}

my_function() {
    trace "$@"
    # Function logic
}

profile() {
    local start=$(date +%s%N)
    "$@"
    local end=$(date +%s%N)
    local duration=$(( (end - start) / 1000000 ))
    echo "[PROFILE] Command '$*' took ${duration}ms" >&2
}

# Usage
profile slow_command arg1 arg2
```

### Common Issues and Solutions

```bash
# Script works in bash but not in sh - check bashisms
checkbashisms script.sh

# Works locally but not on server - check PATH/env
env
echo "$PATH"

# Whitespace in filenames breaking script - always quote
for file in *.txt; do
    process "$file"           # not: process $file
done

# Different behavior in cron - set PATH explicitly
PATH=/usr/local/bin:/usr/bin:/bin
export PATH
```

## Advanced Patterns

### Configuration File Parsing

```bash
# Simple sourcing (dangerous if file not trusted)
load_config() {
    local config_file="$1"
    if [[ ! -f "$config_file" ]]; then
        echo "Error: Config file not found: $config_file" >&2
        return 1
    fi
    # shellcheck source=/dev/null
    source "$config_file"
}

# Safe parsing - no code execution
read_config() {
    local config_file="$1"
    while IFS='=' read -r key value; do
        [[ "$key" =~ ^[[:space:]]*# ]] && continue
        [[ -z "$key" ]] && continue
        key=$(echo "$key" | tr -d ' ')
        value=$(echo "$value" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
        declare -g "$key=$value"
    done < "$config_file"
}
```

### Parallel Processing

```bash
process_files_parallel() {
    local max_jobs=4
    local job_count=0

    for file in *.txt; do
        process_file "$file" &
        ((job_count++))
        if [[ $job_count -ge $max_jobs ]]; then
            wait -n
            ((job_count--))
        fi
    done

    wait
}

# Using GNU Parallel
parallel_with_gnu() {
    parallel -j 4 process_file ::: *.txt
}
```

### Signal Handling

```bash
shutdown_requested=false

handle_sigterm() {
    echo "Received SIGTERM, shutting down gracefully..." >&2
    shutdown_requested=true
}

trap handle_sigterm SIGTERM SIGINT

main_loop() {
    while [[ "$shutdown_requested" == "false" ]]; do
        sleep 1
    done
    echo "Shutdown complete" >&2
}

main_loop
```

### Retries with Exponential Backoff

```bash
retry_with_backoff() {
    local max_attempts=5
    local timeout=1
    local attempt=1
    local exitCode=0

    while [[ $attempt -le $max_attempts ]]; do
        if "$@"; then
            return 0
        else
            exitCode=$?
        fi
        echo "Attempt $attempt failed! Retrying in $timeout seconds..." >&2
        sleep "$timeout"
        attempt=$((attempt + 1))
        timeout=$((timeout * 2))
    done

    echo "Command failed after $max_attempts attempts!" >&2
    return "$exitCode"
}

# Usage
retry_with_backoff curl -f https://api.example.com/health
```
