# Bash Scripting Resources

Comprehensive directory of authoritative sources, tools, and learning resources for bash scripting.

---

## Table of Contents

1. [Official Documentation](#official-documentation)
2. [Style Guides and Standards](#style-guides-and-standards)
3. [Tools and Utilities](#tools-and-utilities)
4. [Learning Resources](#learning-resources)
5. [Community Resources](#community-resources)
6. [Books](#books)
7. [Cheat Sheets and Quick References](#cheat-sheets-and-quick-references)
8. [Testing and Quality](#testing-and-quality)
9. [Platform-Specific Resources](#platform-specific-resources)
10. [Advanced Topics](#advanced-topics)

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

## Official Documentation

### Bash Manual

**GNU Bash Reference Manual**
- **URL:** https://www.gnu.org/software/bash/manual/
- **Description:** The authoritative reference for bash features, syntax, and built-ins
- **Use for:** Detailed feature documentation, syntax clarification, version-specific features

**Bash Man Page**
```bash
man bash        # Complete bash documentation
man bash-builtins  # Built-in commands
```
- **Use for:** Quick reference on local system, offline documentation

### POSIX Standards

**POSIX Shell Command Language**
- **URL:** https://pubs.opengroup.org/onlinepubs/9699919799/utilities/V3_chap02.html
- **Description:** IEEE/Open Group specification for portable shell scripting
- **Use for:** Writing portable scripts, understanding sh vs bash differences

**POSIX Utilities**
- **URL:** https://pubs.opengroup.org/onlinepubs/9699919799/idx/utilities.html
- **Description:** Standard utilities available in POSIX-compliant systems
- **Use for:** Portable command usage, cross-platform compatibility

### Command Documentation

**GNU Coreutils Manual**
- **URL:** https://www.gnu.org/software/coreutils/manual/
- **Description:** Documentation for core GNU utilities (ls, cat, grep, etc.)
- **Use for:** Understanding Linux command behavior, GNU-specific features

**Man Pages Online**
- **URL:** https://man7.org/linux/man-pages/
- **URL:** https://www.freebsd.org/cgi/man.cgi (BSD/macOS)
- **Description:** Online searchable man pages
- **Use for:** Quick online reference, comparing Linux vs BSD commands

---

## Style Guides and Standards

### Google Shell Style Guide

**URL:** https://google.github.io/styleguide/shellguide.html

**Key Points:**
- Industry-standard practices from Google
- Covers naming conventions, formatting, best practices
- When to use shell vs other languages
- Safety and portability guidelines

**Use for:** Professional code style, team standards, code reviews

### Defensive Bash Programming

**URL:** https://kfirlavi.herokuapp.com/blog/2012/11/14/defensive-bash-programming

**Key Points:**
- Writing robust bash scripts
- Error handling patterns
- Safe coding practices
- Code organization

**Use for:** Improving script reliability, avoiding common pitfalls

### Shell Style Guide (GitHub)

**URL:** https://github.com/bahamas10/bash-style-guide

**Key Points:**
- Community-driven style guidelines
- Practical examples
- Modern bash features

**Use for:** Alternative perspectives on style, community standards

---

## Tools and Utilities

### ShellCheck

**Website:** https://www.shellcheck.net/
**GitHub:** https://github.com/koalaman/shellcheck
**Online Tool:** https://www.shellcheck.net/ (paste code for instant feedback)

**Description:** Static analysis tool for shell scripts

**Installation:**
```bash
# Ubuntu/Debian
apt-get install shellcheck

# macOS
brew install shellcheck

# Windows (Scoop)
scoop install shellcheck

# Via Docker
docker run --rm -v "$PWD:/mnt" koalaman/shellcheck script.sh
```

**Usage:**
```bash
shellcheck script.sh                 # Check script
shellcheck -x script.sh              # Follow source statements
shellcheck -f json script.sh         # JSON output
shellcheck -e SC2086 script.sh       # Exclude specific warnings
```

**ShellCheck Wiki:** https://www.shellcheck.net/wiki/
- Detailed explanations of every warning
- **Use for:** Understanding and fixing ShellCheck warnings

### shfmt

**GitHub:** https://github.com/mvdan/sh

**Description:** Shell script formatter

**Installation:**
```bash
# macOS
brew install shfmt

# Go
go install mvdan.cc/sh/v3/cmd/shfmt@latest
```

**Usage:**
```bash
shfmt -i 4 -w script.sh             # Format with 4-space indent
shfmt -d script.sh                  # Show diff without modifying
shfmt -l script.sh                  # List files that would be changed
```

**Use for:** Consistent code formatting, automated formatting in CI

### BATS (Bash Automated Testing System)

**GitHub:** https://github.com/bats-core/bats-core

**Description:** Testing framework for bash scripts

**Installation:**
```bash
git clone https://github.com/bats-core/bats-core.git
cd bats-core
./install.sh /usr/local
```

**Usage:**
```bash
bats test/                          # Run all tests
bats test/script.bats               # Run specific test file
bats --tap test/                    # TAP output format
```

**Documentation:** https://bats-core.readthedocs.io/

**Use for:** Unit testing bash scripts, CI/CD integration

### bashate

**GitHub:** https://github.com/openstack/bashate

**Description:** Style checker (used by OpenStack)

**Installation:**
```bash
pip install bashate
```

**Usage:**
```bash
bashate script.sh
bashate -i E006 script.sh           # Ignore specific errors
```

**Use for:** Additional style checking beyond ShellCheck

### checkbashisms

**Package:** devscripts (Debian)

**Description:** Checks for bashisms in sh scripts

**Installation:**
```bash
apt-get install devscripts          # Ubuntu/Debian
```

**Usage:**
```bash
checkbashisms script.sh
checkbashisms -f script.sh          # Force check even if #!/bin/bash
```

**Use for:** Ensuring POSIX compliance, portable scripts

---

## Learning Resources

### Interactive Tutorials

**Bash Academy**
- **URL:** https://www.bash.academy/
- **Description:** Modern, comprehensive bash tutorial
- **Topics:** Basics, scripting, advanced features
- **Use for:** Learning bash from scratch, structured learning path

**Learn Shell**
- **URL:** https://www.learnshell.org/
- **Description:** Interactive bash tutorial with exercises
- **Use for:** Hands-on practice, beginners

**Bash Scripting Tutorial**
- **URL:** https://linuxconfig.org/bash-scripting-tutorial
- **Description:** Comprehensive tutorial series
- **Use for:** Step-by-step learning, examples

### Guides and Documentation

**Bash Guide for Beginners**
- **URL:** https://tldp.org/LDP/Bash-Beginners-Guide/html/
- **Author:** The Linux Documentation Project
- **Description:** Comprehensive guide covering basics to intermediate
- **Use for:** Structured learning, reference material

**Advanced Bash-Scripting Guide**
- **URL:** https://tldp.org/LDP/abs/html/
- **Description:** In-depth coverage of advanced bash topics
- **Topics:** Complex scripting, text processing, system administration
- **Use for:** Advanced techniques, real-world examples

**Bash Hackers Wiki**
- **URL:** https://wiki.bash-hackers.org/
- **Alternative:** https://flokoe.github.io/bash-hackers-wiki/ (maintained mirror)
- **Description:** Community-driven bash documentation
- **Use for:** In-depth explanations, advanced topics, edge cases

**Greg's Wiki (Wooledge)**
- **URL:** https://mywiki.wooledge.org/
- **Key Pages:**
  - https://mywiki.wooledge.org/BashFAQ
  - https://mywiki.wooledge.org/BashPitfalls
  - https://mywiki.wooledge.org/BashGuide
- **Description:** High-quality bash Q&A and guides
- **Use for:** Common questions, avoiding pitfalls, best practices

### Video Courses

**Bash Scripting on Linux (Udemy)**
- **Description:** Comprehensive video course
- **Use for:** Visual learners

**Shell Scripting: Discover How to Automate Command Line Tasks (Udemy)**
- **Description:** Practical shell scripting course
- **Use for:** Automation-focused learning

**LinkedIn Learning - Learning Bash Scripting**
- **Description:** Professional development course
- **Use for:** Structured corporate training

---

## Community Resources

### Stack Overflow

**Bash Tag**
- **URL:** https://stackoverflow.com/questions/tagged/bash
- **Use for:** Specific problems, code review, troubleshooting

**Top Questions:**
- **URL:** https://stackoverflow.com/questions/tagged/bash?tab=Votes
- **Use for:** Common problems and solutions

### Unix & Linux Stack Exchange

**URL:** https://unix.stackexchange.com/

**Shell Tag:** https://unix.stackexchange.com/questions/tagged/shell
**Bash Tag:** https://unix.stackexchange.com/questions/tagged/bash

**Use for:** Unix/Linux-specific questions, system administration

### Reddit

**/r/bash**
- **URL:** https://www.reddit.com/r/bash/
- **Description:** Bash scripting community
- **Use for:** Discussions, learning resources, help

**/r/commandline**
- **URL:** https://www.reddit.com/r/commandline/
- **Description:** Command-line interface community
- **Use for:** CLI tips, tools, productivity

### IRC/Chat

**Freenode #bash**
- **URL:** irc://irc.freenode.net/bash
- **Description:** Real-time bash help channel
- **Use for:** Live help, quick questions

**Libera.Chat #bash**
- **URL:** irc://irc.libera.chat/bash
- **Description:** Alternative IRC channel
- **Use for:** Live community support

---

## Books

### "Classic Shell Scripting" by Arnold Robbins & Nelson Beebe

**Publisher:** O'Reilly
**ISBN:** 978-0596005955

**Topics:**
- Shell basics and portability
- Text processing and filters
- Shell programming patterns

**Use for:** Comprehensive reference, professional development

### "Learning the bash Shell" by Cameron Newham

**Publisher:** O'Reilly
**ISBN:** 978-0596009656

**Topics:**
- Bash basics
- Command-line editing
- Shell programming

**Use for:** Systematic learning, reference

### "Bash Cookbook" by Carl Albing & JP Vossen

**Publisher:** O'Reilly
**ISBN:** 978-1491975336

**Topics:**
- Solutions to common problems
- Recipes and patterns
- Real-world examples

**Use for:** Problem-solving, practical examples

### "Wicked Cool Shell Scripts" by Dave Taylor & Brandon Perry

**Publisher:** No Starch Press
**ISBN:** 978-1593276027

**Topics:**
- Creative shell scripting
- System administration
- Fun and practical scripts

**Use for:** Inspiration, practical applications

### "The Linux Command Line" by William Shotts

**Publisher:** No Starch Press
**ISBN:** 978-1593279523
**Free PDF:** https://linuxcommand.org/tlcl.php

**Topics:**
- Command-line basics
- Shell scripting fundamentals
- Linux system administration

**Use for:** Beginners, comprehensive introduction

---

## Cheat Sheets and Quick References

### Bash Cheat Sheet (DevHints)

**URL:** https://devhints.io/bash

**Content:**
- Quick syntax reference
- Common patterns
- Parameter expansion
- Conditionals and loops

**Use for:** Quick lookups, syntax reminders

### Bash Scripting Cheat Sheet (GitHub)

**URL:** https://github.com/LeCoupa/awesome-cheatsheets/blob/master/languages/bash.sh

**Content:**
- Comprehensive syntax guide
- Examples and explanations
- Best practices

**Use for:** Single-file reference

### explainshell.com

**URL:** https://explainshell.com/

**Description:** Interactive tool that explains shell commands

**Example:** Paste `tar -xzvf file.tar.gz` to get detailed explanation of each flag

**Use for:** Understanding complex commands, learning command options

### Command Line Fu

**URL:** https://www.commandlinefu.com/

**Description:** Community-contributed command-line snippets

**Use for:** One-liners, clever solutions, learning new commands

### tldr Pages

**URL:** https://tldr.sh/
**GitHub:** https://github.com/tldr-pages/tldr

**Description:** Simplified man pages with examples

**Installation:**
```bash
npm install -g tldr
# Or
brew install tldr
```

**Usage:**
```bash
tldr tar
tldr grep
tldr find
```

**Use for:** Quick command examples, practical usage

---

## Testing and Quality

### Testing Frameworks

**BATS (Bash Automated Testing System)**
- **URL:** https://github.com/bats-core/bats-core
- **Documentation:** https://bats-core.readthedocs.io/
- **Use for:** Unit testing

**shUnit2**
- **URL:** https://github.com/kward/shunit2
- **Description:** xUnit-based unit testing framework
- **Use for:** Alternative to BATS

**Bash Unit**
- **URL:** https://github.com/pgrange/bash_unit
- **Description:** Bash unit testing
- **Use for:** Lightweight testing

### CI/CD Integration

**GitHub Actions Example**
```yaml
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Install ShellCheck
        run: sudo apt-get install -y shellcheck
      - name: Run ShellCheck
        run: find . -name "*.sh" -exec shellcheck {} +
      - name: Install BATS
        run: |
          git clone https://github.com/bats-core/bats-core.git
          cd bats-core
          sudo ./install.sh /usr/local
      - name: Run Tests
        run: bats test/
```

**GitLab CI Example**
```yaml
test:
  image: koalaman/shellcheck-alpine
  script:
    - find . -name "*.sh" -exec shellcheck {} +

bats:
  image: bats/bats
  script:
    - bats test/
```

### Code Coverage

**bashcov**
- **URL:** https://github.com/infertux/bashcov
- **Description:** Code coverage for bash
- **Installation:** `gem install bashcov`
- **Use for:** Measuring test coverage

---

## Platform-Specific Resources

### Linux

**Linux Man Pages**
- **URL:** https://man7.org/linux/man-pages/
- **Use for:** Linux-specific command documentation

**systemd Documentation**
- **URL:** https://www.freedesktop.org/software/systemd/man/
- **Use for:** systemd service management

### macOS

**macOS Man Pages**
- **URL:** https://www.freebsd.org/cgi/man.cgi
- **Description:** BSD-based commands (similar to macOS)
- **Use for:** macOS command differences

**Homebrew**
- **URL:** https://brew.sh/
- **Use for:** Installing GNU tools on macOS

### Windows

**Git for Windows**
- **URL:** https://gitforwindows.org/
- **Documentation:** https://github.com/git-for-windows/git/wiki
- **Use for:** Git Bash on Windows

**WSL Documentation**
- **URL:** https://docs.microsoft.com/en-us/windows/wsl/
- **Use for:** Windows Subsystem for Linux

**Cygwin**
- **URL:** https://www.cygwin.com/
- **Use for:** POSIX environment on Windows

### Containers

**Docker Bash Best Practices**
- **URL:** https://docs.docker.com/develop/develop-images/dockerfile_best-practices/
- **Use for:** Bash in containers

**Container Best Practices**
- **URL:** https://cloud.google.com/architecture/best-practices-for-building-containers
- **Use for:** Production container scripts

---

## Advanced Topics

### Process Substitution

**Greg's Wiki:**
- **URL:** https://mywiki.wooledge.org/ProcessSubstitution
- **Use for:** Understanding `<()` syntax

### Parameter Expansion

**Bash Hackers Wiki:**
- **URL:** https://wiki.bash-hackers.org/syntax/pe
- **Use for:** Complete parameter expansion reference

### Regular Expressions

**Bash Regex:**
- **URL:** https://mywiki.wooledge.org/RegularExpression
- **Use for:** Regex in bash `[[ =~ ]]`

**PCRE vs POSIX:**
- **URL:** https://www.regular-expressions.info/posix.html
- **Use for:** Understanding regex flavors

### Parallel Processing

**GNU Parallel:**
- **URL:** https://www.gnu.org/software/parallel/
- **Tutorial:** https://www.gnu.org/software/parallel/parallel_tutorial.html
- **Use for:** Parallel command execution

### Job Control

**Bash Job Control:**
- **URL:** https://www.gnu.org/software/bash/manual/html_node/Job-Control.html
- **Use for:** Background jobs, job management

---

## Troubleshooting Resources

### Debugging Tools

**bashdb**
- **URL:** http://bashdb.sourceforge.net/
- **Description:** Bash debugger
- **Use for:** Step-by-step debugging

**xtrace**
```bash
set -x  # Enable
set +x  # Disable
```
- **Use for:** Trace command execution

**PS4 for Better Trace Output**
```bash
export PS4='+(${BASH_SOURCE}:${LINENO}): ${FUNCNAME[0]:+${FUNCNAME[0]}(): }'
set -x
```

### Common Issues

**Bash Pitfalls**
- **URL:** https://mywiki.wooledge.org/BashPitfalls
- **Description:** 50+ common mistakes in bash
- **Use for:** Avoiding and fixing common errors

**Bash FAQ**
- **URL:** https://mywiki.wooledge.org/BashFAQ
- **Description:** Frequently asked questions
- **Use for:** Quick answers to common questions

---

## Summary: Where to Find Information

| Question Type | Resource |
|---------------|----------|
| Syntax reference | Bash Manual, DevHints cheat sheet |
| Best practices | Google Shell Style Guide, ShellCheck |
| Portable scripting | POSIX specification, checkbashisms |
| Quick examples | tldr, explainshell.com |
| Common mistakes | Bash Pitfalls, ShellCheck Wiki |
| Advanced topics | Bash Hackers Wiki, Greg's Wiki |
| Testing | BATS documentation |
| Platform differences | Platform-specific docs, Stack Overflow |
| Troubleshooting | Stack Overflow, Unix & Linux SE |
| Learning path | Bash Academy, TLDP guides |

---

## Quick Resource Lookup

**When writing a new script:**
1. Start with template from Google Style Guide
2. Use ShellCheck while developing
3. Reference Bash Manual for specific features
4. Check Bash Pitfalls for common mistakes

**When debugging:**
1. Use `set -x` for tracing
2. Check ShellCheck warnings
3. Search Bash Pitfalls
4. Search Stack Overflow for specific error

**When learning:**
1. Start with Bash Academy or TLDP
2. Use explainshell.com for commands
3. Read Greg's Wiki for in-depth topics
4. Practice with BATS tests

**When ensuring quality:**
1. Run ShellCheck
2. Run shellcheck
3. Format with shfmt
4. Write BATS tests
5. Review against Google Style Guide

These resources provide authoritative, up-to-date information for all aspects of bash scripting.
