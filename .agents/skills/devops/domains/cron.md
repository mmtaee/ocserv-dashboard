# Domain: Cron & Scheduled Tasks

Use this domain when the user asks about cron jobs, scheduled tasks, or systemd timers. Full cron management (creating, editing, removing jobs) is out of scope — this domain covers inspection only.

## Cron Jobs

```bash
# List crontab for current user
ssh "$DEVOPS_HOST" "crontab -l 2>/dev/null || echo 'No crontab for this user'"

# List system-wide cron jobs
ssh "$DEVOPS_HOST" "ls /etc/cron.d/ /etc/cron.daily/ /etc/cron.weekly/ /etc/cron.monthly/ 2>/dev/null"

# View a specific system cron file
ssh "$DEVOPS_HOST" "cat '/etc/cron.d/<name>'"
```

## systemd Timers

```bash
# List all active timers with next trigger time
ssh "$DEVOPS_HOST" "systemctl list-timers --no-pager"

# Status of a specific timer
ssh "$DEVOPS_HOST" "sudo systemctl status <timer-name>.timer --no-pager"
```

Report any timers that have never triggered (`n/a` in the LAST column) or that are overdue by more than their expected interval.
