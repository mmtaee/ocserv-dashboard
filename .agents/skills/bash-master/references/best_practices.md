# Bash Scripting Best Practices & Industry Standards

Comprehensive guide to professional bash scripting following industry standards including Google Shell Style Guide, ShellCheck recommendations, and community best practices.

---

## Table of Contents

1. [Script Structure](#script-structure)
2. [Safety and Robustness](#safety-and-robustness)
3. [Style Guidelines](#style-guidelines)
4. [Functions](#functions)
5. [Variables](#variables)
6. [Error Handling](#error-handling)
7. [Input/Output](#inputoutput)
8. [Security](#security)
9. [Performance](#performance)
10. [Documentation](#documentation)
11. [Testing](#testing)
12. [Maintenance](#maintenance)

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

## Script Structure

### Standard Template

```bash
#!/usr/bin/env bash
#
# Script Name: script_name.sh
# Description: Brief description of what this script does
# Author: Your Name
# Date: 2024-01-01
# Version: 1.0.0
#
# Usage: script_name.sh [OPTIONS] <arguments>
#
# Options:
#   -h, --help    Show help message
#   -v, --verbose Enable verbose output
#
# Dependencies:
#   - bash >= 4.0
#   - jq
#   - curl
#
# Exit Codes:
#   0 - Success
#   1 - General error
#   2 - Invalid arguments
#   3 - Missing dependency
#

set -euo pipefail
IFS=$'\n\t'

# Script metadata
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_NAME="$(basename "${BASH_SOURCE[0]}")"
readonly SCRIPT_VERSION="1.0.0"

# Global constants
readonly DEFAULT_TIMEOUT=30
readonly CONFIG_FILE="${CONFIG_FILE:-$SCRIPT_DIR/config.conf}"

# Global variables
VERBOSE=false
DRY_RUN=false

#------------------------------------------------------------------------------
# Functions
#------------------------------------------------------------------------------

# Show usage information
usage() {
    cat <<EOF
Usage: $SCRIPT_NAME [OPTIONS] <command>

Description of what the script does.

OPTIONS:
    -h, --help          Show this help message
    -v, --verbose       Enable verbose output
    -n, --dry-run       Show what would be done without doing it
    -V, --version       Show version

COMMANDS:
    build               Build the project
    test                Run tests
    deploy              Deploy to production

EXAMPLES:
    $SCRIPT_NAME build
    $SCRIPT_NAME --verbose test
    $SCRIPT_NAME deploy --dry-run

EOF
}

# Cleanup function
cleanup() {
    local exit_code=$?
    # Remove temporary files
    [[ -n "${TEMP_DIR:-}" ]] && rm -rf "$TEMP_DIR"
    exit "$exit_code"
}

# Main function
main() {
    # Parse arguments
    parse_arguments "$@"

    # Validate dependencies
    check_dependencies

    # Main script logic here
    echo "Script execution complete"
}

#------------------------------------------------------------------------------
# Script execution
#------------------------------------------------------------------------------

# Set up cleanup trap
trap cleanup EXIT INT TERM

# Run main function with all arguments
main "$@"
```

### File Organization

```bash
# For larger projects, organize code into modules

# project/
# ├── bin/
# │   └── main.sh           # Entry point
# ├── lib/
# │   ├── common.sh         # Shared utilities
# │   ├── config.sh         # Configuration handling
# │   └── logger.sh         # Logging functions
# ├── config/
# │   └── default.conf      # Default configuration
# ├── test/
# │   ├── test_common.bats  # Unit tests
# │   └── test_config.bats
# └── README.md

# In main.sh:
# Source library files
readonly LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../lib" && pwd)"

# shellcheck source=lib/common.sh
source "$LIB_DIR/common.sh"
# shellcheck source=lib/logger.sh
source "$LIB_DIR/logger.sh"
```

---

## Safety and Robustness

### Essential Safety Settings

```bash
# ALWAYS use these at the start of scripts
set -e          # Exit immediately if a command exits with a non-zero status
set -u          # Treat unset variables as an error
set -o pipefail # Return value of a pipeline is status of last command to exit with non-zero status
set -E          # ERR trap is inherited by shell functions

# Optionally add:
set -x          # Print commands before executing (debugging)
set -C          # Prevent output redirection from overwriting existing files
```

### Safe Word Splitting

```bash
# Default IFS causes issues with filenames containing spaces
# OLD IFS: space, tab, newline
IFS=$' \t\n'

# SAFE IFS: only tab and newline
IFS=$'\n\t'

# This prevents word splitting on spaces, which is a common source of bugs:
files="file1.txt file2.txt"
for file in $files; do  # Without proper IFS, this splits on spaces!
    echo "$file"
done
```

### Quoting Rules

```bash
# ALWAYS quote variable expansions
command "$variable"           # ✓ CORRECT
command $variable             # ✗ WRONG (word splitting and globbing)

# Arrays: Proper expansion
files=("file1.txt" "file 2.txt" "file 3.txt")
process "${files[@]}"         # ✓ CORRECT (each element separate)
process "${files[*]}"         # ✗ WRONG (all elements as one string)
process ${files[@]}           # ✗ WRONG (unquoted, word splitting)

# Command substitution: Quote the result
result="$(command)"           # ✓ CORRECT
result=$(command)             # ✗ WRONG (unless word splitting is desired)

# Glob patterns: Don't quote when you want globbing
for file in *.txt; do         # ✓ CORRECT (globbing intended)
    echo "$file"              # ✓ CORRECT (no globbing inside)
done

for file in "*.txt"; do       # ✗ WRONG (literal "*.txt", no globbing)
    echo "$file"
done
```

### Handling Special Characters

```bash
# Filenames with special characters
# Use quotes and proper escaping

# Create array from find output
mapfile -t files < <(find . -name "*.txt" -print0 | xargs -0)

# Or modern bash:
files=()
while IFS= read -r -d '' file; do
    files+=("$file")
done < <(find . -name "*.txt" -print0)

# Process files safely
for file in "${files[@]}"; do
    [[ -f "$file" ]] && process "$file"
done
```

---

## Style Guidelines

Based on Google Shell Style Guide and community standards.

### Naming Conventions

```bash
# Constants: UPPER_CASE with underscores
readonly MAX_RETRIES=3
readonly DEFAULT_TIMEOUT=30
readonly CONFIG_DIR="/etc/myapp"

# Environment variables: UPPER_CASE (by convention)
export DATABASE_URL="postgres://localhost/db"
export LOG_LEVEL="INFO"

# Global variables: UPPER_CASE or lower_case (be consistent in your project)
GLOBAL_COUNTER=0
current_state="initialized"

# Local variables: lower_case with underscores
local user_name="john"
local file_count=0
local error_message=""

# Functions: lower_case with underscores
function_name() {
    local var="value"
}

# Private functions: Prefix with underscore
_internal_function() {
    # Helper function not meant to be called externally
}
```

### Indentation and Formatting

```bash
# Use 4 spaces for indentation (not tabs)
# Or 2 spaces (be consistent)

# Function definition
my_function() {
    local arg="$1"

    if [[ -n "$arg" ]]; then
        echo "Processing $arg"
    else
        echo "No argument provided"
        return 1
    fi

    return 0
}

# Conditional blocks
if [[ condition ]]; then
    # code
elif [[ other_condition ]]; then
    # code
else
    # code
fi

# Loops
for item in "${array[@]}"; do
    # code
done

while [[ condition ]]; do
    # code
done

# Case statement
case "$variable" in
    pattern1)
        # code
        ;;
    pattern2)
        # code
        ;;
    *)
        # default
        ;;
esac

# Line length: Prefer < 80 characters, max 100
# Break long lines with backslash
long_command \
    --option1 value1 \
    --option2 value2 \
    --option3 value3

# Or use arrays for readability
command_args=(
    --option1 value1
    --option2 value2
    --option3 value3
)
command "${command_args[@]}"
```

### Comments

```bash
# Single-line comments: Start with # followed by space
# This is a comment

# Function documentation (before function definition)
#######################################
# Description of what this function does
# Globals:
#   GLOBAL_VAR - Description
# Arguments:
#   $1 - First argument description
#   $2 - Second argument description (optional)
# Outputs:
#   Writes result to stdout
# Returns:
#   0 on success, non-zero on error
#######################################
my_function() {
    # Implementation
}

# Inline comments: Use sparingly, only when necessary
result=$(complex_calculation)  # Result in milliseconds

# TODO comments
# TODO(username): Description of what needs to be done
# FIXME(username): Description of what needs to be fixed
# HACK(username): Description of workaround and why it's needed

# Section separators for long scripts
#------------------------------------------------------------------------------
# Configuration Section
#------------------------------------------------------------------------------

#######################################
# Database Functions
#######################################
```

### Test Constructs

```bash
# Prefer [[ ]] over [ ] for tests in bash
# [[ ]] is a bash keyword with better behavior:
# - No word splitting
# - No pathname expansion
# - More operators available

# String comparison
if [[ "$string1" == "$string2" ]]; then    # ✓ CORRECT
if [ "$string1" = "$string2" ]; then       # ✓ CORRECT (POSIX)
if [ $string1 == $string2 ]; then          # ✗ WRONG (word splitting, not POSIX)

# String matching with patterns
if [[ "$file" == *.txt ]]; then            # ✓ CORRECT (pattern matching)
if [[ "$file" =~ \.txt$ ]]; then           # ✓ CORRECT (regex)

# Numeric comparison
if [[ $num -gt 10 ]]; then                 # ✓ CORRECT
if (( num > 10 )); then                    # ✓ CORRECT (arithmetic context)

# File tests
if [[ -f "$file" ]]; then                  # ✓ CORRECT (regular file)
if [[ -d "$dir" ]]; then                   # ✓ CORRECT (directory)
if [[ -e "$path" ]]; then                  # ✓ CORRECT (exists)
if [[ -r "$file" ]]; then                  # ✓ CORRECT (readable)
if [[ -w "$file" ]]; then                  # ✓ CORRECT (writable)
if [[ -x "$file" ]]; then                  # ✓ CORRECT (executable)

# Logical operators
if [[ condition1 && condition2 ]]; then    # ✓ CORRECT (AND)
if [[ condition1 || condition2 ]]; then    # ✓ CORRECT (OR)
if [[ ! condition ]]; then                 # ✓ CORRECT (NOT)

# Empty/non-empty string
if [[ -z "$var" ]]; then                   # ✓ CORRECT (empty)
if [[ -n "$var" ]]; then                   # ✓ CORRECT (non-empty)
```

---

## Functions

### Function Best Practices

```bash
# Good function structure
process_file() {
    # 1. Declare local variables
    local file="$1"
    local output_dir="${2:-.}"  # Default to current directory
    local result=""

    # 2. Input validation
    if [[ ! -f "$file" ]]; then
        echo "Error: File not found: $file" >&2
        return 1
    fi

    if [[ ! -d "$output_dir" ]]; then
        echo "Error: Output directory not found: $output_dir" >&2
        return 1
    fi

    # 3. Main logic
    result=$(perform_operation "$file")

    # 4. Output
    echo "$result" > "$output_dir/result.txt"

    # 5. Return status
    return 0
}

# Use return codes to indicate success/failure
# 0 = success, non-zero = error
validate_input() {
    local input="$1"

    if [[ ! "$input" =~ ^[a-zA-Z0-9]+$ ]]; then
        return 1  # Invalid input
    fi

    return 0  # Valid input
}

# Usage
if validate_input "$user_input"; then
    process "$user_input"
else
    echo "Invalid input" >&2
    exit 1
fi
```

### Function Documentation

```bash
#######################################
# Process a file and generate output
# Globals:
#   OUTPUT_FORMAT - Output format (json/xml/csv)
# Arguments:
#   $1 - Input file path (required)
#   $2 - Output directory (optional, default: .)
# Outputs:
#   Writes processed data to stdout
#   Writes result file to output directory
# Returns:
#   0 on success
#   1 if file not found
#   2 if processing fails
# Example:
#   process_file "input.txt" "/tmp/output"
#######################################
process_file() {
    # Implementation
}
```

### Local Variables

```bash
# ALWAYS use local for function variables
bad_function() {
    counter=0  # ✗ WRONG - Global variable!
}

good_function() {
    local counter=0  # ✓ CORRECT - Local to function
}

# Declare local before assignment
good_practice() {
    local result
    result=$(command_that_might_fail) || return 1
    echo "$result"
}

# This won't catch command failure:
bad_practice() {
    local result=$(command_that_might_fail)  # ✗ WRONG
    echo "$result"
}
```

---

## Variables

### Variable Declaration

```bash
# Readonly for constants
readonly MAX_RETRIES=3
declare -r MAX_RETRIES=3  # Alternative syntax

# Arrays
files=("file1.txt" "file2.txt" "file3.txt")
declare -a files=("file1.txt" "file2.txt")

# Associative arrays (bash 4+)
declare -A config
config[host]="localhost"
config[port]="8080"

# Integer variables
declare -i count=0
count+=1  # Arithmetic operation

# Export for environment
export DATABASE_URL="postgres://localhost/db"
declare -x DATABASE_URL="postgres://localhost/db"
```

### Variable Expansion

```bash
# Default values
value="${var:-default}"        # Use default if var is unset or empty
value="${var-default}"         # Use default only if var is unset
value="${var:=default}"        # Assign default if var is unset or empty
value="${var+alternative}"     # Use alternative if var is set

# String length
length="${#string}"

# Substring
substring="${string:0:5}"      # First 5 characters
substring="${string:5}"        # From 5th character to end

# Pattern matching (prefix removal)
filename="/path/to/file.txt"
basename="${filename##*/}"     # file.txt (remove longest match of */)
dirname="${filename%/*}"       # /path/to (remove shortest match of /*)

# Pattern matching (suffix removal)
file="document.tar.gz"
name="${file%.gz}"             # document.tar (remove shortest .gz)
name="${file%%.*}"             # document (remove longest .*)

# Search and replace
string="hello world"
new_string="${string/world/universe}"     # First occurrence
new_string="${string//o/0}"               # All occurrences
new_string="${string/#hello/hi}"          # Prefix match
new_string="${string/%world/earth}"       # Suffix match

# Case modification (bash 4+)
upper="${string^^}"            # TO UPPERCASE
lower="${string,,}"            # to lowercase
capitalize="${string^}"        # Capitalize first letter
```

### Command Substitution

```bash
# Modern syntax: $()
result=$(command)              # ✓ CORRECT (preferred)
result=`command`               # ✓ CORRECT (old style, avoid)

# Nested command substitution
outer=$(echo "$(echo inner)")  # ✓ CORRECT (easy to nest)
outer=`echo \`echo inner\``    # ✗ WRONG (hard to nest, requires escaping)

# Process substitution
diff <(command1) <(command2)   # Compare outputs
while read -r line; do
    echo "$line"
done < <(command)              # Read command output
```

---

## Error Handling

### Exit Codes

```bash
# Standard exit codes
readonly EXIT_SUCCESS=0
readonly EXIT_ERROR=1
readonly EXIT_INVALID_ARGS=2
readonly EXIT_MISSING_DEPENDENCY=3

# Use meaningful exit codes
validate_args() {
    if [[ $# -lt 1 ]]; then
        echo "Error: Missing required argument" >&2
        exit "$EXIT_INVALID_ARGS"
    fi
}

# Check command success
if ! command_that_might_fail; then
    echo "Error: Command failed" >&2
    exit "$EXIT_ERROR"
fi

# Alternative syntax
command_that_might_fail || {
    echo "Error: Command failed" >&2
    exit "$EXIT_ERROR"
}
```

### Error Messages

```bash
# ALWAYS write errors to stderr
echo "Error: Something went wrong" >&2

# Use consistent error message format
error() {
    local message="$1"
    local code="${2:-$EXIT_ERROR}"

    echo "ERROR: $message" >&2
    return "$code"
}

# Usage
if ! validate_input "$input"; then
    error "Invalid input: $input" "$EXIT_INVALID_ARGS"
fi
```

### Trap Handlers

```bash
# Cleanup on exit
cleanup() {
    local exit_code=$?

    # Cleanup operations
    [[ -n "${TEMP_DIR:-}" ]] && rm -rf "$TEMP_DIR"
    [[ -n "${LOCKFILE:-}" ]] && rm -f "$LOCKFILE"

    # Don't mask errors
    exit "$exit_code"
}

trap cleanup EXIT

# Handle specific signals
handle_sigterm() {
    echo "Received SIGTERM, shutting down..." >&2
    # Graceful shutdown logic
    exit 143  # 128 + 15 (SIGTERM)
}

trap handle_sigterm TERM

# ERR trap (bash 4.1+)
error_handler() {
    local line="$1"
    echo "Error on line $line" >&2
}

trap 'error_handler ${LINENO}' ERR
```

### Defensive Programming

```bash
# Validate all inputs
process_file() {
    local file="$1"

    # Check file exists
    if [[ ! -f "$file" ]]; then
        echo "Error: File not found: $file" >&2
        return 1
    fi

    # Check file is readable
    if [[ ! -r "$file" ]]; then
        echo "Error: File not readable: $file" >&2
        return 1
    fi

    # Process file
}

# Check dependencies before use
check_dependencies() {
    local deps=(curl jq awk sed)
    local missing=()

    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            missing+=("$dep")
        fi
    done

    if [[ ${#missing[@]} -gt 0 ]]; then
        echo "Error: Missing dependencies: ${missing[*]}" >&2
        exit "$EXIT_MISSING_DEPENDENCY"
    fi
}

# Validate environment
if [[ -z "${REQUIRED_VAR:-}" ]]; then
    echo "Error: REQUIRED_VAR must be set" >&2
    exit 1
fi
```

---

## Input/Output

### Reading User Input

```bash
# Simple read
read -rp "Enter your name: " name
echo "Hello, $name"

# Read with timeout
if read -rt 10 -p "Enter value (10s timeout): " value; then
    echo "You entered: $value"
else
    echo "Timeout or error"
fi

# Read password (no echo)
read -rsp "Enter password: " password
echo  # New line after password input

# Read confirmation
confirm() {
    local prompt="${1:-Are you sure?}"
    local response

    read -rp "$prompt [y/N] " response
    case "$response" in
        [yY][eE][sS]|[yY])
            return 0
            ;;
        *)
            return 1
            ;;
    esac
}

# Usage
if confirm "Delete all files?"; then
    rm -rf *
fi
```

### Reading Files

```bash
# Read file line by line
while IFS= read -r line; do
    echo "Line: $line"
done < file.txt

# Skip empty lines and comments
while IFS= read -r line || [[ -n "$line" ]]; do
    # Skip empty lines
    [[ -z "$line" ]] && continue

    # Skip comments
    [[ "$line" =~ ^[[:space:]]*# ]] && continue

    echo "Processing: $line"
done < file.txt

# Read into array
mapfile -t lines < file.txt
# Or
readarray -t lines < file.txt

# Read with null delimiter (for filenames with spaces)
while IFS= read -r -d '' file; do
    echo "File: $file"
done < <(find . -type f -print0)
```

### Writing Output

```bash
# Stdout vs stderr
echo "Normal output"              # stdout
echo "Error message" >&2          # stderr

# Redirect output
command > output.txt              # Overwrite
command >> output.txt             # Append
command 2> errors.txt             # Stderr only
command &> all_output.txt         # Both stdout and stderr
command > output.txt 2>&1         # Both (POSIX way)

# Here documents
cat <<EOF > file.txt
Line 1
Line 2
Variables are expanded: $VAR
EOF

# Here documents (no expansion)
cat <<'EOF' > file.txt
Line 1
Line 2
Variables are NOT expanded: $VAR
EOF

# Here strings
grep "pattern" <<< "$variable"
```

---

## Security

### Input Validation

```bash
# Validate input format
validate_email() {
    local email="$1"
    local regex="^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$"

    if [[ "$email" =~ $regex ]]; then
        return 0
    else
        return 1
    fi
}

# Sanitize file paths
sanitize_path() {
    local path="$1"

    # Remove directory traversal attempts
    path="${path//..\/}"

    # Remove leading slashes (if restricting to relative paths)
    path="${path#/}"

    echo "$path"
}

# Whitelist validation (preferred over blacklist)
validate_action() {
    local action="$1"
    local valid_actions=("start" "stop" "restart" "status")

    for valid in "${valid_actions[@]}"; do
        if [[ "$action" == "$valid" ]]; then
            return 0
        fi
    done

    return 1
}
```

### Command Injection Prevention

```bash
# NEVER use eval with user input
# ✗ DANGEROUS
eval "$user_input"

# NEVER concatenate user input into commands
# ✗ DANGEROUS
grep "$user_pattern" file.txt  # If pattern contains flags, command injection!

# ✓ SAFE - Use -- to separate options from arguments
grep -- "$user_pattern" file.txt

# ✓ SAFE - Use arrays for complex commands
command_args=(
    --option1 "$user_value1"
    --option2 "$user_value2"
)
command "${command_args[@]}"

# ✓ SAFE - Use printf %q for shell escaping
safe_value=$(printf %q "$user_input")
eval "command $safe_value"  # Now safe, but avoid if possible
```

### Temporary Files

```bash
# Use mktemp for secure temporary files
TEMP_FILE=$(mktemp) || {
    echo "Error: Cannot create temp file" >&2
    exit 1
}

# Cleanup on exit
trap 'rm -f "$TEMP_FILE"' EXIT

# Secure temp file (mode 600)
SECURE_TEMP=$(mktemp)
chmod 600 "$SECURE_TEMP"

# Temporary directory
TEMP_DIR=$(mktemp -d) || {
    echo "Error: Cannot create temp directory" >&2
    exit 1
}

trap 'rm -rf "$TEMP_DIR"' EXIT
```

### Secrets Management

```bash
# Don't hardcode secrets
# ✗ WRONG
PASSWORD="secret123"

# ✓ CORRECT - Read from environment
PASSWORD="${DATABASE_PASSWORD:-}"
if [[ -z "$PASSWORD" ]]; then
    echo "Error: DATABASE_PASSWORD must be set" >&2
    exit 1
fi

# ✓ CORRECT - Read from file
if [[ -f "$HOME/.config/app/password" ]]; then
    PASSWORD=$(cat "$HOME/.config/app/password")
fi

# ✓ CORRECT - Prompt user
read -rsp "Enter password: " PASSWORD
echo

# Don't log secrets
# ✗ WRONG
echo "Connecting with password: $PASSWORD"

# ✓ CORRECT
echo "Connecting to database..."

# Mask secrets in process list
# ✗ WRONG - Password visible in ps
mysql -pSecret123

# ✓ CORRECT - Use config file or environment variable
export MYSQL_PWD="$PASSWORD"
mysql

# Clear secrets from environment when done
unset PASSWORD
```

---

## Performance

### Avoid Unnecessary Subshells

```bash
# ✗ SLOW - Creates subshell
value=$(expr $a + $b)

# ✓ FAST - Bash arithmetic
value=$((a + b))

# ✗ SLOW - External command
value=$(echo "$string" | wc -c)

# ✓ FAST - Parameter expansion
value=${#string}
```

### Use Bash Built-ins

```bash
# ✗ SLOW - External commands
basename=$(basename "$path")
dirname=$(dirname "$path")

# ✓ FAST - Parameter expansion
basename="${path##*/}"
dirname="${path%/*}"

# ✗ SLOW - grep
if echo "$string" | grep -q "pattern"; then

# ✓ FAST - Bash regex
if [[ "$string" =~ pattern ]]; then

# ✗ SLOW - awk/cut
field=$(echo "$line" | awk '{print $3}')

# ✓ FAST - Read into array
read -ra fields <<< "$line"
field="${fields[2]}"
```

### Efficient Loops

```bash
# ✗ SLOW - Running external command in loop
for i in {1..1000}; do
    result=$(date +%s)
done

# ✓ FAST - Call once
timestamp=$(date +%s)
for i in {1..1000}; do
    result=$timestamp
done

# ✗ SLOW - Multiple passes
cat file | grep pattern | sort | uniq

# ✓ FAST - Single pass where possible
grep pattern file | sort -u
```

---

## Documentation

### Script Header

```bash
#!/usr/bin/env bash
#
# backup.sh - Automated backup script
#
# Description:
#   Creates incremental backups of specified directories
#   to a remote server using rsync.
#
# Usage:
#   backup.sh [OPTIONS] <source> <destination>
#
# Options:
#   -h, --help              Show this help message
#   -v, --verbose           Enable verbose output
#   -n, --dry-run          Show what would be done
#   -c, --config FILE      Use alternative config file
#
# Arguments:
#   source                  Directory to backup
#   destination            Remote destination (user@host:/path)
#
# Examples:
#   backup.sh /home/user user@backup:/backups/
#   backup.sh -v -c custom.conf /data remote:/store/
#
# Dependencies:
#   - rsync >= 3.0
#   - ssh
#
# Environment Variables:
#   BACKUP_CONFIG          Path to configuration file
#   BACKUP_VERBOSE         Enable verbose mode if set
#
# Exit Codes:
#   0   Success
#   1   General error
#   2   Invalid arguments
#   3   Missing dependency
#   4   Backup failed
#
# Author: Your Name <email@example.com>
# Version: 1.2.0
# Date: 2024-01-01
# License: MIT
#
```

### Inline Documentation

```bash
# Document complex logic
# This algorithm uses binary search to find the optimal value
# Time complexity: O(log n)
# Space complexity: O(1)

# Explain workarounds
# HACK: Sleep needed because API has rate limiting without proper headers
sleep 1

# Document assumptions
# Assumes file is in CSV format with header row

# Link to external resources
# See: https://docs.example.com/api for API documentation
```

### README and CHANGELOG

Every non-trivial script should have:

1. **README.md** - Installation, usage, examples
2. **CHANGELOG.md** - Version history
3. **LICENSE** - Licensing information

---

## Testing

### Unit Tests with BATS

```bash
# test/backup.bats
#!/usr/bin/env bats

# Setup runs before each test
setup() {
    # Create temp directory for tests
    TEST_DIR="$(mktemp -d)"
    export TEST_DIR
}

# Teardown runs after each test
teardown() {
    rm -rf "$TEST_DIR"
}

@test "backup creates archive" {
    run ./backup.sh "$TEST_DIR" backup.tar.gz
    [ "$status" -eq 0 ]
    [ -f backup.tar.gz ]
}

@test "backup fails with invalid source" {
    run ./backup.sh /nonexistent backup.tar.gz
    [ "$status" -eq 1 ]
    [ "${lines[0]}" = "Error: Source directory not found" ]
}

@test "backup validates dependencies" {
    # Mock missing dependency
    function tar() { return 127; }
    export -f tar

    run ./backup.sh "$TEST_DIR" backup.tar.gz
    [ "$status" -eq 3 ]
}
```

### Integration Tests

```bash
# integration_test.sh
#!/usr/bin/env bash
set -euo pipefail

# Test end-to-end workflow
test_full_workflow() {
    echo "Testing full workflow..."

    # Setup
    local test_dir="/tmp/test_$$"
    mkdir -p "$test_dir"

    # Execute
    ./script.sh create "$test_dir/output"
    ./script.sh process "$test_dir/output"
    ./script.sh verify "$test_dir/output"

    # Verify
    if [[ -f "$test_dir/output/result.txt" ]]; then
        echo "✓ Full workflow test passed"
        rm -rf "$test_dir"
        return 0
    else
        echo "✗ Full workflow test failed"
        rm -rf "$test_dir"
        return 1
    fi
}

# Run all tests
main() {
    local failed=0

    test_full_workflow || ((failed++))

    if [[ $failed -eq 0 ]]; then
        echo "All tests passed"
        exit 0
    else
        echo "$failed test(s) failed"
        exit 1
    fi
}

main
```

---

## Maintenance

### Version Control

```bash
# Include version in script
readonly VERSION="1.2.0"

show_version() {
    echo "$SCRIPT_NAME version $VERSION"
}

# Semantic versioning: MAJOR.MINOR.PATCH
# - MAJOR: Breaking changes
# - MINOR: New features (backward compatible)
# - PATCH: Bug fixes
```

### Deprecation

```bash
# Deprecation warning
deprecated_function() {
    echo "Warning: deprecated_function is deprecated, use new_function instead" >&2
    new_function "$@"
}

# Version-based deprecation
if [[ "${SCRIPT_VERSION%%.*}" -ge 2 ]]; then
    # Remove deprecated feature in version 2.0
    unset deprecated_function
fi
```

### Backward Compatibility

```bash
# Support old parameter names
if [[ -n "${OLD_PARAM:-}" && -z "${NEW_PARAM:-}" ]]; then
    echo "Warning: OLD_PARAM is deprecated, use NEW_PARAM" >&2
    NEW_PARAM="$OLD_PARAM"
fi

# Support multiple config file locations
for config in "$XDG_CONFIG_HOME/app/config" "$HOME/.config/app/config" "$HOME/.apprc"; do
    if [[ -f "$config" ]]; then
        CONFIG_FILE="$config"
        break
    fi
done
```

---

## Summary Checklist

Before considering a bash script production-ready:

- [ ] Passes ShellCheck with no warnings
- [ ] Uses `set -euo pipefail`
- [ ] All variables quoted properly
- [ ] Functions use local variables
- [ ] Has usage/help message
- [ ] Validates all inputs
- [ ] Checks dependencies
- [ ] Proper error messages (to stderr)
- [ ] Uses meaningful exit codes
- [ ] Includes cleanup trap
- [ ] Has inline documentation
- [ ] Follows consistent style
- [ ] Has unit tests (BATS)
- [ ] Has integration tests
- [ ] Tested on target platforms
- [ ] Has README documentation
- [ ] Version controlled (git)
- [ ] Reviewed by peer

**Additional for production:**
- [ ] Has CI/CD pipeline
- [ ] Logging implemented
- [ ] Monitoring/alerting configured
- [ ] Security reviewed
- [ ] Performance tested
- [ ] Disaster recovery plan
- [ ] Runbook/operational docs

This ensures professional, maintainable, and robust bash scripts.
