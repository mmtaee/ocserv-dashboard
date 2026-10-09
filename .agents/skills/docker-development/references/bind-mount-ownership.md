# Bind-Mount Ownership: Root-Owned Artifacts on the Host

## The Problem

Containers that run as root and write into a bind-mounted project directory
leave **root-owned files on the host**. Typical producers:

```bash
docker compose run --rm app npm install     # node_modules/ now root-owned
docker compose run --rm app composer install
docker compose run --rm app npm run build  # dist/, public/build/ root-owned
```

The host user then hits failures that look unrelated:

```
npm error EACCES: permission denied, rename '.../node_modules/@babel/code-frame' -> ...
rm: cannot remove 'node_modules/...': Permission denied
```

Host-side `npm install`, build-tool cleanup steps (e.g. webpack/Encore
`cleanupOutputBeforeBuild`), and even `git clean -fdx` fail on these files.

## Diagnosis

```bash
find node_modules public/build -maxdepth 2 -user root | head
```

Any hit means a containerized process wrote there as root.

It rarely announces itself that way, though. What you see first is the tool that
runs next, failing for reasons that read like the application's fault:

- a test suite red with `Permission denied` inside a library's file writer, one
  failure per test that writes output — an application bug, until you look at who
  owns the output directory
- `composer install` / `npm ci` aborting on "Could not delete …" for a path the
  host user never created

The common shape: the container run succeeded, and the *following* host command
is the one that fails. Suspect ownership before debugging the failure it reports.

## Cleanup (no sudo required)

Use a throwaway container — root inside the container can act on what root
created, and the mount scopes it to the project.

**Give the files back** when the container wrote something you want to keep, or
touched a tracked file. Deleting a `composer.lock` the container rewrote loses
the state; deleting a test's output directory only postpones the question:

```bash
docker run --rm -v "$PWD:/work" -w /work alpine \
  chown -R "$(id -u):$(id -g)" /work
```

**Delete** when the artifacts are disposable and gitignored:

```bash
docker run --rm -v "$PWD:/work" -w /work alpine \
  sh -c 'rm -rf node_modules public/build dist'
```

Then reinstall/rebuild as the host user. Either way, verify before trusting the
next run — `find . -not -user "$(id -un)"` should come back empty, and a tracked
file the container rewrote wants `git checkout --` on top of the `chown`.

## Prevention

| Approach | How |
|---|---|
| Run as the host user | `docker compose run --rm --user "$(id -u):$(id -g)" -e HOME=/tmp app npm ci` — the arbitrary UID has no writable home in the container, and npm writes its cache to `$HOME`; point `HOME` (or `npm_config_cache`) at a writable path |
| Fix the UID in the image | `adduser -u 1000 ...` + `USER app` matching the typical host UID |
| Compose-wide | `user: "${UID:-1000}:${GID:-1000}"` on dev services — note `UID`/`GID` are **not** exported environment variables in most shells (bash's `UID` is shell-only); set them in the project `.env` file or `export UID GID` before composing |
| Keep artifacts out of the mount | named volume over `node_modules/`, or build inside the image (multi-stage) instead of into the mount |

Rootless Docker / userns-remap avoids the issue entirely but changes
semantics for the whole daemon.

## Related Gotcha: Named Volumes Mask Image Content

A named volume mounted over a path (e.g. `public/`) is populated from the
image **only on first use**. After deploying a new image, the volume still
holds the **old** content — refresh it explicitly (temp container +
`docker cp`/rsync) or recreate the volume as part of the deploy.
