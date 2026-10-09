# Multi-stage caching: keep code-independent installs off the code-copy lineage

**Symptom:** a dev/CI image rebuilds heavy tooling (apt, pecl/xdebug, browser installs, `npm ci`) on **every** source change, even though none of it depends on the code.

**Cause:** the tooling stage is `FROM <the code stage>`, and the code stage ends with `COPY . .`. Docker invalidates every layer downstream of that copy on any tracked-file content change — so the tooling, layered on top, re-runs each time.

**Fix — re-parent, don't reorder.** Put the code-independent installs in a *sibling* stage `FROM base` (a stage with **no** code copy), then pull the built tree into the leaf via one `COPY --from=<code-stage>`:

```dockerfile
# composer/npm install + COPY . . + build (code-dependent)
FROM base AS deps
# apt, xdebug, symfony-cli, npm ci, chromium — NO code copy
FROM base AS devtools
FROM devtools AS dev
# the ONE code-dependent layer of dev
COPY --from=deps --chown=app:app /app /app
```

A source-only edit now invalidates `deps` (and dev's copy) but leaves the whole `devtools` lineage CACHED.

## Guardrails

- **Isolation by lineage:** keep xdebug/chromium in the `devtools → dev → e2e` branch only. A `production`/`profiling`/`tools` stage that is `FROM base`/`FROM deps` must not inherit `devtools`, or it gains xdebug (skews profiling timings) and browser bloat. Verify with the actual `FROM` chain, not by hoping.
- **Same-layer cleanup:** a transient `node_modules` needed only to run `npx playwright install` should be `rm -rf`'d in the *same* `RUN` (the leaf's `COPY --from=deps` overwrites it anyway) so it never bloats the layer. The browser binary lives in `~/.cache/ms-playwright`, outside `node_modules`, so it survives.
- **Verify the cache, not just the build:** `docker build --check` + a cold build prove it *builds*; only a **content** change to a source file + rebuild proves the *caching* — BuildKit hashes content, not mtime, so a bare `touch` won't invalidate. Look for `CACHED` on the tooling layers.
