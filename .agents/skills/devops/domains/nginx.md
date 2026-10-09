# Domain: Nginx & SSL

## Nginx Config Audit
```bash
# Test current config
ssh "$DEVOPS_HOST" "sudo nginx -t"

# List enabled sites
ssh "$DEVOPS_HOST" "ls /etc/nginx/sites-enabled/"

# View a site config
ssh "$DEVOPS_HOST" "sudo cat /etc/nginx/sites-enabled/<site>"

# Check nginx version and modules
ssh "$DEVOPS_HOST" "nginx -v"
```

## Let's Encrypt / Certbot SSL Renewal

Before running any certbot command, verify certbot is installed:
```bash
if ! ssh "$DEVOPS_HOST" "dpkg -l certbot 2>/dev/null | grep -q '^ii' || [ -L /usr/bin/certbot ] || [ -f /usr/bin/certbot ]" 2>/dev/null; then
  echo "certbot not installed — SSL provisioning not available; install with: sudo apt-get install -y certbot python3-certbot-nginx (or on systems with snapd: sudo snap install --classic certbot && sudo ln -sf /snap/bin/certbot /usr/bin/certbot)"
  # stop
fi
```

```bash
# List all certs and expiry
ssh "$DEVOPS_HOST" "sudo certbot certificates"

# Dry-run renewal (safe, no changes)
ssh "$DEVOPS_HOST" "sudo certbot renew --dry-run"

# Actually renew (ask user first)
ssh "$DEVOPS_HOST" "sudo certbot renew"

# Renew specific cert
ssh "$DEVOPS_HOST" "sudo certbot renew --cert-name <domain>"
```
After renewal, always reload nginx: `ssh "$DEVOPS_HOST" "sudo systemctl reload nginx"`

## Add a New vHost (Guided Workflow)
When user asks to add a new site, follow these steps:

1. Ask: domain name, document root (or reverse proxy target), SSL (yes/no)
2. Generate nginx config (show it to user for review before writing)
3. Use Config Edit Workflow to write to `/etc/nginx/sites-available/<domain>`
4. Enable site: `ssh "$DEVOPS_HOST" "sudo ln -s /etc/nginx/sites-available/<domain> /etc/nginx/sites-enabled/"`
5. Test: `ssh "$DEVOPS_HOST" "sudo nginx -t"`
6. If SSL requested:
   - Ask the user: "What email should certbot use for expiry notifications? (required for Let's Encrypt — must be a valid email address; certbot will reject malformed addresses with a confusing error)"
   - Once you have the email, run:
     ```bash
     ssh "$DEVOPS_HOST" "sudo certbot --nginx -d '${domain}' --non-interactive --agree-tos -m '${email}'"
     ```
   - If certbot has never been run on this server before, it may still prompt for ToS on the first run despite `--non-interactive`. In that case, instruct the user to run this command manually in their own terminal:
     ```
     ssh <host> "sudo certbot --nginx -d <domain> --agree-tos -m <email>"
     ```
     and return to /devops once it completes.
7. Reload: `ssh "$DEVOPS_HOST" "sudo systemctl reload nginx"`
8. Log: EDIT to audit log

**Standard reverse proxy config template** (show this when adding a new proxied site):
```nginx
server {
    listen 80;
    server_name <domain>;

    location / {
        proxy_pass http://localhost:<port>;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```
Note: certbot will add the SSL block automatically when run with --nginx.

## Remove a vHost
1. Show current config to user, ask to confirm removal
2. Disable symlink: `ssh "$DEVOPS_HOST" "sudo rm /etc/nginx/sites-enabled/<domain>"`
3. Ask: "Also remove the config file at `/etc/nginx/sites-available/<domain>`? (yes/no)"
   If yes: `ssh "$DEVOPS_HOST" "sudo rm /etc/nginx/sites-available/<domain>"`
4. Log the removal to `~/.agents/state/devops/audit.log`:
   ```bash
   echo "[$(date '+%Y-%m-%d %H:%M:%S')] EDIT | target: nginx/sites-enabled/<domain> | detail: removed vhost symlink" >> ~/.agents/state/devops/audit.log
   ```
5. Test and reload nginx
6. Optionally revoke cert: `ssh "$DEVOPS_HOST" "sudo certbot delete --cert-name <domain>"`

## Common Nginx Issues to Check
- Missing `proxy_set_header` causing 502s behind proxy
- Missing `client_max_body_size` for file upload sites
- Rate limiting not configured (security gap)
- `server_tokens off` should be set (hides nginx version)
