# Common Bash Patterns and Anti-Patterns

Collection of proven patterns and common mistakes in bash scripting with explanations and solutions.

---

## Table of Contents

1. [Variable Handling](#variable-handling)
2. [Command Execution](#command-execution)
3. [File Operations](#file-operations)
4. [String Processing](#string-processing)
5. [Arrays and Loops](#arrays-and-loops)
6. [Conditionals and Tests](#conditionals-and-tests)
7. [Functions](#functions)
8. [Error Handling](#error-handling)
9. [Process Management](#process-management)
10. [Security Patterns](#security-patterns)

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

## Variable Handling

### Pattern: Safe Variable Expansion

```bash
# ✓ GOOD: Always quote variables
echo "$variable"
cp "$source" "$destination"
rm -rf "$directory"

# ✗ BAD: Unquoted variables
echo $variable           # Word splitting and globbing
cp $source $destination  # Breaks with spaces
rm -rf $directory        # VERY DANGEROUS unquoted
```

**Why:** Unquoted variables undergo word splitting and pathname expansion, leading to unexpected behavior.

### Pattern: Default Values

```bash
# ✓ GOOD: Use parameter expansion for defaults
timeout="${TIMEOUT:-30}"
config="${CONFIG_FILE:-$HOME/.config/app.conf}"

# ✗ BAD: Manual check
if [ -z "$TIMEOUT" ]; then
    timeout=30
else
    timeout="$TIMEOUT"
fi
```

**Why:** Parameter expansion is concise, readable, and handles edge cases correctly.

### Anti-Pattern: Confusing Assignment and Comparison

```bash
# ✗ VERY BAD: Using = instead of ==
if [ "$var" = "value" ]; then  # Assignment in POSIX test!
    echo "Match"
fi

# ✓ GOOD: Use == or = correctly
if [[ "$var" == "value" ]]; then  # Comparison in bash
    echo "Match"
fi

# ✓ GOOD: POSIX-compliant
if [ "$var" = "value" ]; then  # Single = is correct in [ ]
    echo "Match"
fi
```

**Why:** In `[[ ]]`, both `=` and `==` work. In `[ ]`, only `=` is POSIX-compliant.

### Anti-Pattern: Unset Variable Access

```bash
# ✗ BAD: Accessing undefined variables
echo "Value: $undefined_variable"  # Silent error, prints "Value: "

# ✓ GOOD: Use set -u
set -u
echo "Value: $undefined_variable"  # Error: undefined_variable: unbound variable

# ✓ GOOD: Provide default
echo "Value: ${undefined_variable:-default}"
```

**Why:** `set -u` catches typos and logic errors early.

---

## Command Execution

### Pattern: Check Command Existence

```bash
# ✓ GOOD: Use command -v
if command -v jq &> /dev/null; then
    echo "jq is installed"
else
    echo "jq is not installed" >&2
    exit 1
fi

# ✗ BAD: Using which
if which jq; then  # Deprecated, not POSIX
    echo "jq is installed"
fi

# ✗ BAD: Using type
if type jq; then  # Verbose output
    echo "jq is installed"
fi
```

**Why:** `command -v` is POSIX-compliant, silent, and reliable.

### Pattern: Command Substitution

```bash
# ✓ GOOD: Modern syntax with $()
result=$(command arg1 arg2)
timestamp=$(date +%s)

# ✗ BAD: Backticks (hard to nest)
result=`command arg1 arg2`
timestamp=`date +%s`

# ✓ GOOD: Nested substitution
result=$(echo "Outer: $(echo "Inner")")

# ✗ BAD: Nested backticks (requires escaping)
result=`echo "Outer: \`echo \"Inner\"\`"`
```

**Why:** `$()` is easier to read, nest, and maintain.

### Anti-Pattern: Useless Use of Cat

```bash
# ✗ BAD: UUOC (Useless Use of Cat)
cat file.txt | grep "pattern"

# ✓ GOOD: Direct input
grep "pattern" file.txt

# ✗ BAD: Multiple cats
cat file1 | grep pattern | cat | sort | cat

# ✓ GOOD: Direct pipeline
grep pattern file1 | sort
```

**Why:** Unnecessary `cat` wastes resources and adds extra processes.

### Anti-Pattern: Using ls in Scripts

```bash
# ✗ BAD: Parsing ls output
for file in $(ls *.txt); do
    echo "$file"
done

# ✓ GOOD: Use globbing
for file in *.txt; do
    [[ -f "$file" ]] || continue  # Skip if no matches
    echo "$file"
done

# ✗ BAD: Counting files with ls
count=$(ls -1 | wc -l)

# ✓ GOOD: Use array
files=(*)
count=${#files[@]}
```

**Why:** `ls` output is meant for humans, not scripts. Parsing it breaks with spaces, newlines, etc.

---

## File Operations

### Pattern: Safe File Reading

```bash
# ✓ GOOD: Preserve leading/trailing whitespace and backslashes
while IFS= read -r line; do
    echo "Line: $line"
done < file.txt

# ✗ BAD: Without IFS= (strips leading/trailing whitespace)
while read -r line; do
    echo "Line: $line"
done < file.txt

# ✗ BAD: Without -r (interprets backslashes)
while IFS= read line; do
    echo "Line: $line"
done < file.txt
```

**Why:** `IFS=` prevents trimming, `-r` prevents backslash interpretation.

### Pattern: Null-Delimited Files

```bash
# ✓ GOOD: For filenames with special characters
find . -name "*.txt" -print0 | while IFS= read -r -d '' file; do
    echo "Processing: $file"
done

# Or with mapfile (bash 4+)
mapfile -d '' -t files < <(find . -name "*.txt" -print0)
for file in "${files[@]}"; do
    echo "Processing: $file"
done

# ✗ BAD: Newline-delimited (breaks with newlines in filenames)
find . -name "*.txt" | while IFS= read -r file; do
    echo "Processing: $file"
done
```

**Why:** Filenames can contain any character except null and slash.

### Anti-Pattern: Testing File Existence Incorrectly

```bash
# ✗ BAD: Using ls to test existence
if ls file.txt &> /dev/null; then
    echo "File exists"
fi

# ✓ GOOD: Use test operators
if [[ -f file.txt ]]; then
    echo "File exists"
fi

# ✓ GOOD: Different tests
[[ -e path ]]   # Exists (file or directory)
[[ -f file ]]   # Regular file
[[ -d dir ]]    # Directory
[[ -L link ]]   # Symbolic link
[[ -r file ]]   # Readable
[[ -w file ]]   # Writable
[[ -x file ]]   # Executable
```

**Why:** Test operators are the correct, efficient way to check file properties.

### Pattern: Temporary Files

```bash
# ✓ GOOD: Secure temporary file
temp_file=$(mktemp)
trap 'rm -f "$temp_file"' EXIT

# Use temp file
echo "data" > "$temp_file"

# ✗ BAD: Insecure temp file
temp_file="/tmp/myapp.$$"
echo "data" > "$temp_file"
# No cleanup!

# ✓ GOOD: Temporary directory
temp_dir=$(mktemp -d)
trap 'rm -rf "$temp_dir"' EXIT
```

**Why:** `mktemp` creates secure, unique files and prevents race conditions.

---

## String Processing

### Pattern: String Manipulation with Parameter Expansion

```bash
# ✓ GOOD: Use bash parameter expansion
filename="document.tar.gz"
basename="${filename%%.*}"      # document
extension="${filename##*.}"     # gz
name="${filename%.gz}"          # document.tar

# ✗ BAD: Using external commands
basename=$(echo "$filename" | sed 's/\..*$//')
extension=$(echo "$filename" | awk -F. '{print $NF}')
```

**Why:** Parameter expansion is faster and doesn't spawn processes.

### Pattern: String Comparison

```bash
# ✓ GOOD: Use [[ ]] for strings
if [[ "$string1" == "$string2" ]]; then
    echo "Equal"
fi

# ✓ GOOD: Pattern matching
if [[ "$filename" == *.txt ]]; then
    echo "Text file"
fi

# ✓ GOOD: Regex matching
if [[ "$email" =~ ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
    echo "Valid email"
fi

# ✗ BAD: Using grep for simple string check
if echo "$string" | grep -q "substring"; then
    echo "Found"
fi

# ✓ GOOD: Use substring matching
if [[ "$string" == *"substring"* ]]; then
    echo "Found"
fi
```

**Why:** `[[ ]]` is bash-native, faster, and more readable.

### Anti-Pattern: Word Splitting Issues

```bash
# ✗ BAD: Unquoted expansion with spaces
var="file1.txt file2.txt"
for file in $var; do  # Splits on spaces!
    echo "$file"      # file1.txt, then file2.txt
done

# ✓ GOOD: Use array
files=("file1.txt" "file2.txt")
for file in "${files[@]}"; do
    echo "$file"
done

# ✗ BAD: Word splitting in command arguments
file="my file.txt"
rm $file  # Tries to remove "my" and "file.txt"!

# ✓ GOOD: Quote variables
rm "$file"
```

**Why:** Word splitting on spaces is a major source of bugs.

---

## Arrays and Loops

### Pattern: Array Declaration and Use

```bash
# ✓ GOOD: Array declaration
files=("file1.txt" "file2.txt" "file 3.txt")

# ✓ GOOD: Array expansion (each element quoted)
for file in "${files[@]}"; do
    echo "$file"
done

# ✗ BAD: Unquoted array expansion
for file in ${files[@]}; do  # Word splitting!
    echo "$file"
done

# ✓ GOOD: Add to array
files+=("file4.txt")

# ✓ GOOD: Array length
echo "Count: ${#files[@]}"

# ✓ GOOD: Array indices
for i in "${!files[@]}"; do
    echo "File $i: ${files[$i]}"
done
```

**Why:** Proper array handling prevents word splitting and globbing issues.

### Pattern: Reading Command Output into Array

```bash
# ✓ GOOD: mapfile/readarray (bash 4+)
mapfile -t lines < file.txt

# ✓ GOOD: With command substitution
mapfile -t files < <(find . -name "*.txt")

# ✗ BAD: Word splitting
files=($(find . -name "*.txt"))  # Breaks with spaces in filenames!

# ✓ GOOD: Alternative (POSIX-compatible)
while IFS= read -r file; do
    files+=("$file")
done < <(find . -name "*.txt")
```

**Why:** `mapfile` is efficient and handles special characters correctly.

### Anti-Pattern: C-Style For Loops for Arrays

```bash
# ✗ BAD: C-style loop for arrays
for ((i=0; i<${#files[@]}; i++)); do
    echo "${files[$i]}"
done

# ✓ GOOD: For-in loop
for file in "${files[@]}"; do
    echo "$file"
done

# ✓ ACCEPTABLE: When you need the index
for i in "${!files[@]}"; do
    echo "Index $i: ${files[$i]}"
done
```

**Why:** For-in loops are simpler and less error-prone.

### Pattern: Loop over Range

```bash
# ✓ GOOD: Brace expansion
for i in {1..10}; do
    echo "$i"
done

# ✓ GOOD: With variables (bash 4+)
start=1
end=10
for i in $(seq $start $end); do
    echo "$i"
done

# ✓ GOOD: C-style (arithmetic)
for ((i=1; i<=10; i++)); do
    echo "$i"
done

# ✗ BAD: Using seq in a loop unnecessarily
for i in $(seq 1 1000000); do  # Creates huge string in memory!
    echo "$i"
done

# ✓ GOOD: Use C-style for large ranges
for ((i=1; i<=1000000; i++)); do
    echo "$i"
done
```

**Why:** Choose the right loop construct based on the use case.

---

## Conditionals and Tests

### Pattern: File Tests

```bash
# ✓ GOOD: Use appropriate test
if [[ -f "$file" ]]; then         # Regular file
if [[ -d "$dir" ]]; then          # Directory
if [[ -e "$path" ]]; then         # Exists (any type)
if [[ -L "$link" ]]; then         # Symbolic link
if [[ -r "$file" ]]; then         # Readable
if [[ -w "$file" ]]; then         # Writable
if [[ -x "$file" ]]; then         # Executable
if [[ -s "$file" ]]; then         # Non-empty file

# ✗ BAD: Incorrect test
if [[ -e "$file" ]]; then         # Exists, but could be directory!
    cat "$file"                   # Fails if directory
fi

# ✓ GOOD: Specific test
if [[ -f "$file" ]]; then
    cat "$file"
fi
```

**Why:** Use the most specific test for your use case.

### Pattern: Numeric Comparison

```bash
# ✓ GOOD: Arithmetic context
if (( num > 10 )); then
    echo "Greater than 10"
fi

# ✓ GOOD: Test operator
if [[ $num -gt 10 ]]; then
    echo "Greater than 10"
fi

# ✗ BAD: String comparison for numbers
if [[ "$num" > "10" ]]; then  # Lexicographic comparison!
    echo "Greater than 10"    # "9" > "10" is true!
fi
```

**Why:** Use numeric comparison operators for numbers.

### Anti-Pattern: Testing Boolean Strings

```bash
# ✗ BAD: Comparing to string "true"
if [[ "$flag" == "true" ]]; then
    do_something
fi

# ✓ GOOD: Use boolean variable directly
flag=false  # or true

if $flag; then
    do_something
fi

# ✓ BETTER: Use integers for flags
flag=0  # false
flag=1  # true

if (( flag )); then
    do_something
fi

# ✓ GOOD: For command success/failure
if command; then
    echo "Success"
fi
```

**Why:** Boolean strings are error-prone; use actual booleans or return codes.

### Pattern: Multiple Conditions

```bash
# ✓ GOOD: Logical operators
if [[ condition1 && condition2 ]]; then
    echo "Both true"
fi

if [[ condition1 || condition2 ]]; then
    echo "At least one true"
fi

if [[ ! condition ]]; then
    echo "False"
fi

# ✗ BAD: Separate tests
if [ condition1 -a condition2 ]; then  # Deprecated
    echo "Both true"
fi

# ✗ BAD: Nested ifs for AND
if [[ condition1 ]]; then
    if [[ condition2 ]]; then
        echo "Both true"
    fi
fi
```

**Why:** `&&` and `||` in `[[ ]]` are clearer and recommended.

---

## Functions

### Pattern: Function Return Values

```bash
# ✓ GOOD: Return status, output to stdout
get_value() {
    local value="result"

    if [[ -n "$value" ]]; then
        echo "$value"
        return 0
    else
        return 1
    fi
}

# Usage
if result=$(get_value); then
    echo "Got: $result"
else
    echo "Failed"
fi

# ✗ BAD: Using return for data
get_value() {
    return 42  # Can only return 0-255!
}
result=$?  # Gets 42, but limited range
```

**Why:** `return` is for exit status (0-255), not data. Output to stdout for data.

### Pattern: Local Variables in Functions

```bash
# ✓ GOOD: Declare local variables
my_function() {
    local arg="$1"
    local result=""

    result=$(process "$arg")
    echo "$result"
}

# ✗ BAD: Global variables
my_function() {
    arg="$1"        # Pollutes global namespace!
    result=""       # Global variable!

    result=$(process "$arg")
    echo "$result"
}
```

**Why:** Local variables prevent unexpected side effects.

### Anti-Pattern: Capturing Local Command Failure

```bash
# ✗ BAD: Local declaration masks command failure
my_function() {
    local result=$(command_that_fails)  # $? is from 'local', not 'command'!
    echo "$result"
}

# ✓ GOOD: Separate declaration and assignment
my_function() {
    local result
    result=$(command_that_fails) || return 1
    echo "$result"
}

# ✓ GOOD: Check command separately
my_function() {
    local result

    if ! result=$(command_that_fails); then
        return 1
    fi

    echo "$result"
}
```

**Why:** Combining `local` and command substitution hides command failure.

---

## Error Handling

### Pattern: Check Command Success

```bash
# ✓ GOOD: Direct check
if ! command; then
    echo "Command failed" >&2
    exit 1
fi

# ✓ GOOD: With logical operator
command || {
    echo "Command failed" >&2
    exit 1
}

# ✓ GOOD: Capture output and check
if ! output=$(command 2>&1); then
    echo "Command failed: $output" >&2
    exit 1
fi

# ✗ BAD: Not checking status
command  # What if it fails?
next_command
```

**Why:** Always check if commands succeed unless failure is acceptable.

### Pattern: Error Messages to stderr

```bash
# ✓ GOOD: Errors to stderr
echo "Error: Invalid argument" >&2

# ✗ BAD: Errors to stdout
echo "Error: Invalid argument"

# ✓ GOOD: Error function
error() {
    echo "ERROR: $*" >&2
}

error "Something went wrong"
```

**Why:** stderr is for errors, stdout is for data output.

### Pattern: Cleanup on Exit

```bash
# ✓ GOOD: Trap for cleanup
temp_file=$(mktemp)

cleanup() {
    rm -f "$temp_file"
}

trap cleanup EXIT

# Do work with temp_file

# ✗ BAD: Manual cleanup (might not run)
temp_file=$(mktemp)

# Do work

rm -f "$temp_file"  # Doesn't run if script exits early!
```

**Why:** Trap ensures cleanup runs on exit, even on errors.

### Anti-Pattern: Silencing Errors

```bash
# ✗ BAD: Silencing errors
command 2>/dev/null  # What if it fails?
next_command

# ✓ GOOD: Check status even if silencing output
if ! command 2>/dev/null; then
    echo "Command failed" >&2
    exit 1
fi

# ✓ ACCEPTABLE: When failure is expected and acceptable
if command 2>/dev/null; then
    echo "Command succeeded"
else
    echo "Command failed (expected)"
fi
```

**Why:** Silencing errors without checking status leads to silent failures.

---

## Process Management

### Pattern: Background Jobs

```bash
# ✓ GOOD: Track background jobs
long_running_task &
pid=$!

# Wait for completion
if wait "$pid"; then
    echo "Task completed successfully"
else
    echo "Task failed" >&2
fi

# ✓ GOOD: Multiple background jobs
job1 &
pid1=$!
job2 &
pid2=$!

wait "$pid1" "$pid2"
```

**Why:** Proper job management prevents zombie processes.

### Pattern: Timeout for Commands

```bash
# ✓ GOOD: Use timeout command (if available)
if timeout 30 long_running_command; then
    echo "Completed within timeout"
else
    echo "Timed out or failed" >&2
fi

# ✓ GOOD: Manual timeout implementation
timeout_command() {
    local timeout=$1
    shift

    "$@" &
    local pid=$!

    ( sleep "$timeout"; kill "$pid" 2>/dev/null ) &
    local killer=$!

    if wait "$pid" 2>/dev/null; then
        kill "$killer" 2>/dev/null
        wait "$killer" 2>/dev/null
        return 0
    else
        return 1
    fi
}

timeout_command 30 long_running_command
```

**Why:** Prevents scripts from hanging indefinitely.

### Anti-Pattern: Killing Processes Unsafely

```bash
# ✗ BAD: kill -9 immediately
kill -9 "$pid"

# ✓ GOOD: Graceful shutdown first
kill -TERM "$pid"
sleep 2

if kill -0 "$pid" 2>/dev/null; then
    echo "Process still running, forcing..." >&2
    kill -KILL "$pid"
fi

# ✓ GOOD: With timeout
graceful_kill() {
    local pid=$1
    local timeout=${2:-10}

    kill -TERM "$pid" 2>/dev/null || return 0

    for ((i=0; i<timeout; i++)); do
        if ! kill -0 "$pid" 2>/dev/null; then
            return 0
        fi
        sleep 1
    done

    echo "Forcing kill of $pid" >&2
    kill -KILL "$pid" 2>/dev/null
}
```

**Why:** SIGTERM allows graceful shutdown; SIGKILL should be last resort.

---

## Security Patterns

### Pattern: Input Validation

```bash
# ✓ GOOD: Whitelist validation
validate_action() {
    local action=$1
    case "$action" in
        start|stop|restart|status)
            return 0
            ;;
        *)
            echo "Error: Invalid action: $action" >&2
            return 1
            ;;
    esac
}

# ✗ BAD: No validation
action="$1"
systemctl "$action" myservice  # User can pass arbitrary commands!

# ✓ GOOD: Validate first
if validate_action "$1"; then
    systemctl "$1" myservice
else
    exit 1
fi
```

**Why:** Whitelist validation prevents command injection.

### Pattern: Avoid eval

```bash
# ✗ BAD: eval with user input
eval "$user_command"  # DANGEROUS!

# ✓ GOOD: Use arrays
command_args=("$arg1" "$arg2" "$arg3")
command "${command_args[@]}"

# ✗ BAD: Dynamic variable names
eval "var_$name=value"

# ✓ GOOD: Associative arrays (bash 4+)
declare -A vars
vars[$name]="value"
```

**Why:** `eval` with user input is a security vulnerability.

### Pattern: Safe PATH

```bash
# ✓ GOOD: Set explicit PATH
export PATH="/usr/local/bin:/usr/bin:/bin"

# ✓ GOOD: Use absolute paths for critical commands
/usr/bin/rm -rf "$directory"

# ✗ BAD: Trusting user's PATH
rm -rf "$directory"  # What if there's a malicious 'rm' in PATH?
```

**Why:** Prevents PATH injection attacks.

---

## Summary

**Most Critical Patterns:**

1. Always quote variable expansions: `"$var"`
2. Use `set -euo pipefail` for safety
3. Prefer `[[ ]]` over `[ ]` in bash
4. Use arrays for lists: `"${array[@]}"`
5. Check command success: `if ! command; then`
6. Use local variables in functions
7. Errors to stderr: `echo "Error" >&2`
8. Use `mktemp` for temporary files
9. Cleanup with traps: `trap cleanup EXIT`
10. Validate all user input

**Most Dangerous Anti-Patterns:**

1. Unquoted variables: `$var`
2. Parsing `ls` output
3. Using `eval` with user input
4. Silencing errors without checking
5. Not using `set -u` or defaults
6. Global variables in functions
7. Word splitting on filenames
8. Testing strings with `>` for numbers
9. `kill -9` without trying graceful shutdown
10. Trusting user PATH

Following these patterns and avoiding anti-patterns will result in robust, secure, and maintainable bash scripts.
